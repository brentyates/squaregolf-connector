// features/CameraManager.js
export class CameraManager {
    constructor(apiClient, eventBus) {
        this.api = apiClient;
        this.eventBus = eventBus;
        this.config = null;
    }

    $(id) {
        return document.getElementById(id);
    }

    async save() {
        const urlField = this.$('cameraURL');
        const enabledField = this.$('cameraEnabled');

        if (!urlField || !enabledField) {
            return { success: false, error: 'Camera controls are not rendered in this UI.' };
        }

        const url = urlField.value.trim();
        const enabled = enabledField.checked;

        if (enabled && !url) {
            this.eventBus.emit('camera:error', 'Please enter a valid camera URL');
            return { success: false };
        }

        try {
            const response = await this.api.post('/api/camera/config', { url, enabled });

            if (!response.ok) {
                throw new Error(`Failed to save config: ${response.statusText}`);
            }

            this.eventBus.emit('camera:saved');
            return { success: true };
        } catch (error) {
            this.eventBus.emit('camera:error', error.message);
            return { success: false, error: error.message };
        }
    }

    updateConfig(config) {
        this.config = config;

        const urlField = this.$('cameraURL');
        const enabledCheckbox = this.$('cameraEnabled');

        if (urlField) {
            urlField.value = config.url || '';
        }

        if (enabledCheckbox) {
            enabledCheckbox.checked = config.enabled;
        }

        this.eventBus.emit('camera:config-updated', config);
    }
}
