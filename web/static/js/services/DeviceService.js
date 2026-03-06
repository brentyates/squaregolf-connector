// services/DeviceService.js
export class DeviceService {
    constructor(apiClient, eventBus) {
        this.api = apiClient;
        this.eventBus = eventBus;
        this.deviceStatus = null;
    }

    async connect(deviceName = '') {
        return this.#submitAction({
            url: '/api/device/connect',
            body: { deviceName },
            successEvent: 'device:connecting',
            errorEvent: 'device:error',
            defaultErrorMessage: 'Failed to initiate connection'
        });
    }

    async disconnect() {
        return this.#submitAction({
            url: '/api/device/disconnect',
            successEvent: 'device:disconnecting',
            errorEvent: 'device:error',
            defaultErrorMessage: 'Failed to disconnect'
        });
    }

    updateStatus(status) {
        this.deviceStatus = status;
        this.eventBus.emit('device:status', status);
    }

    getStatus() {
        return this.deviceStatus;
    }

    async #submitAction({ url, body, successEvent, errorEvent, defaultErrorMessage }) {
        try {
            const response = await this.api.post(url, body);

            if (!response.ok) {
                throw new Error(`${defaultErrorMessage}: ${response.statusText}`);
            }

            this.eventBus.emit(successEvent);
            return { success: true };
        } catch (error) {
            this.eventBus.emit(errorEvent, error.message);
            return { success: false, error: error.message };
        }
    }
}
