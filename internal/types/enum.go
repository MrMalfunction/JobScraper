package types

type ScrapableWebsites string

const (
	Workday ScrapableWebsites = "workday"
)

func AllScrapableWebsites() []ScrapableWebsites {
	return []ScrapableWebsites{
		Workday,
	}
}
