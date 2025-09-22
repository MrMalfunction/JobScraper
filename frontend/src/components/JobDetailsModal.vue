<template>
  <div class="modal-overlay" @click="closeModal">
    <div class="modal-container" @click.stop>
      <div class="modal-header">
        <h2>{{ job.job_role }}</h2>
        <button @click="closeModal" class="close-btn">×</button>
      </div>
      
      <div class="modal-content">
        <div class="job-info">
          <div class="info-item">
            <strong>Company:</strong>
            <span>{{ job.company_name }}</span>
          </div>
          
          <div class="info-item">
            <strong>Posted Date:</strong>
            <span>{{ formatDate(job.job_post_date) }}</span>
          </div>
          
          <div v-if="job.job_id" class="info-item">
            <strong>Job ID:</strong>
            <span>{{ job.job_id }}</span>
          </div>
        </div>
        
        <div class="job-description">
          <h3>Job Description</h3>
          <div class="description-content" v-html="formatJobDetails(job.job_details)"></div>
        </div>
        
        <div v-if="job.job_ai_summary" class="ai-summary">
          <h3>AI Summary</h3>
          <p>{{ job.job_ai_summary }}</p>
        </div>
      </div>
      
      <div class="modal-footer">
        <a 
          :href="job.job_link" 
          target="_blank" 
          class="apply-btn"
          @click="trackApplication"
        >
          Apply Now
        </a>
        <button @click="closeModal" class="close-btn-secondary">
          Close
        </button>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'JobDetailsModal',
  props: {
    job: {
      type: Object,
      required: true
    }
  },
  
  mounted() {
    // Prevent body scroll when modal is open
    document.body.style.overflow = 'hidden'
    
    // Close modal on Escape key
    document.addEventListener('keydown', this.handleKeydown)
  },
  
  beforeUnmount() {
    // Restore body scroll
    document.body.style.overflow = 'auto'
    
    // Remove event listener
    document.removeEventListener('keydown', this.handleKeydown)
  },
  
  methods: {
    closeModal() {
      this.$emit('close')
    },
    
    handleKeydown(event) {
      if (event.key === 'Escape') {
        this.closeModal()
      }
    },
    
    formatDate(dateString) {
      return new Date(dateString).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'long',
        day: 'numeric'
      })
    },
    
    formatJobDetails(details) {
      if (!details) return ''
      
      // Convert plain text to HTML with basic formatting
      return details
        .replace(/\n\n/g, '</p><p>')
        .replace(/\n/g, '<br>')
        .replace(/^/, '<p>')
        .replace(/$/, '</p>')
    },
    
    trackApplication() {
      // Track application click for analytics
      console.log('Application clicked for job:', this.job.job_hash)
    }
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.modal-container {
  background: white;
  border-radius: 12px;
  max-width: 800px;
  width: 100%;
  max-height: 90vh;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  animation: slideIn 0.3s ease;
  display: flex;
  flex-direction: column;
}

@keyframes slideIn {
  from { 
    opacity: 0; 
    transform: translateY(-50px) scale(0.95); 
  }
  to { 
    opacity: 1; 
    transform: translateY(0) scale(1); 
  }
}

.modal-header {
  padding: 2rem 2rem 1rem;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.modal-header h2 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
  line-height: 1.3;
  padding-right: 2rem;
}

.close-btn {
  background: none;
  border: none;
  font-size: 2rem;
  color: white;
  cursor: pointer;
  padding: 0;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: all 0.3s ease;
  flex-shrink: 0;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.2);
  transform: scale(1.1);
}

.modal-content {
  padding: 2rem;
  overflow-y: auto;
  flex: 1;
}

.job-info {
  margin-bottom: 2rem;
  display: grid;
  gap: 1rem;
}

.info-item {
  display: flex;
  gap: 1rem;
}

.info-item strong {
  color: #4a5568;
  min-width: 120px;
  font-weight: 600;
}

.info-item span {
  color: #2d3748;
}

.job-description,
.ai-summary {
  margin-bottom: 2rem;
}

.job-description h3,
.ai-summary h3 {
  color: #2d3748;
  font-size: 1.25rem;
  font-weight: 700;
  margin-bottom: 1rem;
  border-bottom: 2px solid #e2e8f0;
  padding-bottom: 0.5rem;
}

.description-content {
  color: #4a5568;
  line-height: 1.8;
}

.description-content :deep(p) {
  margin-bottom: 1rem;
}

.ai-summary p {
  color: #4a5568;
  line-height: 1.8;
  background: #f7fafc;
  padding: 1rem;
  border-radius: 8px;
  border-left: 4px solid #667eea;
}

.modal-footer {
  padding: 1.5rem 2rem;
  border-top: 1px solid #e2e8f0;
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  background: #f7fafc;
}

.apply-btn {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  text-decoration: none;
  padding: 0.75rem 2rem;
  border-radius: 8px;
  font-weight: 600;
  transition: all 0.3s ease;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.apply-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.apply-btn::after {
  content: '↗';
  font-size: 1.2em;
}

.close-btn-secondary {
  background: #e2e8f0;
  color: #4a5568;
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.close-btn-secondary:hover {
  background: #cbd5e0;
}

@media (max-width: 768px) {
  .modal-container {
    margin: 0;
    max-height: 100vh;
    border-radius: 0;
  }
  
  .modal-header {
    padding: 1.5rem 1.5rem 1rem;
  }
  
  .modal-header h2 {
    font-size: 1.25rem;
  }
  
  .modal-content {
    padding: 1.5rem;
  }
  
  .modal-footer {
    padding: 1rem 1.5rem;
    flex-direction: column;
  }
  
  .info-item {
    flex-direction: column;
    gap: 0.25rem;
  }
  
  .info-item strong {
    min-width: auto;
  }
}
</style>