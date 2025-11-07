package common

import (
	"crypto/sha256"
	"encoding/hex"
	"job-scraper/internal/db"
	"log/slog"
	"regexp"
	"strings"
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
