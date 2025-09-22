package internal

import (
	"job-scraper/internal/services/service_jobs"
	"job-scraper/internal/services/service_scraper"

	"github.com/labstack/echo/v4"
)

// Starts Job scrapping for companies which allow scraping
// Method: [POST]
func SubmitScrapeRequest(c echo.Context) error {
	return service_scraper.StartJobScrapping(c)
}

func SubmitWorkdayCompanyToScrape(c echo.Context) error {
	return service_scraper.AddWorkdayCompanyToScrapeList(c)
}

// Job search endpoints
func SearchJobs(c echo.Context) error {
	return service_jobs.SearchJobs(c)
}

func GetLatestJobs(c echo.Context) error {
	return service_jobs.GetLatestJobs(c)
}

func GetCompanies(c echo.Context) error {
	return service_jobs.GetCompanies(c)
}
