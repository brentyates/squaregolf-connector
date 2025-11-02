// services/ApiClient.js
export class ApiClient {
    async get(url) {
        return fetch(url);
    }

    async post(url, data = null) {
        const options = {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' }
        };

        if (data) {
            options.body = JSON.stringify(data);
        }

        return fetch(url, options);
    }

    async put(url, data) {
        return fetch(url, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });
    }

    async delete(url) {
        return fetch(url, { method: 'DELETE' });
    }
}
