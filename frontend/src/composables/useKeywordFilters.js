// job-scraper/frontend/src/composables/useKeywordFilters.js
import { ref, computed } from 'vue'

export function useKeywordFilters() {
    const includeKeywords = ref('')
    const excludeKeywords = ref('')

    const hasActiveKeywords = computed(() => {
        return includeKeywords.value.trim() || excludeKeywords.value.trim()
    })

    const buildKeywordParams = () => {
        const params = {}

        // Add include keywords
        if (includeKeywords.value) {
            const keywords = includeKeywords.value
                .split(',')
                .map(k => k.trim())
                .filter(k => k)
            keywords.forEach(keyword => {
                params.include_keywords = keyword
            })
        }

        // Add exclude keywords
        if (excludeKeywords.value) {
            const keywords = excludeKeywords.value
                .split(',')
                .map(k => k.trim())
                .filter(k => k)
            keywords.forEach(keyword => {
                params.exclude_keywords = keyword
            })
        }

        return params
    }

    const clearKeywords = () => {
        includeKeywords.value = ''
        excludeKeywords.value = ''
    }

    return {
        includeKeywords,
        excludeKeywords,
        hasActiveKeywords,
        buildKeywordParams,
        clearKeywords
    }
}
