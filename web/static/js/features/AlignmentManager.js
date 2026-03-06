// features/AlignmentManager.js
export class AlignmentManager {
    constructor(apiClient, eventBus) {
        this.api = apiClient;
        this.eventBus = eventBus;
        this.currentHandedness = 'right';
    }

    async start() {
        return this.#runCommand({
            url: '/api/alignment/start',
            successEvent: 'alignment:started',
            errorEvent: 'alignment:error',
            stopOnError: true,
            errorMessage: 'Failed to start alignment'
        });
    }

    async stop() {
        return this.#runCommand({
            url: '/api/alignment/stop',
            successEvent: 'alignment:stopped',
            errorEvent: 'alignment:error',
            emitError: false,
            errorMessage: 'Failed to stop alignment'
        });
    }

    async save() {
        return this.#runCommand({
            url: '/api/alignment/stop',
            successEvent: 'alignment:saved',
            errorEvent: 'alignment:error',
            stopOnError: true,
            errorMessage: 'Failed to save calibration'
        });
    }

    async cancel() {
        return this.#runCommand({
            url: '/api/alignment/cancel',
            successEvent: 'alignment:cancelled',
            errorEvent: 'alignment:error',
            stopOnError: true,
            errorMessage: 'Failed to cancel alignment'
        });
    }

    async setHandedness(handedness) {
        try {
            const response = await this.api.post('/api/alignment/handedness', { handedness });

            if (!response.ok) {
                throw new Error('Failed to set handedness');
            }

            this.currentHandedness = handedness;
            this.eventBus.emit('alignment:handedness-changed', handedness);
            return { success: true };
        } catch (error) {
            this.eventBus.emit('alignment:error', 'Failed to set handedness');
            this.eventBus.emit('alignment:stopped');
            return { success: false, error: error.message };
        }
    }

    updateDisplay(angle, isAligned) {
        this.eventBus.emit('alignment:update', { angle, isAligned });
    }

    async #runCommand({
        url,
        successEvent,
        errorEvent,
        errorMessage,
        stopOnError = false,
        emitError = true
    }) {
        try {
            const response = await this.api.post(url);

            if (!response.ok) {
                throw new Error(errorMessage);
            }

            this.eventBus.emit(successEvent);
            return { success: true };
        } catch (error) {
            if (emitError) {
                this.eventBus.emit(errorEvent, errorMessage);
            }
            if (stopOnError) {
                this.eventBus.emit('alignment:stopped');
            }
            return { success: false, error: error.message };
        }
    }
}
