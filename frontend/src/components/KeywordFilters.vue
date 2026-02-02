job-scraper/frontend/src/components/KeywordFilters.vue
<template>
    <div class="keyword-filters">
        <div class="form-group">
            <label for="includeKeywords">
                <span class="label-icon">✅</span>
                Include Keywords
            </label>
            <input
                id="includeKeywords"
                v-model="includeKeywords"
                type="text"
                placeholder="e.g. go, python, remote (comma-separated)"
            />
            <span class="helper-text">Jobs must contain at least one of these keywords</span>
        </div>

        <div class="form-group">
            <label for="excludeKeywords">
                <span class="label-icon">❌</span>
                Exclude Keywords
            </label>
            <input
                id="excludeKeywords"
                v-model="excludeKeywords"
                type="text"
                placeholder="e.g. senior, manager (comma-separated)"
            />
            <span class="helper-text">Jobs must NOT contain any of these keywords</span>
        </div>
    </div>
</template>

<script>
export default {
    name: "KeywordFilters",
    props: {
        modelValue: {
            type: Object,
            default: () => ({
                includeKeywords: "",
                excludeKeywords: "",
            }),
        },
    },
    computed: {
        includeKeywords: {
            get() {
                return this.modelValue.includeKeywords || "";
            },
            set(value) {
                this.$emit("update:modelValue", {
                    ...this.modelValue,
                    includeKeywords: value,
                });
            },
        },
        excludeKeywords: {
            get() {
                return this.modelValue.excludeKeywords || "";
            },
            set(value) {
                this.$emit("update:modelValue", {
                    ...this.modelValue,
                    excludeKeywords: value,
                });
            },
        },
    },
};
</script>

<style scoped>
.keyword-filters {
    display: contents;
}

.form-group {
    margin-bottom: 0;
}

.form-group label {
    display: flex;
    align-items: center;
    margin-bottom: 0.5rem;
    font-weight: 600;
    color: #4a5568;
    font-size: 0.95rem;
}

.label-icon {
    margin-right: 0.5rem;
    font-size: 1.1rem;
}

.form-group input {
    width: 100%;
    padding: 0.75rem;
    border: 2px solid #e2e8f0;
    border-radius: 8px;
    font-size: 0.95rem;
    transition:
        border-color 0.3s ease,
        box-shadow 0.3s ease;
    background: white;
}

.form-group input:focus {
    outline: none;
    border-color: #667eea;
    box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.form-group input::placeholder {
    color: #a0aec0;
    font-size: 0.9rem;
}

.helper-text {
    display: block;
    margin-top: 0.375rem;
    font-size: 0.8rem;
    color: #718096;
    font-style: italic;
}
</style>
