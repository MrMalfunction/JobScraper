<template>
  <div class="companies">
    <div class="page-header">
      <h1>Companies</h1>
      <p>Manage registered companies and their scraping status</p>
    </div>

    <!-- Company Stats -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon">🏢</div>
        <div class="stat-content">
          <h3>{{ companies.length }}</h3>
          <p>Total Companies</p>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon">✅</div>
        <div class="stat-content">
          <h3>{{ activeCompanies }}</h3>
          <p>Active Scrapers</p>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon">⏸️</div>
        <div class="stat-content">
          <h3>{{ inactiveCompanies }}</h3>
          <p>Inactive</p>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon">🏷️</div>
        <div class="stat-content">
          <h3>{{ workdayCompanies }}</h3>
          <p>Workday Sites</p>
        </div>
      </div>
    </div>

    <!-- Add Company Section -->
    <div class="card">
      <h2>Add New Company</h2>
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
          <small class="form-help">
            Configure the JSON payload for the Workday API. Adjust parameters like locations, jobFamilies, and search text as needed.
          </small>
        </div>
        
        <div class="form-actions">
          <button type="submit" class="btn" :disabled="isAddingCompany">
            <span v-if="isAddingCompany" class="spinner"></span>
            {{ isAddingCompany ? 'Adding Company...' : 'Add Company' }}
          </button>
          <button type="button" @click="resetForm" class="btn btn-secondary">
            Reset Form
          </button>
        </div>
      </form>
      
      <div v-if="addCompanyMessage" :class="['message', addCompanyMessageType]">
        {{ addCompanyMessage }}
      </div>
    </div>

    <!-- Companies List -->
    <div class="card">
      <div class="section-header">
        <h2>Registered Companies</h2>
        <button @click="loadCompanies" class="btn btn-secondary" :disabled="isLoadingCompanies">
          <span v-if="isLoadingCompanies" class="spinner"></span>
          Refresh
        </button>
      </div>
      
      <div v-if="companies.length > 0" class="companies-grid">
        <div 
          v-for="company in companies" 
          :key="company.name" 
          class="company-card"
          :class="{ active: company.to_scrape }"
        >
          <div class="company-header">
            <h3 class="company-name">{{ company.name }}</h3>
            <div class="company-status">
              <span v-if="company.to_scrape" class="status-badge active">Active</span>
              <span v-else class="status-badge inactive">Inactive</span>
            </div>
          </div>
          
          <div class="company-details">
            <div class="detail-item">
              <strong>Type:</strong>
              <span>{{ company.career_site_type }}</span>
            </div>
            
            <div class="detail-item">
              <strong>URL:</strong>
              <a :href="company.base_url" target="_blank" class="company-url">
                {{ company.base_url }}
              </a>
            </div>
          </div>
          
          <div class="company-actions">
            <button 
              @click="toggleCompanyStatus(company)"
              :class="['action-btn', company.to_scrape ? 'btn-danger' : 'btn-success']"
              :disabled="isTogglingStatus"
            >
              {{ company.to_scrape ? 'Disable' : 'Enable' }}
            </button>
            
            <button @click="viewCompanyJobs(company)" class="action-btn btn-secondary">
              View Jobs
            </button>
          </div>
        </div>
      </div>
      
      <div v-else-if="!isLoadingCompanies" class="no-companies">
        <div class="no-companies-icon">🏢</div>
        <h3>No companies registered</h3>
        <p>Add your first company above to start scraping jobs.</p>
      </div>
      
      <div v-if="isLoadingCompanies" class="loading-container">
        <div class="loading-spinner"></div>
        <p>Loading companies...</p>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'Companies',
  data() {
    return {
      companies: [],
      companyForm: {
        name: '',
        baseUrl: '',
        reqBody: '{"searchText":"","locations":[],"jobFamilies":[],"postedWithin":"","limit":20,"offset":0}'
      },
      isLoadingCompanies: false,
      isAddingCompany: false,
      isTogglingStatus: false,
      addCompanyMessage: '',
      addCompanyMessageType: ''
    }
  },
  
  computed: {
    activeCompanies() {
      return this.companies.filter(c => c.to_scrape).length
    },
    
    inactiveCompanies() {
      return this.companies.filter(c => !c.to_scrape).length
    },
    
    workdayCompanies() {
      return this.companies.filter(c => c.career_site_type === 'workday').length
    }
  },
  
  async mounted() {
    await this.loadCompanies()
  },
  
  methods: {
    async loadCompanies() {
      this.isLoadingCompanies = true
      
      try {
        const response = await axios.get('/api/companies')
        this.companies = response.data.data
      } catch (error) {
        console.error('Error loading companies:', error)
        this.companies = []
      }
      
      this.isLoadingCompanies = false
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
        this.resetForm()
        
        // Refresh companies list
        await this.loadCompanies()
        
      } catch (error) {
        console.error('Error adding company:', error)
        this.addCompanyMessage = error.response?.data?.message || 'Failed to add company'
        this.addCompanyMessageType = 'error'
      }
      
      this.isAddingCompany = false
    },
    
    resetForm() {
      this.companyForm.name = ''
      this.companyForm.baseUrl = ''
      this.companyForm.reqBody = '{"searchText":"","locations":[],"jobFamilies":[],"postedWithin":"","limit":20,"offset":0}'
      this.addCompanyMessage = ''
    },
    
    async toggleCompanyStatus(company) {
      // Note: This would require a new API endpoint to update company status
      // For now, we'll just show a message
      this.isTogglingStatus = true
      
      try {
        // Simulate API call
        await new Promise(resolve => setTimeout(resolve, 1000))
        
        // Update local state
        company.to_scrape = !company.to_scrape
        
        this.showNotification('success', `Company ${company.name} ${company.to_scrape ? 'enabled' : 'disabled'}`)
        
      } catch (error) {
        console.error('Error toggling company status:', error)
        this.showNotification('error', 'Failed to update company status')
      }
      
      this.isTogglingStatus = false
    },
    
    viewCompanyJobs(company) {
      // Navigate to all jobs with company filter
      this.$router.push({
        name: 'AllJobs',
        query: { company: company.name }
      })
    },
    
    showNotification(type, message) {
      // Use the parent app's notification system
      this.$parent.showNotification(type, message)
    }
  }
}
</script>

