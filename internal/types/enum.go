package types

type ScrapableWebsites string

const (
	Workday     ScrapableWebsites = "workday"
	Greenhouse  ScrapableWebsites = "greenhouse"
	OracleCloud ScrapableWebsites = "oraclecloud"
)

func AllScrapableWebsites() []ScrapableWebsites {
	return []ScrapableWebsites{
		Workday,
		Greenhouse,
		OracleCloud,
	}
}
