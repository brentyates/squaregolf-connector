// services/ApiClient.js
export class ApiClient {
    async get(url, options = {}) {
        return this.#request(url, options);
    }

    async post(url, data = null, options = {}) {
        return this.#request(url, {
            ...options,
            method: 'POST',
            body: data
        });
    }

    async put(url, data, options = {}) {
        return this.#request(url, {
            ...options,
            method: 'PUT',
            body: data
        });
    }

    async delete(url, options = {}) {
        return this.#request(url, {
            ...options,
            method: 'DELETE'
        });
    }

    async #request(url, options = {}) {
        const {
            headers,
            body,
            ...fetchOptions
        } = options;

        const requestHeaders = new Headers(headers ?? {});
        let requestBody = body;

        if (body !== null && body !== undefined && !(body instanceof FormData) && !(body instanceof Blob)) {
            if (!requestHeaders.has('Content-Type')) {
                requestHeaders.set('Content-Type', 'application/json');
            }
            requestBody = JSON.stringify(body);
        }

        return fetch(url, {
            ...fetchOptions,
            headers: requestHeaders,
            body: requestBody
        });
    }
}
