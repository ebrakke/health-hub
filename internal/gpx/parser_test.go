package gpx

import (
	"math"
	"testing"

	"health-hub/internal/config"
	"health-hub/internal/models"
)

func TestCalculateSmoothedElevation(t *testing.T) {
	tests := []struct {
		name     string
		points   []models.GPXPoint
		config   *config.Config
		expected float64
	}{
		{
			name: "Simple elevation gain without smoothing",
			points: []models.GPXPoint{
				{Lat: 40.0, Lon: -74.0, Elevation: 10.0},
				{Lat: 40.1, Lon: -74.1, Elevation: 15.0},
				{Lat: 40.2, Lon: -74.2, Elevation: 20.0},
			},
			config: &config.Config{
				ElevationSmoothingEnabled: false,
			},
			expected: 10.0, // 5.0 + 5.0
		},
		{
			name: "Noisy elevation data with smoothing enabled",
			points: []models.GPXPoint{
				{Lat: 40.0, Lon: -74.0, Elevation: 10.0},
				{Lat: 40.1, Lon: -74.1, Elevation: 12.0}, // +2.0 (noise)
				{Lat: 40.2, Lon: -74.2, Elevation: 11.5}, // -0.5 (noise)
				{Lat: 40.3, Lon: -74.3, Elevation: 14.0}, // +2.5 (noise)
				{Lat: 40.4, Lon: -74.4, Elevation: 13.8}, // -0.2 (noise)
				{Lat: 40.5, Lon: -74.5, Elevation: 18.0}, // +4.2 (significant gain)
			},
			config: &config.Config{
				ElevationSmoothingEnabled:  true,
				ElevationSmoothingWindow:   3,
				ElevationMinGain:          3.0,
			},
			expected: 0.0, // No gains above 3.0m after smoothing small dataset
		},
		{
			name: "Large elevation gain that should be counted",
			points: []models.GPXPoint{
				{Lat: 40.0, Lon: -74.0, Elevation: 100.0},
				{Lat: 40.1, Lon: -74.1, Elevation: 102.0},
				{Lat: 40.2, Lon: -74.2, Elevation: 105.0},
				{Lat: 40.3, Lon: -74.3, Elevation: 108.0},
				{Lat: 40.4, Lon: -74.4, Elevation: 115.0}, // Clear elevation gain
			},
			config: &config.Config{
				ElevationSmoothingEnabled:  true,
				ElevationSmoothingWindow:   3,
				ElevationMinGain:          2.0,
			},
			expected: 9.17, // Moving average smoothed elevation gain (rounded)
		},
		{
			name: "Empty points slice",
			points: []models.GPXPoint{},
			config: &config.Config{
				ElevationSmoothingEnabled: true,
				ElevationSmoothingWindow:  5,
				ElevationMinGain:         3.0,
			},
			expected: 0.0,
		},
		{
			name: "Single point",
			points: []models.GPXPoint{
				{Lat: 40.0, Lon: -74.0, Elevation: 10.0},
			},
			config: &config.Config{
				ElevationSmoothingEnabled: true,
				ElevationSmoothingWindow:  5,
				ElevationMinGain:         3.0,
			},
			expected: 0.0,
		},
		{
			name: "Consistent elevation gain with lenient threshold",
			points: []models.GPXPoint{
				{Lat: 40.0, Lon: -74.0, Elevation: 10.0},
				{Lat: 40.1, Lon: -74.1, Elevation: 11.5},
				{Lat: 40.2, Lon: -74.2, Elevation: 13.0},
				{Lat: 40.3, Lon: -74.3, Elevation: 14.5},
				{Lat: 40.4, Lon: -74.4, Elevation: 16.0},
			},
			config: &config.Config{
				ElevationSmoothingEnabled:  true,
				ElevationSmoothingWindow:   3,
				ElevationMinGain:          1.0,
			},
			expected: 3.0, // Moving average smoothed elevation gain
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateSmoothedElevation(tt.points, tt.config)
			tolerance := 0.2
			if math.Abs(result-tt.expected) > tolerance {
				t.Errorf("calculateSmoothedElevation() = %v, expected %v (tolerance: %v)", result, tt.expected, tolerance)
			}
		})
	}
}

