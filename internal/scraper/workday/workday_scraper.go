package workday

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"job-scraper/internal/db"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"unicode/utf8"

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

type WorkdayListJobs struct {
	Title         string `json:"title"`
	ExternalPath  string `json:"externalPath"`
	LocationsText string `json:"locationsText"`
	PostedOn      string `json:"postedOn"`
}

type WorkdayResponse struct {
	Total       int               `json:"total"`
	JobPostings []WorkdayListJobs `json:"jobPostings"`
}

func parsePostedDate(postedOn string) time.Time {
	// Parse "Posted X Days Ago" format
	re := regexp.MustCompile(`Posted\s+(\d+)\s+Days?\s+Ago`)
	matches := re.FindStringSubmatch(postedOn)

	if len(matches) >= 2 {
		days, err := strconv.Atoi(matches[1])
		if err == nil {
			return time.Now().AddDate(0, 0, -days)
		}
	}

	// Handle "Posted Today" or similar
	if strings.Contains(strings.ToLower(postedOn), "today") {
		return time.Now()
	}

	// Handle "Posted Yesterday" or similar
	// Subtract more than 24 hours to ensure proper truncation handling
	if strings.Contains(strings.ToLower(postedOn), "yesterday") {
		return time.Now().Add(-25 * time.Hour)
	}

	// Default to current time if parsing fails
	return time.Now()
}

type WorkdayScraper struct{}

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

func listJobsAndStartDetailsScrape(company db.Companies, scrapeDateLimitTruncated time.Time, jobDetailScrapeChannel chan<- *db.Jobs) {
	rClient := resty.New()
	rClient.SetHeader("User-Agent", "")
	defer rClient.Close()

	var req_body map[string]interface{}

	err := json.Unmarshal([]byte(company.ApiRequestBody), &req_body)
	if err != nil {
		slog.Error("error unmarshaling JSON")
	}
	slog.Info("req_body", "req_body", fmt.Sprint(req_body))

	offset := 0
	for {
		req_body["offset"] = offset
		var workdayResp WorkdayResponse

		resp, err := rClient.R().
			// SetDebug(true).
			SetHeaders(
				map[string]string{
					"cache-control":  "no-cache",
					"sec-fetch-dest": "empty",
					"sec-fetch-mode": "cors",
					"sec-fetch-site": "same-origin",
				}).
			SetBody(req_body).
			SetResult(&workdayResp).
			Post(company.BaseUrl + "/jobs")

		if err != nil {
			slog.Error("Failed to fetch jobs", "company", company.Name, "error", err)
			break
		}

		result := resp.Result().(*WorkdayResponse)

		// Get the parsed result
		slog.Info("Successfully fetched jobs", "company", company.Name, "offset", offset, "jobs_in_response", len(result.JobPostings))
		// break

		if len(result.JobPostings) == 0 {
			slog.Info("No more jobs found, stopping pagination", "company", company.Name)
			break
		}
		allJobsTooOld := true
		for _, posting := range result.JobPostings {
			jobPostDate := parsePostedDate(posting.PostedOn)
			jobPostDateTruncated := jobPostDate.Truncate(24 * time.Hour)

			if jobPostDateTruncated.After(scrapeDateLimitTruncated) || jobPostDateTruncated.Equal(scrapeDateLimitTruncated) {
				allJobsTooOld = false
				job := &db.Jobs{
					JobHash:      "",
					JobId:        "",
					JobRole:      posting.Title,
					JobDetails:   "",
					JobPostDate:  jobPostDate.Format("2006-01-02"),
					JobLink:      company.BaseUrl + posting.ExternalPath,
					JobAISummary: "",
					CompanyName:  company.Name,
				}

				jobDetailScrapeChannel <- job
			} else {
				// Job out of range skipped scraping
			}
		}
		if allJobsTooOld {
			break
		}

		offset += len(result.JobPostings)
	}

}

func (ws WorkdayScraper) jobDetailsScraperWorker(jobChannel <-chan *db.Jobs) {
	slog.Debug("[Workday_Scraper] Worker started to scrape Job Details")
	rClient := resty.New()
	rClient.SetHeader("User-Agent", "")
	defer rClient.Close()

	for job := range jobChannel {
		var jobDetailsResp WorkdayJobDetailsResponse
		resp, err := rClient.R().
			SetHeaders(map[string]string{
				"cache-control":  "no-cache",
				"sec-fetch-dest": "document",
				"sec-fetch-mode": "navigate",
				"sec-fetch-site": "same-origin",
			}).
			SetResult(&jobDetailsResp).
			Get(job.JobLink)
		if err != nil {
			slog.Error("[Workday_Scraper_Worker] Failed to fetch job details", "jobLink", job.JobLink, "error", err)
			continue
		}

		// Get the parsed result
		result := resp.Result().(*WorkdayJobDetailsResponse)

		// Update job details
		job.JobId = (result.JobPostingInfo.JobReqId)
		job.JobLink = (result.JobPostingInfo.ExternalUrl)
		job.JobDetails = removeExtraNewlines(cleanUTF8String(html2text.HTML2Text(result.JobPostingInfo.JobDescription)))
		job.JobRole = (job.JobRole)
		job.CompanyName = (job.CompanyName)
		job.JobHash = getSHA256Hash(job.JobLink)

		// Insert job into database only if it doesn't exist
		dbResult := db.DB.FirstOrCreate(job, db.Jobs{JobHash: job.JobHash})
		if dbResult.Error != nil {
			slog.Error("[Workday_Scraper_Worker] Failed to insert job into database",
				"jobLink", job.JobLink,
				"jobId", job.JobId,
				"error", dbResult.Error)
		} else if dbResult.RowsAffected > 0 {
			slog.Info("[Workday_Scraper_Worker] Job inserted into database",
				"jobLink", job.JobLink,
				"jobId", job.JobId)
		} else {
			slog.Info("[Workday_Scraper_Worker] Job already exists, skipping insert",
				"jobLink", job.JobLink,
				"jobId", job.JobId)
		}

		slog.Info("[Workday_Scraper_Worker] Job Scraped", "JobLink", job.JobLink)
	}
	slog.Info("[Workday_Scraper_Worker] Job Details Worker shutting down")
}

func (ws WorkdayScraper) StartScraping(companiesToScrape <-chan db.Companies, scrapeDayLimit time.Time) {
	scrapeDateLimitTruncated := scrapeDayLimit.Truncate(24 * time.Hour)
	// now := time.Now().Truncate(24 * time.Hour)

	jobDetailScrapeChannel := make(chan *db.Jobs, 10000)
	slog.Info("[Workday_Scraper] Workday Jobs Details channel created")
	scraperWorkerCount := 4
	for range make([]struct{}, scraperWorkerCount) {
		go ws.jobDetailsScraperWorker(jobDetailScrapeChannel)
		slog.Info("[Workday_Scraper] Job details scraper started")
	}

	// Use WaitGroup to track company listing workers
	var wg sync.WaitGroup
	for company := range companiesToScrape {
		wg.Go(func() {
			listJobsAndStartDetailsScrape(company, scrapeDateLimitTruncated, jobDetailScrapeChannel)
		})
	}

	// Wait for all companies to finish listing jobs, then close the channel
	wg.Wait()
	close(jobDetailScrapeChannel)
	slog.Info("[Workday_Scraper] Workday Companies Job list complete.")

}
