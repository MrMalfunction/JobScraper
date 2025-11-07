package greenhouse

import (
	"testing"
	"time"
)

// Note: GetSHA256Hash, CleanUTF8String, and RemoveExtraNewlines tests
// are in job-scraper/internal/scraper/common/utils_test.go
// This test file only contains Greenhouse-specific function tests.

func TestParseGreenhouseDate(t *testing.T) {
	tests := []struct {
		name        string
		dateStr     string
		expectError bool
		checkFn     func(time.Time) bool
	}{
		{
			name:        "Valid RFC3339 format with timezone",
			dateStr:     "2025-10-29T09:22:45-04:00",
			expectError: false,
			checkFn: func(result time.Time) bool {
				expected := time.Date(2025, 10, 29, 9, 22, 45, 0, time.FixedZone("EDT", -4*3600))
				return result.Equal(expected)
			},
		},
		{
			name:        "Valid RFC3339 format with UTC",
			dateStr:     "2025-11-06T14:30:00Z",
			expectError: false,
			checkFn: func(result time.Time) bool {
				expected := time.Date(2025, 11, 6, 14, 30, 0, 0, time.UTC)
				return result.Equal(expected)
			},
		},
		{
			name:        "Valid RFC3339 format with positive timezone",
			dateStr:     "2025-12-25T18:00:00+05:30",
			expectError: false,
			checkFn: func(result time.Time) bool {
				expected := time.Date(2025, 12, 25, 18, 0, 0, 0, time.FixedZone("IST", 5*3600+30*60))
				return result.Equal(expected)
			},
		},
		{
			name:        "Invalid date format",
			dateStr:     "not-a-date",
			expectError: true,
			checkFn:     nil,
		},
		{
			name:        "Empty string",
			dateStr:     "",
			expectError: true,
			checkFn:     nil,
		},
		{
			name:        "Invalid format - missing timezone",
			dateStr:     "2025-10-29T09:22:45",
			expectError: true,
			checkFn:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseGreenhouseDate(tt.dateStr)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if tt.checkFn != nil && !tt.checkFn(result) {
					t.Errorf("Date check failed for input %q. Result: %v", tt.dateStr, result)
				}
			}
		})
	}
}

func TestIsJobPublishedRecently(t *testing.T) {
	now := time.Now()
	today := now.Truncate(24 * time.Hour)

	tests := []struct {
		name           string
		firstPublished string
		scrapeLimit    time.Time
		expected       bool
	}{
		{
			name:           "Job published today",
			firstPublished: now.Format(time.RFC3339),
			scrapeLimit:    today,
			expected:       true,
		},
		{
			name:           "Job published yesterday",
			firstPublished: now.AddDate(0, 0, -1).Format(time.RFC3339),
			scrapeLimit:    today,
			expected:       false,
		},
		{
			name:           "Job published on scrape limit date",
			firstPublished: today.Format(time.RFC3339),
			scrapeLimit:    today,
			expected:       true,
		},
		{
			name:           "Job published 7 days ago",
			firstPublished: now.AddDate(0, 0, -7).Format(time.RFC3339),
			scrapeLimit:    today.AddDate(0, 0, -6),
			expected:       false,
		},
		{
			name:           "Job published within range",
			firstPublished: now.AddDate(0, 0, -5).Format(time.RFC3339),
			scrapeLimit:    today.AddDate(0, 0, -7),
			expected:       true,
		},
		{
			name:           "Invalid date format returns false",
			firstPublished: "invalid-date",
			scrapeLimit:    today,
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isJobPublishedRecently(tt.firstPublished, tt.scrapeLimit)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for date %s with limit %v",
					tt.expected, result, tt.firstPublished, tt.scrapeLimit)
			}
		})
	}
}

func TestExtractJobIDFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "Standard Greenhouse URL",
			url:      "https://job-boards.greenhouse.io/sonyinteractiveentertainmentglobal/jobs/5686915004",
			expected: "5686915004",
		},
		{
			name:     "URL with trailing slash",
			url:      "https://job-boards.greenhouse.io/company/jobs/12345/",
			expected: "",
		},
		{
			name:     "Short URL",
			url:      "https://example.com/jobs/999",
			expected: "999",
		},
		{
			name:     "URL with query parameters",
			url:      "https://example.com/jobs/12345?ref=source",
			expected: "12345?ref=source",
		},
		{
			name:     "Empty string",
			url:      "",
			expected: "",
		},
		{
			name:     "Just a number",
			url:      "12345",
			expected: "12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJobIDFromURL(tt.url)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
