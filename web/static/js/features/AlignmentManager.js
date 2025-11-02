// features/AlignmentManager.js
export class AlignmentManager {
    constructor(apiClient, eventBus) {
        this.api = apiClient;
        this.eventBus = eventBus;
        this.currentHandedness = 'right';
        this.explicitlyStopped = false;
    }

    async start() {
        try {
            const response = await this.api.post('/api/alignment/start');

            if (!response.ok) {
                throw new Error('Failed to start alignment');
            }

            console.log('Alignment started');
            this.eventBus.emit('alignment:started');
        } catch (error) {
            console.error('Error starting alignment:', error);
            this.eventBus.emit('alignment:error', 'Failed to start alignment');
        }
    }

    async stop() {
        try {
            const response = await this.api.post('/api/alignment/stop');

            if (!response.ok) {
                throw new Error('Failed to stop alignment');
            }

            console.log('Alignment stopped');
            this.eventBus.emit('alignment:stopped');
        } catch (error) {
            console.error('Error stopping alignment:', error);
        }
    }

    async save() {
        try {
            this.explicitlyStopped = true;

            const response = await this.api.post('/api/alignment/stop');

            if (!response.ok) {
                throw new Error('Failed to save alignment');
            }

            console.log('Alignment saved');
            this.eventBus.emit('alignment:saved');
        } catch (error) {
            console.error('Error saving alignment:', error);
            this.eventBus.emit('alignment:error', 'Failed to save calibration');
        }
    }

    async cancel(skipNavigation = false) {
        try {
            this.explicitlyStopped = true;

            const response = await this.api.post('/api/alignment/cancel');

            if (!response.ok) {
                throw new Error('Failed to cancel alignment');
            }

            console.log('Alignment cancelled');
            this.eventBus.emit('alignment:cancelled', { skipNavigation });
        } catch (error) {
            console.error('Error cancelling alignment:', error);
            this.eventBus.emit('alignment:error', 'Failed to cancel alignment');
        }
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
            console.error('Error setting handedness:', error);
            this.eventBus.emit('alignment:error', 'Failed to set handedness');
            return { success: false };
        }
    }

    updateDisplay(angle, isAligned) {
        this.eventBus.emit('alignment:update', { angle, isAligned });
    }
}
