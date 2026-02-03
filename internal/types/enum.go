package types

type ScrapableWebsites string

const (
	Workday     ScrapableWebsites = "workday"
	Greenhouse  ScrapableWebsites = "greenhouse"
	OracleCloud ScrapableWebsites = "oraclecloud"
	Generic     ScrapableWebsites = "generic"
)

func AllScrapableWebsites() []ScrapableWebsites {
	return []ScrapableWebsites{
		Workday,
		Greenhouse,
		OracleCloud,
		Generic,
	}
}
