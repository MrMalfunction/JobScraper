package service_scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"job-scraper/internal/api_models"
	"job-scraper/internal/db"
	"job-scraper/internal/scraper"
	"job-scraper/internal/scraper/oraclecloud"
	"job-scraper/internal/types"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
	"resty.dev/v3"
)

func StartJobScrapping(c echo.Context) error {

	var companies []db.Companies
	db.DB.Where(&db.Companies{ToScrape: true}).Order("career_site_type").Find(&companies)

	slog.Info("Starting job scraping session", "total_companies_enabled", len(companies))
	if len(companies) == 0 {
		slog.Warn("No companies enabled for scraping")
		return c.JSON(http.StatusAccepted, api_models.StdResponse{
			Message: "No companies enabled for scraping",
			Data:    nil,
		})
	}

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
		case string(types.Greenhouse):
			if scraper_lister_channels[types.Greenhouse] == nil {
				scraper_lister_channels[types.Greenhouse] = make(chan db.Companies, len(companies))
				greenhouseScraper := scraper.JobScraperFactory(types.Greenhouse)
				go greenhouseScraper.StartScraping(scraper_lister_channels[types.Greenhouse], time.Now().Truncate(24*time.Hour))
			}
			scraper_lister_channels[types.Greenhouse] <- company
		case string(types.OracleCloud):
			if scraper_lister_channels[types.OracleCloud] == nil {
				scraper_lister_channels[types.OracleCloud] = make(chan db.Companies, len(companies))
				oraclecloudScraper := scraper.JobScraperFactory(types.OracleCloud)
				go oraclecloudScraper.StartScraping(scraper_lister_channels[types.OracleCloud], time.Now().Truncate(24*time.Hour))
			}
			scraper_lister_channels[types.OracleCloud] <- company
		case string(types.Generic):
			if scraper_lister_channels[types.Generic] == nil {
				scraper_lister_channels[types.Generic] = make(chan db.Companies, len(companies))
				genericScraper := scraper.JobScraperFactory(types.Generic)
				go genericScraper.StartScraping(scraper_lister_channels[types.Generic], time.Now().Truncate(24*time.Hour))
			}
			scraper_lister_channels[types.Generic] <- company
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
		Message: fmt.Sprintf("Scraping started for %d companies", len(companies)),
		Data: map[string]interface{}{
			"companies_count": len(companies),
		},
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

func AddGreenhouseCompanyToScrapeList(c echo.Context) error {
	var greenhouseCompData api_models.AddGreenhouseCompanyScrapeList

	if err := c.Bind(&greenhouseCompData); err != nil {
		return c.JSON(http.StatusBadRequest, api_models.StdResponse{
			Message: "Invalid request body",
			Data:    nil,
		})
	}

	companyDBData := db.Companies{
		Name:           greenhouseCompData.Name,
		BaseUrl:        greenhouseCompData.BaseUrl,
		CareerSiteType: string(types.Greenhouse),
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

	slog.Info("Inserted Greenhouse Company to DB.")
	return c.JSON(http.StatusAccepted, api_models.StdResponse{
		Message: fmt.Sprintf("Added %s company to scrape list", greenhouseCompData.Name),
		Data:    nil,
	})
}

func AddOracleCloudCompanyToScrapeList(c echo.Context) error {
	var oracleCloudCompData api_models.AddOracleCloudCompanyScrapeList

	if err := c.Bind(&oracleCloudCompData); err != nil {
		return c.JSON(http.StatusBadRequest, api_models.StdResponse{
			Message: "Invalid request body",
			Data:    nil,
		})
	}

	// Transform browser URL to API URL
	apiURL, err := oraclecloud.TransformBrowserURLToAPIURL(oracleCloudCompData.BrowserUrl)
	if err != nil {
		slog.Error("Failed to transform Oracle Cloud URL",
			"error", err,
			"browserUrl", oracleCloudCompData.BrowserUrl,
		)
		return c.JSON(http.StatusBadRequest, api_models.StdResponse{
			Message: fmt.Sprintf("Failed to transform URL: %s", err.Error()),
			Data:    nil,
		})
	}

	companyDBData := db.Companies{
		Name:           oracleCloudCompData.Name,
		BaseUrl:        apiURL,
		CareerSiteType: string(types.OracleCloud),
		ToScrape:       true,
	}

	if err := db.DB.Create(&companyDBData).Error; err != nil {
		slog.Error("Failed to insert Oracle Cloud company into database",
			"error", err,
			"company", companyDBData,
		)
		return c.JSON(http.StatusInternalServerError, api_models.StdResponse{
			Message: "Failed to insert company.",
			Data:    nil,
		})
	}

	slog.Info("Inserted Oracle Cloud Company to DB.")
	return c.JSON(http.StatusAccepted, api_models.StdResponse{
		Message: fmt.Sprintf("Added %s company to scrape list", oracleCloudCompData.Name),
		Data:    nil,
	})
}

func UpdateCompany(c echo.Context) error {
	companyName := c.Param("name")
	if companyName == "" {
		return c.JSON(http.StatusBadRequest, api_models.StdResponse{
			Message: "Company name is required",
			Data:    nil,
		})
	}

	var updateReq api_models.UpdateCompanyRequest
	if err := c.Bind(&updateReq); err != nil {
		return c.JSON(http.StatusBadRequest, api_models.StdResponse{
			Message: "Invalid request body",
			Data:    nil,
		})
	}

	// Find the company
	var company db.Companies
	if err := db.DB.Where("name = ?", companyName).First(&company).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, api_models.StdResponse{
				Message: fmt.Sprintf("Company '%s' not found", companyName),
				Data:    nil,
			})
		}
		slog.Error("Failed to fetch company", "error", err, "company", companyName)
		return c.JSON(http.StatusInternalServerError, api_models.StdResponse{
			Message: "Failed to fetch company",
			Data:    nil,
		})
	}

	// Update fields if provided
	updateMap := make(map[string]interface{})

	if updateReq.BaseUrl != "" {
		updateMap["base_url"] = updateReq.BaseUrl
	}

	if updateReq.CareerSiteType != "" {
		updateMap["career_site_type"] = updateReq.CareerSiteType
	}

	if len(updateReq.ApiRequestBody) > 0 {
		// Compact the JSON if it's for Workday
		var compactBuf bytes.Buffer
		if err := json.Compact(&compactBuf, updateReq.ApiRequestBody); err != nil {
			return c.JSON(http.StatusBadRequest, api_models.StdResponse{
				Message: "Invalid JSON in api_request_body",
				Data:    nil,
			})
		}
		updateMap["api_request_body"] = compactBuf.String()
	}

	if updateReq.ApiRequestQueryParam != "" {
		updateMap["api_request_query_param"] = updateReq.ApiRequestQueryParam
	}

	if updateReq.ApiRequestHeaders != "" {
		updateMap["api_request_headers"] = updateReq.ApiRequestHeaders
	}

	if updateReq.ApiRequestMethod != "" {
		updateMap["api_request_method"] = updateReq.ApiRequestMethod
	}

	if updateReq.PaginationKey != "" {
		updateMap["pagination_key"] = updateReq.PaginationKey
	}

	if updateReq.ResponseJsonPath != "" {
		updateMap["response_json_path"] = updateReq.ResponseJsonPath
	}

	if updateReq.JobIdJsonPath != "" {
		updateMap["job_id_json_path"] = updateReq.JobIdJsonPath
	}

	if updateReq.JobTitleJsonPath != "" {
		updateMap["job_title_json_path"] = updateReq.JobTitleJsonPath
	}

	if updateReq.JobDetailsJsonPath != "" {
		updateMap["job_details_json_path"] = updateReq.JobDetailsJsonPath
	}

	if updateReq.JobLinkJsonPath != "" {
		updateMap["job_link_json_path"] = updateReq.JobLinkJsonPath
	}

	if updateReq.JobDateJsonPath != "" {
		updateMap["job_date_json_path"] = updateReq.JobDateJsonPath
	}

	if updateReq.ToScrape != nil {
		updateMap["to_scrape"] = *updateReq.ToScrape
	}

	// Rename is handled separately since it's the primary key
	if updateReq.Name != "" && updateReq.Name != companyName {
		updateMap["name"] = updateReq.Name
	}

	if len(updateMap) == 0 {
		return c.JSON(http.StatusBadRequest, api_models.StdResponse{
			Message: "No fields to update",
			Data:    nil,
		})
	}

	// Update the company
	if err := db.DB.Model(&company).Updates(updateMap).Error; err != nil {
		slog.Error("Failed to update company", "error", err, "company", companyName)
		return c.JSON(http.StatusInternalServerError, api_models.StdResponse{
			Message: "Failed to update company",
			Data:    nil,
		})
	}

	slog.Info("Company updated", "company", companyName, "to_scrape", updateMap["to_scrape"])

	return c.JSON(http.StatusOK, api_models.StdResponse{
		Message: fmt.Sprintf("Company '%s' updated successfully", companyName),
		Data:    nil,
	})
}

