package main

import (
	"encoding/csv"
	"fmt"
	"job-scraper/internal"
	"job-scraper/internal/types"
	"log/slog"
	"os"
)

func main() {
	// Set max days to scrape - jobs posted within the last X days will be scraped
	// maxDaysToScrape := 1 // Scrape jobs posted within the last 1 day
	// scrapeDateLimit := time.Now().AddDate(0, 0, -maxDaysToScrape)
	internal.StartServer()

}

func writeJobsToCSVChannel(jobsChannel <-chan *types.JobDetails, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		slog.Error("Failed to create CSV file", "error", err)
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	header := []string{"ULid", "JobID", "JobRole", "JobDetails", "JobPostDate", "JobLink", "JobCompanyName"}
	if err := writer.Write(header); err != nil {
		slog.Error("Failed to write CSV header", "error", err)
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write job data as it arrives from the channel
	jobCount := 0
	for job := range jobsChannel {
		record := []string{
			job.JUlid.String(),
			job.JobID,
			job.JobRole,
			job.JobDetails,
			job.JobPostDate.Format("2006-01-02 15:04:05"),
			job.JobLink,
			job.JobCompanyName,
		}
		if err := writer.Write(record); err != nil {
			slog.Error("Failed to write job record", "error", err, "jobID", job.JobID)
			return fmt.Errorf("failed to write job record: %w", err)
		}
		jobCount++

		// Flush periodically to ensure data is written
		if jobCount%10 == 0 {
			writer.Flush()
		}

		slog.Info("Job written to CSV", "jobID", job.JobID, "title", job.JobRole, "total_written", jobCount)
	}

	slog.Info("Finished writing jobs to CSV", "total_jobs", jobCount, "filename", filename)
	return nil
}
