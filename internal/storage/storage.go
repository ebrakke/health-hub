package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"health-hub/internal/models"
)

type Storage interface {
	SaveActivity(activity *models.Activity) error
	GetActivities() ([]*models.Activity, error)
	SaveHealthMetric(metric *models.HealthMetric) error
	GetHealthMetrics() ([]*models.HealthMetric, error)
	SaveGPXTrack(track *models.GPXTrack) error
	GetGPXTracks() ([]*models.GPXTrack, error)
	SaveFile(filename string, data []byte) error
}

type FileStorage struct {
	basePath string
}

func NewFileStorage(basePath string) *FileStorage {
	os.MkdirAll(basePath, 0755)
	os.MkdirAll(filepath.Join(basePath, "activities"), 0755)
	os.MkdirAll(filepath.Join(basePath, "health"), 0755)
	os.MkdirAll(filepath.Join(basePath, "gpx"), 0755)
	os.MkdirAll(filepath.Join(basePath, "uploads"), 0755)
	
	return &FileStorage{basePath: basePath}
}

func (fs *FileStorage) SaveActivity(activity *models.Activity) error {
	if activity.ID == "" {
		activity.ID = fmt.Sprintf("activity_%d", time.Now().UnixNano())
	}
	activity.CreatedAt = time.Now()
	
	filename := filepath.Join(fs.basePath, "activities", activity.ID+".json")
	return fs.saveJSON(filename, activity)
}

func (fs *FileStorage) GetActivities() ([]*models.Activity, error) {
	var activities []*models.Activity
	
	files, err := ioutil.ReadDir(filepath.Join(fs.basePath, "activities"))
	if err != nil {
		return activities, nil
	}
	
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			var activity models.Activity
			if err := fs.loadJSON(filepath.Join(fs.basePath, "activities", file.Name()), &activity); err == nil {
				activities = append(activities, &activity)
			}
		}
	}
	
	return activities, nil
}

func (fs *FileStorage) SaveHealthMetric(metric *models.HealthMetric) error {
	if metric.ID == "" {
		metric.ID = fmt.Sprintf("health_%d", time.Now().UnixNano())
	}
	metric.CreatedAt = time.Now()
	
	filename := filepath.Join(fs.basePath, "health", metric.ID+".json")
	return fs.saveJSON(filename, metric)
}

func (fs *FileStorage) GetHealthMetrics() ([]*models.HealthMetric, error) {
	var metrics []*models.HealthMetric
	
	files, err := ioutil.ReadDir(filepath.Join(fs.basePath, "health"))
	if err != nil {
		return metrics, nil
	}
	
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			var metric models.HealthMetric
			if err := fs.loadJSON(filepath.Join(fs.basePath, "health", file.Name()), &metric); err == nil {
				metrics = append(metrics, &metric)
			}
		}
	}
	
	return metrics, nil
}

func (fs *FileStorage) SaveGPXTrack(track *models.GPXTrack) error {
	if track.ID == "" {
		track.ID = fmt.Sprintf("gpx_%d", time.Now().UnixNano())
	}
	track.CreatedAt = time.Now()
	
	filename := filepath.Join(fs.basePath, "gpx", track.ID+".json")
	return fs.saveJSON(filename, track)
}

func (fs *FileStorage) GetGPXTracks() ([]*models.GPXTrack, error) {
	var tracks []*models.GPXTrack
	
	files, err := ioutil.ReadDir(filepath.Join(fs.basePath, "gpx"))
	if err != nil {
		return tracks, nil
	}
	
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			var track models.GPXTrack
			if err := fs.loadJSON(filepath.Join(fs.basePath, "gpx", file.Name()), &track); err == nil {
				tracks = append(tracks, &track)
			}
		}
	}
	
	return tracks, nil
}

func (fs *FileStorage) SaveFile(filename string, data []byte) error {
	return ioutil.WriteFile(filepath.Join(fs.basePath, "uploads", filename), data, 0644)
}

func (fs *FileStorage) saveJSON(filename string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

func (fs *FileStorage) loadJSON(filename string, v interface{}) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}