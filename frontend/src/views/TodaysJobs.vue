<template>
    <div class="todays-jobs">
        <div class="page-header">
            <h1>Today's Jobs</h1>
            <p>Jobs posted today across all companies</p>
        </div>

        <!-- Filters -->
        <div class="card filters-card">
            <h3>Filters</h3>
            <div class="filters">
                <div class="form-group">
                    <label for="filterCompany">Company:</label>
                    <input
                        id="filterCompany"
                        v-model="filters.company"
                        type="text"
                        placeholder="Filter by company name"
                    />
                </div>

                <div class="form-group">
                    <label for="filterTitle">Job Title:</label>
                    <input
                        id="filterTitle"
                        v-model="filters.title"
                        type="text"
                        placeholder="Filter by job title"
                    />
                </div>

                <div class="form-group">
                    <button @click="applyFilters" class="btn" :disabled="isLoading">
                        <span v-if="isLoading" class="spinner"></span>
                        Apply Filters
                    </button>
                    <button @click="clearFilters" class="btn btn-secondary">Clear</button>
                </div>
            </div>
        </div>

        <!-- Results Summary -->
        <div v-if="!isLoading" class="results-summary">
            <p>{{ jobsData.total }} jobs found for today {{ getCurrentDate() }}</p>
        </div>

        <!-- Jobs List -->
        <div v-if="jobsData.jobs.length > 0" class="jobs-container">
            <div class="jobs-grid">
                <div
                    v-for="job in jobsData.jobs"
                    :key="job.job_hash"
                    class="job-card"
                    @click="openJobDetails(job)"
                >
                    <div class="job-header">
                        <h3 class="job-title">{{ job.job_role }}</h3>
                        <div class="job-company">{{ job.company_name }}</div>
                    </div>

                    <div class="job-meta">
                        <span class="job-date">{{ formatDate(job.job_post_date) }}</span>
                    </div>

                    <div class="job-preview">
                        {{ truncateText(job.job_details, 150) }}
                    </div>

                    <div class="job-actions">
                        <button class="view-details-btn">View Details</button>
                    </div>
                </div>
            </div>

            <!-- Pagination -->
            <div v-if="jobsData.total > jobsData.limit" class="pagination">
                <button
                    @click="changePage(currentPage - 1)"
                    :disabled="currentPage <= 1 || isLoading"
                    class="pagination-btn"
                >
                    ← Previous
                </button>

                <span class="pagination-info"> Page {{ currentPage }} of {{ totalPages }} </span>

                <button
                    @click="changePage(currentPage + 1)"
                    :disabled="!jobsData.has_more || isLoading"
                    class="pagination-btn"
                >
                    Next →
                </button>
            </div>
        </div>

        <!-- Loading State -->
        <div v-if="isLoading" class="loading-container">
            <div class="loading-spinner"></div>
            <p>Loading today's jobs...</p>
        </div>

        <!-- No Results -->
        <div v-if="!isLoading && jobsData.jobs.length === 0" class="no-results">
            <div class="no-results-icon">📭</div>
            <h3>No jobs found for today</h3>
            <p>There are no jobs posted today matching your criteria.</p>
        </div>

        <!-- Job Details Modal -->
        <JobDetailsModal v-if="selectedJob" :job="selectedJob" @close="closeJobDetails" />
    </div>
</template>

<script>
import axios from "axios";
import JobDetailsModal from "../components/JobDetailsModal.vue";

export default {
    name: "TodaysJobs",
    components: {
        JobDetailsModal,
    },
    data() {
        return {
            jobsData: {
                jobs: [],
                total: 0,
                page: 1,
                limit: 20,
                has_more: false,
            },
            filters: {
                company: "",
                title: "",
            },
            currentPage: 1,
            isLoading: false,
            selectedJob: null,
        };
    },

    computed: {
        totalPages() {
            return Math.ceil(this.jobsData.total / this.jobsData.limit);
        },
    },

    async mounted() {
        await this.loadTodaysJobs();
    },

    methods: {
        async loadTodaysJobs() {
            this.isLoading = true;

            try {
                const params = {
                    limit: 20,
                    offset: (this.currentPage - 1) * 20,
                };

                if (this.filters.company) {
                    params.company = this.filters.company;
                }

                if (this.filters.title) {
                    params.title = this.filters.title;
                }

                const response = await axios.get("/api/jobs/today", { params });
                this.jobsData = response.data.data;
            } catch (error) {
                console.error("Error loading today's jobs:", error);
                this.jobsData = { jobs: [], total: 0, page: 1, limit: 20, has_more: false };
            }

            this.isLoading = false;
        },

        async applyFilters() {
            this.currentPage = 1;
            await this.loadTodaysJobs();
        },

        async clearFilters() {
            this.filters.company = "";
            this.filters.title = "";
            this.currentPage = 1;
            await this.loadTodaysJobs();
        },

        async changePage(page) {
            if (page < 1 || page === this.currentPage) return;

            this.currentPage = page;
            await this.loadTodaysJobs();

            // Scroll to top
            window.scrollTo({ top: 0, behavior: "smooth" });
        },

        openJobDetails(job) {
            this.selectedJob = job;
        },

        closeJobDetails() {
            this.selectedJob = null;
        },

        truncateText(text, maxLength) {
            if (!text) return "";
            if (text.length <= maxLength) return text;
            return text.substr(0, maxLength) + "...";
        },

        getCurrentDate() {
            return new Date().toLocaleDateString();
        },

        formatDate(dateString) {
            return new Date(dateString).toLocaleDateString();
        },
    },
};
</script>