func TestCalculateSimpleElevation(t *testing.T) {
	tests := []struct {
		name     string
		points   []models.GPXPoint
		expected float64
	}{
		{
			name: "Simple ascending elevation",
			points: []models.GPXPoint{
				{Lat: 40.0, Lon: -74.0, Elevation: 10.0},
				{Lat: 40.1, Lon: -74.1, Elevation: 15.0},
				{Lat: 40.2, Lon: -74.2, Elevation: 20.0},
			},
			expected: 10.0,
		},
		{
			name: "Mixed elevation changes",
			points: []models.GPXPoint{
				{Lat: 40.0, Lon: -74.0, Elevation: 10.0},
				{Lat: 40.1, Lon: -74.1, Elevation: 15.0}, // +5.0
				{Lat: 40.2, Lon: -74.2, Elevation: 12.0}, // -3.0 (ignored)
				{Lat: 40.3, Lon: -74.3, Elevation: 18.0}, // +6.0
			},
			expected: 11.0, // 5.0 + 6.0
		},
		{
			name: "All descending elevation",
			points: []models.GPXPoint{
				{Lat: 40.0, Lon: -74.0, Elevation: 20.0},
				{Lat: 40.1, Lon: -74.1, Elevation: 15.0},
				{Lat: 40.2, Lon: -74.2, Elevation: 10.0},
			},
			expected: 0.0,
		},
		{
			name: "Empty points",
			points: []models.GPXPoint{},
			expected: 0.0,
		},
		{
			name: "Single point",
			points: []models.GPXPoint{
				{Lat: 40.0, Lon: -74.0, Elevation: 10.0},
			},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateSimpleElevation(tt.points)
			if result != tt.expected {
				t.Errorf("calculateSimpleElevation() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateMedian(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{
			name:     "Odd number of values",
			values:   []float64{1.0, 3.0, 2.0, 5.0, 4.0},
			expected: 3.0,
		},
		{
			name:     "Even number of values",
			values:   []float64{1.0, 2.0, 3.0, 4.0},
			expected: 2.5,
		},
		{
			name:     "Single value",
			values:   []float64{5.0},
			expected: 5.0,
		},
		{
			name:     "Empty slice",
			values:   []float64{},
			expected: 0.0,
		},
		{
			name:     "Two values",
			values:   []float64{2.0, 8.0},
			expected: 5.0,
		},
		{
			name:     "Identical values",
			values:   []float64{3.0, 3.0, 3.0},
			expected: 3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateMedian(tt.values)
			if result != tt.expected {
				t.Errorf("calculateMedian() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestParseGPXElevationIntegration(t *testing.T) {
	testGPX := `<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1" creator="Test">
  <trk>
    <name>Test Track</name>
    <trkseg>
      <trkpt lat="40.7128" lon="-74.0060">
        <ele>10.0</ele>
        <time>2023-01-01T10:00:00Z</time>
      </trkpt>
      <trkpt lat="40.7129" lon="-74.0061">
        <ele>12.0</ele>
        <time>2023-01-01T10:00:30Z</time>
      </trkpt>
      <trkpt lat="40.7130" lon="-74.0062">
        <ele>11.5</ele>
        <time>2023-01-01T10:01:00Z</time>
      </trkpt>
      <trkpt lat="40.7131" lon="-74.0063">
        <ele>14.0</ele>
        <time>2023-01-01T10:01:30Z</time>
      </trkpt>
      <trkpt lat="40.7132" lon="-74.0064">
        <ele>13.8</ele>
        <time>2023-01-01T10:02:00Z</time>
      </trkpt>
      <trkpt lat="40.7133" lon="-74.0065">
        <ele>16.0</ele>
        <time>2023-01-01T10:02:30Z</time>
      </trkpt>
    </trkseg>
  </trk>
</gpx>`

	t.Run("Parse GPX with default smoothing", func(t *testing.T) {
		track, activity, err := ParseGPX(testGPX)
		if err != nil {
			t.Fatalf("ParseGPX() error = %v", err)
		}

		if len(track.Points) != 6 {
			t.Errorf("Expected 6 points, got %d", len(track.Points))
		}

		if activity.Name != "Test Track" {
			t.Errorf("Expected activity name 'Test Track', got '%s'", activity.Name)
		}

		// With default smoothing (window=5, min_gain=3.0), elevation should be filtered
		// The exact value will depend on the smoothing algorithm
		if activity.TotalElevation < 0 {
			t.Errorf("Total elevation should not be negative, got %f", activity.TotalElevation)
		}
	})

	t.Run("Parse GPX without smoothing", func(t *testing.T) {
		cfg := &config.Config{
			ElevationSmoothingEnabled: false,
		}

		track, activity, err := ParseGPXWithConfig(testGPX, cfg)
		if err != nil {
			t.Fatalf("ParseGPXWithConfig() error = %v", err)
		}

		if len(track.Points) != 6 {
			t.Errorf("Expected 6 points, got %d", len(track.Points))
		}

		// Without smoothing, should count all positive elevation changes
		// 12-10 + 14-11.5 + 16-13.8 = 2 + 2.5 + 2.2 = 6.7
		expectedElevation := 6.7
		tolerance := 0.01
		if math.Abs(activity.TotalElevation-expectedElevation) > tolerance {
			t.Errorf("Expected total elevation %f, got %f", expectedElevation, activity.TotalElevation)
		}
	})

	t.Run("Parse GPX with custom smoothing", func(t *testing.T) {
		cfg := &config.Config{
			ElevationSmoothingEnabled:  true,
			ElevationSmoothingWindow:   3,
			ElevationMinGain:          1.0,
		}

		track, activity, err := ParseGPXWithConfig(testGPX, cfg)
		if err != nil {
			t.Fatalf("ParseGPXWithConfig() error = %v", err)
		}

		if len(track.Points) != 6 {
			t.Errorf("Expected 6 points, got %d", len(track.Points))
		}

		// With lenient smoothing, should detect some elevation gain
		if activity.TotalElevation <= 0 {
			t.Errorf("Expected some elevation gain with lenient smoothing, got %f", activity.TotalElevation)
		}
	})
}

// Benchmark tests for performance
func BenchmarkCalculateSmoothedElevation(b *testing.B) {
	// Create a large dataset for benchmarking
	points := make([]models.GPXPoint, 1000)
	for i := 0; i < 1000; i++ {
		points[i] = models.GPXPoint{
			Lat:       40.0 + float64(i)*0.001,
			Lon:       -74.0 + float64(i)*0.001,
			Elevation: 100.0 + float64(i%20)*2.0, // Some elevation variation
		}
	}

	cfg := &config.Config{
		ElevationSmoothingEnabled:  true,
		ElevationSmoothingWindow:   5,
		ElevationMinGain:          3.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateSmoothedElevation(points, cfg)
	}
}

func BenchmarkCalculateSimpleElevation(b *testing.B) {
	// Create a large dataset for benchmarking
	points := make([]models.GPXPoint, 1000)
	for i := 0; i < 1000; i++ {
		points[i] = models.GPXPoint{
			Lat:       40.0 + float64(i)*0.001,
			Lon:       -74.0 + float64(i)*0.001,
			Elevation: 100.0 + float64(i%20)*2.0,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateSimpleElevation(points)
	}
}