func DeleteCompany(c echo.Context) error {
	companyName := c.Param("name")
	if companyName == "" {
		return c.JSON(http.StatusBadRequest, api_models.StdResponse{
			Message: "Company name is required",
			Data:    nil,
		})
	}

	// Find the company first
	var company db.Companies
	if err := db.DB.Where("name = ?", companyName).First(&company).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, api_models.StdResponse{
				Message: fmt.Sprintf("Company '%s' not found", companyName),
				Data:    nil,
			})
		}
		slog.Error("Failed to fetch company", "error", err, "company", companyName)
		return c.JSON(http.StatusInternalServerError, api_models.StdResponse{
			Message: "Failed to fetch company",
			Data:    nil,
		})
	}

	// Delete the company (cascade will delete associated jobs)
	if err := db.DB.Delete(&company).Error; err != nil {
		slog.Error("Failed to delete company", "error", err, "company", companyName)
		return c.JSON(http.StatusInternalServerError, api_models.StdResponse{
			Message: "Failed to delete company",
			Data:    nil,
		})
	}

	slog.Info("Deleted company", "company", companyName)
	return c.JSON(http.StatusOK, api_models.StdResponse{
		Message: fmt.Sprintf("Company '%s' and associated jobs deleted successfully", companyName),
		Data:    nil,
	})
}

