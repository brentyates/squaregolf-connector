// features/SettingsManager.js
export class SettingsManager {
    constructor(apiClient, eventBus) {
        this.api = apiClient;
        this.eventBus = eventBus;
        this.settings = {};
    }

    async load() {
        try {
            const response = await this.api.get('/api/settings');
            if (response.ok) {
                this.settings = await response.json();
                this.eventBus.emit('settings:loaded', this.settings);
            }
        } catch (error) {
            console.error('Failed to load settings:', error);
        }
    }

    async save(newSettings) {
        try {
            const response = await this.api.post('/api/settings', newSettings);

            if (response.ok) {
                this.settings = { ...this.settings, ...newSettings };
                this.eventBus.emit('settings:saved', this.settings);
                return { success: true };
            } else {
                throw new Error(`Failed to save settings: ${response.statusText}`);
            }
        } catch (error) {
            this.eventBus.emit('settings:error', error.message);
            return { success: false, error: error.message };
        }
    }

    get(key) {
        return this.settings[key];
    }

    getAll() {
        return { ...this.settings };
    }
}
