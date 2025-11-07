package oraclecloud

import (
	"fmt"
	"job-scraper/internal/db"
	"job-scraper/internal/scraper/common"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/k3a/html2text"
	"resty.dev/v3"
)

type OracleCloudRequisitionItem struct {
	Id                     string `json:"Id"`
	Title                  string `json:"Title"`
	PostedDate             string `json:"PostedDate"`
	PrimaryLocation        string `json:"PrimaryLocation"`
	ShortDescriptionStr    string `json:"ShortDescriptionStr"`
	Relevancy              int    `json:"Relevancy"`
	Distance               int    `json:"Distance"`
	PrimaryLocationCountry string `json:"PrimaryLocationCountry"`
}

type OracleCloudJobListResponse struct {
	Items []struct {
		RequisitionList []OracleCloudRequisitionItem `json:"requisitionList"`
		TotalJobsCount  int                          `json:"TotalJobsCount"`
		Offset          int                          `json:"Offset"`
		Limit           int                          `json:"Limit"`
	} `json:"items"`
	Count   int  `json:"count"`
	HasMore bool `json:"hasMore"`
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
}

type OracleCloudJobDetailsResponse struct {
	Items []struct {
		Id                         string `json:"Id"`
		Title                      string `json:"Title"`
		Category                   string `json:"Category"`
		RequisitionId              int64  `json:"RequisitionId"`
		ExternalPostedStartDate    string `json:"ExternalPostedStartDate"`
		ExternalDescriptionStr     string `json:"ExternalDescriptionStr"`
		CorporateDescriptionStr    string `json:"CorporateDescriptionStr"`
		OrganizationDescriptionStr string `json:"OrganizationDescriptionStr"`
		ShortDescriptionStr        string `json:"ShortDescriptionStr"`
		PrimaryLocation            string `json:"PrimaryLocation"`
		PrimaryLocationCountry     string `json:"PrimaryLocationCountry"`
		JobFunction                string `json:"JobFunction"`
		BusinessUnit               string `json:"BusinessUnit"`
	} `json:"items"`
	Count   int  `json:"count"`
	HasMore bool `json:"hasMore"`
}

type OracleCloudScraper struct{}

