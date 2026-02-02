<template>
    <div class="todays-jobs">
        <div class="page-header">
            <h1>Today's Jobs</h1>
            <p>{{ jobsData.total }} jobs added today</p>
        </div>

        <!-- Compact Filters -->
        <div class="filters-compact">
            <input
                v-model="filters.company"
                type="text"
                placeholder="🏢 Company"
                class="filter-input"
            />
            <input
                v-model="filters.title"
                type="text"
                placeholder="💼 Job Title"
                class="filter-input"
            />
            <input
                v-model="includeKeywords"
                type="text"
                placeholder="✅ Include: go, python, remote"
                class="filter-input"
            />
            <input
                v-model="excludeKeywords"
                type="text"
                placeholder="❌ Exclude: senior, manager"
                class="filter-input"
            />
            <button @click="applyFilters" class="btn-search" :disabled="isLoading">
                {{ isLoading ? "..." : "Search" }}
            </button>
            <button @click="clearFilters" class="btn-clear" :disabled="isLoading">Clear</button>
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

                    <div v-if="job.job_insert_time" class="job-insert-time">
                        <span class="insert-time-badge"
                            >Added {{ getRelativeTime(job.job_insert_time) }}</span
                        >
                    </div>

                    <div class="job-preview">
                        {{ truncateText(job.job_details, 150) }}
                    </div>

                    <div class="job-actions">
                        <a :href="job.job_link" target="_blank" class="apply-btn">Apply</a>
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
            <h3>No jobs added today</h3>
            <p>There are no jobs added to the database today matching your criteria.</p>
        </div>

        <!-- Job Details Modal -->
        <JobDetailsModal v-if="selectedJob" :job="selectedJob" @close="closeJobDetails" />
    </div>
</template>

<script>
import axios from "axios";
import JobDetailsModal from "../components/JobDetailsModal.vue";
import KeywordFilters from "../components/KeywordFilters.vue";
import { useKeywordFilters } from "../composables/useKeywordFilters.js";

export default {
    name: "TodaysJobs",
    components: {
        JobDetailsModal,
        KeywordFilters,
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

    setup() {
        const keywordFilters = useKeywordFilters();
        return {
            ...keywordFilters,
        };
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

                // Add keyword filters
                Object.assign(params, this.buildKeywordParams());

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
            this.clearKeywords();
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
            // Extract date portion only (YYYY-MM-DD) to avoid timezone conversion
            // Handles both "2025-10-24" and "2025-10-24T00:00:00Z" formats
            const datePart = dateString.includes("T") ? dateString.split("T")[0] : dateString;
            const [year, month, day] = datePart.split("-");
            return new Date(year, month - 1, day).toLocaleDateString();
        },

        getRelativeTime(dateTimeString) {
            const date = new Date(dateTimeString);
            const now = new Date();
            const diffMs = now - date;
            const diffSec = Math.floor(diffMs / 1000);
            const diffMin = Math.floor(diffSec / 60);
            const diffHr = Math.floor(diffMin / 60);
            const diffDay = Math.floor(diffHr / 24);

            if (diffSec < 60) return "just now";
            if (diffMin < 60) return `${diffMin} minute${diffMin > 1 ? "s" : ""} ago`;
            if (diffHr < 24) return `${diffHr} hour${diffHr > 1 ? "s" : ""} ago`;
            if (diffDay < 7) return `${diffDay} day${diffDay > 1 ? "s" : ""} ago`;
            return date.toLocaleDateString("en-US", { month: "short", day: "numeric" });
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
    margin-bottom: 1.5rem;
}

.page-header h1 {
    font-size: 2rem;
    font-weight: 700;
    color: #2d3748;
    margin-bottom: 0.25rem;
}

.page-header p {
    font-size: 0.95rem;
    color: #718096;
}

.filters-compact {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 1.5rem;
    padding: 1rem;
    background: white;
    border-radius: 8px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    flex-wrap: wrap;
}

.filter-input {
    flex: 1;
    min-width: 180px;
    padding: 0.625rem 0.875rem;
    border: 1px solid #e2e8f0;
    border-radius: 6px;
    font-size: 0.9rem;
    transition: border-color 0.2s;
}

.filter-input:focus {
    outline: none;
    border-color: #667eea;
}

.filter-input::placeholder {
    color: #a0aec0;
}

.btn-search {
    padding: 0.625rem 1.5rem;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    border: none;
    border-radius: 6px;
    font-weight: 600;
    font-size: 0.9rem;
    cursor: pointer;
    transition: all 0.2s;
    white-space: nowrap;
}

.btn-search:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.btn-search:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

.btn-clear {
    padding: 0.625rem 1rem;
    background: white;
    color: #718096;
    border: 1px solid #e2e8f0;
    border-radius: 6px;
    font-weight: 600;
    font-size: 0.9rem;
    cursor: pointer;
    transition: all 0.2s;
    white-space: nowrap;
}

.btn-clear:hover:not(:disabled) {
    background: #f7fafc;
    border-color: #cbd5e0;
}

.btn-clear:disabled {
    opacity: 0.6;
    cursor: not-allowed;
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
    margin-bottom: 0.5rem;
}

.job-date {
    color: #718096;
    font-size: 0.9rem;
    background: #f7fafc;
    padding: 0.25rem 0.5rem;
    border-radius: 6px;
}

.job-insert-time {
    margin-bottom: 1rem;
}

.insert-time-badge {
    display: inline-block;
    color: #667eea;
    font-size: 0.85rem;
    background: #eef2ff;
    padding: 0.25rem 0.6rem;
    border-radius: 12px;
    font-weight: 500;
    border: 1px solid #dce4ff;
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

.apply-btn {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    text-decoration: none;
    padding: 0.5rem 1rem;
    border-radius: 6px;
    font-weight: 600;
    transition: all 0.3s ease;
    display: inline-block;
    margin-right: 0.5rem;
}

.apply-btn:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
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
    .filters-compact {
        flex-direction: column;
    }

    .filter-input {
        min-width: 100%;
    }

    .btn-search,
    .btn-clear {
        width: 100%;
    }

    .jobs-grid {
        grid-template-columns: 1fr;
    }

    .pagination {
        flex-direction: column;
        gap: 1rem;
    }

    .page-header h1 {
        font-size: 1.75rem;
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
