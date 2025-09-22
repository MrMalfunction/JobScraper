<template>
  <div class="container">
    <div class="header">
      <h1>Job Scraper Dashboard</h1>
      <p>Manage companies and search for jobs</p>
    </div>

    <!-- Add Company Section -->
    <div class="card">
      <h2>Add Workday Company</h2>
      <form @submit.prevent="addCompany">
        <div class="form-group">
          <label for="companyName">Company Name:</label>
          <input
            id="companyName"
            v-model="companyForm.name"
            type="text"
            placeholder="e.g., Google"
            required
          />
        </div>
        
        <div class="form-group">
          <label for="baseUrl">Base URL:</label>
          <input
            id="baseUrl"
            v-model="companyForm.baseUrl"
            type="url"
            placeholder="https://company.workdayapp.com"
            required
          />
        </div>
        
        <div class="form-group">
          <label for="reqBody">Request Body (JSON):</label>
          <textarea
            id="reqBody"
            v-model="companyForm.reqBody"
            rows="6"
            placeholder='{"searchText":"","locations":[],"jobFamilies":[],"postedWithin":"","limit":20,"offset":0}'
            required
          ></textarea>
        </div>
        
        <button type="submit" class="btn" :disabled="isAddingCompany">
          {{ isAddingCompany ? 'Adding...' : 'Add Company' }}
        </button>
      </form>
      
      <div v-if="addCompanyMessage" :class="addCompanyMessageType">
        {{ addCompanyMessage }}
      </div>
    </div>

    <!-- Job Search Section -->
    <div class="card">
      <h2>Search Jobs</h2>
      <div class="search-filters">
        <div class="form-group">
          <label for="searchCompany">Company:</label>
          <input
            id="searchCompany"
            v-model="searchForm.company"
            type="text"
            placeholder="Company name"
          />
        </div>
        
        <div class="form-group">
          <label for="searchTitle">Job Title:</label>
          <input
            id="searchTitle"
            v-model="searchForm.title"
            type="text"
            placeholder="Job title"
          />
        </div>
        
        <div class="form-group">
          <button @click="searchJobs" class="btn" :disabled="isSearching">
            {{ isSearching ? 'Searching...' : 'Search' }}
          </button>
        </div>
      </div>
      
      <div style="margin-top: 1rem;">
        <button @click="getLatestJobs" class="btn" :disabled="isSearching">
          {{ isSearching ? 'Loading...' : 'Get Latest Jobs' }}
        </button>
      </div>
    </div>

    <!-- Results Section -->
    <div v-if="searchResults.jobs.length > 0" class="card">
      <h2>Search Results ({{ searchResults.total }} jobs found)</h2>
      <div class="job-list">
        <div v-for="job in searchResults.jobs" :key="job.job_hash" class="job-item">
          <div class="job-title">{{ job.job_role }}</div>
          <div class="job-company">{{ job.company_name }}</div>
          <div class="job-date">Posted: {{ job.job_post_date }}</div>
          <div class="job-details">{{ truncateText(job.job_details, 200) }}</div>
          <a :href="job.job_link" target="_blank" class="job-link">View Job</a>
        </div>
      </div>
      
      <!-- Pagination -->
      <div v-if="searchResults.total > searchResults.limit" class="pagination">
        <button 
          @click="changePage(currentPage - 1)" 
          :disabled="currentPage <= 1"
        >
          Previous
        </button>
        
        <button 
          v-for="page in visiblePages" 
          :key="page"
          @click="changePage(page)"
          :class="{ active: page === currentPage }"
        >
          {{ page }}
        </button>
        
        <button 
          @click="changePage(currentPage + 1)" 
          :disabled="!searchResults.has_more"
        >
          Next
        </button>
      </div>
    </div>

    <!-- Loading state -->
    <div v-if="isSearching" class="loading">
      <p>Loading jobs...</p>
    </div>

    <!-- No results -->
    <div v-if="!isSearching && searchResults.jobs.length === 0 && hasSearched" class="card">
      <p>No jobs found. Try adjusting your search criteria.</p>
    </div>

    <!-- Companies Section -->
    <div class="card">
      <h2>Registered Companies</h2>
      <button @click="loadCompanies" class="btn" style="margin-bottom: 1rem;">
        Refresh Companies
      </button>
      
      <div v-if="companies.length > 0">
        <div v-for="company in companies" :key="company.name" class="job-item">
          <div class="job-title">{{ company.name }}</div>
          <div class="job-company">{{ company.career_site_type }}</div>
          <div class="job-details">{{ company.base_url }}</div>
          <div style="color: #666; font-size: 0.9rem;">
            Status: {{ company.to_scrape ? 'Active' : 'Inactive' }}
          </div>
        </div>
      </div>
      
      <div v-else-if="!isLoadingCompanies">
        <p>No companies registered yet.</p>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'App',
  data() {
    return {
      // Company form data
      companyForm: {
        name: '',
        baseUrl: '',
        reqBody: '{"searchText":"","locations":[],"jobFamilies":[],"postedWithin":"","limit":20,"offset":0}'
      },
      
      // Search form data
      searchForm: {
        company: '',
        title: ''
      },
      
      // Results
      searchResults: {
        jobs: [],
        total: 0,
        page: 1,
        limit: 10,
        has_more: false
      },
      
      companies: [],
      
      // UI state
      isAddingCompany: false,
      isSearching: false,
      isLoadingCompanies: false,
      hasSearched: false,
      currentPage: 1,
      addCompanyMessage: '',
      addCompanyMessageType: ''
    }
  },
  
  computed: {
    visiblePages() {
      const totalPages = Math.ceil(this.searchResults.total / this.searchResults.limit)
      const pages = []
      const start = Math.max(1, this.currentPage - 2)
      const end = Math.min(totalPages, this.currentPage + 2)
      
      for (let i = start; i <= end; i++) {
        pages.push(i)
      }
      
      return pages
    }
  },
  
  mounted() {
    this.loadCompanies()
  },
  
  methods: {
    async addCompany() {
      this.isAddingCompany = true
      this.addCompanyMessage = ''
      
      try {
        // Validate JSON
        const reqBodyObj = JSON.parse(this.companyForm.reqBody)
        
        const response = await axios.post('/add_scrape_company/workday', {
          name: this.companyForm.name,
          base_url: this.companyForm.baseUrl,
          req_body: reqBodyObj
        })
        
        this.addCompanyMessage = response.data.message
        this.addCompanyMessageType = 'success'
        
        // Reset form
        this.companyForm.name = ''
        this.companyForm.baseUrl = ''
        this.companyForm.reqBody = '{"searchText":"","locations":[],"jobFamilies":[],"postedWithin":"","limit":20,"offset":0}'
        
        // Refresh companies list
        this.loadCompanies()
        
      } catch (error) {
        console.error('Error adding company:', error)
        this.addCompanyMessage = error.response?.data?.message || 'Failed to add company'
        this.addCompanyMessageType = 'error'
      }
      
      this.isAddingCompany = false
    },
    
    async searchJobs() {
      this.isSearching = true
      this.hasSearched = true
      this.currentPage = 1
      
      try {
        const params = {
          limit: 10,
          offset: 0
        }
        
        if (this.searchForm.company) {
          params.company = this.searchForm.company
        }
        
        if (this.searchForm.title) {
          params.title = this.searchForm.title
        }
        
        const response = await axios.get('/api/jobs/search', { params })
        this.searchResults = response.data.data
        
      } catch (error) {
        console.error('Error searching jobs:', error)
        this.searchResults = { jobs: [], total: 0, page: 1, limit: 10, has_more: false }
      }
      
      this.isSearching = false
    },
    
    async getLatestJobs() {
      this.isSearching = true
      this.hasSearched = true
      this.currentPage = 1
      
      try {
        const response = await axios.get('/api/jobs/latest', {
          params: { limit: 20 }
        })
        
        // Convert latest jobs response to search results format
        this.searchResults = {
          jobs: response.data.data,
          total: response.data.data.length,
          page: 1,
          limit: 20,
          has_more: false
        }
        
      } catch (error) {
        console.error('Error getting latest jobs:', error)
        this.searchResults = { jobs: [], total: 0, page: 1, limit: 10, has_more: false }
      }
      
      this.isSearching = false
    },
    
    async loadCompanies() {
      this.isLoadingCompanies = true
      
      try {
        const response = await axios.get('/api/companies')
        this.companies = response.data.data
      } catch (error) {
        console.error('Error loading companies:', error)
      }
      
      this.isLoadingCompanies = false
    },
    
    async changePage(page) {
      if (page < 1 || page === this.currentPage) return
      
      this.currentPage = page
      this.isSearching = true
      
      try {
        const params = {
          limit: 10,
          offset: (page - 1) * 10
        }
        
        if (this.searchForm.company) {
          params.company = this.searchForm.company
        }
        
        if (this.searchForm.title) {
          params.title = this.searchForm.title
        }
        
        const response = await axios.get('/api/jobs/search', { params })
        this.searchResults = response.data.data
        
      } catch (error) {
        console.error('Error loading page:', error)
      }
      
      this.isSearching = false
    },
    
    truncateText(text, maxLength) {
      if (!text) return ''
      if (text.length <= maxLength) return text
      return text.substr(0, maxLength) + '...'
    }
  }
}
</script>