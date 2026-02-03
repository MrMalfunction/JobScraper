package scraper

import (
	"job-scraper/internal/db"
	"job-scraper/internal/scraper/generic"
	"job-scraper/internal/scraper/greenhouse"
	"job-scraper/internal/scraper/oraclecloud"
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
	case types.Greenhouse:
		return greenhouse.GreenhouseScraper{}
	case types.OracleCloud:
		return oraclecloud.OracleCloudScraper{}
	case types.Generic:
		return generic.GenericScraper{}
	default:
		return nil
	}
}
