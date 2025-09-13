package test_shit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	// API Configuration
	nvidiaAPIURL     = "https://nvidia.wd5.myworkdayjobs.com/wday/cxs/nvidia/NVIDIAExternalCareerSite"
	workdayBaseURL   = "https://nvidia.wd5.myworkdayjobs.com"
	requestTimeout   = 30 * time.Second
	defaultJobsLimit = 20
)

// StandardJob represents the standardized job output format
type StandardJob struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Location    string `json:"location"`
	Company     string `json:"company"`
	Department  string `json:"department"`
	PostedDate  string `json:"posted_date"`
	StartDate   string `json:"start_date"`
	TimeType    string `json:"time_type"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Salary      string `json:"salary"`
}

// WorkdayJobDetail represents the structure of Workday API job detail responses
type WorkdayJobDetail struct {
	JobPostingInfo struct {
		ID             string `json:"id"`
		Title          string `json:"title"`
		JobDescription string `json:"jobDescription"`
		Location       string `json:"location"`
		PostedOn       string `json:"postedOn"`
		StartDate      string `json:"startDate"`
		TimeType       string `json:"timeType"`
		JobReqID       string `json:"jobReqId"`
		ExternalURL    string `json:"externalUrl"`
	} `json:"jobPostingInfo"`
	HiringOrganization struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"hiringOrganization"`
}

// NVIDIAJobsResponse represents NVIDIA's API response structure for job listings
type NVIDIAJobsResponse struct {
	Total int `json:"total"`
	Jobs  []struct {
		Title        string `json:"title"`
		Location     string `json:"location"`
		PostedOn     string `json:"postedOn"`
		JobReqID     string `json:"jobReqId"`
		ExternalPath string `json:"externalPath"`
	} `json:"jobPostings"`
}

// NVIDIAJobRequest represents the request payload for NVIDIA jobs API
type NVIDIAJobRequest struct {
	AppliedFacets map[string][]string `json:"appliedFacets"`
	Limit         int                 `json:"limit"`
	Offset        int                 `json:"offset"`
	SearchText    string              `json:"searchText"`
}

// JobScraper handles job scraping operations
type JobScraper struct {
	client *http.Client
	logger *log.Logger
}

// NewJobScraper creates a new JobScraper instance
func NewJobScraper() *JobScraper {
	return &JobScraper{
		client: &http.Client{
			Timeout: requestTimeout,
		},
		logger: log.Default(),
	}
}

func main() {
	scraper := NewJobScraper()

	ctx := context.Background()

	// Fetch NVIDIA jobs
	jobs, err := scraper.FetchNVIDIAJobs(ctx)
	if err != nil {
		log.Printf("Error fetching NVIDIA jobs: %v", err)
		return
	}

	fmt.Printf("Found %d NVIDIA jobs\n", len(jobs))

	// Display first 2 jobs as examples
	for i, job := range jobs {
		if i >= 2 {
			break
		}
		scraper.PrintJob(job)
	}
}

