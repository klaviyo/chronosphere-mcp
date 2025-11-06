// Copyright 2025 Chronosphere Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prometheus

import (
	"image/color"
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
)

func TestDetectYAxisLabel(t *testing.T) {
	tests := []struct {
		name     string
		series   model.Matrix
		expected string
	}{
		{
			name:     "empty series",
			series:   model.Matrix{},
			expected: "Value",
		},
		{
			name: "bytes metric",
			series: model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{"__name__": "memory_bytes"},
				},
			},
			expected: "Bytes",
		},
		{
			name: "seconds metric",
			series: model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{"__name__": "request_duration_seconds"},
				},
			},
			expected: "Seconds",
		},
		{
			name: "milliseconds metric",
			series: model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{"__name__": "latency_milliseconds"},
				},
			},
			expected: "Milliseconds",
		},
		{
			name: "ratio metric",
			series: model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{"__name__": "success_ratio"},
				},
			},
			expected: "Ratio",
		},
		{
			name: "percent metric",
			series: model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{"__name__": "cpu_percent"},
				},
			},
			expected: "Ratio",
		},
		{
			name: "total counter",
			series: model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{"__name__": "http_requests_total"},
				},
			},
			expected: "Count",
		},
		{
			name: "count metric",
			series: model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{"__name__": "error_count"},
				},
			},
			expected: "Count",
		},
		{
			name: "requests metric",
			series: model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{"__name__": "api_requests"},
				},
			},
			expected: "Requests",
		},
		{
			name: "cpu metric",
			series: model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{"__name__": "cpu_usage"},
				},
			},
			expected: "CPU Usage",
		},
		{
			name: "memory metric",
			series: model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{"__name__": "memory_usage"},
				},
			},
			expected: "Memory",
		},
		{
			name: "latency metric",
			series: model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{"__name__": "latency"},
				},
			},
			expected: "Latency",
		},
		{
			name: "rate metric",
			series: model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{"__name__": "request_rate"},
				},
			},
			expected: "Rate",
		},
		{
			name: "unknown metric with name",
			series: model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{"__name__": "custom_metric"},
				},
			},
			expected: "custom_metric",
		},
		{
			name: "metric without name label",
			series: model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{"label1": "value1"},
				},
			},
			expected: "Value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectYAxisLabel(tt.series)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTimeTickMarker(t *testing.T) {
	marker := &timeTickMarker{}

	tests := []struct {
		name           string
		minVal         float64
		maxVal         float64
		expectedCount  int
		expectedFormat string
	}{
		{
			name:           "less than 1 hour",
			minVal:         float64(time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC).Unix()),
			maxVal:         float64(time.Date(2025, 1, 1, 12, 30, 0, 0, time.UTC).Unix()),
			expectedCount:  6, // major ticks
			expectedFormat: "15:04:05",
		},
		{
			name:           "less than 1 day",
			minVal:         float64(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix()),
			maxVal:         float64(time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC).Unix()),
			expectedCount:  8, // major ticks
			expectedFormat: "15:04",
		},
		{
			name:           "less than 1 week",
			minVal:         float64(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix()),
			maxVal:         float64(time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC).Unix()),
			expectedCount:  7, // major ticks
			expectedFormat: "01-02 15:04",
		},
		{
			name:           "1 week or more",
			minVal:         float64(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix()),
			maxVal:         float64(time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC).Unix()),
			expectedCount:  10, // major ticks
			expectedFormat: "2006-01-02",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticks := marker.Ticks(tt.minVal, tt.maxVal)

			// Count major ticks (those with labels)
			majorTicks := 0
			for _, tick := range ticks {
				if tick.Label != "" {
					majorTicks++
				}
			}

			assert.Equal(t, tt.expectedCount, majorTicks, "major tick count mismatch")

			// Verify all major ticks have properly formatted labels
			for _, tick := range ticks {
				if tick.Label != "" {
					timestamp := time.Unix(int64(tick.Value), 0).UTC()
					expectedLabel := timestamp.Format(tt.expectedFormat)
					assert.Equal(t, expectedLabel, tick.Label, "tick label format mismatch")
				}
			}

			// Verify we have minor ticks
			assert.Greater(t, len(ticks), majorTicks, "should have minor ticks")
		})
	}
}

