/**
 * Utility to parse curl commands and extract URL, method, headers, and body
 */
export function parseCurlCommand(curlCommand) {
    try {
        // Remove line breaks and extra spaces
        let cmd = curlCommand.trim().replace(/\\\n/g, ' ').replace(/\s+/g, ' ');

        // Extract URL
        let url = '';
        const urlMatch = cmd.match(/curl\s+(?:'([^']+)'|"([^"]+)"|([^\s]+))/);
        if (urlMatch) {
            url = urlMatch[1] || urlMatch[2] || urlMatch[3];
        }

        // If URL not found in first position, try to find it after flags
        if (!url || url.startsWith('-')) {
            const urlPattern = /(?:--url\s+|'|")?(https?:\/\/[^\s'"]+)/i;
            const match = cmd.match(urlPattern);
            if (match) {
                url = match[1];
            }
        }

        // Extract method
        let method = 'GET';
        const methodMatch = cmd.match(/-X\s+(['"]?)(\w+)\1/i) || cmd.match(/--request\s+(['"]?)(\w+)\1/i);
        if (methodMatch) {
            method = methodMatch[2].toUpperCase();
        }

        // Extract headers
        const headers = {};
        const headerRegex = /-H\s+['"]([^:]+):\s*([^'"]+)['"]/g;
        let headerMatch;
        while ((headerMatch = headerRegex.exec(cmd)) !== null) {
            const key = headerMatch[1].trim();
            const value = headerMatch[2].trim();
            headers[key] = value;
        }

        // Also try --header format
        const headerRegex2 = /--header\s+['"]([^:]+):\s*([^'"]+)['"]/g;
        while ((headerMatch = headerRegex2.exec(cmd)) !== null) {
            const key = headerMatch[1].trim();
            const value = headerMatch[2].trim();
            headers[key] = value;
        }

        // Extract body/data
        let body = null;
        const bodyMatch = cmd.match(/(?:-d|--data|--data-raw|--data-binary)\s+['"](.+?)['"]/s);
        if (bodyMatch) {
            try {
                body = JSON.parse(bodyMatch[1]);
            } catch (e) {
                // If not JSON, try to parse as form data
                body = bodyMatch[1];
            }
        }

        // Extract query parameters from URL
        let baseUrl = url;
        let queryParams = '';
        const queryIndex = url.indexOf('?');
        if (queryIndex > -1) {
            baseUrl = url.substring(0, queryIndex);
            queryParams = url.substring(queryIndex + 1);
        }

        return {
            success: true,
            url: baseUrl,
            fullUrl: url,
            method: method,
            headers: headers,
            body: body,
            queryParams: queryParams,
        };
    } catch (error) {
        return {
            success: false,
            error: 'Failed to parse curl command: ' + error.message,
        };
    }
}

/**
 * Example usage:
 * const result = parseCurlCommand(`
 *   curl 'https://api.example.com/jobs?limit=20' \
 *   -H 'Content-Type: application/json' \
 *   -H 'Authorization: Bearer token123' \
 *   -X POST \
 *   -d '{"page":1,"filters":{"active":true}}'
 * `);
 */
