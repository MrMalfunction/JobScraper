package service_jobs

import (
	"job-scraper/internal/api_models"
	"job-scraper/internal/db"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// SearchJobs searches for jobs by company name and/or job title
func SearchJobs(c echo.Context) error {
	// Parse query parameters
	company := strings.TrimSpace(c.QueryParam("company"))
	title := strings.TrimSpace(c.QueryParam("title"))

	limitStr := c.QueryParam("limit")
	if limitStr == "" {
		limitStr = "10"
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}

	offsetStr := c.QueryParam("offset")
	if offsetStr == "" {
		offsetStr = "0"
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Build query
	query := db.DB.Model(&db.Jobs{})

	if company != "" {
		query = query.Where("LOWER(company_name) ILIKE ?", "%"+strings.ToLower(company)+"%")
	}

	if title != "" {
		query = query.Where("LOWER(job_role) ILIKE ?", "%"+strings.ToLower(title)+"%")
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, api_models.StdResponse{
			Message: "Failed to count jobs",
			Data:    nil,
		})
	}

	// Get jobs with pagination
	var jobs []db.Jobs
	if err := query.Order("job_insert_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&jobs).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, api_models.StdResponse{
			Message: "Failed to fetch jobs",
			Data:    nil,
		})
	}

	// Convert to response format
	jobResponses := make([]api_models.JobResponse, len(jobs))
	for i, job := range jobs {
		jobResponses[i] = api_models.JobResponse{
			JobHash:       job.JobHash,
			JobId:         job.JobId,
			JobRole:       job.JobRole,
			JobDetails:    job.JobDetails,
			JobPostDate:   job.JobPostDate,
			JobInsertTime: job.JobInsertTime.Format(time.RFC3339),
			JobLink:       job.JobLink,
			JobAISummary:  job.JobAISummary,
			CompanyName:   job.CompanyName,
		}
	}

	// Calculate pagination info
	page := (offset / limit) + 1
	hasMore := int64(offset+limit) < totalCount

	response := api_models.JobSearchResponse{
		Jobs:    jobResponses,
		Total:   totalCount,
		Page:    page,
		Limit:   limit,
		HasMore: hasMore,
	}

	return c.JSON(http.StatusOK, api_models.StdResponse{
		Message: "Jobs retrieved successfully",
		Data:    response,
	})
}

// GetLatestJobs gets the most recent jobs
func GetLatestJobs(c echo.Context) error {
	limitStr := c.QueryParam("limit")
	if limitStr == "" {
		limitStr = "20"
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	var jobs []db.Jobs
	if err := db.DB.Order("job_insert_time DESC").
		Limit(limit).
		Find(&jobs).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, api_models.StdResponse{
			Message: "Failed to fetch latest jobs",
			Data:    nil,
		})
	}

	// Convert to response format
	jobResponses := make([]api_models.JobResponse, len(jobs))
	for i, job := range jobs {
		jobResponses[i] = api_models.JobResponse{
			JobHash:       job.JobHash,
			JobId:         job.JobId,
			JobRole:       job.JobRole,
			JobDetails:    job.JobDetails,
			JobPostDate:   job.JobPostDate,
			JobInsertTime: job.JobInsertTime.Format(time.RFC3339),
			JobLink:       job.JobLink,
			JobAISummary:  job.JobAISummary,
			CompanyName:   job.CompanyName,
		}
	}

	return c.JSON(http.StatusOK, api_models.StdResponse{
		Message: "Latest jobs retrieved successfully",
		Data:    jobResponses,
	})
}

// GetTodaysJobs gets jobs posted today with pagination and filtering
func GetTodaysJobs(c echo.Context) error {
	// Parse query parameters
	company := strings.TrimSpace(c.QueryParam("company"))
	title := strings.TrimSpace(c.QueryParam("title"))

	limitStr := c.QueryParam("limit")
	if limitStr == "" {
		limitStr = "20"
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offsetStr := c.QueryParam("offset")
	if offsetStr == "" {
		offsetStr = "0"
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get today's date range (start and end of today in local timezone)
	now := time.Now().Local()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1)

	// Build query for jobs inserted today
	query := db.DB.Model(&db.Jobs{}).Where("job_insert_time >= ? AND job_insert_time < ?", startOfDay, endOfDay)

	if company != "" {
		query = query.Where("LOWER(company_name) ILIKE ?", "%"+strings.ToLower(company)+"%")
	}

	if title != "" {
		query = query.Where("LOWER(job_role) ILIKE ?", "%"+strings.ToLower(title)+"%")
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, api_models.StdResponse{
			Message: "Failed to count today's jobs",
			Data:    nil,
		})
	}

	// Get jobs with pagination
	var jobs []db.Jobs
	if err := query.Order("job_insert_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&jobs).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, api_models.StdResponse{
			Message: "Failed to fetch today's jobs",
			Data:    nil,
		})
	}

	// Convert to response format
	jobResponses := make([]api_models.JobResponse, len(jobs))
	for i, job := range jobs {
		jobResponses[i] = api_models.JobResponse{
			JobHash:       job.JobHash,
			JobId:         job.JobId,
			JobRole:       job.JobRole,
			JobDetails:    job.JobDetails,
			JobPostDate:   job.JobPostDate,
			JobInsertTime: job.JobInsertTime.Format(time.RFC3339),
			JobLink:       job.JobLink,
			JobAISummary:  job.JobAISummary,
			CompanyName:   job.CompanyName,
		}
	}

	// Calculate pagination info
	page := (offset / limit) + 1
	hasMore := int64(offset+limit) < totalCount

	response := api_models.JobSearchResponse{
		Jobs:    jobResponses,
		Total:   totalCount,
		Page:    page,
		Limit:   limit,
		HasMore: hasMore,
	}

	return c.JSON(http.StatusOK, api_models.StdResponse{
		Message: "Today's jobs retrieved successfully",
		Data:    response,
	})
}

