package greenhouse

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"job-scraper/internal/db"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"

	"unicode/utf8"

	"github.com/k3a/html2text"
	"resty.dev/v3"
)

type GreenhouseJobLocation struct {
	Name string `json:"name"`
}

type GreenhouseJobListItem struct {
	ID             int                   `json:"id"`
	Title          string                `json:"title"`
	AbsoluteURL    string                `json:"absolute_url"`
	Location       GreenhouseJobLocation `json:"location"`
	UpdatedAt      string                `json:"updated_at"`
	RequisitionID  string                `json:"requisition_id"`
	InternalJobID  int                   `json:"internal_job_id"`
	FirstPublished string                `json:"first_published"`
	Language       string                `json:"language"`
}

type GreenhouseJobListResponse struct {
	Jobs []GreenhouseJobListItem `json:"jobs"`
}

type GreenhouseJobDetail struct {
	ID            int                   `json:"id"`
	Title         string                `json:"title"`
	Content       string                `json:"content"`
	UpdatedAt     string                `json:"updated_at"`
	RequisitionID string                `json:"requisition_id"`
	Location      GreenhouseJobLocation `json:"location"`
	AbsoluteURL   string                `json:"absolute_url"`
	InternalJobID int                   `json:"internal_job_id"`
	Language      string                `json:"language"`
}

type GreenhouseScraper struct{}

func getSHA256Hash(strToHash string) string {
	h := sha256.New()
	h.Write([]byte(strToHash))
	return hex.EncodeToString(h.Sum(nil))
}

func cleanUTF8String(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	// Remove invalid UTF-8 sequences
	return strings.ToValidUTF8(s, "")
}

func removeExtraNewlines(s string) string {
	// Replace multiple consecutive newlines (with possible spaces/tabs between) with double newlines
	re := regexp.MustCompile(`\n[\s]*\n[\s\n]*`)
	cleaned := re.ReplaceAllString(s, "\n\n")

	// Replace any remaining multiple newlines with double newlines
	re2 := regexp.MustCompile(`\n{3,}`)
	cleaned = re2.ReplaceAllString(cleaned, "\n\n")

	return strings.TrimSpace(cleaned)
}

func parseGreenhouseDate(dateStr string) (time.Time, error) {
	// Parse ISO 8601 format: "2025-10-29T09:22:45-04:00"
	parsedTime, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return time.Time{}, err
	}
	return parsedTime, nil
}

func isJobPublishedRecently(firstPublished string, scrapeDateLimitTruncated time.Time) bool {
	publishedTime, err := parseGreenhouseDate(firstPublished)
	if err != nil {
		slog.Error("Failed to parse first_published date", "date", firstPublished, "error", err)
		return false
	}

	publishedDateTruncated := publishedTime.Truncate(24 * time.Hour)

	// Check if the job was published within the last 24 hours or on the scrape date
	return publishedDateTruncated.After(scrapeDateLimitTruncated) || publishedDateTruncated.Equal(scrapeDateLimitTruncated)
}

func listJobsAndStartDetailsScrape(company db.Companies, scrapeDateLimitTruncated time.Time, jobDetailScrapeChannel chan<- *db.Jobs) {
	rClient := resty.New()
	rClient.SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	defer rClient.Close()

	// BaseUrl should be in format: https://boards-api.greenhouse.io/{board_token}
	if company.BaseUrl == "" {
		slog.Error("Base URL not found for company", "company", company.Name)
		return
	}

	apiURL := company.BaseUrl + "/jobs"
	slog.Info(apiURL)

	var jobListResp GreenhouseJobListResponse

	resp, err := rClient.R().
		SetHeaders(map[string]string{
			"Accept":         "application/json",
			"cache-control":  "no-cache",
			"sec-fetch-dest": "empty",
			"sec-fetch-mode": "cors",
			"sec-fetch-site": "cross-site",
		}).
		SetResult(&jobListResp).
		Get(apiURL)

	if err != nil {
		slog.Error("Failed to fetch jobs", "company", company.Name, "error", err)
		return
	}

	result := resp.Result().(*GreenhouseJobListResponse)

	slog.Info("Successfully fetched jobs", "company", company.Name, "total_jobs", len(result.Jobs))

	if len(result.Jobs) == 0 {
		slog.Info("No jobs found", "company", company.Name)
		return
	}

	recentJobsCount := 0
	for _, jobItem := range result.Jobs {
		// Check if job was published within the last 24 hours
		if isJobPublishedRecently(jobItem.FirstPublished, scrapeDateLimitTruncated) {
			recentJobsCount++

			publishedTime, _ := parseGreenhouseDate(jobItem.FirstPublished)

			job := &db.Jobs{
				JobHash:      "",
				JobId:        jobItem.RequisitionID,
				JobRole:      jobItem.Title,
				JobDetails:   "",
				JobPostDate:  publishedTime.Format("2006-01-02"),
				JobLink:      jobItem.AbsoluteURL,
				JobAISummary: "",
				CompanyName:  company.Name,
			}

			jobDetailScrapeChannel <- job
		}
	}

	slog.Info("Recent jobs queued for detailed scraping",
		"company", company.Name,
		"recent_jobs", recentJobsCount,
		"total_jobs", len(result.Jobs))
}

