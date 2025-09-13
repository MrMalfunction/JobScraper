package scrapers

import (
	"job-scraper/manifests"
	"job-scraper/scrapers/workday"
	"job-scraper/types"
	"time"
)

type JobLister interface {
	ListJobs(jobsChannel chan<- *types.JobDetails, scrapeDateLimit time.Time)
}

type JobScraper interface {
	ScrapeJobDetails(jobsChannel <-chan *types.JobDetails, resultsChannel chan<- *types.JobDetails, numWorkers int)
}

type ScrapeProviders string

const (
	Workday string = "workday"
)

func JobListerFactory(providerName ScrapeProviders, configPath string) JobLister {
	switch providerName {
	case ScrapeProviders(Workday):
		config := manifests.LoadWorkdayCompanies(configPath)
		return workday.JobLister{Companies: config}
	default:
		return nil
	}
}

func JobScraperFactory(providerName ScrapeProviders) JobScraper {
	switch providerName {
	case ScrapeProviders(Workday):
		return workday.JobScraper{}
	default:
		return nil
	}
}
