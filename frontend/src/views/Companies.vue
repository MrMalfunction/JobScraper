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

            <div class="stat-card">
                <div class="stat-icon">🌿</div>
                <div class="stat-content">
                    <h3>{{ greenhouseCompanies }}</h3>
                    <p>Greenhouse Sites</p>
                </div>
            </div>

            <div class="stat-card">
                <div class="stat-icon">☁️</div>
                <div class="stat-content">
                    <h3>{{ oraclecloudCompanies }}</h3>
                    <p>Oracle Cloud Sites</p>
                </div>
            </div>

            <div class="stat-card">
                <div class="stat-icon">🔧</div>
                <div class="stat-content">
                    <h3>{{ genericCompanies }}</h3>
                    <p>Generic API Sites</p>
                </div>
            </div>
        </div>

        <!-- Companies List -->
        <div class="card">
            <div class="section-header">
                <h2>Registered Companies</h2>
                <div class="header-actions">
                    <button @click="showAddModal = true" class="btn btn-primary">
                        ➕ Add Company
                    </button>
                    <button
                        @click="loadCompanies"
                        class="btn btn-secondary"
                        :disabled="isLoadingCompanies"
                    >
                        <span v-if="isLoadingCompanies" class="spinner"></span>
                        Refresh
                    </button>
                </div>
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
                        <div class="detail-row">
                            <div class="detail-badge">
                                <span class="badge-icon">🏷️</span>
                                <span class="badge-text">{{ company.career_site_type }}</span>
                            </div>
                        </div>

                        <div class="detail-row url-row">
                            <div class="url-display">
                                <span class="url-icon">🔗</span>
                                <a
                                    :href="company.base_url"
                                    target="_blank"
                                    class="company-url"
                                    :title="company.base_url"
                                >
                                    {{ truncateUrl(company.base_url) }}
                                </a>
                            </div>
                            <button
                                @click="copyToClipboard(company.base_url)"
                                class="copy-btn"
                                title="Copy URL"
                            >
                                📋
                            </button>
                        </div>
                    </div>

                    <div class="company-actions">
                        <button
                            @click="toggleCompanyStatus(company)"
                            :class="[
                                'action-btn',
                                company.to_scrape ? 'btn-danger' : 'btn-success',
                            ]"
                            :disabled="togglingCompanies[company.name]"
                        >
                            {{ company.to_scrape ? "Disable" : "Enable" }}
                        </button>

                        <button @click="editCompany(company)" class="action-btn btn-secondary">
                            ✏️ Edit
                        </button>

                        <button @click="viewCompanyJobs(company)" class="action-btn btn-secondary">
                            📋 View Jobs
                        </button>

                        <button
                            @click="confirmDeleteCompany(company)"
                            class="action-btn btn-danger-outline"
                            :disabled="isDeletingCompany"
                        >
                            🗑️ Delete
                        </button>
                    </div>
                </div>
            </div>

            <div v-else-if="!isLoadingCompanies" class="no-companies">
                <div class="no-companies-icon">🏢</div>
                <h3>No companies registered</h3>
                <p>Click "Add Company" to start scraping jobs.</p>
            </div>

            <div v-if="isLoadingCompanies" class="loading-container">
                <div class="loading-spinner"></div>
                <p>Loading companies...</p>
            </div>
        </div>

        <!-- Add Company Modal -->
        <div v-if="showAddModal" class="modal-overlay" @click.self="closeModal">
            <div class="modal-content">
                <div class="modal-header">
                    <h2>Add Company</h2>
                    <button @click="closeModal" class="close-btn">✕</button>
                </div>

                <form @submit.prevent="addCompany" class="company-form">
                    <div class="form-group">
                        <label for="companyType">Company Type:</label>
                        <select
                            id="companyType"
                            v-model="companyForm.type"
                            @change="onTypeChange"
                            required
                        >
                            <option value="workday">Workday</option>
                            <option value="greenhouse">Greenhouse</option>
                            <option value="oraclecloud">Oracle Cloud</option>
                            <option value="generic">Generic (Custom API)</option>
                        </select>
                        <small class="form-help">
                            Select the career site platform used by the company
                        </small>
                    </div>

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

                        <div class="form-group" v-if="companyForm.type !== 'oraclecloud'">
                            <label for="baseUrl">Base URL:</label>
                            <input
                                id="baseUrl"
                                v-model="companyForm.baseUrl"
                                type="url"
                                :placeholder="
                                    companyForm.type === 'workday'
                                        ? 'https://company.workdayapp.com'
                                        : 'https://boards-api.greenhouse.io/v1/boards/companyname'
                                "
                                required
                            />
                            <small v-if="companyForm.type === 'greenhouse'" class="form-help">
                                Example:
                                https://boards-api.greenhouse.io/v1/boards/sonyinteractiveentertainmentglobal
                            </small>
                        </div>
                    </div>

                    <!-- Oracle Cloud specific fields -->
                    <div v-if="companyForm.type === 'oraclecloud'" class="form-group">
                        <label for="browserUrl">Browser URL (from Oracle Career Site):</label>
                        <textarea
                            id="browserUrl"
                            v-model="companyForm.browserUrl"
                            rows="4"
                            placeholder="https://jpmc.fa.oraclecloud.com/hcmUI/CandidateExperience/en/sites/CX_1001/jobs?lastSelectedFacet=CATEGORIES&location=United+States&locationId=300000000289738&locationLevel=country&mode=location&selectedCategoriesFacet=300000086152753&selectedPostingDatesFacet=7"
                            :required="companyForm.type === 'oraclecloud'"
                        ></textarea>
                        <div class="form-help warning-box">
                            <strong>⚠️ Important:</strong>
                            <ul style="margin: 0.5rem 0; padding-left: 1.5rem">
                                <li>
                                    Copy the URL from your browser. You can optionally select
                                    category and location filters on the Oracle career site before
                                    copying
                                </li>
                                <li>
                                    The URL must be from an Oracle Cloud career site (contains
                                    <code>/sites/CX_XXXX/jobs</code>)
                                </li>
                                <li>
                                    Categories (<code>selectedCategoriesFacet</code>) are optional -
                                    if provided, multiple categories will be handled automatically
                                </li>
                                <li>
                                    If <code>selectedPostingDatesFacet</code> is missing, it will
                                    default to 7 days
                                </li>
                            </ul>
                        </div>
                    </div>

                    <!-- Workday specific fields -->
                    <div v-if="companyForm.type === 'workday'" class="form-group">
                        <label for="reqBody">Request Body (JSON):</label>
                        <textarea
                            id="reqBody"
                            v-model="companyForm.reqBody"
                            rows="6"
                            placeholder='{"searchText":"","locations":[],"jobFamilies":[],"postedWithin":"","limit":20,"offset":0}'
                            :required="companyForm.type === 'workday'"
                        ></textarea>
                        <small class="form-help">
                            Configure the JSON payload for the Workday API. Adjust parameters like
                            locations, jobFamilies, and search text as needed.
                        </small>
                    </div>

                    <!-- Generic scraper specific fields -->
                    <div v-if="companyForm.type === 'generic'" class="generic-scraper-form">
                        <div class="form-group">
                            <label for="curlCommand">Paste Curl Command (Optional):</label>
                            <textarea
                                id="curlCommand"
                                v-model="companyForm.curlCommand"
                                rows="4"
                                placeholder="curl 'https://api.example.com/jobs' -H 'Content-Type: application/json' -d '{...}'"
                            ></textarea>
                            <button type="button" @click="parseCurl" class="btn btn-secondary btn-small" style="margin-top: 0.5rem;">
                                🔄 Parse Curl
                            </button>
                            <small class="form-help">
                                Paste a curl command to auto-populate fields below, or fill them manually.
                            </small>
                        </div>

                        <div class="form-row">
                            <div class="form-group">
                                <label for="genericMethod">HTTP Method:</label>
                                <select id="genericMethod" v-model="companyForm.genericMethod" required>
                                    <option value="GET">GET</option>
                                    <option value="POST">POST</option>
                                </select>
                            </div>
                            <div class="form-group">
                                <label for="genericUrl">API URL:</label>
                                <input
                                    id="genericUrl"
                                    v-model="companyForm.genericUrl"
                                    type="url"
                                    placeholder="https://api.example.com/jobs"
                                    required
                                />
                            </div>
                        </div>

                        <div class="form-group">
                            <label for="genericHeaders">Headers (JSON):</label>
                            <textarea
                                id="genericHeaders"
                                v-model="companyForm.genericHeaders"
                                rows="4"
                                placeholder='{"Content-Type": "application/json", "Authorization": "Bearer token"}'
                            ></textarea>
                            <small class="form-help">
                                Custom headers as JSON object. Leave empty if not needed.
                            </small>
                        </div>

                        <div class="form-group">
                            <label for="genericBody">Request Body (JSON):</label>
                            <textarea
                                id="genericBody"
                                v-model="companyForm.genericBody"
                                rows="4"
                                placeholder='{"page": 1, "limit": 20}'
                            ></textarea>
                            <small class="form-help">
                                Request body as JSON. For GET requests, this will be converted to query parameters.
                            </small>
                        </div>

                        <div class="form-group">
                            <label for="genericQueryParams">Query Parameters (Optional):</label>
                            <input
                                id="genericQueryParams"
                                v-model="companyForm.genericQueryParams"
                                type="text"
                                placeholder="limit=20&sort=date"
                            />
                            <small class="form-help">
                                Additional query string parameters (without ?).
                            </small>
                        </div>

                        <div class="form-group">
                            <label for="paginationKey">Pagination Key:</label>
                            <input
                                id="paginationKey"
                                v-model="companyForm.paginationKey"
                                type="text"
                                placeholder="page or offset"
                                required
                            />
                            <small class="form-help">
                                The parameter name used for pagination (e.g., "page", "offset", "skip").
                            </small>
                        </div>

                        <div class="form-group">
                            <label>JSON Paths for Data Extraction:</label>
                            <div class="json-paths-grid">
                                <div class="json-path-item">
                                    <label for="responseJsonPath">Jobs Array Path:</label>
                                    <input
                                        id="responseJsonPath"
                                        v-model="companyForm.responseJsonPath"
                                        type="text"
                                        placeholder="data.jobs or jobs"
                                        required
                                    />
                                    <small>Path to the array containing job listings</small>
                                </div>
                                <div class="json-path-item">
                                    <label for="jobIdPath">Job ID Path:</label>
                                    <input
                                        id="jobIdPath"
                                        v-model="companyForm.jobIdPath"
                                        type="text"
                                        placeholder="id or job_id"
                                        required
                                    />
                                </div>
                                <div class="json-path-item">
                                    <label for="jobTitlePath">Job Title Path:</label>
                                    <input
                                        id="jobTitlePath"
                                        v-model="companyForm.jobTitlePath"
                                        type="text"
                                        placeholder="title or name"
                                        required
                                    />
                                </div>
                                <div class="json-path-item">
                                    <label for="jobLinkPath">Job Link Path:</label>
                                    <input
                                        id="jobLinkPath"
                                        v-model="companyForm.jobLinkPath"
                                        type="text"
                                        placeholder="url or link"
                                        required
                                    />
                                    <small>Path to job URL, or first field if using template</small>
                                </div>
                                <div class="json-path-item">
                                    <label for="jobLinkTemplate">Job Link Template (Optional):</label>
                                    <input
                                        id="jobLinkTemplate"
                                        v-model="companyForm.jobLinkTemplate"
                                        type="text"
                                        placeholder="{base_url}{job_path}"
                                    />
                                    <small>Combine fields: {field1}{field2}. Use JSON paths in {}</small>
                                </div>
                                <div class="json-path-item">
                                    <label for="jobDetailsPath">Job Details Path (Optional):</label>
                                    <input
                                        id="jobDetailsPath"
                                        v-model="companyForm.jobDetailsPath"
                                        type="text"
                                        placeholder="description or details[*]"
                                    />
                                    <small>Supports arrays: use [*] for all items or [0] for first</small>
                                </div>
                                <div class="json-path-item">
                                    <label for="jobDatePath">Job Date Path (Optional):</label>
                                    <input
                                        id="jobDatePath"
                                        v-model="companyForm.jobDatePath"
                                        type="text"
                                        placeholder="posted_at or created_date"
                                    />
                                </div>
                            </div>
                            <small class="form-help">
                                <strong>JSON Path Features:</strong><br>
                                • Dot notation for nested fields: "data.jobs", "attributes.title"<br>
                                • Arrays: use [*] for all items, [0] for first: "details[*]", "items[0].text"<br>
                                • Link template (optional): if provided, replaces direct path. Combine fields with {field1}{field2}
                            </small>
                        </div>

                        <div class="form-group">
                            <button type="button" @click="testGenericConfig" class="btn btn-primary btn-full" :disabled="isTestingConfig">
                                <span v-if="isTestingConfig" class="spinner"></span>
                                {{ isTestingConfig ? "Testing..." : "🧪 Test Configuration (Dry Run)" }}
                            </button>
                            <small class="form-help">
                                Test your configuration before saving. This will fetch data and validate JSON paths.
                            </small>
                        </div>

                        <div v-if="dryRunResult" :class="['dry-run-result', dryRunResult.valid ? 'success' : 'error']">
                            <h4>{{ dryRunResult.valid ? "✅ Test Successful" : "❌ Test Failed" }}</h4>
                            <p>{{ dryRunResult.message }}</p>
                            <div v-if="dryRunResult.valid && dryRunResult.sample_data">
                                <p><strong>Found {{ dryRunResult.extracted_jobs }} jobs. Sample data:</strong></p>
                                <pre>{{ JSON.stringify(dryRunResult.sample_data, null, 2) }}</pre>
                            </div>
                            <div v-if="!dryRunResult.valid && dryRunResult.error_details">
                                <p><strong>Error Details:</strong></p>
                                <pre>{{ dryRunResult.error_details }}</pre>
                            </div>
                        </div>
                    </div>

                    <div v-if="addCompanyMessage" :class="['message', addCompanyMessageType]">
                        {{ addCompanyMessage }}
                    </div>

                    <div class="modal-actions">
                        <button type="submit" class="btn btn-primary" :disabled="isAddingCompany">
                            <span v-if="isAddingCompany" class="spinner"></span>
                            {{ isAddingCompany ? "Adding Company..." : "Add Company" }}
                        </button>
                        <button type="button" @click="resetForm" class="btn btn-secondary">
                            Reset Form
                        </button>
                        <button type="button" @click="closeModal" class="btn btn-tertiary">
                            Cancel
                        </button>
                    </div>
                </form>
            </div>
        </div>

        <!-- Edit Company Modal -->
        <div v-if="showEditModal" class="modal-overlay" @click.self="closeEditModal">
            <div class="modal-content edit-modal">
                <div class="modal-header">
                    <div class="modal-title">
                        <span class="modal-icon">✏️</span>
                        <h2>Edit Company</h2>
                    </div>
                    <button @click="closeEditModal" class="close-btn">✕</button>
                </div>

                <form @submit.prevent="updateCompany" class="company-form compact-form">
                    <div class="form-row-compact">
                        <div class="form-group">
                            <label for="editCompanyName">
                                <span class="label-icon">🏢</span>
                                Company Name
                            </label>
                            <input
                                id="editCompanyName"
                                v-model="editForm.name"
                                type="text"
                                placeholder="e.g., Google"
                                required
                            />
                        </div>

                        <div class="form-group">
                            <label for="editCompanyType">
                                <span class="label-icon">🏷️</span>
                                Company Type
                            </label>
                            <select
                                id="editCompanyType"
                                v-model="editForm.career_site_type"
                                disabled
                            >
                                <option value="workday">Workday</option>
                                <option value="greenhouse">Greenhouse</option>
                                <option value="oraclecloud">Oracle Cloud</option>
                            </select>
                            <small class="form-help info-help"> ℹ️ Type cannot be changed </small>
                        </div>
                    </div>

                    <div class="form-group">
                        <label for="editBaseUrl">
                            <span class="label-icon">🔗</span>
                            Base URL
                        </label>
                        <textarea
                            id="editBaseUrl"
                            v-model="editForm.base_url"
                            rows="2"
                            required
                            class="url-textarea"
                        ></textarea>
                    </div>

                    <div v-if="editForm.career_site_type === 'workday'" class="form-group">
                        <label for="editReqBody">
                            <span class="label-icon">📄</span>
                            Request Body (JSON)
                        </label>
                        <textarea
                            id="editReqBody"
                            v-model="editForm.api_request_body"
                            rows="6"
                            class="json-textarea"
                            placeholder='{"searchText":"","locations":[],"jobFamilies":[]}'
                        ></textarea>
                    </div>

                    <div class="form-group checkbox-group-compact">
                        <label class="checkbox-label">
                            <input type="checkbox" v-model="editForm.to_scrape" />
                            <span class="checkbox-text">
                                <strong>Enable scraping for this company</strong>
                            </span>
                        </label>
                    </div>

                    <div v-if="editCompanyMessage" :class="['message', editCompanyMessageType]">
                        {{ editCompanyMessage }}
                    </div>

                    <div class="modal-actions">
                        <button type="submit" class="btn btn-primary" :disabled="isEditingCompany">
                            <span v-if="isEditingCompany" class="spinner"></span>
                            {{ isEditingCompany ? "Updating..." : "💾 Update Company" }}
                        </button>
                        <button type="button" @click="closeEditModal" class="btn btn-tertiary">
                            Cancel
                        </button>
                    </div>
                </form>
            </div>
        </div>

        <!-- Delete Confirmation Modal -->
        <div v-if="showDeleteModal" class="modal-overlay" @click.self="showDeleteModal = false">
            <div class="modal-content modal-small">
                <div class="modal-header">
                    <h2>⚠️ Confirm Delete</h2>
                    <button @click="showDeleteModal = false" class="close-btn">✕</button>
                </div>

                <div class="delete-warning">
                    <p>
                        Are you sure you want to delete
                        <strong>{{ companyToDelete?.name }}</strong
                        >?
                    </p>
                    <p class="warning-text">
                        This will permanently delete the company and all associated jobs. This
                        action cannot be undone.
                    </p>
                </div>

                <div class="modal-actions">
                    <button
                        @click="deleteCompany"
                        class="btn btn-danger"
                        :disabled="isDeletingCompany"
                    >
                        <span v-if="isDeletingCompany" class="spinner"></span>
                        {{ isDeletingCompany ? "Deleting..." : "Yes, Delete" }}
                    </button>
                    <button
                        @click="showDeleteModal = false"
                        class="btn btn-tertiary"
                        :disabled="isDeletingCompany"
                    >
                        Cancel
                    </button>
                </div>
            </div>
        </div>

        <!-- Component Notification -->
        <div v-if="notification" :class="['notification', notification.type]">
            {{ notification.message }}
        </div>
    </div>
