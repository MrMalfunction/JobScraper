package api_models

type StdResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type JobResponse struct {
	JobHash       string `json:"job_hash"`
	JobId         string `json:"job_id"`
	JobRole       string `json:"job_role"`
	JobDetails    string `json:"job_details"`
	JobPostDate   string `json:"job_post_date"`
	JobInsertTime string `json:"job_insert_time"`
	JobLink       string `json:"job_link"`
	JobAISummary  string `json:"job_ai_summary"`
	CompanyName   string `json:"company_name"`
}

type JobSearchResponse struct {
	Jobs    []JobResponse `json:"jobs"`
	Total   int64         `json:"total"`
	Page    int           `json:"page"`
	Limit   int           `json:"limit"`
	HasMore bool          `json:"has_more"`
}

type CompanyResponse struct {
	Name                 string `json:"name"`
	BaseUrl              string `json:"base_url"`
	CareerSiteType       string `json:"career_site_type"`
	ApiRequestBody       string `json:"api_request_body"`
	ApiRequestQueryParam string `json:"api_request_query_param"`
	ToScrape             bool   `json:"to_scrape"`
}
