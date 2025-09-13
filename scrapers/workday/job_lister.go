package workday

import (
	"job-scraper/manifests"
	"job-scraper/types"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"resty.dev/v3"
)

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

type JobLister struct {
	Companies manifests.WorkdayYAML
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

	// Default to current time if parsing fails
	return time.Now()
}

func (lister JobLister) ListJobs(jobsChannel chan<- *types.JobDetails, scrapeDateLimit time.Time) {
	defer close(jobsChannel)

	scrapeDateLimitTruncated := scrapeDateLimit.Truncate(24 * time.Hour)
	now := time.Now().Truncate(24 * time.Hour)

	slog.Info("Starting job listing for companies", "total_companies", len(lister.Companies.Companies), "scraping_from", scrapeDateLimit.Format("2006-01-02"), "to", now.Format("2006-01-02"))

	// Create a single client for all companies
	client := resty.New()
	client.SetHeader("User-Agent", "")
	defer client.Close()

	for companyKey, company := range lister.Companies.Companies {
		slog.Info("Listing all jobs for company", "company", company.Name, "key", companyKey)

		offset := 0

		for {
			// Temporary: send a fixed JSON body for testing
			fixedBody := map[string]any{
				"appliedFacets": map[string]any{
					"locationHierarchy1": []string{"2fcb99c455831013ea52fb338f2932d8"},
				},
				"limit":      20,
				"offset":     0,
				"searchText": "",
			}

			var workdayResp WorkdayResponse

			resp, err := client.R().
				SetHeaders(map[string]string{
					"origin":         "https://nvidia.wd5.myworkdayjobs.com",
					"cache-control":  "no-cache",
					"sec-fetch-dest": "empty",
					"sec-fetch-mode": "cors",
					"sec-fetch-site": "same-origin",
				}).
				SetBody(fixedBody).
				SetResult(&workdayResp).
				Post(company.BaseUrl + "/jobs")

			if err != nil {
				slog.Error("Failed to fetch jobs", "company", company.Name, "error", err)
				break
			}

			// Get the parsed result
			result := resp.Result().(*WorkdayResponse)
			slog.Info("Successfully fetched jobs", "company", company.Name, "offset", offset, "jobs_in_response", len(result.JobPostings))

			if len(result.JobPostings) == 0 {
				slog.Info("No more jobs found, stopping pagination", "company", company.Name)
				break
			}

			jobsAddedFromThisBatch := 0
			allJobsTooOld := true

			// Convert to JobDetails and filter for jobs within date limit
			for _, posting := range result.JobPostings {
				jobPostDate := parsePostedDate(posting.PostedOn)
				jobPostDateTruncated := jobPostDate.Truncate(24 * time.Hour)

				// Check if this job is posted within our date range (not older than the limit and not in future)
				if jobPostDateTruncated.After(scrapeDateLimitTruncated) || jobPostDateTruncated.Equal(scrapeDateLimitTruncated) {
					// Job is within our date range
					allJobsTooOld = false
					job := &types.JobDetails{
						JUlid:          ulid.Make(),
						JobID:          "", // Will be filled by job scraper
						JobRole:        posting.Title,
						JobDetails:     "", // Will be filled by job scraper
						JobPostDate:    jobPostDate,
						JobLink:        company.BaseUrl + posting.ExternalPath,
						JobCompanyName: company.Name,
					}

					// Send job to channel
					jobsChannel <- job
					jobsAddedFromThisBatch++
					slog.Info("Added job to channel", "company", company.Name, "title", posting.Title, "location", posting.LocationsText, "posted_date", jobPostDate.Format("2006-01-02"))
				} else {
					// Job is older than our limit, but we continue processing this batch
					// since jobs might not be in perfect chronological order
					slog.Debug("Skipping job outside date range", "company", company.Name, "title", posting.Title, "posted_date", jobPostDate.Format("2006-01-02"), "oldest_allowed", scrapeDateLimit.Format("2006-01-02"))
				}
			}

			slog.Info("Processed batch", "company", company.Name, "jobs_added", jobsAddedFromThisBatch, "total_in_batch", len(result.JobPostings))

			// If all jobs in this batch are older than our scrape date limit, stop pagination
			// This assumes that job postings are generally ordered by date (newest first)
			if allJobsTooOld {
				slog.Info("All jobs in batch are older than date range, stopping pagination", "company", company.Name, "oldest_date", scrapeDateLimit.Format("2006-01-02"))
				break
			}

			// Move to next page
			offset += len(result.JobPostings)
		}

		slog.Info("Completed job listing for company", "company", company.Name)
	}

	slog.Info("Completed job listing for all companies")
}
