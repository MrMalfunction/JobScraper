# Job Scraper with Vue Frontend

This application provides a job scraping service with a Vue.js frontend for managing companies and searching jobs.

## Features

- **Add Workday Companies**: Add companies using Workday as their career site
- **Add Greenhouse Companies**: Add companies using Greenhouse as their career site
- **Add Oracle Cloud Companies**: Add companies using Oracle Cloud as their career site
- **Job Search**: Search jobs by company name and/or job title
- **Latest Jobs**: View the most recently posted jobs
- **Company Management**: View all registered companies

## API Endpoints

### Company Management
- `POST /add_scrape_company/workday` - Add a new Workday company
- `POST /add_scrape_company/greenhouse` - Add a new Greenhouse company
- `POST /add_scrape_company/oraclecloud` - Add a new Oracle Cloud company
- `GET /api/companies` - Get all registered companies

### Job Search
- `GET /api/jobs/search?company=&title=&limit=&offset=` - Search jobs with optional filters
- `GET /api/jobs/latest?limit=` - Get latest jobs

### Scraping
- `GET /start_scrape` - Start scraping jobs for all registered companies

## Setup

### Prerequisites
- Go 1.24+ 
- Node.js 18+
- PostgreSQL database

### Backend Setup
1. Set up your database DSN in environment variables or config
2. Build the application:
   ```bash
   go build -o job-scraper main.go
   ```
3. Run the server:
   ```bash
   ./job-scraper
   ```

### Frontend Setup
1. Navigate to the frontend directory:
   ```bash
   cd frontend
   ```
2. Install dependencies:
   ```bash
   npm install
   ```
3. Build the frontend:
   ```bash
   npm run build
   ```

The built frontend will be served automatically by the Go server at `http://localhost:8080`.

### Development Mode
For frontend development with hot reload:
```bash
cd frontend
npm run dev
```
This will start the frontend on `http://localhost:3000` with API proxy to the Go server.

## Usage

### Adding a Workday Company

To add a company that uses Workday for job postings, you need:

1. **Company Name**: The display name for the company
2. **Base URL**: The Workday careers URL (e.g., `https://company.workdayapp.com`)
3. **Request Body**: JSON configuration for the job search API

**Sample Request Body:**
```json
{
  "searchText": "",
  "locations": [],
  "jobFamilies": [],
  "postedWithin": "",
  "limit": 20,
  "offset": 0
}
```

**Sample curl command:**
```bash
curl -X POST http://localhost:8080/add_scrape_company/workday \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Example Company",
    "base_url": "https://example.workdayapp.com",
    "req_body": {
      "searchText": "",
      "locations": [],
      "jobFamilies": [],
      "postedWithin": "",
      "limit": 20,
      "offset": 0
    }
  }'
```

### Adding a Greenhouse Company

To add a company that uses Greenhouse for job postings, you need:

1. **Company Name**: The display name for the company
2. **Base URL**: The Greenhouse boards API URL (e.g., `https://boards-api.greenhouse.io/v1/boards/companyname`)

**Sample curl command:**
```bash
curl -X POST http://localhost:8080/add_scrape_company/greenhouse \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Example Company",
    "base_url": "https://boards-api.greenhouse.io/v1/boards/examplecompany"
  }'
```

### Adding an Oracle Cloud Company

To add a company that uses Oracle Cloud for job postings:

1. **Company Name**: The display name for the company
2. **Browser URL**: Copy the full URL from your browser after selecting category and location filters on the Oracle career site

**Important Notes:**
- Navigate to the Oracle career site in your browser
- Optionally select desired **category** and **location** filters (categories are not required)
- Copy the complete URL from your browser's address bar
- The URL must contain the site number (e.g., `/sites/CX_1001/`)
- If `selectedPostingDatesFacet` is missing, it will default to 7 days
- Multiple categories are supported if provided (separated by semicolons in the URL)

**Example Browser URLs:**

With categories:
```
https://jpmc.fa.oraclecloud.com/hcmUI/CandidateExperience/en/sites/CX_1001/jobs?selectedCategoriesFacet=300000086152753&selectedPostingDatesFacet=7
```

Without categories (scrapes all jobs):
```
https://jpmc.fa.oraclecloud.com/hcmUI/CandidateExperience/en/sites/CX_1001/jobs
```

**Sample curl command:**
```bash
curl -X POST http://localhost:8080/add_scrape_company/oraclecloud \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Example Company",
    "browser_url": "https://company.fa.oraclecloud.com/hcmUI/CandidateExperience/en/sites/CX_1001/jobs"
  }'
```

Or with specific categories:
```bash
curl -X POST http://localhost:8080/add_scrape_company/oraclecloud \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Example Company",
    "browser_url": "https://company.fa.oraclecloud.com/hcmUI/CandidateExperience/en/sites/CX_1001/jobs?selectedCategoriesFacet=300000086152753"
  }'
```

The backend will automatically transform the browser URL to the appropriate REST API endpoint.

### Searching Jobs

Use the web interface at `http://localhost:8080` to:
- Add new companies (Workday, Greenhouse, or Oracle Cloud)
- Search for jobs by company and/or title
- View latest job postings
- Browse registered companies

## Database Schema

### Companies Table
- `name` (Primary Key): Company name
- `base_url`: Career site URL
- `career_site_type`: Type of career site (e.g., "workday", "greenhouse", "oraclecloud")
- `api_request_body`: JSON configuration for API requests (optional, used by Workday)
- `to_scrape`: Boolean indicating if company should be scraped

### Jobs Table
- `job_hash` (Primary Key): Unique job identifier
- `job_id`: Job ID from the source
- `job_role`: Job title/role
- `job_details`: Job description
- `job_post_date`: Date job was posted
- `job_link`: URL to the job posting
- `job_ai_summary`: AI-generated summary (optional)
- `company_name`: Foreign key to Companies table

## API Response Format

All API endpoints return responses in this format:
```json
{
  "message": "Success message",
  "data": { /* response data */ }
}
```

For job searches, the data includes pagination information:
```json
{
  "message": "Jobs retrieved successfully",
  "data": {
    "jobs": [...],
    "total": 150,
    "page": 1,
    "limit": 10,
    "has_more": true
  }
}
```