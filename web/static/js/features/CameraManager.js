// features/CameraManager.js
export class CameraManager {
    constructor(apiClient, eventBus) {
        this.api = apiClient;
        this.eventBus = eventBus;
        this.config = null;
    }

    async save() {
        const url = document.getElementById('cameraURL').value.trim();
        const enabled = document.getElementById('cameraEnabled').checked;

        if (!url) {
            this.eventBus.emit('camera:error', 'Please enter a valid camera URL');
            return { success: false };
        }

        try {
            const response = await this.api.post('/api/camera/config', { url, enabled });

            if (response.ok) {
                this.eventBus.emit('camera:saved');
                return { success: true };
            } else {
                throw new Error(`Failed to save config: ${response.statusText}`);
            }
        } catch (error) {
            this.eventBus.emit('camera:error', error.message);
            return { success: false, error: error.message };
        }
    }

    updateConfig(config) {
        this.config = config;

        // Update UI elements
        const urlField = document.getElementById('cameraURL');
        const enabledCheckbox = document.getElementById('cameraEnabled');

        if (urlField && config.url) {
            urlField.value = config.url;
        }

        if (enabledCheckbox) {
            enabledCheckbox.checked = config.enabled;
        }

        this.eventBus.emit('camera:config-updated', config);
    }
}
