# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build and Run
- `make build` - Build the application binary
- `make run` - Build and run the application locally
- `make dev` - Run with hot reload using air (auto-installs if missing)
- `make serve` - Build and run with network info display
- `make start` - Full development setup (deps + serve)

### Dependencies
- `make deps` - Install and tidy Go dependencies
- `go mod tidy` - Clean up dependencies

### S3 Storage
- `make run-s3` - Run with S3 storage enabled (requires S3_BUCKET env var)
- Set `USE_S3=true` and `S3_BUCKET=bucket-name` for S3 backup storage

### Development Tools
- `make install-air` - Install air for hot reload
- `make network-info` - Show local and Tailscale network information
- `make clean` - Remove build artifacts and data directory

## Architecture

This is a Go web application for managing personal health and fitness data with the following structure:

### Core Components
- **main.go** - Entry point, server setup, and Tailscale integration
- **internal/config/** - Environment-based configuration management
- **internal/handlers/** - HTTP handlers with embedded HTML templates
- **internal/models/** - Data models for activities, health metrics, and GPX tracks
- **internal/storage/** - Storage abstraction with file and S3 implementations

### Storage System
- **File Storage** - Default local JSON file storage in `./data/` directory
- **S3 Storage** - Optional S3 backup with local caching (hybrid approach)
- **Directory Structure**: `data/{activities,health,gpx,uploads}/`

### Data Models
- **Activity** - Fitness activities with GPX file references
- **HealthMetric** - Generic health data points (heart rate, sleep, etc.)
- **GPXTrack** - Parsed GPS track data from GPX files
- **Specialized Models** - SleepData and HeartRateData for specific metrics

### HTTP Handlers
- Single-page web interface with embedded CSS/JS in Go template
- RESTful API endpoints for activities and health metrics
- File upload handlers for GPX and JSON health data
- Basic GPX parsing (simplified, not full XML parser)

## Frontend Architecture

### HTMX Integration
- **Use HTMX for all interactive elements** - Replace vanilla JavaScript with HTMX attributes
- **Forms** - Use `hx-post`, `hx-encoding="multipart/form-data"` for file uploads
- **Dynamic Content** - Use `hx-get`, `hx-target`, `hx-swap` for loading stats and data
- **Upload Feedback** - Use `hx-indicator` and `hx-swap="outerHTML"` for user feedback
- **Auto-refresh** - Use `hx-trigger="every 30s"` for periodic data updates

### Tailwind CSS Styling
- **Use Tailwind utility classes** - Replace all custom CSS with Tailwind classes
- **CDN Integration** - Include Tailwind CSS via CDN for simplicity
- **Responsive Design** - Use Tailwind's responsive prefixes (`sm:`, `md:`, `lg:`)
- **Component Classes** - Use Tailwind's component-friendly class combinations

### Frontend Standards
- **No Custom CSS** - Use only Tailwind utility classes
- **No Vanilla JavaScript** - Use HTMX attributes for all interactions
- **Embedded Templates** - Keep HTML templates in Go handlers for simplicity
- **Progressive Enhancement** - Ensure forms work without JavaScript

### Template Structure
```html
<!-- Include HTMX and Tailwind CSS -->
<script src="https://unpkg.com/htmx.org@1.9.10"></script>
<script src="https://cdn.tailwindcss.com"></script>

<!-- Use HTMX for forms -->
<form hx-post="/api/upload/gpx" hx-encoding="multipart/form-data" 
      hx-target="#upload-status" hx-swap="innerHTML">
  <input type="file" name="gpx" accept=".gpx" required 
         class="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100">
  <button type="submit" class="mt-4 bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">
    Upload GPX
  </button>
</form>

<!-- Use HTMX for dynamic content -->
<div hx-get="/api/activities/count" hx-trigger="load, every 30s" 
     hx-target="this" hx-swap="innerHTML"
     class="bg-white p-6 rounded-lg shadow-md">
  <h3 class="text-lg font-semibold text-gray-900">Activities</h3>
  <p class="text-gray-500">Loading...</p>
</div>
```

### API Endpoints for HTMX
- **HTML Fragments** - Create endpoints that return HTML fragments for HTMX
- **Status Responses** - Return appropriate HTTP status codes for HTMX handling
- **Partial Updates** - Design endpoints to return specific UI components

## Configuration

Environment variables:
- `PORT` - Server port (default: 8080)
- `DATA_PATH` - Data storage path (default: ./data)
- `USE_S3` - Enable S3 storage (default: false)
- `S3_BUCKET` - S3 bucket name for backups
- `AWS_REGION` - AWS region (default: us-east-1)
- `ENVIRONMENT` - Environment (development/production)

## Key Features

- **Tailscale Integration** - Automatic Tailscale IP detection and display
- **Hybrid Storage** - Local-first with optional S3 backup
- **GPX Processing** - Basic GPX file parsing and activity creation
- **Health Data Import** - JSON-based health metric importing
- **HTMX Web Interface** - Interactive single-page application with HTMX
- **Tailwind Styling** - Utility-first CSS framework for consistent design

## Development Notes

- Uses Go 1.23.0 with minimal dependencies (only AWS SDK)
- No external database - JSON file-based storage
- Simplified GPX parsing (not full XML parser)
- HTMX for interactivity, Tailwind for styling
- Templates embedded in handler code
- Static file serving from `./static/` directory (if needed)
- No authentication system currently implemented

## Frontend Migration Notes

Current implementation uses:
- Custom CSS in embedded templates
- Vanilla JavaScript for form handling and AJAX
- Manual DOM manipulation for stats loading

Should be converted to:
- Tailwind utility classes for all styling
- HTMX attributes for all interactivity
- Server-side HTML fragment responses for dynamic content