// TransformBrowserURLToAPIURL converts Oracle Cloud browser URL to REST API URL
func TransformBrowserURLToAPIURL(browserURL string) (string, error) {
	parsedURL, err := url.Parse(browserURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Extract base domain and protocol
	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	// Extract siteNumber from path
	// Path format: /hcmUI/CandidateExperience/en/sites/CX_1001/jobs
	pathParts := strings.Split(parsedURL.Path, "/")
	var siteNumber string
	for i, part := range pathParts {
		if part == "sites" && i+1 < len(pathParts) {
			siteNumber = pathParts[i+1]
			break
		}
	}

	if siteNumber == "" {
		return "", fmt.Errorf("could not extract siteNumber from URL path")
	}

	// Parse query parameters - use RawQuery to preserve URL encoding
	queryParams := parsedURL.Query()

	// Get selectedCategoriesFacet (optional, can be multiple separated by semicolon)
	// Extract raw value to preserve URL encoding like %3B
	categoriesFacet := ""
	if parsedURL.RawQuery != "" {
		re := regexp.MustCompile(`selectedCategoriesFacet=([^&]*)`)
		matches := re.FindStringSubmatch(parsedURL.RawQuery)
		if len(matches) > 1 {
			categoriesFacet = matches[1]
		}
	}

	// Get selectedPostingDatesFacet, default to 7 if not present
	postingDatesFacet := queryParams.Get("selectedPostingDatesFacet")
	if postingDatesFacet == "" {
		postingDatesFacet = "7"
	}

	// Build API URL with finder parameters
	finderParts := []string{
		"findReqs",
		fmt.Sprintf("siteNumber=%s", siteNumber),
		"limit=25",
		"offset=0",
	}

	// Add optional category facet if provided
	if categoriesFacet != "" {
		finderParts = append(finderParts, fmt.Sprintf("selectedCategoriesFacet=%s", categoriesFacet))
	}

	// Add posting dates facet
	finderParts = append(finderParts, fmt.Sprintf("selectedPostingDatesFacet=%s", postingDatesFacet))

	// Add sort by
	finderParts = append(finderParts, "sortBy=POSTING_DATES_DESC")

	finder := strings.Join(finderParts, ";")

	apiURL := fmt.Sprintf(
		"%s/hcmRestApi/resources/latest/recruitingCEJobRequisitions?onlyData=true&expand=requisitionList&finder=%s",
		baseURL,
		finder,
	)

	return apiURL, nil
}

// ParseOracleAPIURL extracts base URL and finder parameters from Oracle Cloud API URL
func ParseOracleAPIURL(apiURL string) (baseURL, siteNumber string, err error) {
	parsedURL, err := url.Parse(apiURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse API URL: %w", err)
	}

	// Extract base domain
	baseURL = fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	// Extract finder parameter - check both decoded and raw query
	queryParams := parsedURL.Query()
	finder := queryParams.Get("finder")

	// If not found in parsed query, try extracting from raw query
	if finder == "" && parsedURL.RawQuery != "" {
		re := regexp.MustCompile(`finder=([^&]*)`)
		matches := re.FindStringSubmatch(parsedURL.RawQuery)
		if len(matches) > 1 {
			// URL decode the finder parameter
			finder, _ = url.QueryUnescape(matches[1])
		}
	}

	if finder == "" {
		return "", "", fmt.Errorf("finder parameter not found in API URL")
	}

	// Extract siteNumber from finder
	finderParts := strings.Split(finder, ";")
	for _, part := range finderParts {
		if strings.HasPrefix(part, "siteNumber=") {
			siteNumber = strings.TrimPrefix(part, "siteNumber=")
			break
		}
	}

	if siteNumber == "" {
		return "", "", fmt.Errorf("siteNumber not found in finder parameter")
	}

	return baseURL, siteNumber, nil
}

func parseOracleDate(dateStr string) time.Time {
	// Oracle Cloud uses ISO 8601 format: 2025-11-06 or 2025-11-06T21:06:24+00:00
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t
		}
	}

	// Default to current time if parsing fails
	slog.Warn("Failed to parse Oracle date", "dateStr", dateStr)
	return time.Now()
}

func buildJobDetailURL(baseURL, siteNumber, jobId string) string {
	// Note: %%22 is used to produce %22 in the output (URL-encoded double quote)
	// Oracle API requires: finder=ById;Id="210630221",siteNumber=CX_1001
	// Which in URL format is: finder=ById;Id=%22210630221%22,siteNumber=CX_1001
	// In fmt.Sprintf, %% escapes a literal %, so %%22 produces %22
	return fmt.Sprintf(
		"%s/hcmRestApi/resources/latest/recruitingCEJobRequisitionDetails?expand=all&onlyData=true&finder=ById;Id=%%22%s%%22,siteNumber=%s",
		baseURL,
		jobId,
		siteNumber,
	)
}

type jobDetailRequest struct {
	requisition *OracleCloudRequisitionItem
	company     db.Companies
	baseURL     string
	siteNumber  string
}