</template>

<script>
import axios from "axios";
import { parseCurlCommand } from "../composables/useCurlParser";

export default {
    name: "Companies",
    data() {
        return {
            companies: [],
            companyForm: {
                type: "workday",
                name: "",
                baseUrl: "",
                browserUrl: "",
                reqBody:
                    '{"searchText":"","locations":[],"jobFamilies":[],"postedWithin":"","limit":20,"offset":0}',
                // Generic scraper fields
                curlCommand: "",
                genericMethod: "GET",
                genericUrl: "",
                genericHeaders: "{}",
                genericBody: "{}",
                genericQueryParams: "",
                paginationKey: "",
                responseJsonPath: "",
                jobIdPath: "",
                jobTitlePath: "",
                jobLinkPath: "",
                jobLinkTemplate: "",
                jobDetailsPath: "",
                jobDatePath: "",
            },
            isLoadingCompanies: false,
            isAddingCompany: false,
            togglingCompanies: {},
            isDeletingCompany: false,
            isEditingCompany: false,
            isTestingConfig: false,
            addCompanyMessage: "",
            addCompanyMessageType: "",
            editCompanyMessage: "",
            editCompanyMessageType: "",
            showAddModal: false,
            showEditModal: false,
            showDeleteModal: false,
            companyToDelete: null,
            notification: null,
            dryRunResult: null,
            editForm: {
                originalName: "",
                name: "",
                base_url: "",
                career_site_type: "",
                api_request_body: "",
                to_scrape: true,
            },
        };
    },

    computed: {
        activeCompanies() {
            return this.companies.filter((c) => c.to_scrape).length;
        },

        inactiveCompanies() {
            return this.companies.filter((c) => !c.to_scrape).length;
        },

        workdayCompanies() {
            return this.companies.filter((c) => c.career_site_type === "workday").length;
        },

        greenhouseCompanies() {
            return this.companies.filter((c) => c.career_site_type === "greenhouse").length;
        },

        oraclecloudCompanies() {
            return this.companies.filter((c) => c.career_site_type === "oraclecloud").length;
        },

        genericCompanies() {
            return this.companies.filter((c) => c.career_site_type === "generic").length;
        },
    },

    async mounted() {
        await this.loadCompanies();
    },

    methods: {
        async loadCompanies() {
            this.isLoadingCompanies = true;

            try {
                const response = await axios.get("/api/companies");
                this.companies = response.data.data;
            } catch (error) {
                console.error("Error loading companies:", error);
                this.companies = [];
            }

            this.isLoadingCompanies = false;
        },

        async addCompany() {
            this.isAddingCompany = true;
            this.addCompanyMessage = "";

            try {
                let response;

                if (this.companyForm.type === "workday") {
                    // Validate JSON for Workday
                    const reqBodyObj = JSON.parse(this.companyForm.reqBody);

                    response = await axios.post("/add_scrape_company/workday", {
                        name: this.companyForm.name,
                        base_url: this.companyForm.baseUrl,
                        req_body: reqBodyObj,
                    });
                } else if (this.companyForm.type === "greenhouse") {
                    response = await axios.post("/add_scrape_company/greenhouse", {
                        name: this.companyForm.name,
                        base_url: this.companyForm.baseUrl,
                    });
                } else if (this.companyForm.type === "oraclecloud") {
                    response = await axios.post("/add_scrape_company/oraclecloud", {
                        name: this.companyForm.name,
                        browser_url: this.companyForm.browserUrl,
                    });
                } else if (this.companyForm.type === "generic") {
                    // Parse headers and body
                    const headers = JSON.parse(this.companyForm.genericHeaders || "{}");
                    const body = this.companyForm.genericBody ? JSON.parse(this.companyForm.genericBody) : {};

                    response = await axios.post("/add_scrape_company/generic", {
                        name: this.companyForm.name,
                        base_url: this.companyForm.genericUrl,
                        method: this.companyForm.genericMethod,
                        headers: headers,
                        body: body,
                        query_params: this.companyForm.genericQueryParams,
                        pagination_key: this.companyForm.paginationKey,
                        response_json_path: this.companyForm.responseJsonPath,
                        job_id_json_path: this.companyForm.jobIdPath,
                        job_title_json_path: this.companyForm.jobTitlePath,
                        job_details_json_path: this.companyForm.jobDetailsPath,
                        job_link_json_path: this.companyForm.jobLinkPath,
                        job_link_template: this.companyForm.jobLinkTemplate,
                        job_date_json_path: this.companyForm.jobDatePath,
                        dry_run: false,
                    });
                }

                this.addCompanyMessage = response.data.message;
                this.addCompanyMessageType = "success";

                // Refresh companies list
                await this.loadCompanies();

                // Close modal and reset form after a short delay
                setTimeout(() => {
                    this.closeModal();
                }, 1500);
            } catch (error) {
                console.error("Error adding company:", error);
                this.addCompanyMessage = error.response?.data?.message || "Failed to add company";
                this.addCompanyMessageType = "error";
            }

            this.isAddingCompany = false;
        },

        parseCurl() {
            if (!this.companyForm.curlCommand) {
                alert("Please paste a curl command first");
                return;
            }

            const result = parseCurlCommand(this.companyForm.curlCommand);
            
            if (result.success) {
                this.companyForm.genericUrl = result.url;
                this.companyForm.genericMethod = result.method;
                this.companyForm.genericHeaders = JSON.stringify(result.headers, null, 2);
                if (result.body) {
                    this.companyForm.genericBody = typeof result.body === 'string' 
                        ? result.body 
                        : JSON.stringify(result.body, null, 2);
                }
                this.companyForm.genericQueryParams = result.queryParams;
                
                this.addCompanyMessage = "Curl command parsed successfully! Please review and fill in the JSON paths below.";
                this.addCompanyMessageType = "success";
            } else {
                this.addCompanyMessage = result.error;
                this.addCompanyMessageType = "error";
            }
        },

        async testGenericConfig() {
            this.isTestingConfig = true;
            this.dryRunResult = null;

            try {
                const headers = JSON.parse(this.companyForm.genericHeaders || "{}");
                const body = this.companyForm.genericBody ? JSON.parse(this.companyForm.genericBody) : {};

                const response = await axios.post("/add_scrape_company/generic", {
                    name: this.companyForm.name || "Test Company",
                    base_url: this.companyForm.genericUrl,
                    method: this.companyForm.genericMethod,
                    headers: headers,
                    body: body,
                    query_params: this.companyForm.genericQueryParams,
                    pagination_key: this.companyForm.paginationKey,
                    response_json_path: this.companyForm.responseJsonPath,
                    job_id_json_path: this.companyForm.jobIdPath,
                    job_title_json_path: this.companyForm.jobTitlePath,
                    job_details_json_path: this.companyForm.jobDetailsPath,
                    job_link_json_path: this.companyForm.jobLinkPath,
                    job_link_template: this.companyForm.jobLinkTemplate,
                    job_date_json_path: this.companyForm.jobDatePath,
                    dry_run: true,
                });

                this.dryRunResult = response.data;
            } catch (error) {
                console.error("Error testing configuration:", error);
                this.dryRunResult = {
                    valid: false,
                    message: error.response?.data?.message || "Failed to test configuration",
                    error_details: error.message,
                };
            }

            this.isTestingConfig = false;
        },

        resetForm() {
            this.companyForm.type = "workday";
            this.companyForm.name = "";
            this.companyForm.baseUrl = "";
            this.companyForm.browserUrl = "";
            this.companyForm.reqBody =
                '{"searchText":"","locations":[],"jobFamilies":[],"postedWithin":"","limit":20,"offset":0}';
            this.companyForm.curlCommand = "";
            this.companyForm.genericMethod = "GET";
            this.companyForm.genericUrl = "";
            this.companyForm.genericHeaders = "{}";
            this.companyForm.genericBody = "{}";
            this.companyForm.genericQueryParams = "";
            this.companyForm.paginationKey = "";
            this.companyForm.responseJsonPath = "";
            this.companyForm.jobIdPath = "";
            this.companyForm.jobTitlePath = "";
            this.companyForm.jobLinkPath = "";
            this.companyForm.jobLinkTemplate = "";
            this.companyForm.jobDetailsPath = "";
            this.companyForm.jobDatePath = "";
            this.dryRunResult = null;
            this.addCompanyMessage = "";
        },

        onTypeChange() {
            // Clear form fields when type changes
            this.companyForm.name = "";
            this.companyForm.baseUrl = "";
            this.companyForm.browserUrl = "";
            this.addCompanyMessage = "";
        },

        closeModal() {
            this.showAddModal = false;
            this.resetForm();
        },

        async toggleCompanyStatus(company) {
            this.togglingCompanies[company.name] = true;

            try {
                const newStatus = !company.to_scrape;
                await axios.put(`/api/companies/${encodeURIComponent(company.name)}`, {
                    to_scrape: newStatus,
                });

                // Update local state - find and update the company in the array
                const companyIndex = this.companies.findIndex((c) => c.name === company.name);
                if (companyIndex !== -1) {
                    this.companies[companyIndex].to_scrape = newStatus;
                }

                this.showNotification(
                    "success",
                    `Company ${company.name} ${newStatus ? "enabled" : "disabled"}`,
                );
            } catch (error) {
                console.error("Error toggling company status:", error);
                this.showNotification(
                    "error",
                    error.response?.data?.message || "Failed to update company status",
                );
            } finally {
                this.togglingCompanies[company.name] = false;
            }
        },

        editCompany(company) {
            this.editForm.originalName = company.name;
            this.editForm.name = company.name;
            this.editForm.base_url = company.base_url;
            this.editForm.career_site_type = company.career_site_type;
            this.editForm.api_request_body = company.api_request_body || "";
            this.editForm.to_scrape = company.to_scrape;
            this.editCompanyMessage = "";
            this.showEditModal = true;
        },

        closeEditModal() {
            this.showEditModal = false;
            this.editForm = {
                originalName: "",
                name: "",
                base_url: "",
                career_site_type: "",
                api_request_body: "",
                to_scrape: true,
            };
            this.editCompanyMessage = "";
        },

        async updateCompany() {
            this.isEditingCompany = true;
            this.editCompanyMessage = "";

            try {
                const updateData = {
                    name:
                        this.editForm.name !== this.editForm.originalName
                            ? this.editForm.name
                            : undefined,
                    base_url: this.editForm.base_url,
                    to_scrape: this.editForm.to_scrape,
                };

                // Add api_request_body for Workday companies
                if (
                    this.editForm.career_site_type === "workday" &&
                    this.editForm.api_request_body
                ) {
                    try {
                        updateData.api_request_body = JSON.parse(this.editForm.api_request_body);
                    } catch (e) {
                        this.editCompanyMessage = "Invalid JSON in Request Body";
                        this.editCompanyMessageType = "error";
                        this.isEditingCompany = false;
                        return;
                    }
                }

                const response = await axios.put(
                    `/api/companies/${encodeURIComponent(this.editForm.originalName)}`,
                    updateData,
                );

                this.editCompanyMessage = response.data.message;
                this.editCompanyMessageType = "success";

                // Refresh companies list
                await this.loadCompanies();

                // Close modal after a short delay
                setTimeout(() => {
                    this.closeEditModal();
                }, 1500);
            } catch (error) {
                console.error("Error updating company:", error);
                this.editCompanyMessage =
                    error.response?.data?.message || "Failed to update company";
                this.editCompanyMessageType = "error";
            }

            this.isEditingCompany = false;
        },

        confirmDeleteCompany(company) {
            this.companyToDelete = company;
            this.showDeleteModal = true;
        },

        async deleteCompany() {
            if (!this.companyToDelete) return;

            this.isDeletingCompany = true;

            try {
                const response = await axios.delete(
                    `/api/companies/${encodeURIComponent(this.companyToDelete.name)}`,
                );

                this.showNotification("success", response.data.message);

                // Refresh companies list
                await this.loadCompanies();

                // Close modal after successful delete and refresh
                this.showDeleteModal = false;
                this.companyToDelete = null;
                this.isDeletingCompany = false;
            } catch (error) {
                console.error("Error deleting company:", error);
                this.showNotification(
                    "error",
                    error.response?.data?.message || "Failed to delete company",
                );
                this.isDeletingCompany = false;
            }
        },

        viewCompanyJobs(company) {
            // Navigate to all jobs with company filter
            this.$router.push({
                name: "AllJobs",
                query: { company: company.name },
            });
        },

        showNotification(type, message) {
            this.notification = { type, message };
            setTimeout(() => {
                this.notification = null;
            }, 5000);
        },

        truncateUrl(url) {
            if (!url) return "";
            const maxLength = 80;
            if (url.length <= maxLength) return url;
            return url.substring(0, maxLength) + "...";
        },

        copyToClipboard(text) {
            if (navigator.clipboard && navigator.clipboard.writeText) {
                navigator.clipboard
                    .writeText(text)
                    .then(() => {
                        this.showNotification("success", "URL copied to clipboard!");
                    })
                    .catch((err) => {
                        console.error("Failed to copy:", err);
                        this.showNotification("error", "Failed to copy URL");
                    });
            } else {
                // Fallback for older browsers
                const textArea = document.createElement("textarea");
                textArea.value = text;
                textArea.style.position = "fixed";
                textArea.style.left = "-999999px";
                document.body.appendChild(textArea);
                textArea.select();
                try {
                    document.execCommand("copy");
                    this.showNotification("success", "URL copied to clipboard!");
                } catch (err) {
                    console.error("Failed to copy:", err);
                    this.showNotification("error", "Failed to copy URL");
                }
                document.body.removeChild(textArea);
            }
        },
    },
};
</script>