func AddGenericCompanyToScrapeList(c echo.Context) error {
	var genericCompData api_models.AddGenericCompanyScrapeList

	if err := c.Bind(&genericCompData); err != nil {
		return c.JSON(http.StatusBadRequest, api_models.StdResponse{
			Message: "Invalid request body",
			Data:    nil,
		})
	}

	// Validate the request
	if err := c.Validate(&genericCompData); err != nil {
		return c.JSON(http.StatusBadRequest, api_models.StdResponse{
			Message: fmt.Sprintf("Validation error: %s", err.Error()),
			Data:    nil,
		})
	}

	// If dry run is requested, test the configuration
	if genericCompData.DryRun {
		dryRunResult := performDryRun(&genericCompData)
		return c.JSON(http.StatusOK, dryRunResult)
	}

	// Convert headers to JSON string
	headersJSON, err := json.Marshal(genericCompData.Headers)
	if err != nil {
		return c.JSON(http.StatusBadRequest, api_models.StdResponse{
			Message: "Invalid headers format",
			Data:    nil,
		})
	}

	// Convert body to JSON string
	var bodyJSON []byte
	if len(genericCompData.Body) > 0 {
		bodyJSON, err = json.Marshal(genericCompData.Body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, api_models.StdResponse{
				Message: "Invalid body format",
				Data:    nil,
			})
		}
	}

	companyDBData := db.Companies{
		Name:               genericCompData.Name,
		BaseUrl:            genericCompData.BaseUrl,
		CareerSiteType:     string(types.Generic),
		ApiRequestMethod:   genericCompData.Method,
		ApiRequestHeaders:  string(headersJSON),
		ApiRequestBody:     string(bodyJSON),
		ApiRequestQueryParam: genericCompData.QueryParams,
		PaginationKey:      genericCompData.PaginationKey,
		ResponseJsonPath:   genericCompData.ResponseJsonPath,
		JobIdJsonPath:      genericCompData.JobIdJsonPath,
		JobTitleJsonPath:   genericCompData.JobTitleJsonPath,
		JobDetailsJsonPath: genericCompData.JobDetailsJsonPath,
		JobLinkJsonPath:    genericCompData.JobLinkJsonPath,
		JobDateJsonPath:    genericCompData.JobDateJsonPath,
		ToScrape:           true,
	}

	if err := db.DB.Create(&companyDBData).Error; err != nil {
		slog.Error("Failed to insert generic company into database",
			"error", err,
			"company", companyDBData,
		)
		return c.JSON(http.StatusInternalServerError, api_models.StdResponse{
			Message: "Failed to insert company.",
			Data:    nil,
		})
	}

	slog.Info("Inserted Generic Company to DB.", "company", genericCompData.Name)
	return c.JSON(http.StatusAccepted, api_models.StdResponse{
		Message: fmt.Sprintf("Added %s company to scrape list", genericCompData.Name),
		Data:    nil,
	})
}

