package types

type ScrapebleWebsites string

const (
	Workday ScrapebleWebsites = "workday"
)

func AllScrapableWebsites() []ScrapebleWebsites {
	return []ScrapebleWebsites{
		Workday,
	}
}
