package storage

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"health-hub/internal/models"
)

type S3Storage struct {
	*FileStorage
	s3Client *s3.S3
	bucket   string
}

func NewS3Storage(basePath, bucket string) (*S3Storage, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %v", err)
	}

	return &S3Storage{
		FileStorage: NewFileStorage(basePath),
		s3Client:    s3.New(sess),
		bucket:      bucket,
	}, nil
}

func (s3s *S3Storage) SaveFile(filename string, data []byte) error {
	// Save locally first
	if err := s3s.FileStorage.SaveFile(filename, data); err != nil {
		return err
	}

	// Upload to S3
	key := fmt.Sprintf("uploads/%s", filename)
	_, err := s3s.s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})

	return err
}

func (s3s *S3Storage) SaveActivity(activity *models.Activity) error {
	// Save locally first
	if err := s3s.FileStorage.SaveActivity(activity); err != nil {
		return err
	}

	// Backup to S3
	return s3s.backupToS3("activities", activity.ID+".json")
}

func (s3s *S3Storage) SaveHealthMetric(metric *models.HealthMetric) error {
	// Save locally first
	if err := s3s.FileStorage.SaveHealthMetric(metric); err != nil {
		return err
	}

	// Backup to S3
	return s3s.backupToS3("health", metric.ID+".json")
}

func (s3s *S3Storage) SaveGPXTrack(track *models.GPXTrack) error {
	// Save locally first
	if err := s3s.FileStorage.SaveGPXTrack(track); err != nil {
		return err
	}

	// Backup to S3
	return s3s.backupToS3("gpx", track.ID+".json")
}

func (s3s *S3Storage) backupToS3(folder, filename string) error {
	localPath := filepath.Join(s3s.basePath, folder, filename)
	data, err := os.ReadFile(localPath)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("data/%s/%s", folder, filename)
	_, err = s3s.s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})

	return err
}