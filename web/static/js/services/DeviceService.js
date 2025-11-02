// services/DeviceService.js
export class DeviceService {
    constructor(apiClient, eventBus) {
        this.api = apiClient;
        this.eventBus = eventBus;
        this.deviceStatus = null;
    }

    async connect(deviceName) {
        try {
            const response = await this.api.post('/api/device/connect', {
                deviceName: deviceName || ""
            });

            if (response.ok) {
                this.eventBus.emit('device:connecting');
                return { success: true };
            } else {
                throw new Error(`Failed to initiate connection: ${response.statusText}`);
            }
        } catch (error) {
            this.eventBus.emit('device:error', error.message);
            return { success: false, error: error.message };
        }
    }

    async disconnect() {
        try {
            const response = await this.api.post('/api/device/disconnect');

            if (response.ok) {
                this.eventBus.emit('device:disconnecting');
                return { success: true };
            }
        } catch (error) {
            this.eventBus.emit('device:error', error.message);
            return { success: false, error: error.message };
        }
    }

    updateStatus(status) {
        this.deviceStatus = status;
        this.eventBus.emit('device:status', status);
    }

    getStatus() {
        return this.deviceStatus;
    }
}
