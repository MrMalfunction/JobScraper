package service_scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"job-scraper/internal/api_models"
	"job-scraper/internal/db"
	"job-scraper/internal/scraper"
	"job-scraper/internal/types"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func StartJobScrapping(c echo.Context) error {

	var companies []db.Companies
	db.DB.Where(&db.Companies{ToScrape: true}).Order("career_site_type").Find(&companies)

	scraper_lister_channels := make(map[types.ScrapableWebsites]chan db.Companies)

	for _, company := range companies {
		switch company.CareerSiteType {
		case string(types.Workday):
			if scraper_lister_channels[types.Workday] == nil {
				scraper_lister_channels[types.Workday] = make(chan db.Companies, len(companies))
				workdayScraper := scraper.JobScraperFactory(types.Workday)
				go workdayScraper.StartScraping(scraper_lister_channels[types.Workday], time.Now().Truncate(24*time.Hour))
			}
			scraper_lister_channels[types.Workday] <- company
		default:
			slog.Debug("This Scraper Logic doesn't exist yet")
		}
	}

	// Close all created channels
	defer func() {
		for _, ch := range scraper_lister_channels {
			close(ch)
		}
	}()

	return c.JSON(http.StatusAccepted, api_models.StdResponse{
		Message: "Scrapping Job Accepted",
		Data:    nil,
	})
}

func AddWorkdayCompanyToScrapeList(c echo.Context) error {
	var workdayCompData api_models.AddWorkdayCompanyScrapeList

	if err := c.Bind(&workdayCompData); err != nil {
		return c.JSON(http.StatusBadRequest, api_models.StdResponse{
			Message: "Invalid request body",
			Data:    nil,
		})
	}

	// compact the JSON
	var compactBuf bytes.Buffer
	if err := json.Compact(&compactBuf, workdayCompData.ApiReqBody); err != nil {
		return c.JSON(http.StatusBadRequest, api_models.StdResponse{
			Message: "Invalid JSON in req_body",
			Data:    nil,
		})
	}

	companyDBData := db.Companies{
		Name:           workdayCompData.Name,
		BaseUrl:        workdayCompData.BaseUrl,
		CareerSiteType: string(types.Workday),
		ApiRequestBody: compactBuf.String(),
		ToScrape:       true,
	}

	if err := db.DB.Create(&companyDBData).Error; err != nil {
		slog.Error("Failed to insert company into database",
			"error", err,
			"company", companyDBData,
		)
		return c.JSON(http.StatusInternalServerError, api_models.StdResponse{
			Message: "Failed to insert company.",
			Data:    nil,
		})
	}

	slog.Info("Inserted Workday Company to DB.")
	return c.JSON(http.StatusAccepted, api_models.StdResponse{
		Message: fmt.Sprintf("Added %s company to scrape list", workdayCompData.Name),
		Data:    nil,
	})
}
