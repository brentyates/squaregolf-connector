// services/InfiniteTeesService.js
export class InfiniteTeesService {
    #errorEvent = 'infinitetees:error';
    #statusEvent = 'infinitetees:status';

    constructor(apiClient, eventBus) {
        this.api = apiClient;
        this.eventBus = eventBus;
        this.status = null;
    }

    async connect(ip, port) {
        if (!ip || !port) {
            this.eventBus.emit(this.#errorEvent, 'Please enter valid IP and port');
            return { success: false };
        }

        return this.#submitAction({
            url: '/api/infinitetees/connect',
            body: { ip, port },
            successEvent: 'infinitetees:connecting',
            defaultErrorMessage: 'Failed to connect'
        });
    }

    async disconnect() {
        return this.#submitAction({
            url: '/api/infinitetees/disconnect',
            successEvent: 'infinitetees:disconnecting',
            defaultErrorMessage: 'Failed to disconnect'
        });
    }

    async saveConfig(ip, port, autoConnect) {
        return this.#submitAction({
            url: '/api/infinitetees/config',
            body: { ip, port, autoConnect },
            defaultErrorMessage: 'Failed to save config'
        });
    }

    updateStatus(status) {
        this.status = status;
        this.eventBus.emit(this.#statusEvent, status);
    }

    async #submitAction({ url, body, successEvent, defaultErrorMessage }) {
        try {
            const response = await this.api.post(url, body);

            if (!response.ok) {
                throw new Error(`${defaultErrorMessage}: ${response.statusText}`);
            }

            if (successEvent) {
                this.eventBus.emit(successEvent);
            }

            return { success: true };
        } catch (error) {
            this.eventBus.emit(this.#errorEvent, error.message);
            return { success: false, error: error.message };
        }
    }
}