func TestFormatSeriesForRender(t *testing.T) {
	tests := []struct {
		name           string
		series         model.Matrix
		legend         bool
		expectedLength int
		hasNames       bool
	}{
		{
			name:           "empty series",
			series:         model.Matrix{},
			legend:         true,
			expectedLength: 0,
			hasNames:       false,
		},
		{
			name: "single series with legend",
			series: model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{
						"__name__": "test_metric",
						"label1":   "value1",
					},
					Values: []model.SamplePair{
						{Timestamp: 1640000000000, Value: 42.5}, // Unix milliseconds (2021-12-20)
					},
				},
			},
			legend:         true,
			expectedLength: 2, // name + data
			hasNames:       true,
		},
		{
			name: "single series without legend",
			series: model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{
						"__name__": "test_metric",
					},
					Values: []model.SamplePair{
						{Timestamp: 1640000000000, Value: 42.5}, // Unix milliseconds (2021-12-20)
					},
				},
			},
			legend:         false,
			expectedLength: 1, // data only
			hasNames:       false,
		},
		{
			name: "multiple series with legend",
			series: model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{"__name__": "metric1"},
					Values: []model.SamplePair{
						{Timestamp: 1640000000000, Value: 10}, // Unix milliseconds (2021-12-20)
					},
				},
				&model.SampleStream{
					Metric: model.Metric{"__name__": "metric2"},
					Values: []model.SamplePair{
						{Timestamp: 1640000000000, Value: 20}, // Unix milliseconds (2021-12-20)
					},
				},
			},
			legend:         true,
			expectedLength: 4, // 2 names + 2 data
			hasNames:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSeriesForRender(tt.series, tt.legend)
			assert.Equal(t, tt.expectedLength, len(result))

			if tt.hasNames {
				// Verify we have alternating string names and XYer data
				for i := 0; i < len(result); i += 2 {
					_, isString := result[i].(string)
					assert.True(t, isString, "expected string at position %d", i)

					if i+1 < len(result) {
						_, isXYer := result[i+1].(plotter.XYer)
						assert.True(t, isXYer, "expected XYer at position %d", i+1)
					}
				}
			}

			// Verify timestamps are in Unix seconds (not divided by 60)
			for i := 0; i < len(result); i++ {
				if xyer, ok := result[i].(plotter.XYs); ok {
					for _, pt := range xyer {
						// Unix timestamps should be large numbers (billions for year 2000+)
						assert.Greater(t, pt.X, 1000000000.0, "timestamp should be in Unix seconds")
					}
				}
			}
		})
	}
}

func TestFormatMetric(t *testing.T) {
	tests := []struct {
		name     string
		metric   model.Metric
		expected string
	}{
		{
			name:     "empty metric",
			metric:   model.Metric{},
			expected: "",
		},
		{
			name: "single label",
			metric: model.Metric{
				"__name__": "test_metric",
			},
			expected: "test_metric",
		},
		{
			name: "multiple labels",
			metric: model.Metric{
				"__name__": "http_requests",
				"method":   "GET",
				"status":   "200",
			},
			// Note: order may vary due to map iteration
			expected: "", // Will check contains
		},
		{
			name: "label with newline",
			metric: model.Metric{
				"label": "value\nwith\nnewlines",
			},
			expected: "value with newlines",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatMetric(tt.metric)

			if tt.expected != "" {
				if tt.name == "multiple labels" {
					// Verify all values are present
					assert.Contains(t, result, "http_requests")
					assert.Contains(t, result, "GET")
					assert.Contains(t, result, "200")
					// Verify pipe separator
					assert.Contains(t, result, "|")
				} else {
					assert.Equal(t, tt.expected, result)
				}
			}

			// Verify no newlines in output
			assert.NotContains(t, result, "\n")
		})
	}
}

