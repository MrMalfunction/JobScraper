<template>
    <div class="dashboard">
        <div class="page-header">
            <h1>Job Scraper Dashboard</h1>
            <p>Your centralized hub for tracking job postings and managing companies</p>
        </div>

        <!-- Interactive Stats Cards -->
        <div class="stats-grid">
            <div class="stat-card clickable" @click="$router.push('/all-jobs')">
                <div
                    class="stat-icon"
                    style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%)"
                >
                    📊
                </div>
                <div class="stat-content">
                    <h3>{{ totalJobs }}</h3>
                    <p>Total Jobs</p>
                </div>
                <div class="stat-arrow">→</div>
            </div>

            <div class="stat-card clickable" @click="$router.push('/companies')">
                <div
                    class="stat-icon"
                    style="background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%)"
                >
                    🏢
                </div>
                <div class="stat-content">
                    <h3>{{ totalCompanies }}</h3>
                    <p>Registered Companies</p>
                </div>
                <div class="stat-arrow">→</div>
            </div>

            <div class="stat-card clickable" @click="$router.push('/todays-jobs')">
                <div
                    class="stat-icon"
                    style="background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)"
                >
                    📅
                </div>
                <div class="stat-content">
                    <h3>{{ todaysJobs }}</h3>
                    <p>Today's Jobs</p>
                </div>
                <div class="stat-arrow">→</div>
            </div>

            <div class="stat-card clickable" @click="$router.push('/companies')">
                <div
                    class="stat-icon"
                    style="background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)"
                >
                    🔄
                </div>
                <div class="stat-content">
                    <h3>{{ activeCompanies }}</h3>
                    <p>Active Scrapers</p>
                </div>
                <div class="stat-arrow">→</div>
            </div>
        </div>

        <!-- Recent Activity -->
        <div class="card activity-card">
            <div class="section-header">
                <h2>Recent Activity</h2>
                <button @click="loadStats" class="btn btn-secondary" :disabled="isLoadingStats">
                    <span v-if="isLoadingStats" class="spinner"></span>
                    Refresh
                </button>
            </div>

            <div v-if="isLoadingStats" class="loading-container">
                <div class="loading-spinner"></div>
                <p>Loading statistics...</p>
            </div>

            <div v-else class="activity-summary">
                <div class="activity-item">
                    <div class="activity-icon">✅</div>
                    <div class="activity-content">
                        <h4>Job Database</h4>
                        <p>
                            {{ totalJobs }} total jobs tracked across {{ totalCompanies }} companies
                        </p>
                    </div>
                </div>

                <div class="activity-item">
                    <div class="activity-icon">🆕</div>
                    <div class="activity-content">
                        <h4>Today's Updates</h4>
                        <p>
                            {{ todaysJobs }} new job{{ todaysJobs !== 1 ? "s" : "" }} posted today
                        </p>
                    </div>
                </div>

                <div class="activity-item">
                    <div class="activity-icon">⚡</div>
                    <div class="activity-content">
                        <h4>Active Monitoring</h4>
                        <p>
                            {{ activeCompanies }} company scraper{{
                                activeCompanies !== 1 ? "s" : ""
                            }}
                            currently active
                        </p>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
import axios from "axios";

export default {
    name: "Dashboard",
    data() {
        return {
            // Stats
            totalJobs: 0,
            totalCompanies: 0,
            todaysJobs: 0,
            activeCompanies: 0,

            // UI state
            isLoadingStats: false,
        };
    },

    async mounted() {
        await this.loadStats();
    },

    methods: {
        async loadStats() {
            this.isLoadingStats = true;

            try {
                // Load companies
                const companiesResponse = await axios.get("/api/companies");
                const companies = companiesResponse.data.data;
                this.totalCompanies = companies.length;
                this.activeCompanies = companies.filter((c) => c.to_scrape).length;

                // Load today's jobs
                const todaysResponse = await axios.get("/api/jobs/today", { params: { limit: 1 } });
                this.todaysJobs = todaysResponse.data.data.total || 0;

                // Load all jobs count
                const allJobsResponse = await axios.get("/api/jobs/all", { params: { limit: 1 } });
                this.totalJobs = allJobsResponse.data.data.total || 0;
            } catch (error) {
                console.error("Error loading stats:", error);
            }

            this.isLoadingStats = false;
        },
    },
};
</script>

<style scoped>
.dashboard {
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
    margin-bottom: 3rem;
}

.page-header h1 {
    font-size: 2.8rem;
    font-weight: 700;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    margin-bottom: 0.5rem;
}

.page-header p {
    font-size: 1.2rem;
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
    border-radius: 16px;
    padding: 2rem;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
    border: 2px solid #e2e8f0;
    display: flex;
    align-items: center;
    gap: 1.5rem;
    transition: all 0.3s ease;
    position: relative;
    overflow: hidden;
}

.stat-card.clickable {
    cursor: pointer;
}

.stat-card.clickable:hover {
    transform: translateY(-8px);
    box-shadow: 0 20px 40px rgba(0, 0, 0, 0.15);
    border-color: #667eea;
}

.stat-card::before {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 4px;
    background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
    opacity: 0;
    transition: opacity 0.3s ease;
}

.stat-card:hover::before {
    opacity: 1;
}

.stat-icon {
    font-size: 2.5rem;
    width: 70px;
    height: 70px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 16px;
    flex-shrink: 0;
}

.stat-content {
    flex: 1;
}

.stat-content h3 {
    font-size: 2.5rem;
    font-weight: 700;
    color: #2d3748;
    margin: 0;
    line-height: 1;
}

.stat-content p {
    color: #718096;
    margin: 0.5rem 0 0 0;
    font-weight: 500;
    font-size: 0.95rem;
}

.stat-arrow {
    font-size: 1.5rem;
    color: #667eea;
    opacity: 0;
    transition: all 0.3s ease;
}

.stat-card:hover .stat-arrow {
    opacity: 1;
    transform: translateX(5px);
}

.section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
}

.activity-card {
    margin-bottom: 2rem;
}

.activity-summary {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
}

.activity-item {
    display: flex;
    align-items: center;
    gap: 1.5rem;
    padding: 1.5rem;
    background: #f7fafc;
    border-radius: 12px;
    border: 1px solid #e2e8f0;
    transition: all 0.3s ease;
}

.activity-item:hover {
    background: white;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.activity-icon {
    font-size: 2rem;
    width: 50px;
    height: 50px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: white;
    border-radius: 10px;
    flex-shrink: 0;
}

.activity-content h4 {
    font-size: 1rem;
    font-weight: 700;
    color: #2d3748;
    margin: 0 0 0.25rem 0;
}

.activity-content p {
    color: #718096;
    margin: 0;
    font-size: 0.9rem;
}

.loading-container {
    text-align: center;
    padding: 3rem 2rem;
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

@media (max-width: 1024px) {
    .stats-grid {
        grid-template-columns: repeat(2, 1fr);
    }
}

@media (max-width: 768px) {
    .page-header h1 {
        font-size: 2rem;
    }

    .page-header p {
        font-size: 1rem;
    }

    .stats-grid {
        grid-template-columns: 1fr;
    }

    .stat-card {
        padding: 1.5rem;
    }

    .stat-content h3 {
        font-size: 2rem;
    }

    .section-header {
        flex-direction: column;
        gap: 1rem;
        align-items: stretch;
    }
}

@media (max-width: 480px) {
    .stat-icon {
        width: 60px;
        height: 60px;
        font-size: 2rem;
    }
}
</style>