<style scoped>
.companies {
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
    margin-bottom: 2rem;
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

.section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
    flex-wrap: wrap;
    gap: 1rem;
}

.header-actions {
    display: flex;
    gap: 0.75rem;
    flex-wrap: wrap;
}

.btn-primary {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    border: none;
    padding: 0.75rem 1.5rem;
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.3s ease;
}

.btn-primary:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 20px rgba(102, 126, 234, 0.4);
}

.btn-primary:disabled {
    opacity: 0.6;
    cursor: not-allowed;
    transform: none;
}

.btn-secondary {
    background: #718096;
    color: white;
    border: none;
    padding: 0.75rem 1.5rem;
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.3s ease;
}

.btn-secondary:hover {
    background: #4a5568;
}

.btn-secondary:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

.spinner {
    display: inline-block;
    width: 14px;
    height: 14px;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-radius: 50%;
    border-top-color: white;
    animation: spin 0.8s ease-in-out infinite;
    margin-right: 0.5rem;
    vertical-align: middle;
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
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 4px;
    background: linear-gradient(135deg, #48bb78 0%, #38a169 100%);
}

.company-card:not(.active)::before {
    content: "";
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
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}

.detail-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.detail-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    background: #f7fafc;
    padding: 0.4rem 0.8rem;
    border-radius: 6px;
    border: 1px solid #e2e8f0;
    font-size: 0.9rem;
}