<style scoped>
.todays-jobs {
    animation: fadeIn 0.5s ease;
}

@keyframes fadeIn {
    from {
        opacity: 0;
        transform: translateY(20px);
    }
    to {
        opacity: 1;
        transform: translateY(0);
    }
}

.page-header {
    text-align: center;
    margin-bottom: 2rem;
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

.filters-card {
    margin-bottom: 2rem;
}

.filters {
    display: grid;
    grid-template-columns: 1fr 1fr auto auto;
    gap: 1rem;
    align-items: end;
}

.results-summary {
    text-align: center;
    margin-bottom: 2rem;
    font-size: 1.1rem;
    color: #4a5568;
    font-weight: 500;
}

.jobs-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 1.5rem;
    margin-bottom: 2rem;
}

@media (max-width: 1400px) {
    .jobs-grid {
        grid-template-columns: repeat(3, 1fr);
    }
}

@media (max-width: 1024px) {
    .jobs-grid {
        grid-template-columns: repeat(2, 1fr);
    }
}

.job-card {
    background: white;
    border-radius: 12px;
    padding: 1.5rem;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
    border: 1px solid #e2e8f0;
    cursor: pointer;
    transition: all 0.3s ease;
    position: relative;
    overflow: hidden;
}

.job-card:hover {
    transform: translateY(-4px);
    box-shadow: 0 12px 30px rgba(0, 0, 0, 0.15);
    border-color: #667eea;
}

.job-card::before {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 4px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.job-header {
    margin-bottom: 1rem;
}

.job-title {
    font-size: 1.25rem;
    font-weight: 700;
    color: #2d3748;
    margin-bottom: 0.5rem;
    line-height: 1.3;
}

.job-company {
    color: #667eea;
    font-weight: 600;
    font-size: 1rem;
}

.job-meta {
    margin-bottom: 1rem;
}

.job-date {
    color: #718096;
    font-size: 0.9rem;
    background: #f7fafc;
    padding: 0.25rem 0.5rem;
    border-radius: 6px;
}

.job-preview {
    color: #4a5568;
    line-height: 1.6;
    margin-bottom: 1rem;
    min-height: 60px;
}

.job-actions {
    text-align: right;
}

.view-details-btn {
    background: transparent;
    color: #667eea;
    border: 2px solid #667eea;
    padding: 0.5rem 1rem;
    border-radius: 6px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.3s ease;
}

.view-details-btn:hover {
    background: #667eea;
    color: white;
}

.pagination {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 2rem;
    margin: 2rem 0;
}

.pagination-btn {
    background: white;
    border: 2px solid #e2e8f0;
    padding: 0.75rem 1.5rem;
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.3s ease;
}

.pagination-btn:hover:not(:disabled) {
    border-color: #667eea;
    color: #667eea;
}

.pagination-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.pagination-info {
    font-weight: 600;
    color: #4a5568;
}

.loading-container {
    text-align: center;
    padding: 4rem 2rem;
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
    to {
        transform: rotate(360deg);
    }
}

.no-results {
    text-align: center;
    padding: 4rem 2rem;
    color: #718096;
}

.no-results-icon {
    font-size: 4rem;
    margin-bottom: 1rem;
}

.no-results h3 {
    font-size: 1.5rem;
    color: #4a5568;
    margin-bottom: 0.5rem;
}

@media (max-width: 768px) {
    .filters {
        grid-template-columns: 1fr;
        gap: 1rem;
    }

    .jobs-grid {
        grid-template-columns: 1fr;
    }

    .pagination {
        flex-direction: column;
        gap: 1rem;
    }

    .page-header h1 {
        font-size: 2rem;
    }
}

.todays-jobs {
    max-width: 1800px;
    margin: 0 auto;
    padding: 2rem 1.5rem;
}

@media (max-width: 768px) {
    .todays-jobs {
        padding: 2rem 0.75rem;
    }
}
</style>
