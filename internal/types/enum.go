package types

type ScrapableWebsites string

const (
	Workday    ScrapableWebsites = "workday"
	Greenhouse ScrapableWebsites = "greenhouse"
)

func AllScrapableWebsites() []ScrapableWebsites {
	return []ScrapableWebsites{
		Workday,
		Greenhouse,
	}
}