func performDryRun(config *api_models.AddGenericCompanyScrapeList) api_models.DryRunResponse {
	slog.Info("Performing dry run", "company", config.Name, "url", config.BaseUrl)

	// Create HTTP client
	rClient := resty.New()
	rClient.SetHeader("User-Agent", "JobScraper/1.0")
	defer rClient.Close()

	// Build request
	req := rClient.R()

	// Add headers
	if len(config.Headers) > 0 {
		req.SetHeaders(config.Headers)
	}

	// Execute request
	var resp *resty.Response
	var err error

	fullUrl := config.BaseUrl
	if config.QueryParams != "" {
		fullUrl += "?" + config.QueryParams
	}

	// For dry run, set pagination to first page/offset
	bodyToSend := make(map[string]interface{})
	for k, v := range config.Body {
		bodyToSend[k] = v
	}
	if config.PaginationKey != "" {
		bodyToSend[config.PaginationKey] = 0
	}

	if config.Method == "POST" {
		resp, err = req.SetBody(bodyToSend).Post(fullUrl)
	} else {
		// Convert body to query params for GET
		if len(bodyToSend) > 0 {
			queryParams := make(map[string]string)
			for k, v := range bodyToSend {
				queryParams[k] = fmt.Sprintf("%v", v)
			}
			req.SetQueryParams(queryParams)
		}
		resp, err = req.Get(fullUrl)
	}

	if err != nil {
		return api_models.DryRunResponse{
			Valid:        false,
			Message:      "Failed to fetch data from API",
			ErrorDetails: err.Error(),
		}
	}

	if resp.StatusCode() != 200 {
		return api_models.DryRunResponse{
			Valid:        false,
			Message:      fmt.Sprintf("API returned non-200 status code: %d", resp.StatusCode()),
			ErrorDetails: resp.String(),
		}
	}

	responseBody := resp.String()

	// Try to extract jobs using JSON path
	jobsResult := gjson.Get(responseBody, config.ResponseJsonPath)
	if !jobsResult.Exists() {
		return api_models.DryRunResponse{
			Valid:        false,
			Message:      "Response JSON path not found in API response",
			ErrorDetails: fmt.Sprintf("Path '%s' does not exist", config.ResponseJsonPath),
		}
	}

	if !jobsResult.IsArray() {
		return api_models.DryRunResponse{
			Valid:        false,
			Message:      "Response JSON path does not point to an array",
			ErrorDetails: fmt.Sprintf("Path '%s' is not an array", config.ResponseJsonPath),
		}
	}

	jobs := jobsResult.Array()
	if len(jobs) == 0 {
		return api_models.DryRunResponse{
			Valid:   true,
			Message: "Configuration is valid but no jobs found in response",
		}
	}

	// Extract sample data from first few jobs
	sampleData := make([]map[string]interface{}, 0)
	maxSamples := 3
	if len(jobs) < maxSamples {
		maxSamples = len(jobs)
	}

	for i := 0; i < maxSamples; i++ {
		jobJSON := jobs[i].Raw
		sample := map[string]interface{}{
			"job_id":      gjson.Get(jobJSON, config.JobIdJsonPath).String(),
			"job_title":   gjson.Get(jobJSON, config.JobTitleJsonPath).String(),
			"job_link":    gjson.Get(jobJSON, config.JobLinkJsonPath).String(),
			"job_details": truncateString(gjson.Get(jobJSON, config.JobDetailsJsonPath).String(), 100),
			"job_date":    gjson.Get(jobJSON, config.JobDateJsonPath).String(),
		}
		sampleData = append(sampleData, sample)
	}

	return api_models.DryRunResponse{
		Valid:         true,
		Message:       "Configuration validated successfully",
		SampleData:    sampleData,
		ExtractedJobs: len(jobs),
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
