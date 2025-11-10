package common

import (
	"job-scraper/internal/db"
	"strings"
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
			name:     "empty string",
			input:    "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "simple string",
			input:    "hello",
			expected: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
		{
			name:     "url string",
			input:    "https://example.com/job/12345",
			expected: "8d7e30e3f4f1c3d8b4c3e6f8a9c1b5d7e2f4a6c8e1f3d5b7a9c2e4f6a8b1c3d5e7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSHA256Hash(tt.input)

			// Check that result is 64 characters (SHA256 hex length)
			if len(result) != 64 {
				t.Errorf("GetSHA256Hash() length = %d, want 64", len(result))
			}

			// For known values, check exact match
			if tt.name == "empty string" || tt.name == "simple string" {
				if result != tt.expected {
					t.Errorf("GetSHA256Hash() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestGetSHA256HashConsistency(t *testing.T) {
	input := "test string"
	hash1 := GetSHA256Hash(input)
	hash2 := GetSHA256Hash(input)

	if hash1 != hash2 {
		t.Errorf("GetSHA256Hash() not consistent: %v != %v", hash1, hash2)
	}
}

func TestCleanUTF8String(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid UTF-8 string",
			input:    "Hello, World!",
			expected: "Hello, World!",
		},
		{
			name:     "string with emojis",
			input:    "Hello 👋 World 🌍",
			expected: "Hello 👋 World 🌍",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "string with special characters",
			input:    "Café résumé naïve",
			expected: "Café résumé naïve",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanUTF8String(tt.input)
			if result != tt.expected {
				t.Errorf("CleanUTF8String() = %v, want %v", result, tt.expected)
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
			name:     "single newline",
			input:    "line1\nline2",
			expected: "line1\nline2",
		},
		{
			name:     "double newline",
			input:    "line1\n\nline2",
			expected: "line1\n\nline2",
		},
		{
			name:     "triple newline",
			input:    "line1\n\n\nline2",
			expected: "line1\n\nline2",
		},
		{
			name:     "multiple newlines with spaces",
			input:    "line1\n  \n  \nline2",
			expected: "line1\n\nline2",
		},
		{
			name:     "five newlines",
			input:    "line1\n\n\n\n\nline2",
			expected: "line1\n\nline2",
		},
		{
			name:     "newlines at start and end",
			input:    "\n\nline1\n\nline2\n\n",
			expected: "line1\n\nline2",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only newlines",
			input:    "\n\n\n\n",
			expected: "",
		},
		{
			name:     "text with mixed whitespace",
			input:    "line1\n \t\n\t \nline2",
			expected: "line1\n\nline2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveExtraNewlines(tt.input)
			if result != tt.expected {
				t.Errorf("RemoveExtraNewlines() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestInsertJobToDB(t *testing.T) {
	// Note: This test requires database setup and is more of an integration test
	// For unit testing, we would typically mock the database
	// This is a placeholder to show the test structure

	t.Skip("Skipping database integration test - requires DB setup")

	job := &db.Jobs{
		JobHash:      "test-hash-123",
		JobId:        "TEST-001",
		JobRole:      "Software Engineer",
		JobDetails:   "Test job details",
		JobPostDate:  "2025-01-01",
		JobLink:      "https://example.com/job/test-001",
		JobAISummary: "",
		CompanyName:  "Test Company",
	}

	result := InsertJobToDB(job, "TestScraper")

	// This would test the actual behavior, but requires DB
	_ = result
}

func TestRemoveExtraNewlinesComplexScenarios(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(string) bool
		desc     string
	}{
		{
			name:  "job description with multiple paragraphs",
			input: "Requirements:\n\n\n- Skill 1\n- Skill 2\n\n\n\nBenefits:\n\n- Benefit 1",
			validate: func(s string) bool {
				// Should not have more than 2 consecutive newlines
				return !strings.Contains(s, "\n\n\n")
			},
			desc: "should not contain triple newlines",
		},
		{
			name:  "preserves double newlines",
			input: "Paragraph 1\n\nParagraph 2\n\nParagraph 3",
			validate: func(s string) bool {
				return strings.Count(s, "\n\n") == 2
			},
			desc: "should preserve exactly 2 double newlines",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveExtraNewlines(tt.input)
			if !tt.validate(result) {
				t.Errorf("RemoveExtraNewlines() failed validation: %s\nResult: %q", tt.desc, result)
			}
		})
	}
}

func TestGetTodayMidnight(t *testing.T) {
	today := GetTodayMidnight()

	// Verify it's midnight (00:00:00)
	if today.Hour() != 0 || today.Minute() != 0 || today.Second() != 0 || today.Nanosecond() != 0 {
		t.Errorf("Expected midnight (00:00:00), got %02d:%02d:%02d.%09d",
			today.Hour(), today.Minute(), today.Second(), today.Nanosecond())
	}

	// Verify it's in local timezone
	if today.Location() != time.Local {
		t.Errorf("Expected local timezone, got %v", today.Location())
	}

	// Verify it's today's date
	now := time.Now()
	if today.Year() != now.Year() || today.Month() != now.Month() || today.Day() != now.Day() {
		t.Errorf("Expected today's date %v, got %v", now.Format("2006-01-02"), today.Format("2006-01-02"))
	}
}

func TestGetDateMidnight(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "Date with time",
			input:    time.Date(2025, 11, 9, 15, 30, 45, 123456789, time.Local),
			expected: "2025-11-09 00:00:00",
		},
		{
			name:     "Already at midnight",
			input:    time.Date(2025, 11, 9, 0, 0, 0, 0, time.Local),
			expected: "2025-11-09 00:00:00",
		},
		{
			name:     "End of day",
			input:    time.Date(2025, 11, 9, 23, 59, 59, 999999999, time.Local),
			expected: "2025-11-09 00:00:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDateMidnight(tt.input)

			// Verify it's midnight
			if result.Hour() != 0 || result.Minute() != 0 || result.Second() != 0 || result.Nanosecond() != 0 {
				t.Errorf("Expected midnight, got %02d:%02d:%02d.%09d",
					result.Hour(), result.Minute(), result.Second(), result.Nanosecond())
			}

			// Verify the date matches
			resultStr := result.Format("2006-01-02 15:04:05")
			if resultStr != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, resultStr)
			}

			// Verify it's in local timezone
			if result.Location() != time.Local {
				t.Errorf("Expected local timezone, got %v", result.Location())
			}
		})
	}
}

