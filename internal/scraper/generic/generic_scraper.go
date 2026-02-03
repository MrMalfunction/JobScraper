package generic

import (
	"encoding/json"
	"fmt"
	"job-scraper/internal/db"
	"job-scraper/internal/scraper/common"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/k3a/html2text"
	"github.com/tidwall/gjson"
	"resty.dev/v3"
)

type GenericScraper struct{}

// extractValueFromJSON extracts value from JSON using gjson path
func extractValueFromJSON(jsonData string, path string) string {
	if path == "" {
		return ""
	}
	result := gjson.Get(jsonData, path)
	if !result.Exists() {
		return ""
	}
	return result.String()
}

// listJobsAndStartDetailsScrape fetches jobs from a generic API endpoint
func listJobsAndStartDetailsScrape(company db.Companies, scrapeDateLimitTruncated time.Time, jobDetailScrapeChannel chan<- *db.Jobs) {
	rClient := resty.New()
	rClient.SetHeader("User-Agent", "JobScraper/1.0")
	defer rClient.Close()

	// Parse headers
	var headers map[string]string
	if company.ApiRequestHeaders != "" {
		if err := json.Unmarshal([]byte(company.ApiRequestHeaders), &headers); err != nil {
			slog.Error("error unmarshaling headers", "error", err, "company", company.Name)
			return
		}
	}

	// Parse body if exists
	var reqBody map[string]interface{}
	if company.ApiRequestBody != "" {
		if err := json.Unmarshal([]byte(company.ApiRequestBody), &reqBody); err != nil {
			slog.Error("error unmarshaling request body", "error", err, "company", company.Name)
			return
		}
	} else {
		reqBody = make(map[string]interface{})
	}

	offset := 0
	pageNum := 1
	for {
		// Set pagination value based on pagination key
		if company.PaginationKey != "" {
			// Support both offset-based and page-based pagination
			if company.PaginationKey == "page" {
				reqBody[company.PaginationKey] = pageNum
			} else {
				// Default to offset-based pagination for other keys
				reqBody[company.PaginationKey] = offset
			}
		}

		// Build request
		req := rClient.R()

		// Add headers
		if len(headers) > 0 {
			req.SetHeaders(headers)
		}

		// Execute request based on method
		var resp *resty.Response
		var err error

		fullUrl := company.BaseUrl
		if company.ApiRequestQueryParam != "" {
			fullUrl += "?" + company.ApiRequestQueryParam
		}

		if company.ApiRequestMethod == "POST" {
			resp, err = req.SetBody(reqBody).Post(fullUrl)
		} else {
			// For GET requests, convert body to query params
			if len(reqBody) > 0 {
				req.SetQueryParams(convertMapToStringMap(reqBody))
			}
			resp, err = req.Get(fullUrl)
		}

		if err != nil {
			slog.Error("Failed to fetch jobs", "company", company.Name, "error", err)
			break
		}

		if resp.StatusCode() != 200 {
			slog.Error("Non-200 status code", "company", company.Name, "status", resp.StatusCode())
			break
		}

		responseBody := resp.String()

		// Extract jobs array using JSON path
		jobsResult := gjson.Get(responseBody, company.ResponseJsonPath)
		if !jobsResult.Exists() {
			slog.Warn("Response JSON path not found", "company", company.Name, "path", company.ResponseJsonPath)
			break
		}

		if !jobsResult.IsArray() {
			slog.Warn("Response JSON path does not point to an array", "company", company.Name)
			break
		}

		jobs := jobsResult.Array()
		slog.Info("Successfully fetched jobs", "company", company.Name, "offset", offset, "page", pageNum, "jobs_in_response", len(jobs))

		if len(jobs) == 0 {
			slog.Info("No more jobs found, stopping pagination", "company", company.Name)
			break
		}

		allJobsTooOld := true
		for _, jobResult := range jobs {
			jobJSON := jobResult.Raw

			// Extract job fields using JSON paths
			jobId := extractValueFromJSON(jobJSON, company.JobIdJsonPath)
			jobTitle := extractValueFromJSON(jobJSON, company.JobTitleJsonPath)
			jobLink := extractValueFromJSON(jobJSON, company.JobLinkJsonPath)
			jobDetails := extractValueFromJSON(jobJSON, company.JobDetailsJsonPath)
			jobDateStr := extractValueFromJSON(jobJSON, company.JobDateJsonPath)

			// Parse job date
			var jobPostDate time.Time
			if jobDateStr != "" {
				// Try to parse various date formats
				jobPostDate = parseJobDate(jobDateStr)
			} else {
				// Default to current time if no date is available
				jobPostDate = time.Now()
			}

			// Check if we should scrape this job
			if common.ShouldScrapeJob(jobPostDate, scrapeDateLimitTruncated) {
				allJobsTooOld = false

				// Clean job details if HTML
				cleanDetails := jobDetails
				if jobDetails != "" {
					cleanDetails = common.RemoveExtraNewlines(common.CleanUTF8String(html2text.HTML2Text(jobDetails)))
				}

				job := &db.Jobs{
					JobHash:      common.GetSHA256Hash(jobLink),
					JobId:        jobId,
					JobRole:      jobTitle,
					JobDetails:   cleanDetails,
					JobPostDate:  jobPostDate.Format("2006-01-02"),
					JobLink:      jobLink,
					JobAISummary: "",
					CompanyName:  company.Name,
				}

				jobDetailScrapeChannel <- job
			}
		}

		if allJobsTooOld {
			slog.Info("All jobs too old, stopping pagination", "company", company.Name)
			break
		}

		offset += len(jobs)
		pageNum++

		// Safety check: prevent infinite loops
		if offset > 10000 || pageNum > 500 {
			slog.Warn("Pagination limit reached", "company", company.Name, "offset", offset, "page", pageNum)
			break
		}
	}
}