.badge-icon {
    font-size: 1rem;
}

.badge-text {
    font-weight: 600;
    color: #4a5568;
    text-transform: capitalize;
}

.url-row {
    background: #f7fafc;
    padding: 0.75rem;
    border-radius: 8px;
    border: 1px solid #e2e8f0;
    display: flex;
    align-items: center;
    gap: 0.75rem;
}

.url-display {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex: 1;
    overflow: hidden;
}

.url-icon {
    font-size: 1rem;
    flex-shrink: 0;
}

.company-url {
    color: #667eea;
    text-decoration: none;
    font-size: 0.85rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-family: "Courier New", monospace;
}

.company-url:hover {
    text-decoration: underline;
}

.copy-btn {
    background: white;
    border: 1px solid #cbd5e0;
    padding: 0.4rem 0.6rem;
    border-radius: 6px;
    cursor: pointer;
    font-size: 1rem;
    transition: all 0.2s ease;
    flex-shrink: 0;
}

.copy-btn:hover {
    background: #e2e8f0;
    border-color: #a0aec0;
    transform: scale(1.05);
}

.copy-btn:active {
    transform: scale(0.95);
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

.btn-danger-outline {
    background: white;
    color: #f56565;
    border: 2px solid #f56565;
}

.btn-danger-outline:hover {
    background: #f56565;
    color: white;
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
    to {
        transform: rotate(360deg);
    }
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

/* Modal Styles */
.modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
    padding: 2rem;
    animation: fadeIn 0.3s ease;
}

.modal-content {
    background: white;
    border-radius: 16px;
    max-width: 700px;
    width: 100%;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
    animation: slideUp 0.3s ease;
}

@keyframes slideUp {
    from {
        opacity: 0;
        transform: translateY(30px);
    }
    to {
        opacity: 1;
        transform: translateY(0);
    }
}

.modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 2rem 2rem 1rem 2rem;
    border-bottom: 2px solid #e2e8f0;
}