func (ocs OracleCloudScraper) fetchJobDetails(rClient *resty.Client, baseURL, siteNumber, jobId string) (*db.Jobs, error) {
	detailURL := buildJobDetailURL(baseURL, siteNumber, jobId)

	var detailsResp OracleCloudJobDetailsResponse

	resp, err := rClient.R().
		SetHeaders(map[string]string{
			"Accept":        "application/json",
			"Cache-Control": "no-cache",
		}).
		SetResult(&detailsResp).
		Get(detailURL)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch job details: %w", err)
	}

	result := resp.Result().(*OracleCloudJobDetailsResponse)

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("no job details found for ID: %s", jobId)
	}

	jobDetail := result.Items[0]

	// Combine description fields
	fullDescription := jobDetail.ExternalDescriptionStr
	if jobDetail.CorporateDescriptionStr != "" {
		fullDescription += "\n\n" + jobDetail.CorporateDescriptionStr
	}
	if jobDetail.OrganizationDescriptionStr != "" {
		fullDescription += "\n\n" + jobDetail.OrganizationDescriptionStr
	}

	// Clean and prepare job details
	jobDetails := common.RemoveExtraNewlines(common.CleanUTF8String(html2text.HTML2Text(fullDescription)))

	// Build external job URL
	externalJobURL := fmt.Sprintf(
		"%s/hcmUI/CandidateExperience/en/sites/%s/job/%s",
		baseURL,
		siteNumber,
		jobId,
	)

	// Parse posted date
	jobPostDate := parseOracleDate(jobDetail.ExternalPostedStartDate)

	job := &db.Jobs{
		JobHash:      common.GetSHA256Hash(externalJobURL),
		JobId:        jobId,
		JobRole:      jobDetail.Title,
		JobDetails:   jobDetails,
		JobPostDate:  jobPostDate.Format("2006-01-02"),
		JobLink:      externalJobURL,
		JobAISummary: "",
		CompanyName:  "", // Will be set by caller
	}

	return job, nil
}

func (ocs OracleCloudScraper) jobDetailsScraperWorker(jobChannel <-chan *jobDetailRequest) {
	slog.Debug("[OracleCloud_Scraper] Worker started to scrape Job Details")
	rClient := resty.New()
	rClient.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	defer rClient.Close()

	for req := range jobChannel {
		job, err := ocs.fetchJobDetails(rClient, req.baseURL, req.siteNumber, req.requisition.Id)
		if err != nil {
			slog.Error("[OracleCloud_Scraper_Worker] Failed to fetch job details",
				"jobId", req.requisition.Id,
				"error", err)
			continue
		}

		// Set company name
		job.CompanyName = req.company.Name

		// Insert job into database
		common.InsertJobToDB(job, "OracleCloud_Scraper")
	}
	slog.Info("[OracleCloud_Scraper_Worker] Job Details Worker shutting down")
}

