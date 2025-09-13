package workday

import (
	"job-scraper/types"
	"log/slog"
	"strings"
	"sync"

	"github.com/k3a/html2text"
	"resty.dev/v3"
)

type WorkdayJobPostingInfo struct {
	JobDescription string `json:"jobDescription"`
	ExternalUrl    string `json:"externalUrl"`
	JobReqId       string `json:"jobReqId"`
}

type WorkdayJobDetailsResponse struct {
	JobPostingInfo WorkdayJobPostingInfo `json:"jobPostingInfo"`
}

type JobScraper struct{}

func (scraper JobScraper) ScrapeJobDetails(jobsChannel <-chan *types.JobDetails, resultsChannel chan<- *types.JobDetails, numWorkers int) {
	var wg sync.WaitGroup

	slog.Info("Starting job detail scraping workers", "num_workers", numWorkers)

	// Start worker goroutines
	for i := range numWorkers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Create a separate client for this worker
			jobClient := resty.New()
			jobClient.SetHeader("User-Agent", "")
			defer jobClient.Close()

			slog.Info("Job scraping worker started", "worker_id", workerID)

			for job := range jobsChannel {
				slog.Info("Worker scraping job details", "worker_id", workerID, "jobLink", job.JobLink, "jobRole", job.JobRole)

				var jobDetailsResp WorkdayJobDetailsResponse
				resp, err := jobClient.R().
					SetHeaders(map[string]string{
						"cache-control":  "no-cache",
						"sec-fetch-dest": "document",
						"sec-fetch-mode": "navigate",
						"sec-fetch-site": "same-origin",
					}).
					SetResult(&jobDetailsResp).
					Get(job.JobLink)

				if err != nil {
					slog.Error("Failed to fetch job details", "worker_id", workerID, "jobLink", job.JobLink, "error", err)
					// Still send the job to results channel even if scraping failed
					resultsChannel <- job
					continue
				}

				// Get the parsed result
				result := resp.Result().(*WorkdayJobDetailsResponse)

				// Update job details
				job.JobID = result.JobPostingInfo.JobReqId
				job.JobLink = result.JobPostingInfo.ExternalUrl
				job.JobDetails = strings.TrimSpace(html2text.HTML2Text(result.JobPostingInfo.JobDescription))

				slog.Info("Successfully updated job details", "worker_id", workerID, "jobID", job.JobID, "title", job.JobRole)

				// Send completed job to results channel
				resultsChannel <- job
			}

			slog.Info("Job scraping worker finished", "worker_id", workerID)
		}(i)
	}

	// Wait for all workers to complete
	wg.Wait()
	close(resultsChannel)
	slog.Info("All job scraping workers completed")
}
