class SquareGolfApp {
    constructor() {
        this.websocket = null;
        this.reconnectInterval = null;
        this.currentScreen = 'device';
        this.deviceStatus = null;
        this.gsproStatus = null;
        this.settings = {};
        this.availableChimeSounds = [];
        
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.connectWebSocket();
        this.loadSettings();
        this.loadChimeSounds();
    }

    setupEventListeners() {
        // Navigation
        document.querySelectorAll('.nav-button').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const screen = e.target.dataset.screen || e.target.closest('[data-screen]').dataset.screen;
                this.showScreen(screen);
            });
        });

        // Device controls
        document.getElementById('connectBtn').addEventListener('click', () => this.connectDevice());
        document.getElementById('disconnectBtn').addEventListener('click', () => this.disconnectDevice());

        // GSPro controls
        document.getElementById('gsproConnectBtn').addEventListener('click', () => this.connectGSPro());
        document.getElementById('gsproDisconnectBtn').addEventListener('click', () => this.disconnectGSPro());
        
        // GSPro settings
        document.getElementById('gsproIP').addEventListener('change', () => this.saveGSProConfig());
        document.getElementById('gsproPort').addEventListener('change', () => this.saveGSProConfig());
        document.getElementById('gsproAutoConnect').addEventListener('change', () => this.saveGSProConfig());

        // Alignment controls
        document.getElementById('startCalibrationBtn').addEventListener('click', () => this.startCalibration());
        document.getElementById('stopCalibrationBtn').addEventListener('click', () => this.stopCalibration());

        // Settings controls
        document.getElementById('forgetDeviceBtn').addEventListener('click', () => this.forgetDevice());
        document.getElementById('playChimeBtn').addEventListener('click', () => this.playChime());
        
        // Settings changes
        document.getElementById('settingsDeviceName').addEventListener('change', () => this.saveSettings());
        document.getElementById('settingsAutoConnect').addEventListener('change', () => this.saveSettings());
        document.querySelectorAll('input[name="spinMode"]').forEach(radio => {
            radio.addEventListener('change', () => this.saveSettings());
        });
        document.getElementById('chimeSound').addEventListener('change', () => this.saveSettings());
        document.getElementById('chimeVolume').addEventListener('input', () => {
            this.updateVolumeDisplay();
            this.saveSettings();
        });

        // Metrics tabs
        document.querySelectorAll('.tab-button').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const tab = e.target.dataset.tab;
                this.showMetricsTab(tab);
            });
        });
    }

    connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;
        
        try {
            this.websocket = new WebSocket(wsUrl);
            
            this.websocket.onopen = () => {
                console.log('WebSocket connected');
                this.updateConnectionIndicator(true);
                if (this.reconnectInterval) {
                    clearInterval(this.reconnectInterval);
                    this.reconnectInterval = null;
                }
            };
            
            this.websocket.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data);
                    this.handleWebSocketMessage(message);
                } catch (error) {
                    console.error('Error parsing WebSocket message:', error);
                }
            };
            
            this.websocket.onclose = () => {
                console.log('WebSocket disconnected');
                this.updateConnectionIndicator(false);
                this.scheduleReconnect();
            };
            
            this.websocket.onerror = (error) => {
                console.error('WebSocket error:', error);
                this.updateConnectionIndicator(false);
            };
        } catch (error) {
            console.error('Failed to connect WebSocket:', error);
            this.scheduleReconnect();
        }
    }

    scheduleReconnect() {
        if (this.reconnectInterval) return;
        
        this.reconnectInterval = setInterval(() => {
            console.log('Attempting to reconnect WebSocket...');
            this.connectWebSocket();
        }, 3000);
    }

    handleWebSocketMessage(message) {
        switch (message.type) {
            case 'deviceStatus':
                this.updateDeviceStatus(message.data);
                break;
            case 'gsproStatus':
                this.updateGSProStatus(message.data);
                break;
            default:
                console.log('Unknown WebSocket message type:', message.type);
        }
    }

    updateConnectionIndicator(connected) {
        const indicator = document.getElementById('connectionIndicator');
        const dot = indicator.querySelector('.status-dot');
        const text = indicator.querySelector('.status-text');
        
        if (connected) {
            dot.classList.add('connected');
            dot.classList.remove('connecting');
            text.textContent = 'Connected';
        } else {
            dot.classList.remove('connected', 'connecting');
            text.textContent = 'Disconnected';
        }
    }

    showScreen(screenName) {
        // Update navigation
        document.querySelectorAll('.nav-button').forEach(btn => {
            btn.classList.remove('active');
        });
        document.querySelector(`[data-screen="${screenName}"]`).classList.add('active');
        
        // Update screens
        document.querySelectorAll('.screen').forEach(screen => {
            screen.classList.remove('active');
        });
        document.getElementById(`${screenName}Screen`).classList.add('active');
        
        this.currentScreen = screenName;
    }

    // Device functions
    async connectDevice() {
        const deviceName = document.getElementById('deviceNameInput').value.trim();
        
        try {
            const response = await fetch('/api/device/connect', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ deviceName: deviceName || "" })
            });
            
            if (response.ok) {
                this.showToast('Connection initiated...', 'info');
                this.showLoading('Connecting to device...');
            } else {
                throw new Error(`Failed to initiate connection: ${response.statusText}`);
            }
        } catch (error) {
            this.showToast(`Connection failed: ${error.message}`, 'error');
        }
    }

    async disconnectDevice() {
        try {
            const response = await fetch('/api/device/disconnect', {
                method: 'POST'
            });
            
            if (response.ok) {
                this.showToast('Disconnection initiated...', 'info');
            }
        } catch (error) {
            this.showToast(`Disconnect failed: ${error.message}`, 'error');
        }
    }

    updateDeviceStatus(status) {
        this.deviceStatus = status;
        
        // Update connection status
        const statusElement = document.getElementById('deviceStatus');
        const errorElement = document.getElementById('deviceError');
        const connectBtn = document.getElementById('connectBtn');
        const disconnectBtn = document.getElementById('disconnectBtn');
        
        statusElement.className = 'status-value';
        statusElement.classList.add(status.connectionStatus);
        
        switch (status.connectionStatus) {
            case 'connected':
                statusElement.textContent = 'Connected';
                connectBtn.disabled = true;
                disconnectBtn.disabled = false;
                errorElement.style.display = 'none';
                this.hideLoading();
                this.showDeviceInfo(true);
                break;
            case 'connecting':
                statusElement.textContent = 'Connecting...';
                connectBtn.disabled = true;
                disconnectBtn.disabled = false;
                errorElement.style.display = 'none';
                this.showLoading('Connecting to device...');
                this.showDeviceInfo(false);
                break;
            case 'disconnected':
                statusElement.textContent = 'Disconnected';
                connectBtn.disabled = false;
                disconnectBtn.disabled = true;
                errorElement.style.display = 'none';
                this.hideLoading();
                this.showDeviceInfo(false);
                break;
            case 'error':
                statusElement.textContent = 'Error';
                connectBtn.disabled = false;
                disconnectBtn.disabled = false;
                if (status.lastError) {
                    errorElement.textContent = status.lastError;
                    errorElement.style.display = 'block';
                }
                this.hideLoading();
                this.showDeviceInfo(false);
                break;
        }
        
        // Update device information
        if (status.deviceName) {
            document.getElementById('connectedDeviceName').textContent = status.deviceName;
        }
        
        if (status.batteryLevel !== null) {
            document.getElementById('batteryLevel').textContent = `${status.batteryLevel}%`;
        }
        
        // Update ball status
        this.updateBallStatus('ballDetected', status.ballDetected);
        this.updateBallStatus('ballReady', status.ballReady);
        
        if (status.ballPosition) {
            const posElement = document.getElementById('ballPosition');
            posElement.textContent = `X:${status.ballPosition.x}, Y:${status.ballPosition.y}, Z:${status.ballPosition.z}`;
            document.getElementById('ballPositionItem').style.display = 'block';
        }
        
        // Update system status
        if (status.club) {
            document.getElementById('clubValue').textContent = status.club.regularCode || status.club.name;
            document.getElementById('clubItem').style.display = 'block';
        }
        
        if (status.handedness !== null) {
            const handedness = status.handedness === 1 ? 'Right' : 'Left';
            document.getElementById('handednessValue').textContent = handedness;
            document.getElementById('handednessItem').style.display = 'block';
        }
        
        // Update metrics
        this.updateMetrics(status.lastBallMetrics, status.lastClubMetrics);
    }

    updateBallStatus(elementId, detected) {
        const element = document.getElementById(elementId);
        const icon = document.getElementById(elementId + 'Icon');
        
        if (detected) {
            element.textContent = 'Yes';
            element.classList.add('yes');
            element.classList.remove('no');
            icon.textContent = '🟢';
        } else {
            element.textContent = 'No';
            element.classList.add('no');
            element.classList.remove('yes');
            icon.textContent = '⚪';
        }
    }

    showDeviceInfo(show) {
        const cards = ['deviceInfoCard', 'ballStatusCard', 'systemStatusCard'];
        cards.forEach(cardId => {
            document.getElementById(cardId).style.display = show ? 'block' : 'none';
        });
    }

    updateMetrics(ballMetrics, clubMetrics) {
        if (ballMetrics || clubMetrics) {
            document.getElementById('metricsCard').style.display = 'block';
            
            if (ballMetrics) {
                this.displayBallMetrics(ballMetrics);
            }
            
            if (clubMetrics) {
                this.displayClubMetrics(clubMetrics);
            }
        }
    }

    displayBallMetrics(metrics) {
        const container = document.getElementById('ballMetrics');
        let html = '';
        
        if (metrics.speed !== undefined) html += this.createMetricItem('Speed', `${metrics.speed.toFixed(1)} mph`);
        if (metrics.launchAngle !== undefined) html += this.createMetricItem('Launch Angle', `${metrics.launchAngle.toFixed(1)}°`);
        if (metrics.backSpin !== undefined) html += this.createMetricItem('Back Spin', `${metrics.backSpin.toFixed(0)} rpm`);
        if (metrics.sideSpin !== undefined) html += this.createMetricItem('Side Spin', `${metrics.sideSpin.toFixed(0)} rpm`);
        
        container.innerHTML = html;
    }

    displayClubMetrics(metrics) {
        const container = document.getElementById('clubMetrics');
        let html = '';
        
        if (metrics.speed !== undefined) html += this.createMetricItem('Club Speed', `${metrics.speed.toFixed(1)} mph`);
        if (metrics.angle !== undefined) html += this.createMetricItem('Club Angle', `${metrics.angle.toFixed(1)}°`);
        if (metrics.path !== undefined) html += this.createMetricItem('Club Path', `${metrics.path.toFixed(1)}°`);
        
        container.innerHTML = html;
    }

    createMetricItem(label, value) {
        return `
            <div class="metrics-item">
                <span class="metrics-label">${label}:</span>
                <span class="metrics-value">${value}</span>
            </div>
        `;
    }

    showMetricsTab(tab) {
        document.querySelectorAll('.tab-button').forEach(btn => {
            btn.classList.remove('active');
        });
        document.querySelector(`[data-tab="${tab}"]`).classList.add('active');
        
        document.querySelectorAll('.metrics-tab').forEach(tabContent => {
            tabContent.classList.remove('active');
        });
        document.getElementById(`${tab}Metrics`).classList.add('active');
    }

    // GSPro functions
    async connectGSPro() {
        const ip = document.getElementById('gsproIP').value.trim();
        const port = parseInt(document.getElementById('gsproPort').value);
        
        if (!ip || !port) {
            this.showToast('Please enter valid IP and port', 'error');
            return;
        }
        
        try {
            const response = await fetch('/api/gspro/connect', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ ip, port })
            });
            
            if (response.ok) {
                this.showToast('GSPro connection initiated...', 'info');
            } else {
                throw new Error(`Failed to connect: ${response.statusText}`);
            }
        } catch (error) {
            this.showToast(`GSPro connection failed: ${error.message}`, 'error');
        }
    }

    async disconnectGSPro() {
        try {
            const response = await fetch('/api/gspro/disconnect', {
                method: 'POST'
            });
            
            if (response.ok) {
                this.showToast('GSPro disconnection initiated...', 'info');
            }
        } catch (error) {
            this.showToast(`GSPro disconnect failed: ${error.message}`, 'error');
        }
    }

    updateGSProStatus(status) {
        this.gsproStatus = status;
        
        const statusElement = document.getElementById('gsproStatus');
        const errorElement = document.getElementById('gsproError');
        const connectBtn = document.getElementById('gsproConnectBtn');
        const disconnectBtn = document.getElementById('gsproDisconnectBtn');
        const ipField = document.getElementById('gsproIP');
        const portField = document.getElementById('gsproPort');
        
        statusElement.className = 'status-value';
        statusElement.classList.add(status.connectionStatus);
        
        switch (status.connectionStatus) {
            case 'connected':
                statusElement.textContent = 'Connected';
                connectBtn.disabled = true;
                disconnectBtn.disabled = false;
                ipField.disabled = true;
                portField.disabled = true;
                errorElement.style.display = 'none';
                break;
            case 'connecting':
                statusElement.textContent = 'Connecting...';
                connectBtn.disabled = true;
                disconnectBtn.disabled = true;
                ipField.disabled = true;
                portField.disabled = true;
                errorElement.style.display = 'none';
                break;
            case 'disconnected':
                statusElement.textContent = 'Disconnected';
                connectBtn.disabled = false;
                disconnectBtn.disabled = true;
                ipField.disabled = false;
                portField.disabled = false;
                errorElement.style.display = 'none';
                break;
            case 'error':
                statusElement.textContent = 'Error';
                connectBtn.disabled = false;
                disconnectBtn.disabled = true;
                ipField.disabled = false;
                portField.disabled = false;
                if (status.lastError) {
                    errorElement.textContent = status.lastError;
                    errorElement.style.display = 'block';
                }
                break;
        }
    }

    async saveGSProConfig() {
        const ip = document.getElementById('gsproIP').value.trim();
        const port = parseInt(document.getElementById('gsproPort').value);
        const autoConnect = document.getElementById('gsproAutoConnect').checked;
        
        try {
            const response = await fetch('/api/gspro/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ ip, port, autoConnect })
            });
            
            if (!response.ok) {
                throw new Error(`Failed to save config: ${response.statusText}`);
            }
        } catch (error) {
            this.showToast(`Failed to save GSPro config: ${error.message}`, 'error');
        }
    }

    // Calibration functions
    startCalibration() {
        // TODO: Implement calibration start
        document.getElementById('calibrationStatus').textContent = 'Calibrating...';
        document.getElementById('startCalibrationBtn').disabled = true;
        document.getElementById('stopCalibrationBtn').disabled = false;
        this.showToast('Calibration started', 'info');
    }

    stopCalibration() {
        // TODO: Implement calibration stop
        document.getElementById('calibrationStatus').textContent = 'Ready to calibrate';
        document.getElementById('startCalibrationBtn').disabled = false;
        document.getElementById('stopCalibrationBtn').disabled = true;
        this.showToast('Calibration stopped', 'info');
    }

    // Settings functions
    async loadSettings() {
        try {
            const response = await fetch('/api/settings');
            if (response.ok) {
                this.settings = await response.json();
                this.applySettings();
            }
        } catch (error) {
            console.error('Failed to load settings:', error);
        }
    }

    applySettings() {
        document.getElementById('settingsDeviceName').value = this.settings.deviceName || '';
        document.getElementById('settingsAutoConnect').checked = this.settings.autoConnect || false;
        
        // Set spin mode
        const spinMode = this.settings.spinMode || 'advanced';
        document.querySelector(`input[name="spinMode"][value="${spinMode}"]`).checked = true;
        
        // Set chime settings
        if (this.settings.chimeSound) {
            document.getElementById('chimeSound').value = this.settings.chimeSound;
        }
        
        const volume = this.settings.chimeVolume || 0.8;
        document.getElementById('chimeVolume').value = volume;
        this.updateVolumeDisplay();
    }

    async saveSettings() {
        const deviceName = document.getElementById('settingsDeviceName').value.trim();
        const autoConnect = document.getElementById('settingsAutoConnect').checked;
        const spinMode = document.querySelector('input[name="spinMode"]:checked').value;
        const chimeSound = document.getElementById('chimeSound').value;
        const chimeVolume = parseFloat(document.getElementById('chimeVolume').value);
        
        const settings = {
            deviceName,
            autoConnect,
            spinMode,
            chimeSound,
            chimeVolume
        };
        
        try {
            const response = await fetch('/api/settings', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(settings)
            });
            
            if (response.ok) {
                this.settings = { ...this.settings, ...settings };
            } else {
                throw new Error(`Failed to save settings: ${response.statusText}`);
            }
        } catch (error) {
            this.showToast(`Failed to save settings: ${error.message}`, 'error');
        }
    }

    forgetDevice() {
        document.getElementById('settingsDeviceName').value = '';
        this.saveSettings();
        this.showToast('Device forgotten', 'success');
    }

    async loadChimeSounds() {
        try {
            const response = await fetch('/api/settings/chime/sounds');
            if (response.ok) {
                this.availableChimeSounds = await response.json();
                this.populateChimeSounds();
            }
        } catch (error) {
            console.error('Failed to load chime sounds:', error);
        }
    }

    populateChimeSounds() {
        const select = document.getElementById('chimeSound');
        select.innerHTML = '';
        
        this.availableChimeSounds.forEach(sound => {
            const option = document.createElement('option');
            option.value = sound;
            option.textContent = sound;
            select.appendChild(option);
        });
    }

    async playChime() {
        const sound = document.getElementById('chimeSound').value;
        
        try {
            const response = await fetch('/api/settings/chime/play', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ sound })
            });
            
            if (!response.ok) {
                throw new Error(`Failed to play sound: ${response.statusText}`);
            }
        } catch (error) {
            this.showToast(`Failed to play chime: ${error.message}`, 'error');
        }
    }

    updateVolumeDisplay() {
        const volume = document.getElementById('chimeVolume').value;
        const percentage = Math.round(volume * 100);
        document.getElementById('volumeValue').textContent = `${percentage}%`;
    }

    // Utility functions
    showLoading(text = 'Loading...') {
        const overlay = document.getElementById('loadingOverlay');
        const loadingText = overlay.querySelector('.loading-text');
        loadingText.textContent = text;
        overlay.classList.add('show');
    }

    hideLoading() {
        document.getElementById('loadingOverlay').classList.remove('show');
    }

    showToast(message, type = 'info') {
        const container = document.getElementById('toastContainer');
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.textContent = message;
        
        container.appendChild(toast);
        
        setTimeout(() => {
            toast.remove();
        }, 5000);
    }
}

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.app = new SquareGolfApp();
});
