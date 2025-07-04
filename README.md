# Health Hub

A Go application for managing and analyzing personal health metrics and fitness data.

## Features

- Upload and store GPX files from fitness trackers
- Import health data from various sources (Oura Ring, Fitbit, etc.)
- Clean web interface for data management
- Local file storage with optional S3 backup
- RESTful API for data access
- Extensible architecture for adding new data sources

## Quick Start

1. Clone the repository
2. Run the application:
   ```bash
   go run main.go
   ```
3. Open your browser to `http://localhost:8080`

## Configuration

Configure the application using environment variables:

- `PORT`: Server port (default: 8080)
- `DATA_PATH`: Local data storage path (default: ./data)
- `USE_S3`: Enable S3 storage (default: false)
- `S3_BUCKET`: S3 bucket name for backups
- `AWS_REGION`: AWS region (default: us-east-1)
- `ENVIRONMENT`: Environment (development/production)

## Data Sources

### GPX Files
Upload GPX files from fitness trackers, running watches, or cycling computers.

### Health Data
Upload JSON files containing health metrics. Expected format:
```json
[
  {
    "type": "heart_rate",
    "value": 72,
    "unit": "bpm",
    "timestamp": "2023-01-01T00:00:00Z",
    "source": "oura"
  }
]
```

## API Endpoints

- `GET /api/activities` - List all activities
- `GET /api/health` - List all health metrics
- `POST /api/upload/gpx` - Upload GPX file
- `POST /api/upload/health` - Upload health data

## Deployment

### Local Development
```bash
go run main.go
```

### Docker
```bash
docker build -t health-hub .
docker run -p 8080:8080 health-hub
```

### With S3 Storage
```bash
export USE_S3=true
export S3_BUCKET=your-bucket-name
export AWS_REGION=us-east-1
go run main.go
```

## Architecture

- `main.go` - Application entry point
- `internal/handlers/` - HTTP request handlers
- `internal/models/` - Data models
- `internal/storage/` - Storage implementations
- `internal/config/` - Configuration management

## Future Enhancements

- Data visualization and analytics
- Correlation analysis between different metrics
- Export functionality
- Integration with more health platforms
- Advanced GPX parsing and analysis
- Database storage option
- User authentication and multi-user support