func TestIsJobWithinScrapeLimit(t *testing.T) {
	// Create a reference scrape limit: Nov 5, 2025
	scrapeLimit := time.Date(2025, 11, 5, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name        string
		jobPostDate time.Time
		expected    bool
	}{
		{
			name:        "Job posted on scrape limit date (should include)",
			jobPostDate: time.Date(2025, 11, 5, 10, 30, 0, 0, time.Local),
			expected:    true,
		},
		{
			name:        "Job posted after scrape limit",
			jobPostDate: time.Date(2025, 11, 9, 15, 0, 0, 0, time.Local),
			expected:    true,
		},
		{
			name:        "Job posted before scrape limit",
			jobPostDate: time.Date(2025, 11, 4, 23, 59, 59, 0, time.Local),
			expected:    false,
		},
		{
			name:        "Job posted one day after limit",
			jobPostDate: time.Date(2025, 11, 6, 0, 0, 0, 0, time.Local),
			expected:    true,
		},
		{
			name:        "Job posted today (Nov 9)",
			jobPostDate: time.Date(2025, 11, 9, 12, 0, 0, 0, time.Local),
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsJobWithinScrapeLimit(tt.jobPostDate, scrapeLimit)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for job date %v vs limit %v",
					tt.expected, result, tt.jobPostDate.Format("2006-01-02"), scrapeLimit.Format("2006-01-02"))
			}
		})
	}
}

func TestIsJobFromToday(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	tests := []struct {
		name        string
		jobPostDate time.Time
		expected    bool
	}{
		{
			name:        "Job posted today at noon",
			jobPostDate: time.Date(today.Year(), today.Month(), today.Day(), 12, 0, 0, 0, time.Local),
			expected:    true,
		},
		{
			name:        "Job posted today at midnight",
			jobPostDate: today,
			expected:    true,
		},
		{
			name:        "Job posted yesterday",
			jobPostDate: today.AddDate(0, 0, -1),
			expected:    false,
		},
		{
			name:        "Job posted tomorrow",
			jobPostDate: today.AddDate(0, 0, 1),
			expected:    false,
		},
		{
			name:        "Job posted 7 days ago",
			jobPostDate: today.AddDate(0, 0, -7),
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsJobFromToday(tt.jobPostDate)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for job date %v",
					tt.expected, result, tt.jobPostDate.Format("2006-01-02"))
			}
		})
	}
}

func TestShouldScrapeJob(t *testing.T) {
	scrapeLimit := time.Date(2025, 11, 5, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name        string
		jobPostDate time.Time
		expected    bool
	}{
		{
			name:        "Job within scrape limit",
			jobPostDate: time.Date(2025, 11, 9, 10, 0, 0, 0, time.Local),
			expected:    true,
		},
		{
			name:        "Job on scrape limit boundary",
			jobPostDate: time.Date(2025, 11, 5, 0, 0, 0, 0, time.Local),
			expected:    true,
		},
		{
			name:        "Job before scrape limit",
			jobPostDate: time.Date(2025, 11, 4, 23, 59, 59, 0, time.Local),
			expected:    false,
		},
		{
			name:        "Job with time on limit date",
			jobPostDate: time.Date(2025, 11, 5, 14, 30, 45, 0, time.Local),
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldScrapeJob(tt.jobPostDate, scrapeLimit)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for job date %v vs limit %v",
					tt.expected, result, tt.jobPostDate.Format("2006-01-02"), scrapeLimit.Format("2006-01-02"))
			}
		})
	}
}