.modal-header h2 {
    margin: 0;
    font-size: 1.5rem;
    color: #2d3748;
    font-weight: 700;
}

.close-btn {
    background: none;
    border: none;
    font-size: 1.5rem;
    color: #718096;
    cursor: pointer;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 6px;
    transition: all 0.3s ease;
}

.close-btn:hover {
    background: #f7fafc;
    color: #2d3748;
}

.company-form {
    padding: 2rem;
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

.modal-actions {
    display: flex;
    gap: 1rem;
    margin-top: 2rem;
    flex-wrap: wrap;
}

.btn-tertiary {
    background: #f7fafc;
    color: #4a5568;
    border: 2px solid #e2e8f0;
    padding: 0.75rem 1.5rem;
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.3s ease;
}

.btn-tertiary:hover {
    background: #e2e8f0;
    border-color: #cbd5e0;
}

.message {
    margin: 1rem 0;
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

.warning-box {
    background: #fffbeb;
    border: 1px solid #fbbf24;
    border-radius: 8px;
    padding: 1rem;
    margin-top: 0.5rem;
}

.warning-box ul {
    font-size: 0.9rem;
    line-height: 1.6;
}

.delete-warning {
    padding: 2rem 2rem 1rem 2rem;
}

.delete-warning p {
    margin-bottom: 1rem;
    line-height: 1.6;
    font-size: 1rem;
    color: #2d3748;
}

.delete-warning p:last-child {
    margin-bottom: 0;
}

.warning-text {
    color: #c53030;
    font-weight: 500;
    background: #fed7d7;
    padding: 1rem;
    border-radius: 8px;
    border: 1px solid #feb2b2;
    font-size: 0.95rem;
}

.modal-small {
    max-width: 500px;
}

.modal-small .modal-actions {
    padding: 0 2rem 2rem 2rem;
    margin-top: 0;
}

/* Edit Modal Specific Styles */
.edit-modal {
    max-width: 650px;
}

.modal-title {
    display: flex;
    align-items: center;
    gap: 0.75rem;
}

.modal-icon {
    font-size: 1.75rem;
}

.compact-form {
    padding: 1.5rem 2rem;
}

.form-row-compact {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
    margin-bottom: 1rem;
}

.label-icon {
    font-size: 1.1rem;
    margin-right: 0.25rem;
}

.form-group label {
    display: flex;
    align-items: center;
    font-size: 0.95rem;
}

.info-help {
    background: #ebf8ff;
    color: #2c5282;
    padding: 0.35rem 0.6rem;
    border-radius: 4px;
    border: 1px solid #bee3f8;
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    font-size: 0.8rem;
    margin-top: 0.25rem;
}

.url-textarea {
    font-family: "Courier New", monospace;
    font-size: 0.85rem;
    resize: vertical;
    min-height: 50px;
}

.json-textarea {
    font-family: "Courier New", monospace;
    font-size: 0.85rem;
    resize: vertical;
    background: #f7fafc;
}

.checkbox-group-compact {
    background: #f7fafc;
    padding: 0.75rem 1rem;
    border-radius: 8px;
    border: 1px solid #e2e8f0;
    margin-top: 0.5rem;
}

.checkbox-label {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    cursor: pointer;
    font-weight: normal;
}

.checkbox-label input[type="checkbox"] {
    width: 18px;
    height: 18px;
    cursor: pointer;
}

.checkbox-text {
    display: flex;
    flex-direction: column;
}

.checkbox-text strong {
    color: #2d3748;
    font-size: 0.95rem;
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

.warning-box code {
    background: #fef3c7;
    padding: 0.2rem 0.4rem;
    border-radius: 4px;
    font-family: monospace;
    font-size: 0.85rem;
    color: #92400e;
}

@media (max-width: 768px) {
    .stats-grid {
        grid-template-columns: repeat(2, 1fr);
    }

    .form-row,
    .form-row-compact {
        grid-template-columns: 1fr;
    }

    .companies-grid {
        grid-template-columns: 1fr;
    }

    .section-header {
        flex-direction: column;
        align-items: stretch;
    }

    .header-actions {
        flex-direction: column;
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

    .modal-content {
        max-height: 95vh;
    }

    .modal-overlay {
        padding: 1rem;
    }
}

@media (max-width: 480px) {
    .stats-grid {
        grid-template-columns: 1fr;
    }

    .modal-actions {
        flex-direction: column;
    }

    .company-actions {
        flex-direction: column;
    }

    .modal-header {
        padding: 1.5rem;
    }

    .company-form {
        padding: 1.5rem;
    }
}

/* Generic Scraper Form Styles */
.generic-scraper-form {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    margin-top: 1rem;
}

.json-paths-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
    margin-top: 0.5rem;
}

.json-path-item {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}

.json-path-item label {
    font-size: 0.9rem;
    font-weight: 500;
    color: #4a5568;
}

.json-path-item input {
    padding: 0.5rem;
    border: 1px solid #cbd5e0;
    border-radius: 6px;
    font-size: 0.9rem;
}

.json-path-item small {
    font-size: 0.75rem;
    color: #718096;
}

.btn-small {
    padding: 0.5rem 1rem;
    font-size: 0.9rem;
}

.btn-full {
    width: 100%;
}

.dry-run-result {
    padding: 1rem;
    border-radius: 8px;
    margin-top: 1rem;
}

.dry-run-result.success {
    background: #f0fdf4;
    border: 1px solid #86efac;
}

.dry-run-result.error {
    background: #fef2f2;
    border: 1px solid #fca5a5;
}

.dry-run-result h4 {
    margin: 0 0 0.5rem 0;
    font-size: 1rem;
}

.dry-run-result p {
    margin: 0.5rem 0;
    font-size: 0.9rem;
}

.dry-run-result pre {
    background: #1a202c;
    color: #e2e8f0;
    padding: 1rem;
    border-radius: 6px;
    overflow-x: auto;
    font-size: 0.85rem;
    margin-top: 0.5rem;
}

@media (max-width: 768px) {
    .json-paths-grid {
        grid-template-columns: 1fr;
    }
}
</style>