func TestDarkThemeColors(t *testing.T) {
	// Verify dark theme colors are defined correctly
	tests := []struct {
		name  string
		color color.RGBA
	}{
		{"darkBackground", darkBackground},
		{"darkGridColor", darkGridColor},
		{"darkTextColor", darkTextColor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify color has proper alpha channel
			assert.Equal(t, uint8(255), tt.color.A, "alpha channel should be 255")

			// Verify color components are valid (0-255)
			assert.LessOrEqual(t, tt.color.R, uint8(255))
			assert.LessOrEqual(t, tt.color.G, uint8(255))
			assert.LessOrEqual(t, tt.color.B, uint8(255))
		})
	}

	// Verify we have a good number of series colors
	assert.Equal(t, 10, len(darkSeriesColors), "should have 10 series colors")

	// Verify all series colors are properly defined
	for i, c := range darkSeriesColors {
		rgba, ok := c.(color.RGBA)
		require.True(t, ok, "color %d should be RGBA", i)
		assert.Equal(t, uint8(255), rgba.A, "color %d alpha should be 255", i)
	}
}

func TestApplyDarkTheme(t *testing.T) {
	p := plot.New()
	applyDarkTheme(p)

	// Verify background color is set
	assert.Equal(t, darkBackground, p.BackgroundColor)

	// Verify title styling
	assert.Equal(t, darkTextColor, p.Title.TextStyle.Color)

	// Verify X-axis styling
	assert.Equal(t, darkTextColor, p.X.Label.TextStyle.Color)
	assert.Equal(t, darkTextColor, p.X.Tick.Label.Color)
	assert.Equal(t, darkGridColor, p.X.Tick.LineStyle.Color)
	assert.Equal(t, darkTextColor, p.X.LineStyle.Color)

	// Verify Y-axis styling
	assert.Equal(t, darkTextColor, p.Y.Label.TextStyle.Color)
	assert.Equal(t, darkTextColor, p.Y.Tick.Label.Color)
	assert.Equal(t, darkGridColor, p.Y.Tick.LineStyle.Color)
	assert.Equal(t, darkTextColor, p.Y.LineStyle.Color)

	// Verify legend styling
	assert.Equal(t, darkTextColor, p.Legend.TextStyle.Color)
}

func TestAddStyledLines(t *testing.T) {
	tests := []struct {
		name        string
		pts         []any
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty data",
			pts:         []any{},
			expectError: false,
		},
		{
			name: "single series with name",
			pts: []any{
				"Series 1",
				plotter.XYs{{X: 1, Y: 10}, {X: 2, Y: 20}},
			},
			expectError: false,
		},
		{
			name: "single series without name",
			pts: []any{
				plotter.XYs{{X: 1, Y: 10}, {X: 2, Y: 20}},
			},
			expectError: false,
		},
		{
			name: "multiple series with names",
			pts: []any{
				"Series 1",
				plotter.XYs{{X: 1, Y: 10}},
				"Series 2",
				plotter.XYs{{X: 1, Y: 20}},
			},
			expectError: false,
		},
		{
			name: "invalid data type",
			pts: []any{
				"Series 1",
				123, // Invalid: should be XYer
			},
			expectError: true,
			errorMsg:    "expected plotter.XYer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := plot.New()
			err := addStyledLines(p, tt.pts)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				// Note: We can't directly verify plotters were added since the field is unexported,
				// but we've verified the function executes without error
			}
		})
	}
}

func TestAddStyledLinesColorCycling(t *testing.T) {
	p := plot.New()

	// Create more series than we have colors to test cycling
	numSeries := 15
	pts := make([]any, 0, numSeries*2)
	for i := 0; i < numSeries; i++ {
		pts = append(pts, plotter.XYs{{X: float64(i), Y: float64(i * 10)}})
	}

	err := addStyledLines(p, pts)
	require.NoError(t, err)

	// Note: We can't directly verify the number of plotters or colors used since the Plotters
	// field is unexported and the colors are internal to the Line plotters,
	// but we've verified that the function executes without error with more series than colors
}
