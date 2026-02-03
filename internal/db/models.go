package db

import "time"

// Companies to be scraped.
type Companies struct {
	Name                 string `gorm:"type:string;primaryKey"`
	BaseUrl              string `gorm:"type:string;not null"`
	CareerSiteType       string `gorm:"type:string;not null"` // Ex: Workday, Generic
	ApiRequestBody       string `gorm:"type:string"`
	ApiRequestQueryParam string `gorm:"type:string"`
	ApiRequestHeaders    string `gorm:"type:text"`        // JSON string for custom headers (Generic)
	ApiRequestMethod     string `gorm:"type:string"`      // GET or POST (Generic)
	PaginationKey        string `gorm:"type:string"`      // Key used for pagination (Generic)
	ResponseJsonPath     string `gorm:"type:string"`      // JSON path to extract job data (Generic)
	JobIdJsonPath        string `gorm:"type:string"`      // JSON path to extract job ID (Generic)
	JobTitleJsonPath     string `gorm:"type:string"`      // JSON path to extract job title (Generic)
	JobDetailsJsonPath   string `gorm:"type:string"`      // JSON path to extract job details - supports arrays with [*] or [0] (Generic)
	JobLinkJsonPath      string `gorm:"type:string"`      // JSON path or template like "{baseUrl}{path}" (Generic)
	JobLinkTemplate      string `gorm:"type:string"`      // Optional template for constructing links (Generic)
	JobDateJsonPath      string `gorm:"type:string"`      // JSON path to extract job date (Generic)
	ToScrape             bool   `gorm:"type:boolean"`
}

type Jobs struct {
	JobHash       string    `gorm:"type:string;primaryKey"`
	JobId         string    `gorm:"type:string;not null"`
	JobRole       string    `gorm:"type:string;not null"`
	JobDetails    string    `gorm:"type:text;not null"`
	JobPostDate   string    `gorm:"type:string;not null;index:idx_job_post"`
	JobInsertTime time.Time `gorm:"type:timestamptz;index:idx_insert_time;default:CURRENT_TIMESTAMP"`
	JobLink       string    `gorm:"type:string;not null"`
	JobAISummary  string    `gorm:"type:text"`
	CompanyName   string    `gorm:"type:string;not null;index:idx_company_name"` // Foreign key to Companies.Name
	Company       Companies `gorm:"foreignKey:CompanyName;references:Name;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
