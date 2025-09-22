package db

// Companies to be scraped.
type Companies struct {
	Name                 string `gorm:"type:string;primaryKey"`
	BaseUrl              string `gorm:"type:string;not null"`
	CareerSiteType       string `gorm:"type:string;not null"` // Ex: Workday
	ApiRequestBody       string `gorm:"type:string"`
	ApiRequestQueryParam string `gorm:"type:string"`
	ToScrape             bool   `gorm:"type:boolean"`
}

type Jobs struct {
	JobHash      string    `gorm:"type:string;primaryKey"`
	JobId        string    `gorm:"type:string;not null"`
	JobRole      string    `gorm:"type:string;not null"`
	JobDetails   string    `gorm:"type:text;not null"`
	JobPostDate  string    `gorm:"type:string;not null;index:idx_job_post"`
	JobLink      string    `gorm:"type:string;not null"`
	JobAISummary string    `gorm:"type:text"`
	CompanyName  string    `gorm:"type:string;not null;index:idx_company_name"` // Foreign key to Companies.Name
	Company      Companies `gorm:"foreignKey:CompanyName;references:Name;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
