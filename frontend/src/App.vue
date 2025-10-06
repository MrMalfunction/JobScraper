<template>
    <div id="app">
        <!-- Navigation Header -->
        <nav class="navbar">
            <div class="nav-container">
                <div class="nav-brand">
                    <h1>Job Scraper</h1>
                </div>
                <ul class="nav-menu">
                    <li class="nav-item">
                        <router-link to="/" class="nav-link" active-class="active"
                            >Dashboard</router-link
                        >
                    </li>
                    <li class="nav-item">
                        <router-link to="/todays-jobs" class="nav-link" active-class="active"
                            >Today's Jobs</router-link
                        >
                    </li>
                    <li class="nav-item">
                        <router-link to="/all-jobs" class="nav-link" active-class="active"
                            >All Jobs</router-link
                        >
                    </li>
                    <li class="nav-item">
                        <router-link to="/companies" class="nav-link" active-class="active"
                            >Companies</router-link
                        >
                    </li>
                    <li class="nav-item">
                        <button
                            @click="startScrape"
                            class="scrape-btn"
                            :disabled="isScrapingStarted"
                        >
                            {{ isScrapingStarted ? "Scraping..." : "Start Scrape" }}
                        </button>
                    </li>
                    <li class="nav-item">
                        <button
                            @click="deleteOldJobs"
                            class="delete-btn"
                            :disabled="isDeletingJobs"
                        >
                            {{ isDeletingJobs ? "Deleting..." : "Delete Old Jobs" }}
                        </button>
                    </li>
                </ul>
            </div>
        </nav>

        <!-- Main Content -->
        <main class="main-content">
            <router-view />
        </main>

        <!-- Global notifications -->
        <div v-if="notification" :class="['notification', notification.type]">
            {{ notification.message }}
        </div>
    </div>
</template>

<script>
import axios from "axios";

export default {
    name: "App",
    data() {
        return {
            isScrapingStarted: false,
            isDeletingJobs: false,
            notification: null,
        };
    },
    methods: {
        async startScrape() {
            this.isScrapingStarted = true;

            try {
                const response = await axios.get("/start_scrape");
                this.showNotification(
                    "success",
                    response.data.message || "Scraping started successfully",
                );
            } catch (error) {
                console.error("Error starting scrape:", error);
                this.showNotification(
                    "error",
                    error.response?.data?.message || "Failed to start scraping",
                );
            }

            // Reset button after 5 seconds
            setTimeout(() => {
                this.isScrapingStarted = false;
            }, 5000);
        },

        async deleteOldJobs() {
            this.isDeletingJobs = true;

            try {
                const response = await axios.delete("/api/jobs/cleanup");
                this.showNotification(
                    "success",
                    response.data.message || "Old jobs deleted successfully",
                );
            } catch (error) {
                console.error("Error deleting old jobs:", error);
                this.showNotification(
                    "error",
                    error.response?.data?.message || "Failed to delete old jobs",
                );
            }

            // Reset button after 3 seconds
            setTimeout(() => {
                this.isDeletingJobs = false;
            }, 3000);
        },

        showNotification(type, message) {
            this.notification = { type, message };
            setTimeout(() => {
                this.notification = null;
            }, 5000);
        },
    },
};
</script>

<style>
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: "Segoe UI", Tahoma, Geneva, Verdana, sans-serif;
    background-color: #f8fafc;
    color: #2d3748;
    line-height: 1.6;
}

#app {
    min-height: 100vh;
}

/* Navigation Styles */
.navbar {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    padding: 0;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
    position: sticky;
    top: 0;
    z-index: 100;
}

.nav-container {
    max-width: 1200px;
    margin: 0 auto;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 2rem;
}

.nav-brand h1 {
    font-size: 1.5rem;
    font-weight: 700;
}

.nav-menu {
    display: flex;
    list-style: none;
    align-items: center;
    gap: 2rem;
}

.nav-link {
    color: white;
    text-decoration: none;
    padding: 0.5rem 1rem;
    border-radius: 6px;
    transition: all 0.3s ease;
    font-weight: 500;
}

.nav-link:hover {
    background-color: rgba(255, 255, 255, 0.1);
    transform: translateY(-1px);
}

.nav-link.active {
    background-color: rgba(255, 255, 255, 0.2);
    font-weight: 600;
}

