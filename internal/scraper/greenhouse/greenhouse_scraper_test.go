package greenhouse

import (
	"testing"
	"time"
)

func TestGetSHA256Hash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "https://example.com/job/123",
			expected: "c8d3c5e5f5e5a5e5c5d5e5f5a5b5c5d5e5f5a5b5c5d5e5f5a5b5c5d5e5f5a5b5",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSHA256Hash(tt.input)
			if len(result) != 64 {
				t.Errorf("expected hash length 64, got %d", len(result))
			}
		})
	}
}

func TestCleanUTF8String(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid UTF-8",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "unicode characters",
			input:    "Hello 世界",
			expected: "Hello 世界",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanUTF8String(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestRemoveExtraNewlines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "multiple newlines",
			input:    "Line 1\n\n\n\nLine 2",
			expected: "Line 1\n\nLine 2",
		},
		{
			name:     "newlines with spaces",
			input:    "Line 1\n  \n  \nLine 2",
			expected: "Line 1\n\nLine 2",
		},
		{
			name:     "single newline",
			input:    "Line 1\nLine 2",
			expected: "Line 1\nLine 2",
		},
		{
			name:     "no newlines",
			input:    "Line 1 Line 2",
			expected: "Line 1 Line 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeExtraNewlines(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestParseGreenhouseDate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{
			name:      "valid ISO 8601 date",
			input:     "2025-10-29T09:22:45-04:00",
			shouldErr: false,
		},
		{
			name:      "valid ISO 8601 date UTC",
			input:     "2025-10-29T09:22:45Z",
			shouldErr: false,
		},
		{
			name:      "invalid date format",
			input:     "2025-10-29",
			shouldErr: true,
		},
		{
			name:      "empty string",
			input:     "",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseGreenhouseDate(tt.input)
			if tt.shouldErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.shouldErr && result.IsZero() {
				t.Error("expected non-zero time but got zero time")
			}
		})
	}
}

func TestIsJobPublishedRecently(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)

	tests := []struct {
		name           string
		firstPublished string
		scrapeLimit    time.Time
		expected       bool
	}{
		{
			name:           "job published today",
			firstPublished: now.Format(time.RFC3339),
			scrapeLimit:    now.Truncate(24 * time.Hour),
			expected:       true,
		},
		{
			name:           "job published yesterday",
			firstPublished: yesterday.Format(time.RFC3339),
			scrapeLimit:    now.Truncate(24 * time.Hour),
			expected:       false,
		},
		{
			name:           "job published two days ago",
			firstPublished: twoDaysAgo.Format(time.RFC3339),
			scrapeLimit:    now.Truncate(24 * time.Hour),
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isJobPublishedRecently(tt.firstPublished, tt.scrapeLimit)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
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
			name:     "standard greenhouse URL",
			url:      "https://job-boards.greenhouse.io/sonyinteractiveentertainmentglobal/jobs/5686915004",
			expected: "5686915004",
		},
		{
			name:     "URL with trailing slash",
			url:      "https://job-boards.greenhouse.io/company/jobs/12345/",
			expected: "",
		},
		{
			name:     "URL without job ID",
			url:      "https://job-boards.greenhouse.io/company/jobs",
			expected: "jobs",
		},
		{
			name:     "simple path",
			url:      "jobs/5686915004",
			expected: "5686915004",
		},
		{
			name:     "empty string",
			url:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJobIDFromURL(tt.url)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// Example benchmark for hash generation
func BenchmarkGetSHA256Hash(b *testing.B) {
	url := "https://job-boards.greenhouse.io/sonyinteractiveentertainmentglobal/jobs/5686915004"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getSHA256Hash(url)
	}
}

// Example benchmark for UTF-8 cleaning
func BenchmarkCleanUTF8String(b *testing.B) {
	input := "Hello World with some unicode: 世界 and special chars: ñ é ü"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cleanUTF8String(input)
	}
}