func listJobsAndStartDetailsScrape(
	company db.Companies,
	scrapeDateLimitTruncated time.Time,
	jobDetailScrapeChannel chan<- *jobDetailRequest,
) {
	// Parse company base URL to extract base URL and site number
	baseURL, siteNumber, err := ParseOracleAPIURL(company.BaseUrl)
	if err != nil {
		slog.Error("[OracleCloud_Scraper] Failed to parse company URL",
			"company", company.Name,
			"error", err)
		return
	}

	rClient := resty.New()
	rClient.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	defer rClient.Close()

	// Parse the stored API URL to get finder parameters
	parsedURL, err := url.Parse(company.BaseUrl)
	if err != nil {
		slog.Error("Failed to parse Oracle API URL", "company", company.Name, "error", err)
		return
	}

	// Extract finder parameter
	queryParams := parsedURL.Query()
	finder := queryParams.Get("finder")
	if finder == "" {
		// Try from raw query
		if parsedURL.RawQuery != "" {
			re := regexp.MustCompile(`finder=([^&]*)`)
			matches := re.FindStringSubmatch(parsedURL.RawQuery)
			if len(matches) > 1 {
				finder, _ = url.QueryUnescape(matches[1])
			}
		}
	}

	if finder == "" {
		slog.Error("Could not extract finder from stored URL", "company", company.Name)
		return
	}

	// Parse finder to extract parameters
	finderParams := make(map[string]string)
	parts := strings.Split(finder, ";")
	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) == 2 {
			finderParams[kv[0]] = kv[1]
		} else if len(kv) == 1 {
			finderParams["action"] = kv[0]
		}
	}

	offset := 0
	limit := 25

	for {
		// Update offset in finder
		finderParams["offset"] = fmt.Sprintf("%d", offset)
		finderParams["limit"] = fmt.Sprintf("%d", limit)

		// Rebuild finder string
		var finderParts []string
		if action, ok := finderParams["action"]; ok {
			finderParts = append(finderParts, action)
		}
		// Add parameters in specific order
		if siteNum, ok := finderParams["siteNumber"]; ok {
			finderParts = append(finderParts, fmt.Sprintf("siteNumber=%s", siteNum))
		}
		finderParts = append(finderParts, fmt.Sprintf("limit=%d", limit))
		finderParts = append(finderParts, fmt.Sprintf("offset=%d", offset))
		if categories, ok := finderParams["selectedCategoriesFacet"]; ok {
			finderParts = append(finderParts, fmt.Sprintf("selectedCategoriesFacet=%s", categories))
		}
		if postingDates, ok := finderParams["selectedPostingDatesFacet"]; ok {
			finderParts = append(finderParts, fmt.Sprintf("selectedPostingDatesFacet=%s", postingDates))
		}
		if sortBy, ok := finderParams["sortBy"]; ok {
			finderParts = append(finderParts, fmt.Sprintf("sortBy=%s", sortBy))
		}

		newFinder := strings.Join(finderParts, ";")

		// Build full API URL
		apiURL := fmt.Sprintf(
			"%s/hcmRestApi/resources/latest/recruitingCEJobRequisitions?onlyData=true&expand=requisitionList&finder=%s",
			baseURL,
			url.QueryEscape(newFinder),
		)

		var oracleResp OracleCloudJobListResponse

		resp, err := rClient.R().
			SetHeaders(map[string]string{
				"Accept":        "application/json",
				"Cache-Control": "no-cache",
			}).
			SetResult(&oracleResp).
			Get(apiURL)

		if err != nil {
			slog.Error("Failed to fetch Oracle Cloud jobs", "company", company.Name, "error", err)
			break
		}

		result := resp.Result().(*OracleCloudJobListResponse)

		if len(result.Items) == 0 {
			slog.Info("No items in response, stopping pagination", "company", company.Name)
			break
		}

		// Get the requisition list from the first item
		requisitionList := result.Items[0].RequisitionList
		totalJobs := result.Items[0].TotalJobsCount

		slog.Info("Successfully fetched Oracle Cloud jobs",
			"company", company.Name,
			"offset", offset,
			"jobs_in_response", len(requisitionList),
			"total_jobs", totalJobs)

		if len(requisitionList) == 0 {
			slog.Info("No more jobs found, stopping pagination", "company", company.Name)
			break
		}

		allJobsTooOld := true
		for _, posting := range requisitionList {
			jobPostDate := parseOracleDate(posting.PostedDate)
			jobPostDateTruncated := jobPostDate.Truncate(24 * time.Hour)

			if jobPostDateTruncated.After(scrapeDateLimitTruncated) || jobPostDateTruncated.Equal(scrapeDateLimitTruncated) {
				allJobsTooOld = false
				// Send to worker for detailed scraping
				jobDetailScrapeChannel <- &jobDetailRequest{
					requisition: &posting,
					company:     company,
					baseURL:     baseURL,
					siteNumber:  siteNumber,
				}
			}
		}

		if allJobsTooOld {
			slog.Info("All jobs in this batch are too old, stopping pagination", "company", company.Name)
			break
		}

		// Check if we've reached the end
		if offset+len(requisitionList) >= totalJobs {
			slog.Info("Reached end of job listings", "company", company.Name)
			break
		}

		offset += len(requisitionList)
	}
}

func (ocs OracleCloudScraper) StartScraping(companiesToScrape <-chan db.Companies, scrapeDayLimit time.Time) {
	scrapeDateLimitTruncated := scrapeDayLimit.Truncate(24 * time.Hour)

	slog.Info("[OracleCloud_Scraper] Starting Oracle Cloud scraper")

	// Create channel for job details scraping
	jobDetailScrapeChannel := make(chan *jobDetailRequest, 10000)
	slog.Info("[OracleCloud_Scraper] Oracle Cloud Jobs Details channel created")

	// Start worker pool - workers will watch the channel until it's closed
	scraperWorkerCount := 4
	for range make([]struct{}, scraperWorkerCount) {
		go ocs.jobDetailsScraperWorker(jobDetailScrapeChannel)
		slog.Info("[OracleCloud_Scraper] Job details scraper started")
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
	slog.Info("[OracleCloud_Scraper] Oracle Cloud companies job scraping complete.")
}