.scrape-btn {
    background: rgba(255, 255, 255, 0.2);
    color: white;
    border: 2px solid rgba(255, 255, 255, 0.3);
    padding: 0.5rem 1rem;
    border-radius: 6px;
    cursor: pointer;
    font-weight: 600;
    transition: all 0.3s ease;
}

.scrape-btn:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.3);
    border-color: rgba(255, 255, 255, 0.5);
    transform: translateY(-1px);
}

.scrape-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
    transform: none;
}

.delete-btn {
    background: rgba(245, 101, 101, 0.8);
    color: white;
    border: 2px solid rgba(245, 101, 101, 0.5);
    padding: 0.5rem 1rem;
    border-radius: 6px;
    cursor: pointer;
    font-weight: 600;
    transition: all 0.3s ease;
}

.delete-btn:hover:not(:disabled) {
    background: rgba(245, 101, 101, 1);
    border-color: rgba(245, 101, 101, 0.7);
    transform: translateY(-1px);
}

.delete-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
    transform: none;
}

/* Main Content */
.main-content {
    max-width: 1800px;
    margin: 0 auto;
    padding: 2rem 1rem;
    min-height: calc(100vh - 80px);
}

/* Shared Card Styles */
.card {
    background: white;
    border-radius: 12px;
    padding: 2rem;
    margin-bottom: 2rem;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
    border: 1px solid #e2e8f0;
    transition: all 0.3s ease;
}

.card:hover {
    box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
    transform: translateY(-2px);
}

.card h2 {
    color: #2d3748;
    margin-bottom: 1.5rem;
    font-size: 1.5rem;
    font-weight: 700;
}

.card h3 {
    color: #4a5568;
    margin-bottom: 1rem;
    font-size: 1.25rem;
    font-weight: 600;
}

/* Form Styles */
.form-group {
    margin-bottom: 1.5rem;
}

.form-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 600;
    color: #4a5568;
}

.form-group input,
.form-group textarea,
.form-group select {
    width: 100%;
    padding: 0.75rem;
    border: 2px solid #e2e8f0;
    border-radius: 8px;
    font-size: 1rem;
    transition:
        border-color 0.3s ease,
        box-shadow 0.3s ease;
}

.form-group input:focus,
.form-group textarea:focus,
.form-group select:focus {
    outline: none;
    border-color: #667eea;
    box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

/* Button Styles */
.btn {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    border: none;
    padding: 0.75rem 1.5rem;
    border-radius: 8px;
    font-size: 1rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.3s ease;
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
}

.btn:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
    transform: none;
}

.btn-secondary {
    background: #718096;
}

.btn-secondary:hover:not(:disabled) {
    background: #4a5568;
    box-shadow: 0 4px 12px rgba(113, 128, 150, 0.4);
}

/* Loading Spinner */
.spinner {
    display: inline-block;
    width: 16px;
    height: 16px;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-radius: 50%;
    border-top-color: white;
    animation: spin 1s ease-in-out infinite;
}

@keyframes spin {
    to {
        transform: rotate(360deg);
    }
}

/* Notification Styles */
.notification {
    position: fixed;
    top: 100px;
    right: 20px;
    padding: 1rem 1.5rem;
    border-radius: 8px;
    color: white;
    font-weight: 600;
    z-index: 1000;
    animation: slideIn 0.3s ease;
}

.notification.success {
    background: #48bb78;
}

.notification.error {
    background: #f56565;
}

@keyframes slideIn {
    from {
        transform: translateX(100%);
        opacity: 0;
    }
    to {
        transform: translateX(0);
        opacity: 1;
    }
}

/* Responsive Design */
@media (max-width: 768px) {
    .nav-container {
        flex-direction: column;
        gap: 1rem;
        padding: 1rem;
    }

    .nav-menu {
        flex-wrap: wrap;
        justify-content: center;
        gap: 1rem;
    }

    .main-content {
        padding: 1rem;
    }

    .card {
        padding: 1.5rem;
    }
}

@media (max-width: 480px) {
    .nav-menu {
        flex-direction: column;
        gap: 0.5rem;
    }

    .nav-link {
        padding: 0.5rem;
        font-size: 0.9rem;
    }
}
</style>
