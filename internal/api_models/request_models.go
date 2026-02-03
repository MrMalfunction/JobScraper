package api_models

import "encoding/json"

// type ListChatsRequest struct {
// 	Page     int `query:"page" validate:"omitempty,min=1" default:"1"`
// 	PageSize int `query:"page_size" validate:"omitempty,min=1,max=100" default:"1"`
// }

type AddWorkdayCompanyScrapeList struct {
	Name       string          `json:"name"`
	BaseUrl    string          `json:"base_url"`
	ApiReqBody json.RawMessage `json:"req_body"`
}

type AddGreenhouseCompanyScrapeList struct {
	Name    string `json:"name"`
	BaseUrl string `json:"base_url"`
}

type AddOracleCloudCompanyScrapeList struct {
	Name       string `json:"name"`
	BrowserUrl string `json:"browser_url"`
}

type AddGenericCompanyScrapeList struct {
	Name               string                 `json:"name" validate:"required"`
	BaseUrl            string                 `json:"base_url" validate:"required,url"`
	Method             string                 `json:"method" validate:"required,oneof=GET POST"`
	Headers            map[string]string      `json:"headers"`
	Body               map[string]interface{} `json:"body"`
	QueryParams        string                 `json:"query_params"`
	PaginationKey      string                 `json:"pagination_key" validate:"required"`
	ResponseJsonPath   string                 `json:"response_json_path" validate:"required"`
	JobIdJsonPath      string                 `json:"job_id_json_path" validate:"required"`
	JobTitleJsonPath   string                 `json:"job_title_json_path" validate:"required"`
	JobDetailsJsonPath string                 `json:"job_details_json_path"`
	JobLinkJsonPath    string                 `json:"job_link_json_path" validate:"required"`
	JobLinkTemplate    string                 `json:"job_link_template"` // Optional template like "{base_url}{job_path}"
	JobDateJsonPath    string                 `json:"job_date_json_path"`
	DryRun             bool                   `json:"dry_run"` // If true, validate config without saving
}

type DryRunResponse struct {
	Valid          bool                     `json:"valid"`
	Message        string                   `json:"message"`
	SampleData     []map[string]interface{} `json:"sample_data,omitempty"`
	ExtractedJobs  int                      `json:"extracted_jobs,omitempty"`
	ErrorDetails   string                   `json:"error_details,omitempty"`
}

type JobSearchRequest struct {
	Company         string   `query:"company" json:"company"`
	Title           string   `query:"title" json:"title"`
	IncludeKeywords []string `query:"include_keywords" json:"include_keywords"`
	ExcludeKeywords []string `query:"exclude_keywords" json:"exclude_keywords"`
	Limit           int      `query:"limit" json:"limit" validate:"omitempty,min=1,max=100" default:"10"`
	Offset          int      `query:"offset" json:"offset" validate:"omitempty,min=0" default:"0"`
}

type UpdateCompanyRequest struct {
	Name                 string          `json:"name"`
	BaseUrl              string          `json:"base_url"`
	CareerSiteType       string          `json:"career_site_type"`
	ApiRequestBody       json.RawMessage `json:"api_request_body"`
	ApiRequestQueryParam string          `json:"api_request_query_param"`
	ApiRequestHeaders    string          `json:"api_request_headers"`
	ApiRequestMethod     string          `json:"api_request_method"`
	PaginationKey        string          `json:"pagination_key"`
	ResponseJsonPath     string          `json:"response_json_path"`
	JobIdJsonPath        string          `json:"job_id_json_path"`
	JobTitleJsonPath     string          `json:"job_title_json_path"`
	JobDetailsJsonPath   string          `json:"job_details_json_path"`
	JobLinkJsonPath      string          `json:"job_link_json_path"`
	JobLinkTemplate      string          `json:"job_link_template"`
	JobDateJsonPath      string          `json:"job_date_json_path"`
	ToScrape             *bool           `json:"to_scrape"`
}