func (gs GreenhouseScraper) jobDetailsScraperWorker(baseURL string, jobChannel <-chan *db.Jobs) {
	slog.Debug("[Greenhouse_Scraper] Worker started to scrape Job Details")
	rClient := resty.New()
	rClient.SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	defer rClient.Close()

	for job := range jobChannel {
		// Extract job ID from the job object
		// We need to get the internal job ID from the absolute URL or use the ID from the list
		// For now, we'll extract it from the URL
		jobID := extractJobIDFromURL(job.JobLink)
		if jobID == "" {
			slog.Error("[Greenhouse_Scraper_Worker] Failed to extract job ID from URL", "jobLink", job.JobLink)
			continue
		}

		apiURL := fmt.Sprintf("%s/jobs/%s", baseURL, jobID)

		var jobDetailsResp GreenhouseJobDetail
		resp, err := rClient.R().
			SetHeaders(map[string]string{
				"Accept":         "application/json",
				"cache-control":  "no-cache",
				"sec-fetch-dest": "empty",
				"sec-fetch-mode": "cors",
				"sec-fetch-site": "cross-site",
			}).
			SetResult(&jobDetailsResp).
			Get(apiURL)

		if err != nil {
			slog.Error("[Greenhouse_Scraper_Worker] Failed to fetch job details", "jobLink", job.JobLink, "error", err)
			continue
		}

		// Get the parsed result
		result := resp.Result().(*GreenhouseJobDetail)

		// Update job details
		job.JobId = result.RequisitionID
		job.JobLink = result.AbsoluteURL
		job.JobDetails = removeExtraNewlines(cleanUTF8String(html2text.HTML2Text(result.Content)))
		job.JobRole = result.Title
		job.JobHash = getSHA256Hash(job.JobLink)

		// Insert job into database only if it doesn't exist
		dbResult := db.DB.FirstOrCreate(job, db.Jobs{JobHash: job.JobHash})
		if dbResult.Error != nil {
			slog.Error("[Greenhouse_Scraper_Worker] Failed to insert job into database",
				"jobLink", job.JobLink,
				"jobId", job.JobId,
				"error", dbResult.Error)
		} else if dbResult.RowsAffected > 0 {
			slog.Info("[Greenhouse_Scraper_Worker] Job inserted into database",
				"jobLink", job.JobLink,
				"jobId", job.JobId)
		} else {
			slog.Info("[Greenhouse_Scraper_Worker] Job already exists, skipping insert",
				"jobLink", job.JobLink,
				"jobId", job.JobId)
		}

		slog.Info("[Greenhouse_Scraper_Worker] Job Scraped", "JobLink", job.JobLink)
	}
	slog.Info("[Greenhouse_Scraper_Worker] Job Details Worker shutting down")
}

// extractJobIDFromURL extracts the job ID from Greenhouse job URL
// Example: https://job-boards.greenhouse.io/sonyinteractiveentertainmentglobal/jobs/5686915004 -> 5686915004
func extractJobIDFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func (gs GreenhouseScraper) StartScraping(companiesToScrape <-chan db.Companies, scrapeDayLimit time.Time) {
	scrapeDateLimitTruncated := scrapeDayLimit.Truncate(24 * time.Hour)

	jobDetailScrapeChannel := make(chan *db.Jobs, 10000)
	slog.Info("[Greenhouse_Scraper] Greenhouse Jobs Details channel created")

	scraperWorkerCount := 4

	// Use WaitGroup to track company listing workers
	var wg sync.WaitGroup

	for company := range companiesToScrape {
		// Start workers for this company's base URL
		if company.BaseUrl == "" {
			slog.Error("[Greenhouse_Scraper] Base URL not found for company", "company", company.Name)
			continue
		}

		// Start workers if not already started
		for range make([]struct{}, scraperWorkerCount) {
			go gs.jobDetailsScraperWorker(company.BaseUrl, jobDetailScrapeChannel)
			slog.Info("[Greenhouse_Scraper] Job details scraper worker started", "company", company.Name)
		}

		wg.Go(func() {
			listJobsAndStartDetailsScrape(company, scrapeDateLimitTruncated, jobDetailScrapeChannel)
		})
	}

	// Wait for all companies to finish listing jobs, then close the channel
	wg.Wait()
	close(jobDetailScrapeChannel)
	slog.Info("[Greenhouse_Scraper] Greenhouse Companies Job list complete.")
}
