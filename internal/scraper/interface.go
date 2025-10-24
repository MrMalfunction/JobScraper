package scraper

import (
	"job-scraper/internal/db"
	"job-scraper/internal/scraper/workday"
	"job-scraper/internal/types"
	"time"
)

type scraper interface {
	StartScraping(companiesToScrape <-chan db.Companies, scrapeDayLimit time.Time)
}

func JobScraperFactory(provider types.ScrapableWebsites) scraper {
	switch provider {
	case types.Workday:
		return workday.WorkdayScraper{}
	default:
		return nil
	}
}
