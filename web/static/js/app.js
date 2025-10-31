class SquareGolfApp {
    constructor() {
        this.websocket = null;
        this.reconnectInterval = null;
        this.currentScreen = 'device';
        this.deviceStatus = null;
        this.gsproStatus = null;
        this.cameraConfig = null;
        this.settings = {};
        this.features = {};
        this.currentHandedness = 'right'; // Default to right-handed

        this.init();
    }

    init() {
        this.loadFeatures().then(() => {
            this.setupEventListeners();
            this.connectWebSocket();
            this.loadSettings();
        });
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

        // Camera controls
        document.getElementById('cameraSaveBtn').addEventListener('click', () => this.saveCameraConfig());

        // Alignment controls
        document.getElementById('leftHandedBtn').addEventListener('click', () => this.setHandedness('left'));
        document.getElementById('rightHandedBtn').addEventListener('click', () => this.setHandedness('right'));

        // Settings controls
        document.getElementById('forgetDeviceBtn').addEventListener('click', () => this.forgetDevice());
        
        // Settings changes
        document.getElementById('settingsDeviceName').addEventListener('change', () => this.saveSettings());
        document.getElementById('settingsAutoConnect').addEventListener('change', () => this.saveSettings());
        document.querySelectorAll('input[name="spinMode"]').forEach(radio => {
            radio.addEventListener('change', () => this.saveSettings());
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
            case 'cameraConfig':
                this.updateCameraConfig(message.data);
                break;
            case 'alignmentData':
                if (message.data) {
                    this.updateAlignmentDisplay(
                        message.data.alignmentAngle || 0,
                        message.data.isAligned || false
                    );
                }
                break;
            default:
                console.log('Unknown WebSocket message type:', message.type);
        }
    }

    updateConnectionIndicator(connected) {
        // Update the small WebSocket debug indicator
        const wsIndicator = document.querySelector('.ws-indicator');
        if (wsIndicator) {
            if (connected) {
                wsIndicator.classList.add('connected');
                wsIndicator.classList.remove('connecting');
                wsIndicator.title = 'WebSocket Connected';
            } else {
                wsIndicator.classList.remove('connected', 'connecting');
                wsIndicator.title = 'WebSocket Disconnected';
                console.log('WebSocket disconnected - attempting reconnection...');
            }
        }
    }

    updateDeviceConnectionIndicator(deviceStatus) {
        const indicator = document.getElementById('connectionIndicator');
        const dot = indicator.querySelector('.status-dot');
        const text = indicator.querySelector('.status-text');
        
        switch (deviceStatus) {
            case 'connected':
                dot.classList.remove('connecting');
                dot.classList.add('connected');
                text.textContent = 'Device Connected';
                break;
            case 'connecting':
                dot.classList.remove('connected');
                dot.classList.add('connecting');
                text.textContent = 'Connecting...';
                break;
            case 'disconnected':
            case 'error':
            default:
                dot.classList.remove('connected', 'connecting');
                text.textContent = 'Device Disconnected';
                break;
        }
    }

    showScreen(screenName) {
        // Check if alignment screen requires device connection
        if (screenName === 'alignment' && (!this.deviceStatus || this.deviceStatus.connectionStatus !== 'connected')) {
            this.showToast('Please connect to device first', 'warning');
            return;
        }

        // Stop alignment if leaving alignment screen
        if (this.currentScreen === 'alignment' && screenName !== 'alignment') {
            this.stopAlignment();
        }

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

        // Start alignment if entering alignment screen
        if (screenName === 'alignment') {
            this.startAlignment();
        }
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
        
        // Update the main navigation device connection indicator
        this.updateDeviceConnectionIndicator(status.connectionStatus);
        
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

        if (status.firmwareVersion !== null) {
            document.getElementById('firmwareVersion').textContent = status.firmwareVersion;
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
            const handedness = status.handedness === 0 ? 'Right' : 'Left';
            document.getElementById('handednessValue').textContent = handedness;
            document.getElementById('handednessItem').style.display = 'block';

            // Update alignment screen handedness display
            this.currentHandedness = handedness.toLowerCase();
            this.updateHandednessDisplay(this.currentHandedness);
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

    // Camera functions
    async saveCameraConfig() {
        const url = document.getElementById('cameraURL').value.trim();
        const enabled = document.getElementById('cameraEnabled').checked;

        if (!url) {
            this.showToast('Please enter a valid camera URL', 'error');
            return;
        }

        try {
            const response = await fetch('/api/camera/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ url, enabled })
            });

            if (response.ok) {
                this.showToast('Camera settings saved successfully', 'success');
            } else {
                throw new Error(`Failed to save config: ${response.statusText}`);
            }
        } catch (error) {
            this.showToast(`Failed to save camera config: ${error.message}`, 'error');
        }
    }

    updateCameraConfig(config) {
        this.cameraConfig = config;

        // Update UI elements
        const urlField = document.getElementById('cameraURL');
        const enabledCheckbox = document.getElementById('cameraEnabled');

        if (urlField && config.url) {
            urlField.value = config.url;
        }

        if (enabledCheckbox) {
            enabledCheckbox.checked = config.enabled;
        }
    }

    // Alignment functions
    async startAlignment() {
        try {
            const response = await fetch('/api/alignment/start', {
                method: 'POST'
            });

            if (!response.ok) {
                throw new Error('Failed to start alignment');
            }

            console.log('Alignment started');
        } catch (error) {
            console.error('Error starting alignment:', error);
            this.showToast('Failed to start alignment', 'error');
        }
    }

    async stopAlignment() {
        try {
            const response = await fetch('/api/alignment/stop', {
                method: 'POST'
            });

            if (!response.ok) {
                throw new Error('Failed to stop alignment');
            }

            console.log('Alignment stopped');

            // Reset display
            this.updateAlignmentDisplay(0, false);
        } catch (error) {
            console.error('Error stopping alignment:', error);
        }
    }

    updateAlignmentDisplay(angle, isAligned) {
        // Update numeric angle
        const angleElement = document.getElementById('alignmentAngle');
        const directionElement = document.getElementById('alignmentDirection');
        const statusElement = document.getElementById('alignedStatus');
        const pointerElement = document.getElementById('aimPointer');

        if (!angleElement) return; // Not on alignment screen

        // Format angle
        const formattedAngle = Math.abs(angle).toFixed(1);
        angleElement.textContent = `${formattedAngle}°`;

        // Update direction text
        if (Math.abs(angle) < 0.5) {
            directionElement.textContent = 'Aimed straight';
        } else if (angle > 0) {
            directionElement.textContent = `Aimed ${formattedAngle}° right`;
        } else {
            directionElement.textContent = `Aimed ${formattedAngle}° left`;
        }

        // Update angle color based on magnitude
        angleElement.classList.remove('aligned', 'close', 'far');
        if (isAligned) {
            angleElement.classList.add('aligned');
        } else if (Math.abs(angle) < 5) {
            angleElement.classList.add('close');
        } else {
            angleElement.classList.add('far');
        }

        // Update status indicator
        statusElement.classList.remove('aligned', 'not-aligned');
        const iconElement = statusElement.querySelector('.aligned-icon');
        const textElement = statusElement.querySelector('.aligned-text');

        if (isAligned) {
            statusElement.classList.add('aligned');
            iconElement.textContent = '✅';
            textElement.textContent = 'Aligned!';
        } else {
            statusElement.classList.add('not-aligned');
            iconElement.textContent = '⚠️';
            textElement.textContent = 'Not aligned';
        }

        // Rotate compass pointer
        if (pointerElement) {
            pointerElement.setAttribute('transform', `rotate(${angle} 100 100)`);
        }
    }

    async setHandedness(handedness) {
        try {
            const response = await fetch('/api/alignment/handedness', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ handedness })
            });

            if (!response.ok) {
                throw new Error('Failed to set handedness');
            }

            this.currentHandedness = handedness;
            this.updateHandednessDisplay(handedness);
            console.log('Handedness set to:', handedness);
        } catch (error) {
            console.error('Error setting handedness:', error);
            this.showToast('Failed to set handedness', 'error');
        }
    }

    updateHandednessDisplay(handedness) {
        const label = document.getElementById('handednessLabel');
        if (label) {
            label.textContent = handedness.toUpperCase();
        }
    }

    // Settings functions
    async loadFeatures() {
        try {
            const response = await fetch('/api/features');
            if (response.ok) {
                this.features = await response.json();
                this.applyFeatures();
            }
        } catch (error) {
            console.error('Failed to load features:', error);
        }
    }

    applyFeatures() {
        // Hide camera tab if external camera feature is disabled
        if (!this.features.externalCamera) {
            const cameraNavButton = document.querySelector('.nav-button[data-screen="camera"]');
            if (cameraNavButton) {
                cameraNavButton.style.display = 'none';
            }
            const cameraScreen = document.getElementById('cameraScreen');
            if (cameraScreen) {
                cameraScreen.style.display = 'none';
            }
        }
    }

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
        
        const spinMode = this.settings.spinMode || 'advanced';
        document.querySelector(`input[name="spinMode"][value="${spinMode}"]`).checked = true;
    }

    async saveSettings() {
        const deviceName = document.getElementById('settingsDeviceName').value.trim();
        const autoConnect = document.getElementById('settingsAutoConnect').checked;
        const spinMode = document.querySelector('input[name="spinMode"]:checked').value;
        
        const settings = {
            deviceName,
            autoConnect,
            spinMode
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
