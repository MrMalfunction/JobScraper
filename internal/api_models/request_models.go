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
	ToScrape             *bool           `json:"to_scrape"`
}