// GetAllJobs gets all jobs with pagination and filtering
func GetAllJobs(c echo.Context) error {
	// Parse query parameters
	company := strings.TrimSpace(c.QueryParam("company"))
	title := strings.TrimSpace(c.QueryParam("title"))

	limitStr := c.QueryParam("limit")
	if limitStr == "" {
		limitStr = "20"
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offsetStr := c.QueryParam("offset")
	if offsetStr == "" {
		offsetStr = "0"
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Build query for all jobs
	query := db.DB.Model(&db.Jobs{})

	if company != "" {
		query = query.Where("LOWER(company_name) ILIKE ?", "%"+strings.ToLower(company)+"%")
	}

	if title != "" {
		query = query.Where("LOWER(job_role) ILIKE ?", "%"+strings.ToLower(title)+"%")
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, api_models.StdResponse{
			Message: "Failed to count jobs",
			Data:    nil,
		})
	}

	// Get jobs with pagination
	var jobs []db.Jobs
	if err := query.Order("job_insert_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&jobs).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, api_models.StdResponse{
			Message: "Failed to fetch jobs",
			Data:    nil,
		})
	}

	// Convert to response format
	jobResponses := make([]api_models.JobResponse, len(jobs))
	for i, job := range jobs {
		jobResponses[i] = api_models.JobResponse{
			JobHash:       job.JobHash,
			JobId:         job.JobId,
			JobRole:       job.JobRole,
			JobDetails:    job.JobDetails,
			JobPostDate:   job.JobPostDate,
			JobInsertTime: job.JobInsertTime.Format(time.RFC3339),
			JobLink:       job.JobLink,
			JobAISummary:  job.JobAISummary,
			CompanyName:   job.CompanyName,
		}
	}

	// Calculate pagination info
	page := (offset / limit) + 1
	hasMore := int64(offset+limit) < totalCount

	response := api_models.JobSearchResponse{
		Jobs:    jobResponses,
		Total:   totalCount,
		Page:    page,
		Limit:   limit,
		HasMore: hasMore,
	}

	return c.JSON(http.StatusOK, api_models.StdResponse{
		Message: "All jobs retrieved successfully",
		Data:    response,
	})
}

// GetCompanies gets all companies
func GetCompanies(c echo.Context) error {
	var companies []db.Companies
	if err := db.DB.Find(&companies).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, api_models.StdResponse{
			Message: "Failed to fetch companies",
			Data:    nil,
		})
	}

	// Convert to response format
	companyResponses := make([]api_models.CompanyResponse, len(companies))
	for i, company := range companies {
		companyResponses[i] = api_models.CompanyResponse{
			Name:           company.Name,
			BaseUrl:        company.BaseUrl,
			CareerSiteType: company.CareerSiteType,
			ToScrape:       company.ToScrape,
		}
	}

	return c.JSON(http.StatusOK, api_models.StdResponse{
		Message: "Companies retrieved successfully",
		Data:    companyResponses,
	})
}

// DeleteOldJobs deletes jobs older than 10 days based on when they were inserted into the database
func DeleteOldJobs(c echo.Context) error {
	// Calculate the timestamp 10 days ago
	tenDaysAgo := time.Now().AddDate(0, 0, -10)

	// Delete jobs inserted more than 10 days ago
	result := db.DB.Where("job_insert_time < ?", tenDaysAgo).Delete(&db.Jobs{})

	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, api_models.StdResponse{
			Message: "Failed to delete old jobs",
			Data:    nil,
		})
	}

	response := map[string]interface{}{
		"deleted_count": result.RowsAffected,
		"cutoff_date":   tenDaysAgo.Format(time.RFC3339),
	}

	return c.JSON(http.StatusOK, api_models.StdResponse{
		Message: "Old jobs deleted successfully",
		Data:    response,
	})
}
