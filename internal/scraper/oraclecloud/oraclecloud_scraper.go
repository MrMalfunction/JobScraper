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

	// Get optional locationId
	locationId := queryParams.Get("locationId")

	// Get optional lastSelectedFacet
	lastSelectedFacet := queryParams.Get("lastSelectedFacet")

	// Build API URL with finder parameters
	// Format: findReqs;siteNumber=CX_1001,param1=val1,param2=val2
	params := []string{
		fmt.Sprintf("siteNumber=%s", siteNumber),
	}

	// Add facetsList - standard list of facets
	params = append(params, "facetsList=LOCATIONS%3BWORK_LOCATIONS%3BWORKPLACE_TYPES%3BTITLES%3BCATEGORIES%3BORGANIZATIONS%3BPOSTING_DATES%3BFLEX_FIELDS")

	// Add limit and offset
	params = append(params, "limit=25", "offset=0")

	// Add optional lastSelectedFacet
	if lastSelectedFacet != "" {
		params = append(params, fmt.Sprintf("lastSelectedFacet=%s", lastSelectedFacet))
	}

	// Add optional locationId
	if locationId != "" {
		params = append(params, fmt.Sprintf("locationId=%s", locationId))
	}

	// Add optional category facet if provided
	if categoriesFacet != "" {
		params = append(params, fmt.Sprintf("selectedCategoriesFacet=%s", categoriesFacet))
	}

	// Add posting dates facet
	params = append(params, fmt.Sprintf("selectedPostingDatesFacet=%s", postingDatesFacet))

	// Always add sort by POSTING_DATES_DESC
	params = append(params, "sortBy=POSTING_DATES_DESC")

	// Build finder: findReqs;param1,param2,param3
	finder := "findReqs;" + strings.Join(params, ",")

	apiURL := fmt.Sprintf(
		"%s/hcmRestApi/resources/latest/recruitingCEJobRequisitions?onlyData=true&expand=requisitionList.workLocation,requisitionList.otherWorkLocations,requisitionList.secondaryLocations,flexFieldsFacet.values,requisitionList.requisitionFlexFields&finder=%s",
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
	// Format: findReqs;siteNumber=CX_1001,param1=val1,param2=val2
	// First split by semicolon to separate action from parameters
	semicolonParts := strings.SplitN(finder, ";", 2)

	// Then split parameters by comma
	if len(semicolonParts) > 1 {
		paramParts := strings.Split(semicolonParts[1], ",")
		for _, part := range paramParts {
			if strings.HasPrefix(part, "siteNumber=") {
				siteNumber = strings.TrimPrefix(part, "siteNumber=")
				break
			}
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
		var t time.Time
		var err error

		// For date-only format, parse in local timezone
		if layout == "2006-01-02" {
			t, err = time.ParseInLocation(layout, dateStr, time.Local)
		} else {
			t, err = time.Parse(layout, dateStr)
		}

		if err == nil {
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
	slog.Debug("[OracleCloud_Scraper_Worker] Worker started")
	rClient := resty.New()
	rClient.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	defer rClient.Close()

	for req := range jobChannel {
		job, err := ocs.fetchJobDetails(rClient, req.baseURL, req.siteNumber, req.requisition.Id)
		if err != nil {
			slog.Error("[OracleCloud_Scraper_Worker] Failed to fetch job details",
				"jobId", req.requisition.Id,
				"company", req.company.Name,
				"error", err)
			continue
		}

		// Set company name
		job.CompanyName = req.company.Name

		// Insert job into database
		common.InsertJobToDB(job, "OracleCloud_Scraper")
	}
	slog.Info("[OracleCloud_Scraper_Worker] Worker shutting down")
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
	// Format: findReqs;siteNumber=CX_1001,param1=val1,param2=val2
	finderParams := make(map[string]string)

	// First split by semicolon to separate action from parameters
	semicolonParts := strings.SplitN(finder, ";", 2)
	if len(semicolonParts) > 0 {
		finderParams["action"] = semicolonParts[0]
	}

	// Then split parameters by comma
	if len(semicolonParts) > 1 {
		paramParts := strings.Split(semicolonParts[1], ",")
		for _, part := range paramParts {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				finderParams[kv[0]] = kv[1]
			}
		}
	}

	offset := 0
	limit := 25

	for {
		// Update offset in finder
		finderParams["offset"] = fmt.Sprintf("%d", offset)
		finderParams["limit"] = fmt.Sprintf("%d", limit)

		// Rebuild finder string
		// Format: findReqs;siteNumber=CX_1001,param1=val1,param2=val2
		action := "findReqs"
		if act, ok := finderParams["action"]; ok {
			action = act
		}

		// Add parameters in specific order to match Oracle API expectations
		var params []string
		if siteNum, ok := finderParams["siteNumber"]; ok {
			params = append(params, fmt.Sprintf("siteNumber=%s", siteNum))
		}
		// Add facetsList if present
		if facetsList, ok := finderParams["facetsList"]; ok {
			params = append(params, fmt.Sprintf("facetsList=%s", facetsList))
		}
		params = append(params, fmt.Sprintf("limit=%d", limit))
		params = append(params, fmt.Sprintf("offset=%d", offset))
		// Add lastSelectedFacet if present
		if lastFacet, ok := finderParams["lastSelectedFacet"]; ok {
			params = append(params, fmt.Sprintf("lastSelectedFacet=%s", lastFacet))
		}
		// Add locationId if present
		if locationId, ok := finderParams["locationId"]; ok {
			params = append(params, fmt.Sprintf("locationId=%s", locationId))
		}
		if categories, ok := finderParams["selectedCategoriesFacet"]; ok {
			params = append(params, fmt.Sprintf("selectedCategoriesFacet=%s", categories))
		}
		if postingDates, ok := finderParams["selectedPostingDatesFacet"]; ok {
			params = append(params, fmt.Sprintf("selectedPostingDatesFacet=%s", postingDates))
		}
		// Always use POSTING_DATES_DESC for sorting
		params = append(params, "sortBy=POSTING_DATES_DESC")

		// Build finder with semicolon after action, comma for params
		newFinder := action + ";" + strings.Join(params, ",")

		// Build full API URL
		apiURL := fmt.Sprintf(
			"%s/hcmRestApi/resources/latest/recruitingCEJobRequisitions?onlyData=true&expand=requisitionList.workLocation,requisitionList.otherWorkLocations,requisitionList.secondaryLocations,flexFieldsFacet.values,requisitionList.requisitionFlexFields&finder=%s",
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
			slog.Info("[OracleCloud_Scraper] No items in response, stopping pagination", "company", company.Name)
			break
		}

		// Get the requisition list from the first item
		requisitionList := result.Items[0].RequisitionList
		totalJobs := result.Items[0].TotalJobsCount

		slog.Debug("[OracleCloud_Scraper] Fetched jobs page",
			"company", company.Name,
			"offset", offset,
			"jobs_in_page", len(requisitionList),
			"total_jobs", totalJobs)

		if len(requisitionList) == 0 {
			slog.Info("[OracleCloud_Scraper] No more jobs found, stopping pagination", "company", company.Name)
			break
		}

		allJobsNotToday := true
		jobsScrapedInPage := 0
		for _, posting := range requisitionList {
			jobPostDate := parseOracleDate(posting.PostedDate)

			// Check if job is from today using centralized function
			if common.IsJobFromToday(jobPostDate) {
				allJobsNotToday = false
			}

			// Check if we should scrape this job using centralized function
			if common.ShouldScrapeJob(jobPostDate, scrapeDateLimitTruncated) {
				jobsScrapedInPage++
				// Send to worker for detailed scraping
				jobDetailScrapeChannel <- &jobDetailRequest{
					requisition: &posting,
					company:     company,
					baseURL:     baseURL,
					siteNumber:  siteNumber,
				}
			}
		}

		if jobsScrapedInPage > 0 {
			slog.Info("[OracleCloud_Scraper] Jobs queued for scraping",
				"company", company.Name,
				"count", jobsScrapedInPage,
				"offset", offset)
		}

		// Stop if we've seen a full page without any jobs from today
		if allJobsNotToday {
			slog.Info("[OracleCloud_Scraper] Full page without today's jobs, stopping pagination", "company", company.Name)
			break
		}

		// Check if we've reached the end
		if offset+len(requisitionList) >= totalJobs {
			slog.Info("[OracleCloud_Scraper] Reached end of job listings", "company", company.Name)
			break
		}

		offset += len(requisitionList)
	}
}

func (ocs OracleCloudScraper) StartScraping(companiesToScrape <-chan db.Companies, scrapeDayLimit time.Time) {
	// Get date at midnight using centralized function
	scrapeDateLimitTruncated := common.GetDateMidnight(scrapeDayLimit)

	slog.Info("[OracleCloud_Scraper] Starting Oracle Cloud scraper")

	// Create channel for job details scraping
	jobDetailScrapeChannel := make(chan *jobDetailRequest, 10000)
	slog.Info("[OracleCloud_Scraper] Oracle Cloud Jobs Details channel created")

	// Start worker pool - workers will watch the channel until it's closed
	scraperWorkerCount := 4
	for range make([]struct{}, scraperWorkerCount) {
		go ocs.jobDetailsScraperWorker(jobDetailScrapeChannel)
	}
	slog.Info("[OracleCloud_Scraper] Started workers", "count", scraperWorkerCount)

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