// FetchNVIDIAJobs retrieves job listings from NVIDIA's Workday API
func (js *JobScraper) FetchNVIDIAJobs(ctx context.Context) ([]StandardJob, error) {
	js.logger.Println("Starting NVIDIA job fetch...")

	// Prepare request payload
	requestBody := NVIDIAJobRequest{
		AppliedFacets: map[string][]string{
			"locationHierarchy1": {"2fcb99c455831013ea52fb338f2932d8"},
		},
		Limit:      defaultJobsLimit,
		Offset:     0,
		SearchText: "",
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	// Create and execute request
	resp, err := js.executeAPIRequest(ctx, "POST", nvidiaAPIURL+"/jobs", jsonBody)
	if err != nil {
		return nil, fmt.Errorf("executing API request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var apiResp NVIDIAJobsResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	js.logger.Printf("Successfully parsed %d jobs from API", len(apiResp.Jobs))

	// Convert to standardized format and enhance with details
	jobs := make([]StandardJob, 0, len(apiResp.Jobs))
	for _, apiJob := range apiResp.Jobs {
		standardJob := js.convertNVIDIAJob(apiJob)

		// Enhance with detailed information
		enhancedJob, err := js.enhanceJobWithDetails(ctx, standardJob, apiJob.ExternalPath)
		if err != nil {
			js.logger.Printf("Warning: Could not enhance job %s: %v", standardJob.ID, err)
			jobs = append(jobs, standardJob)
			continue
		}

		jobs = append(jobs, enhancedJob)
	}

	return jobs, nil
}

// FetchWorkdayJobByURL retrieves a single job from a Workday URL
func (js *JobScraper) FetchWorkdayJobByURL(ctx context.Context, url string) (StandardJob, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return StandardJob{}, fmt.Errorf("creating request: %w", err)
	}

	js.setHTMLHeaders(req)

	resp, err := js.client.Do(req)
	if err != nil {
		return StandardJob{}, fmt.Errorf("fetching job page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return StandardJob{}, fmt.Errorf("received status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return StandardJob{}, fmt.Errorf("parsing HTML: %w", err)
	}

	// Try to extract JSON-LD data first
	job := js.extractJobFromJSONLD(doc)
	if job.Title == "" {
		// Fallback to HTML extraction
		job = js.extractJobFromHTML(doc, url)
	}

	return job, nil
}

// executeAPIRequest performs an HTTP request with proper headers and error handling
func (js *JobScraper) executeAPIRequest(ctx context.Context, method, url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	js.setAPIHeaders(req)

	resp, err := js.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return resp, nil
}

// convertNVIDIAJob converts NVIDIA API response to StandardJob format
func (js *JobScraper) convertNVIDIAJob(apiJob struct {
	Title        string `json:"title"`
	Location     string `json:"location"`
	PostedOn     string `json:"postedOn"`
	JobReqID     string `json:"jobReqId"`
	ExternalPath string `json:"externalPath"`
}) StandardJob {
	jobURL := js.constructJobURL(apiJob.ExternalPath)

	return StandardJob{
		ID:         apiJob.JobReqID,
		Title:      apiJob.Title,
		Location:   apiJob.Location,
		Company:    "NVIDIA",
		PostedDate: apiJob.PostedOn,
		URL:        jobURL,
	}
}

// constructJobURL builds a complete URL from an external path
func (js *JobScraper) constructJobURL(externalPath string) string {
	if strings.HasPrefix(externalPath, "http") {
		return externalPath
	}

	if !strings.HasPrefix(externalPath, "/") {
		externalPath = "/" + externalPath
	}

	return workdayBaseURL + externalPath
}

// enhanceJobWithDetails fetches additional job details from the API
func (js *JobScraper) enhanceJobWithDetails(ctx context.Context, job StandardJob, externalPath string) (StandardJob, error) {
	apiURL := nvidiaAPIURL + externalPath

	resp, err := js.executeAPIRequest(ctx, "GET", apiURL, nil)
	if err != nil {
		return job, fmt.Errorf("fetching job details: %w", err)
	}
	defer resp.Body.Close()

	var jobDetail WorkdayJobDetail
	if err := json.NewDecoder(resp.Body).Decode(&jobDetail); err != nil {
		return job, fmt.Errorf("decoding job details: %w", err)
	}

	// Merge details into the existing job
	enhanced := js.mergeJobDetails(job, jobDetail)
	return enhanced, nil
}

// mergeJobDetails combines basic job info with detailed information
func (js *JobScraper) mergeJobDetails(base StandardJob, detail WorkdayJobDetail) StandardJob {
	return StandardJob{
		ID:          detail.JobPostingInfo.ID,
		Title:       detail.JobPostingInfo.Title,
		Location:    detail.JobPostingInfo.Location,
		Company:     detail.HiringOrganization.Name,
		PostedDate:  detail.JobPostingInfo.PostedOn,
		StartDate:   detail.JobPostingInfo.StartDate,
		TimeType:    detail.JobPostingInfo.TimeType,
		URL:         detail.JobPostingInfo.ExternalURL,
		Description: js.cleanHTMLText(detail.JobPostingInfo.JobDescription),
		Salary:      js.extractSalaryFromText(detail.JobPostingInfo.JobDescription),
	}
}

// extractJobFromJSONLD attempts to extract job data from JSON-LD script tags
func (js *JobScraper) extractJobFromJSONLD(doc *goquery.Document) StandardJob {
	var job StandardJob

	doc.Find("script[type='application/json']").Each(func(i int, s *goquery.Selection) {
		jsonText := s.Text()
		var workdayResp WorkdayJobDetail
		if err := json.Unmarshal([]byte(jsonText), &workdayResp); err == nil {
			job = js.mergeJobDetails(StandardJob{}, workdayResp)
		}
	})

	return job
}

// extractJobFromHTML extracts job information from HTML content
func (js *JobScraper) extractJobFromHTML(doc *goquery.Document, url string) StandardJob {
	job := StandardJob{URL: url}

	// Extract basic information using common selectors
	job.Title = js.extractTextBySelector(doc, "h1, [data-automation-id='jobPostingHeader']")
	job.Location = js.extractTextBySelector(doc, "[data-automation-id='locations']")
	job.Company = js.extractTextBySelector(doc, "[data-automation-id='company']")

	// Extract description
	var description strings.Builder
	doc.Find("[data-automation-id='jobPostingDescription'], .jobDescription, #jobDescriptionText").Each(func(i int, s *goquery.Selection) {
		description.WriteString(s.Text())
	})

	job.Description = js.cleanHTMLText(description.String())
	job.Salary = js.extractSalaryFromText(job.Description)

	return job
}

// extractTextBySelector safely extracts text from the first matching element
func (js *JobScraper) extractTextBySelector(doc *goquery.Document, selector string) string {
	return strings.TrimSpace(doc.Find(selector).First().Text())
}

// extractSalaryFromText finds salary information in text using regex patterns
func (js *JobScraper) extractSalaryFromText(text string) string {
	patterns := []string{
		`\$[\d,]+(?:\.\d{2})?(?:\s*-\s*\$[\d,]+(?:\.\d{2})?)?(?:\s*/\s*(?:hr|hour|year|annually))?`,
		`[\d,]+(?:\.\d{2})?\s*-\s*[\d,]+(?:\.\d{2})?\s*(?:per\s+)?(?:hour|hr|year|annually)`,
		`\$[\d,]+(?:\.\d{2})?(?:/hr|/hour|/year|/annually)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if match := re.FindString(text); match != "" {
			return strings.TrimSpace(match)
		}
	}

	return ""
}

// cleanHTMLText removes HTML tags and normalizes text content
func (js *JobScraper) cleanHTMLText(html string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(html, "")

	// Normalize whitespace
	text = regexp.MustCompile(`\n\s*\n\s*\n`).ReplaceAllString(text, "\n\n")
	text = regexp.MustCompile(`[ \t]+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	// Decode HTML entities
	text = js.decodeHTMLEntities(text)

	// Remove/replace unicode emojis
	text = js.replaceEmojis(text)

	return text
}

// decodeHTMLEntities converts HTML entities to their text equivalents
func (js *JobScraper) decodeHTMLEntities(text string) string {
	// Named entities
	entities := map[string]string{
		"&amp;":    "&",
		"&lt;":     "<",
		"&gt;":     ">",
		"&quot;":   "\"",
		"&apos;":   "'",
		"&nbsp;":   " ",
		"&copy;":   "©",
		"&reg;":    "®",
		"&trade;":  "™",
		"&ndash;":  "-",
		"&mdash;":  "-",
		"&lsquo;":  "'",
		"&rsquo;":  "'",
		"&ldquo;":  "\"",
		"&rdquo;":  "\"",
		"&hellip;": "...",
		"&bull;":   "*",
	}

	for entity, replacement := range entities {
		text = strings.ReplaceAll(text, entity, replacement)
	}

	// Numeric entities (decimal)
	re := regexp.MustCompile(`&#(\d+);`)
	text = re.ReplaceAllStringFunc(text, func(match string) string {
		numStr := match[2 : len(match)-1]
		if num, err := strconv.Atoi(numStr); err == nil {
			return string(rune(num))
		}
		return match
	})

	// Numeric entities (hexadecimal)
	re = regexp.MustCompile(`&#[xX]([0-9a-fA-F]+);`)
	text = re.ReplaceAllStringFunc(text, func(match string) string {
		hexStr := match[3 : len(match)-1]
		if num, err := strconv.ParseInt(hexStr, 16, 32); err == nil {
			return string(rune(num))
		}
		return match
	})

	return text
}

// replaceEmojis removes or replaces common emojis with text equivalents
func (js *JobScraper) replaceEmojis(text string) string {
	emojiReplacements := map[string]string{
		"📍": "Location:",
		"🌎": "",
		"✨": "",
		"💬": "",
		"🎓": "",
		"💼": "",
		"🎉": "",
	}

	for emoji, replacement := range emojiReplacements {
		text = strings.ReplaceAll(text, emoji, replacement)
	}

	// Remove remaining emojis
	re := regexp.MustCompile(`[\x{1F600}-\x{1F64F}]|[\x{1F300}-\x{1F5FF}]|[\x{1F680}-\x{1F6FF}]|[\x{1F1E0}-\x{1F1FF}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]`)
	text = re.ReplaceAllString(text, "")

	return text
}

// setAPIHeaders sets appropriate headers for API requests
func (js *JobScraper) setAPIHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; JobScraper/1.0)")
	req.Header.Set("Referer", workdayBaseURL)
}

// setHTMLHeaders sets appropriate headers for HTML requests
func (js *JobScraper) setHTMLHeaders(req *http.Request) {
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; JobScraper/1.0)")
}

// PrintJob outputs job information in a formatted way
func (js *JobScraper) PrintJob(job StandardJob) {
	const separator = "="
	const lineLength = 80

	fmt.Println(strings.Repeat(separator, lineLength))
	fmt.Printf("ID: %s\n", job.ID)
	fmt.Printf("Title: %s\n", job.Title)
	fmt.Printf("Company: %s\n", job.Company)
	fmt.Printf("Location: %s\n", job.Location)
	fmt.Printf("Department: %s\n", job.Department)
	fmt.Printf("Posted: %s\n", job.PostedDate)
	fmt.Printf("Start Date: %s\n", job.StartDate)
	fmt.Printf("Time Type: %s\n", job.TimeType)
	fmt.Printf("Salary: %s\n", job.Salary)
	fmt.Printf("URL: %s\n", job.URL)

	if job.Description != "" {
		fmt.Println("\nFULL JOB DESCRIPTION:")
		fmt.Println(strings.Repeat("-", lineLength))
		fmt.Println(job.Description)
		fmt.Println(strings.Repeat("-", lineLength))
	}

	fmt.Println()
}
