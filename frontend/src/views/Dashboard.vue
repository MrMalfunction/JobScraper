<template>
  <div class="dashboard">
    <div class="page-header">
      <h1>Dashboard</h1>
      <p>Manage companies and get quick access to job data</p>
    </div>

    <!-- Quick Stats Cards -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon">📊</div>
        <div class="stat-content">
          <h3>{{ totalJobs }}</h3>
          <p>Total Jobs</p>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon">🏢</div>
        <div class="stat-content">
          <h3>{{ totalCompanies }}</h3>
          <p>Registered Companies</p>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon">📅</div>
        <div class="stat-content">
          <h3>{{ todaysJobs }}</h3>
          <p>Today's Jobs</p>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon">🔄</div>
        <div class="stat-content">
          <h3>{{ activeCompanies }}</h3>
          <p>Active Scrapers</p>
        </div>
      </div>
    </div>

    <!-- Add Company Section -->
    <div class="card">
      <h2>Add Workday Company</h2>
      <form @submit.prevent="addCompany" class="company-form">
        <div class="form-row">
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
          <span v-if="isAddingCompany" class="spinner"></span>
          {{ isAddingCompany ? 'Adding Company...' : 'Add Company' }}
        </button>
      </form>
      
      <div v-if="addCompanyMessage" :class="['message', addCompanyMessageType]">
        {{ addCompanyMessage }}
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="actions-grid">
      <div class="action-card" @click="$router.push('/todays-jobs')">
        <div class="action-icon">📅</div>
        <h3>Today's Jobs</h3>
        <p>View all jobs posted today</p>
        <span class="action-arrow">→</span>
      </div>
      
      <div class="action-card" @click="$router.push('/all-jobs')">
        <div class="action-icon">📋</div>
        <h3>All Jobs</h3>
        <p>Browse complete job database</p>
        <span class="action-arrow">→</span>
      </div>
      
      <div class="action-card" @click="$router.push('/companies')">
        <div class="action-icon">🏢</div>
        <h3>Companies</h3>
        <p>Manage registered companies</p>
        <span class="action-arrow">→</span>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'Dashboard',
  data() {
    return {
      // Company form data
      companyForm: {
        name: '',
        baseUrl: '',
        reqBody: '{"searchText":"","locations":[],"jobFamilies":[],"postedWithin":"","limit":20,"offset":0}'
      },
      
      // UI state
      isAddingCompany: false,
      addCompanyMessage: '',
      addCompanyMessageType: '',
      
      // Stats
      totalJobs: 0,
      totalCompanies: 0,
      todaysJobs: 0,
      activeCompanies: 0
    }
  },
  
  async mounted() {
    await this.loadStats()
  },
  
  methods: {
    async loadStats() {
      try {
        // Load companies
        const companiesResponse = await axios.get('/api/companies')
        const companies = companiesResponse.data.data
        this.totalCompanies = companies.length
        this.activeCompanies = companies.filter(c => c.to_scrape).length
        
        // Load today's jobs
        const todaysResponse = await axios.get('/api/jobs/today', { params: { limit: 1 } })
        this.todaysJobs = todaysResponse.data.data.total || 0
        
        // Load all jobs count
        const allJobsResponse = await axios.get('/api/jobs/all', { params: { limit: 1 } })
        this.totalJobs = allJobsResponse.data.data.total || 0
        
      } catch (error) {
        console.error('Error loading stats:', error)
      }
    },
    
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
        
        // Refresh stats
        await this.loadStats()
        
      } catch (error) {
        console.error('Error adding company:', error)
        this.addCompanyMessage = error.response?.data?.message || 'Failed to add company'
        this.addCompanyMessageType = 'error'
      }
      
      this.isAddingCompany = false
    }
  }
}
</script>

<style scoped>
.dashboard {
  animation: fadeIn 0.5s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.page-header {
  text-align: center;
  margin-bottom: 3rem;
}

.page-header h1 {
  font-size: 2.5rem;
  font-weight: 700;
  color: #2d3748;
  margin-bottom: 0.5rem;
}

.page-header p {
  font-size: 1.1rem;
  color: #718096;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
  margin-bottom: 3rem;
}

.stat-card {
  background: white;
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
  border: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  gap: 1rem;
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
}

.stat-icon {
  font-size: 2rem;
  width: 60px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  color: white;
}

.stat-content h3 {
  font-size: 2rem;
  font-weight: 700;
  color: #2d3748;
  margin: 0;
}

.stat-content p {
  color: #718096;
  margin: 0;
  font-weight: 500;
}

.company-form {
  max-width: none;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1.5rem;
  margin-top: 3rem;
}

.action-card {
  background: white;
  border-radius: 12px;
  padding: 2rem;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
  border: 1px solid #e2e8f0;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
  text-align: center;
}

.action-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.15);
  border-color: #667eea;
}

.action-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.action-card h3 {
  font-size: 1.25rem;
  font-weight: 700;
  color: #2d3748;
  margin-bottom: 0.5rem;
}

.action-card p {
  color: #718096;
  margin-bottom: 1rem;
}

.action-arrow {
  position: absolute;
  top: 1rem;
  right: 1rem;
  font-size: 1.5rem;
  color: #667eea;
  opacity: 0;
  transition: all 0.3s ease;
}

.action-card:hover .action-arrow {
  opacity: 1;
  transform: translateX(5px);
}

.message {
  margin-top: 1rem;
  padding: 1rem;
  border-radius: 8px;
  font-weight: 500;
}

.message.success {
  background: #f0fff4;
  color: #38a169;
  border: 1px solid #9ae6b4;
}

.message.error {
  background: #fed7d7;
  color: #e53e3e;
  border: 1px solid #feb2b2;
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .form-row {
    grid-template-columns: 1fr;
  }
  
  .actions-grid {
    grid-template-columns: 1fr;
  }
  
  .page-header h1 {
    font-size: 2rem;
  }
}
</style>