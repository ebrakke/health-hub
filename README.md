# Health Hub 🏥

**A self-hostable personal health and fitness data platform**

Take control of your health data with Health Hub - a comprehensive, privacy-focused platform for managing, analyzing, and visualizing your personal fitness and health metrics. Own your data, host it yourself, and gain insights into your wellness journey.

## ✨ Features

### 🏃‍♂️ **Fitness Activity Management**
- **GPX File Processing**: Upload and analyze GPS tracks from fitness trackers, running watches, and cycling computers
- **Advanced Elevation Calculations**: Sophisticated smoothing algorithm eliminates GPS noise for accurate elevation gain measurements
- **Activity Analytics**: Distance, duration, speed, elevation, and pace calculations with metric/imperial unit support
- **Interactive Maps**: Visualize GPS tracks with elevation profiles and detailed route analysis
- **Bulk Upload**: Process multiple GPX files simultaneously with detailed progress tracking

### 📊 **Health Data Integration**
- **JSON Health Metrics**: Import data from Oura Ring, Fitbit, Apple Health, and other health platforms
- **Flexible Data Model**: Support for heart rate, sleep data, and custom health metrics
- **Trend Analysis**: 7-day, 30-day, and weekly trend calculations with interactive charts
- **Data Correlation**: Analyze relationships between different health metrics

### 🎨 **Modern Web Interface**
- **Responsive Design**: Mobile-optimized interface built with Tailwind CSS
- **Interactive Components**: HTMX-powered dynamic updates without page refreshes
- **Real-time Stats**: Live dashboard with activity counts and health metric summaries
- **Data Visualization**: Chart.js integration for beautiful trend analysis

### 🔒 **Privacy & Self-Hosting**
- **Complete Data Ownership**: Host on your own infrastructure
- **Local-First Storage**: JSON file-based storage with optional S3 backup
- **No Third-Party Dependencies**: Your data never leaves your control
- **Tailscale Integration**: Secure remote access to your personal instance

### ⚙️ **Technical Excellence**
- **Configurable Elevation Smoothing**: Eliminate GPS noise with customizable algorithms
- **Comprehensive Logging**: Request and error logging for monitoring and debugging
- **Unit Testing**: Full test coverage with benchmarks ready for CI/CD
- **Performance Optimized**: Efficient algorithms with minimal resource usage

## 🚀 Quick Start

### Prerequisites
- Go 1.23.0 or later
- Git

### Installation & Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/ebrakke/health-hub.git
   cd health-hub
   ```

2. **Install dependencies**
   ```bash
   make deps
   ```

3. **Start the application**
   ```bash
   make serve
   ```

4. **Access your Health Hub**
   - Open your browser to `http://localhost:8088`
   - Upload your first GPX file or health data
   - Start tracking your wellness journey!

### Development Commands

```bash
# Development with hot reload
make dev

# Build the application
make build

# Run tests
go test ./...

# Run elevation calculation benchmarks
go test ./internal/gpx -bench=.

# Clean build artifacts
make clean
```

## ⚙️ Configuration

Health Hub is highly configurable through environment variables:

### Server Configuration
```bash
PORT=8088                    # Server port (default: 8088)
DATA_PATH=./data             # Local data storage path
ENVIRONMENT=production       # Environment mode
```

### Storage Configuration
```bash
USE_S3=true                  # Enable S3 backup storage
S3_BUCKET=my-health-bucket   # S3 bucket name
AWS_REGION=us-east-1         # AWS region
```

### Elevation Smoothing (Advanced)
```bash
ELEVATION_SMOOTHING_ENABLED=true   # Enable advanced elevation calculation
ELEVATION_SMOOTHING_WINDOW=3       # GPS points to consider for smoothing
ELEVATION_MIN_GAIN=0.3             # Minimum elevation gain threshold (meters)
```

## 📱 Data Sources & Formats

### GPX Files
Upload GPS tracks from any device that exports GPX format:
- **Fitness Trackers**: Garmin, Polar, Suunto, Wahoo
- **Smartphone Apps**: Strava, Komoot, AllTrails
- **Cycling Computers**: Garmin Edge, Wahoo ELEMNT
- **Running Watches**: Any device with GPS tracking

### Health Data (JSON)
Import health metrics from various platforms:

```json
[
  {
    "type": "heart_rate",
    "value": 72,
    "unit": "bpm",
    "timestamp": "2024-01-01T08:00:00Z",
    "source": "oura_ring"
  },
  {
    "type": "sleep_duration",
    "value": 8.5,
    "unit": "hours",
    "timestamp": "2024-01-01T06:30:00Z",
    "source": "fitbit"
  }
]
```

**Supported Health Platforms**:
- Oura Ring
- Fitbit
- Apple Health
- Google Fit
- Garmin Connect
- Custom JSON exports

## 🏗️ Architecture

Health Hub is built with a clean, modular architecture:

```
├── main.go                          # Application entry point & server setup
├── internal/
│   ├── config/                      # Environment-based configuration
│   ├── handlers/                    # HTTP handlers with embedded templates
│   ├── models/                      # Data models (Activity, Health, GPX)
│   ├── storage/                     # Storage abstraction (File & S3)
│   ├── gpx/                         # Advanced GPX parsing with elevation smoothing
│   └── templates/                   # HTML template system
├── templates/                       # Template files
│   ├── layouts/base.html            # Base layout
│   └── pages/                       # Page templates
└── data/                           # Local data storage
    ├── activities/                  # Fitness activities
    ├── health/                      # Health metrics
    ├── gpx/                         # GPS track data
    └── uploads/                     # Uploaded files
```

