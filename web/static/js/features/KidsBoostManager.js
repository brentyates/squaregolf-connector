export class KidsBoostManager {
    constructor(api, eventBus) {
        this.api = api;
        this.eventBus = eventBus;
        this.config = {
            enabled: false,
            speedMultiplier: 10.0,
            heightMultiplier: 2.0,
            straightnessBoost: 0.8,
            puttStraightness: 0.5
        };
        this.setupEventListeners();
    }

    setupEventListeners() {
        document.getElementById('kidsBoostEnabled')?.addEventListener('change', (e) => {
            this.config.enabled = e.target.checked;
            this.save();
        });

        document.getElementById('speedMultiplier')?.addEventListener('input', (e) => {
            document.getElementById('speedMultiplierValue').textContent = e.target.value;
        });
        document.getElementById('speedMultiplier')?.addEventListener('change', () => this.save());

        document.getElementById('heightMultiplier')?.addEventListener('input', (e) => {
            document.getElementById('heightMultiplierValue').textContent = parseFloat(e.target.value).toFixed(1);
        });
        document.getElementById('heightMultiplier')?.addEventListener('change', () => this.save());

        document.getElementById('straightnessBoost')?.addEventListener('input', (e) => {
            document.getElementById('straightnessBoostValue').textContent = e.target.value;
        });
        document.getElementById('straightnessBoost')?.addEventListener('change', () => this.save());

        document.getElementById('puttStraightness')?.addEventListener('input', (e) => {
            document.getElementById('puttStraightnessValue').textContent = e.target.value;
        });
        document.getElementById('puttStraightness')?.addEventListener('change', () => this.save());
    }

    async load() {
        try {
            const response = await this.api.get('/api/kidsboost/config');
            if (response.ok) {
                this.config = await response.json();
                this.updateUI();
            }
        } catch (error) {
            console.error('Failed to load kids boost config:', error);
        }
    }

    updateUI() {
        const enabledCheckbox = document.getElementById('kidsBoostEnabled');
        if (enabledCheckbox) {
            enabledCheckbox.checked = this.config.enabled;
        }

        const speedSlider = document.getElementById('speedMultiplier');
        const heightSlider = document.getElementById('heightMultiplier');
        const straightnessSlider = document.getElementById('straightnessBoost');
        const puttSlider = document.getElementById('puttStraightness');

        if (speedSlider) {
            speedSlider.value = this.config.speedMultiplier;
            document.getElementById('speedMultiplierValue').textContent = Math.round(this.config.speedMultiplier);
        }
        if (heightSlider) {
            heightSlider.value = this.config.heightMultiplier;
            document.getElementById('heightMultiplierValue').textContent = this.config.heightMultiplier.toFixed(1);
        }
        if (straightnessSlider) {
            straightnessSlider.value = this.config.straightnessBoost * 100;
            document.getElementById('straightnessBoostValue').textContent = Math.round(this.config.straightnessBoost * 100);
        }
        if (puttSlider) {
            puttSlider.value = this.config.puttStraightness * 100;
            document.getElementById('puttStraightnessValue').textContent = Math.round(this.config.puttStraightness * 100);
        }
    }

    async save() {
        const speedSlider = document.getElementById('speedMultiplier');
        const heightSlider = document.getElementById('heightMultiplier');
        const straightnessSlider = document.getElementById('straightnessBoost');
        const puttSlider = document.getElementById('puttStraightness');
        const enabledCheckbox = document.getElementById('kidsBoostEnabled');

        this.config.enabled = enabledCheckbox?.checked || false;
        this.config.speedMultiplier = parseFloat(speedSlider?.value || 10);
        this.config.heightMultiplier = parseFloat(heightSlider?.value || 2.0);
        this.config.straightnessBoost = parseFloat(straightnessSlider?.value || 80) / 100;
        this.config.puttStraightness = parseFloat(puttSlider?.value || 50) / 100;

        try {
            const response = await this.api.post('/api/kidsboost/config', this.config);
            if (!response.ok) {
                this.eventBus.emit('kidsboost:error', 'Failed to save settings');
            }
        } catch (error) {
            console.error('Failed to save kids boost config:', error);
            this.eventBus.emit('kidsboost:error', 'Failed to save settings');
        }
    }
}