// jobDetailsScraperWorker processes jobs from the channel and inserts them into the database
func (gs GenericScraper) jobDetailsScraperWorker(jobChannel <-chan *db.Jobs) {
	slog.Debug("[Generic_Scraper] Worker started to process jobs")

	for job := range jobChannel {
		// For generic scraper, all data is already extracted, just insert
		common.InsertJobToDB(job, "Generic_Scraper")
		slog.Info("[Generic_Scraper_Worker] Job processed", "JobLink", job.JobLink, "JobId", job.JobId)
	}

	slog.Info("[Generic_Scraper_Worker] Job processing worker shutting down")
}

// StartScraping implements the scraper interface
func (gs GenericScraper) StartScraping(companiesToScrape <-chan db.Companies, scrapeDayLimit time.Time) {
	scrapeDateLimitTruncated := common.GetDateMidnight(scrapeDayLimit)

	jobDetailScrapeChannel := make(chan *db.Jobs, 10000)
	slog.Info("[Generic_Scraper] Generic jobs channel created")

	scraperWorkerCount := 4
	for range make([]struct{}, scraperWorkerCount) {
		go gs.jobDetailsScraperWorker(jobDetailScrapeChannel)
		slog.Info("[Generic_Scraper] Job processing worker started")
	}

	var wg sync.WaitGroup
	for company := range companiesToScrape {
		wg.Add(1)
		go func(c db.Companies) {
			defer wg.Done()
			listJobsAndStartDetailsScrape(c, scrapeDateLimitTruncated, jobDetailScrapeChannel)
		}(company)
	}

	wg.Wait()
	close(jobDetailScrapeChannel)
	slog.Info("[Generic_Scraper] Generic companies job list complete")
}

// parseJobDate attempts to parse various date formats
func parseJobDate(dateStr string) time.Time {
	// Try common date formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"01/02/2006",
		"02-01-2006",
		"Jan 2, 2006",
		"January 2, 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}

	// If parsing fails, return current time
	slog.Warn("Failed to parse date, using current time", "dateStr", dateStr)
	return time.Now()
}

// convertMapToStringMap converts map[string]interface{} to map[string]string
func convertMapToStringMap(m map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		switch val := v.(type) {
		case string:
			result[k] = val
		case int:
			result[k] = strconv.Itoa(val)
		case int64:
			result[k] = strconv.FormatInt(val, 10)
		case float64:
			result[k] = strconv.FormatFloat(val, 'f', -1, 64)
		case bool:
			result[k] = strconv.FormatBool(val)
		default:
			result[k] = fmt.Sprintf("%v", v)
		}
	}
	return result
}