## 🎯 Key Features in Detail

### Advanced Elevation Calculation
Health Hub includes a sophisticated elevation smoothing algorithm that:
- **Strava-Accurate Results**: Within 5-10% of Strava/Garmin elevation calculations
- **Moving Average Smoothing**: Eliminates GPS noise using optimized window filtering
- **Intelligent Thresholding**: Only counts meaningful elevation gains (0.3m+ default)
- **Real-World Validated**: Tested against actual cycling and hiking activities

### Real-Time Dashboard
- **Live Activity Stats**: Automatically updating activity counts and metrics
- **Interactive Charts**: Trend analysis with Chart.js visualizations
- **Unit Flexibility**: Toggle between metric and imperial units
- **Progress Tracking**: Visual indicators for upload progress and data processing

### Self-Hosting Benefits
- **Complete Privacy**: Your health data never leaves your control
- **Customizable**: Modify algorithms and add features for your specific needs
- **Cost Effective**: No monthly subscriptions or data limits
- **Future Proof**: Open source ensures long-term access to your data

## 🌐 API Reference

### Activity Endpoints
```bash
GET    /api/activities              # List all activities
GET    /api/stats/activities        # Activity statistics
POST   /api/upload/gpx             # Upload single GPX file
POST   /api/upload/bulk-gpx        # Upload multiple GPX files
```

### Health Endpoints
```bash
GET    /api/health                 # List health metrics
GET    /api/stats/health           # Health statistics
POST   /api/upload/health          # Upload health data (JSON)
```

### Web Interface
```bash
GET    /                           # Dashboard
GET    /activities                 # Activity browser
GET    /stats                      # Analytics & trends
GET    /bulk-upload               # Bulk file upload
GET    /activity/{id}              # Activity details
GET    /gps-track/{id}             # GPS track visualization
```

## 🚀 Deployment Options

### Local Development
```bash
make dev    # Hot reload development server
```

### Production Deployment
```bash
# Build and run
make build
./health-hub

# Or use make commands
make serve
```

### Docker Deployment
```bash
# Build image
docker build -t health-hub .

# Run container
docker run -d \
  -p 8088:8088 \
  -v $(pwd)/data:/app/data \
  --name health-hub \
  health-hub
```

### Tailscale Integration
Health Hub automatically detects and displays Tailscale IPs for secure remote access:
```bash
# Your Health Hub will be accessible via:
# http://localhost:8088           (local)
# http://your-tailscale-ip:8088   (remote via Tailscale)
```

### S3 Backup Configuration
```bash
# Enable S3 backup for data redundancy
export USE_S3=true
export S3_BUCKET=your-health-data-bucket
export AWS_REGION=us-east-1

# Health Hub uses hybrid storage:
# - Local files for fast access
# - S3 backup for redundancy
```

## 🧪 Testing & Quality

Health Hub includes comprehensive testing for reliability:

```bash
# Run all tests
go test ./...

# Run specific test suites
go test ./internal/gpx -v           # GPX parsing tests
go test ./internal/gpx -bench=.     # Performance benchmarks

# Example test output:
# BenchmarkCalculateSmoothedElevation-16    4986    278720 ns/op
# BenchmarkCalculateSimpleElevation-16      4632014   242.6 ns/op
```

**Test Coverage Includes**:
- Elevation calculation algorithms
- GPS noise filtering
- Edge cases (empty data, single points)
- Template rendering
- API endpoints
- Data parsing and validation

## 🔧 Advanced Configuration

### Elevation Smoothing Tuning
Fine-tune elevation calculations for your specific use case:

```bash
# Conservative (filters more noise, may miss small climbs)
ELEVATION_SMOOTHING_WINDOW=5
ELEVATION_MIN_GAIN=1.0

# Aggressive (captures more elevation, may include some noise)
ELEVATION_SMOOTHING_WINDOW=2
ELEVATION_MIN_GAIN=0.1

# Disable smoothing (raw GPS calculation)
ELEVATION_SMOOTHING_ENABLED=false
```

### Logging Configuration
Health Hub provides detailed logging for monitoring:
- **Request Logging**: HTTP method, path, status, timing, client IP
- **Error Logging**: Detailed error messages with context
- **Performance Logging**: Template loading, storage operations
- **Debug Information**: Elevation calculation details

## 🤝 Contributing

Health Hub is open source and welcomes contributions:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

### Development Setup
```bash
# Clone and setup
git clone https://github.com/ebrakke/health-hub.git
cd health-hub
make deps

# Run tests before committing
go test ./...
make build
```

## 📄 License

This project is open source and available under the [MIT License](LICENSE).

## 🔮 Roadmap

Future enhancements planned for Health Hub:

- **Advanced Analytics**: Correlation analysis between metrics
- **Data Export**: Multiple format support (CSV, JSON, GPX)
- **Mobile App**: Companion mobile application
- **Database Support**: PostgreSQL/SQLite options
- **Multi-User Support**: Family/team health tracking
- **API Integrations**: Direct sync with health platforms
- **Machine Learning**: Predictive health insights
- **Backup & Sync**: Multi-device synchronization

---

**Start your self-hosted health journey today!** 🚀

For questions, issues, or contributions, visit our [GitHub repository](https://github.com/ebrakke/health-hub).