<style scoped>
.companies {
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
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
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

.form-help {
  display: block;
  margin-top: 0.5rem;
  color: #718096;
  font-size: 0.9rem;
  line-height: 1.4;
}

.form-actions {
  display: flex;
  gap: 1rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.companies-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 1.5rem;
}

.company-card {
  background: white;
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
  border: 1px solid #e2e8f0;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.company-card.active::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(135deg, #48bb78 0%, #38a169 100%);
}

.company-card:not(.active)::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(135deg, #ed8936 0%, #dd6b20 100%);
}

.company-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
}

.company-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1rem;
}

.company-name {
  font-size: 1.25rem;
  font-weight: 700;
  color: #2d3748;
  margin: 0;
  flex: 1;
}

.company-status {
  margin-left: 1rem;
}

.status-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.status-badge.active {
  background: #c6f6d5;
  color: #22543d;
}

.status-badge.inactive {
  background: #fed7cc;
  color: #744210;
}

.company-details {
  margin-bottom: 1.5rem;
}

.detail-item {
  display: flex;
  margin-bottom: 0.5rem;
  align-items: center;
  gap: 0.5rem;
}

.detail-item strong {
  color: #4a5568;
  min-width: 60px;
  font-weight: 600;
}

.company-url {
  color: #667eea;
  text-decoration: none;
  word-break: break-all;
  font-size: 0.9rem;
}

.company-url:hover {
  text-decoration: underline;
}

.company-actions {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.action-btn {
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  border: none;
  font-size: 0.9rem;
}

.btn-success {
  background: #48bb78;
  color: white;
}

.btn-success:hover {
  background: #38a169;
}

.btn-danger {
  background: #f56565;
  color: white;
}

.btn-danger:hover {
  background: #e53e3e;
}

.loading-container {
  text-align: center;
  padding: 2rem;
  color: #718096;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #e2e8f0;
  border-radius: 50%;
  border-top-color: #667eea;
  animation: spin 1s ease-in-out infinite;
  margin: 0 auto 1rem;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.no-companies {
  text-align: center;
  padding: 4rem 2rem;
  color: #718096;
}

.no-companies-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
}

.no-companies h3 {
  font-size: 1.5rem;
  color: #4a5568;
  margin-bottom: 0.5rem;
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
    grid-template-columns: repeat(2, 1fr);
  }
  
  .form-row {
    grid-template-columns: 1fr;
  }
  
  .companies-grid {
    grid-template-columns: 1fr;
  }
  
  .section-header {
    flex-direction: column;
    gap: 1rem;
    align-items: stretch;
  }
  
  .company-header {
    flex-direction: column;
    gap: 0.5rem;
  }
  
  .company-status {
    margin-left: 0;
    align-self: flex-start;
  }
  
  .page-header h1 {
    font-size: 2rem;
  }
}

@media (max-width: 480px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .form-actions {
    flex-direction: column;
  }
  
  .company-actions {
    flex-direction: column;
  }
}
</style>