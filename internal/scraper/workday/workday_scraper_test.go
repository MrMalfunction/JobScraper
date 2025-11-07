package workday

import (
	"testing"
	"time"
)

// Note: GetSHA256Hash, CleanUTF8String, and RemoveExtraNewlines tests
// are in job-scraper/internal/scraper/common/utils_test.go
// This test file only contains Workday-specific function tests.

func TestParsePostedDate(t *testing.T) {
	tests := []struct {
		name     string
		postedOn string
		checkFn  func(time.Time) bool // Custom check function for flexibility
	}{
		{
			name:     "Posted Today",
			postedOn: "Posted Today",
			checkFn: func(result time.Time) bool {
				// Should be within last 24 hours
				return time.Since(result).Hours() < 24
			},
		},
		{
			name:     "Posted today (lowercase)",
			postedOn: "Posted today",
			checkFn: func(result time.Time) bool {
				return time.Since(result).Hours() < 24
			},
		},
		{
			name:     "Posted Yesterday",
			postedOn: "Posted Yesterday",
			checkFn: func(result time.Time) bool {
				// Should be between 24-48 hours ago
				hours := time.Since(result).Hours()
				return hours >= 24 && hours < 48
			},
		},
		{
			name:     "Posted yesterday (lowercase)",
			postedOn: "Posted yesterday",
			checkFn: func(result time.Time) bool {
				hours := time.Since(result).Hours()
				return hours >= 24 && hours < 48
			},
		},
		{
			name:     "Posted 2 Days Ago",
			postedOn: "Posted 2 Days Ago",
			checkFn: func(result time.Time) bool {
				// Should be approximately 2 days ago (within 12 hours margin)
				expected := time.Now().AddDate(0, 0, -2)
				diff := expected.Sub(result).Hours()
				return diff >= -12 && diff <= 12
			},
		},
		{
			name:     "Posted 5 Days Ago",
			postedOn: "Posted 5 Days Ago",
			checkFn: func(result time.Time) bool {
				expected := time.Now().AddDate(0, 0, -5)
				diff := expected.Sub(result).Hours()
				return diff >= -12 && diff <= 12
			},
		},
		{
			name:     "Posted 10 Days Ago",
			postedOn: "Posted 10 Days Ago",
			checkFn: func(result time.Time) bool {
				expected := time.Now().AddDate(0, 0, -10)
				diff := expected.Sub(result).Hours()
				return diff >= -12 && diff <= 12
			},
		},
		{
			name:     "Posted 30 Days Ago",
			postedOn: "Posted 30 Days Ago",
			checkFn: func(result time.Time) bool {
				expected := time.Now().AddDate(0, 0, -30)
				diff := expected.Sub(result).Hours()
				return diff >= -12 && diff <= 12
			},
		},
		{
			name:     "Posted 7 Days Ago (with extra spaces)",
			postedOn: "Posted  7  Days  Ago",
			checkFn: func(result time.Time) bool {
				expected := time.Now().AddDate(0, 0, -7)
				diff := expected.Sub(result).Hours()
				return diff >= -12 && diff <= 12
			},
		},
		{
			name:     "Posted 15 Day Ago (singular Day)",
			postedOn: "Posted 15 Day Ago",
			checkFn: func(result time.Time) bool {
				expected := time.Now().AddDate(0, 0, -15)
				diff := expected.Sub(result).Hours()
				return diff >= -12 && diff <= 12
			},
		},
		{
			name:     "Invalid format - returns current time",
			postedOn: "Some random text",
			checkFn: func(result time.Time) bool {
				// Should be very recent (within 1 second)
				return time.Since(result).Seconds() < 1
			},
		},
		{
			name:     "Empty string - returns current time",
			postedOn: "",
			checkFn: func(result time.Time) bool {
				return time.Since(result).Seconds() < 1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePostedDate(tt.postedOn)

			if !tt.checkFn(result) {
				t.Errorf("Date check failed for input %q. Result: %v", tt.postedOn, result)
			}
		})
	}
}
