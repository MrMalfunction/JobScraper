package common

import (
	"crypto/sha256"
	"encoding/hex"
	"job-scraper/internal/db"
	"log/slog"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// GetSHA256Hash generates a SHA256 hash of the input string
func GetSHA256Hash(strToHash string) string {
	h := sha256.New()
	h.Write([]byte(strToHash))
	return hex.EncodeToString(h.Sum(nil))
}

// CleanUTF8String removes invalid UTF-8 sequences from a string
func CleanUTF8String(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	return strings.ToValidUTF8(s, "")
}

// RemoveExtraNewlines normalizes multiple consecutive newlines to double newlines
func RemoveExtraNewlines(s string) string {
	// Replace multiple consecutive newlines (with possible spaces/tabs between) with double newlines
	re := regexp.MustCompile(`\n[\s]*\n[\s\n]*`)
	cleaned := re.ReplaceAllString(s, "\n\n")

	// Replace any remaining multiple newlines with double newlines
	re2 := regexp.MustCompile(`\n{3,}`)
	cleaned = re2.ReplaceAllString(cleaned, "\n\n")

	return strings.TrimSpace(cleaned)
}

// InsertJobToDB inserts a job into the database if it doesn't already exist
// Returns true if the job was inserted, false if it already existed
func InsertJobToDB(job *db.Jobs, scraperName string) bool {
	dbResult := db.DB.FirstOrCreate(job, db.Jobs{JobHash: job.JobHash})

	if dbResult.Error != nil {
		slog.Error("["+scraperName+"_Worker] Failed to insert job into database",
			"jobLink", job.JobLink,
			"jobId", job.JobId,
			"error", dbResult.Error)
		return false
	}

	if dbResult.RowsAffected > 0 {
		slog.Info("["+scraperName+"_Worker] Job inserted into database",
			"jobLink", job.JobLink,
			"jobId", job.JobId,
			"company", job.CompanyName)
		return true
	}

	slog.Debug("["+scraperName+"_Worker] Job already exists, skipping insert",
		"jobLink", job.JobLink,
		"jobId", job.JobId)
	return false
}

// GetTodayMidnight returns today's date at midnight in local timezone
func GetTodayMidnight() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
}

// GetDateMidnight returns the given time at midnight in local timezone
func GetDateMidnight(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

// IsJobWithinScrapeLimit checks if a job's posted date is within the scrape date limit
// jobPostDate: the date when the job was posted
// scrapeDateLimit: the oldest date we want to scrape (inclusive)
// Returns true if the job should be scraped, false otherwise
func IsJobWithinScrapeLimit(jobPostDate time.Time, scrapeDateLimit time.Time) bool {
	jobDateMidnight := GetDateMidnight(jobPostDate)
	scrapeLimitMidnight := GetDateMidnight(scrapeDateLimit)

	return jobDateMidnight.After(scrapeLimitMidnight) || jobDateMidnight.Equal(scrapeLimitMidnight)
}

// IsJobFromToday checks if a job was posted today
func IsJobFromToday(jobPostDate time.Time) bool {
	jobDateMidnight := GetDateMidnight(jobPostDate)
	todayMidnight := GetTodayMidnight()

	return jobDateMidnight.Equal(todayMidnight)
}

// ShouldScrapeJob is a convenience function that checks if a job should be scraped
// based on the scrape date limit. This is the main function scrapers should use.
func ShouldScrapeJob(jobPostDate time.Time, scrapeDateLimit time.Time) bool {
	return IsJobWithinScrapeLimit(jobPostDate, scrapeDateLimit)
}
