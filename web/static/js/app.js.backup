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

        // Status bar navigation
        document.getElementById('statusDevice')?.addEventListener('click', () => this.showScreen('device'));
        document.getElementById('statusGSPro')?.addEventListener('click', () => this.showScreen('gspro'));
        document.getElementById('statusBallReady')?.addEventListener('click', () => this.showScreen('shotMonitor'));

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
        document.getElementById('saveAlignmentBtn').addEventListener('click', () => this.saveAlignment());
        document.getElementById('cancelAlignmentBtn').addEventListener('click', () => this.cancelAlignment());

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
        // Update the global status bar WebSocket indicator
        const statusWebSocket = document.getElementById('statusWebSocket');
        if (statusWebSocket) {
            if (connected) {
                statusWebSocket.classList.add('connected');
                statusWebSocket.classList.remove('disconnected');
            } else {
                statusWebSocket.classList.remove('connected');
                statusWebSocket.classList.add('disconnected');
                console.log('WebSocket disconnected - attempting reconnection...');
            }
        }
    }

    updateDeviceConnectionIndicator(deviceStatus) {
        // Update the global status bar Device indicator
        const statusDevice = document.getElementById('statusDevice');
        if (statusDevice) {
            if (deviceStatus === 'connected') {
                statusDevice.classList.add('connected');
                statusDevice.classList.remove('disconnected');
            } else {
                statusDevice.classList.remove('connected');
                statusDevice.classList.add('disconnected');
            }
        }
    }

    showScreen(screenName) {
        // Check if alignment screen requires device connection
        if (screenName === 'alignment' && (!this.deviceStatus || this.deviceStatus.connectionStatus !== 'connected')) {
            this.showToast('Please connect to device first', 'warning');
            return;
        }

        // Auto-cancel alignment if leaving alignment screen without explicit save/cancel
        if (this.currentScreen === 'alignment' && screenName !== 'alignment') {
            if (!this.alignmentExplicitlyStopped) {
                this.cancelAlignment();
            }
            this.alignmentExplicitlyStopped = false; // Reset for next time
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

        // Update page title
        const pageTitles = {
            device: 'Device Connection',
            shotMonitor: 'Shot Monitor',
            gspro: 'GSPro Connection',
            camera: 'Swing Camera',
            alignment: 'Device Alignment',
            settings: 'Settings'
        };
        document.getElementById('pageTitle').textContent = pageTitles[screenName] || screenName;

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
            const batteryElement = document.getElementById('batteryLevel');
            const level = status.batteryLevel;
            let icon = '';
            let className = '';

            if (level >= 80) {
                icon = '🔋';
                className = 'battery-high';
            } else if (level >= 50) {
                icon = '🔋';
                className = 'battery-medium';
            } else if (level >= 20) {
                icon = '⚠️';
                className = 'battery-medium';
            } else {
                icon = '🪫';
                className = 'battery-low';
            }

            batteryElement.innerHTML = `<span class="battery-indicator"><span class="battery-icon ${className}">${icon}</span> ${level}%</span>`;
        }

        // Update LM Version (FirmwareVersion)
        const firmwareElement = document.getElementById('firmwareVersion');
        if (firmwareElement) {
            firmwareElement.textContent = status.firmwareVersion !== null ? status.firmwareVersion : '-';
        }

        // Update Launcher Version
        const launcherElement = document.getElementById('launcherVersion');
        if (launcherElement) {
            launcherElement.textContent = status.launcherVersion !== null ? status.launcherVersion : '-';
        }

        // Update MMI Version
        const mmiElement = document.getElementById('mmiVersion');
        if (mmiElement) {
            mmiElement.textContent = status.mmiVersion !== null ? status.mmiVersion : '-';
        }

        // Update system status
        if (status.club) {
            const clubValueElement = document.getElementById('clubValue');
            const clubItemElement = document.getElementById('clubItem');
            if (clubValueElement) {
                clubValueElement.textContent = status.club.regularCode || status.club.name;
            }
            if (clubItemElement) {
                clubItemElement.style.display = 'block';
            }
        }

        if (status.handedness !== null) {
            const handedness = status.handedness === 0 ? 'Right' : 'Left';
            const handednessValueElement = document.getElementById('handednessValue');
            const handednessItemElement = document.getElementById('handednessItem');

            if (handednessValueElement) {
                handednessValueElement.textContent = handedness;
            }
            if (handednessItemElement) {
                handednessItemElement.style.display = 'block';
            }

            // Update alignment screen handedness display
            this.currentHandedness = handedness.toLowerCase();
            this.updateHandednessDisplay(this.currentHandedness);
        }
        
        // Update metrics
        this.updateMetrics(status.lastBallMetrics, status.lastClubMetrics);

        // Update Shot Monitor if available
        if (window.shotMonitor) {
            window.shotMonitor.updateStatus(status);

            // If we have new shot data, update current shot and add to history
            if (status.lastBallMetrics && Object.keys(status.lastBallMetrics).length > 0) {
                window.shotMonitor.updateCurrentShot(status.lastBallMetrics, status.lastClubMetrics);
                window.shotMonitor.addShotToHistory(status.lastBallMetrics, status.lastClubMetrics || {});
            }
        }

        // Update alignment display if alignment data is present
        if (status.isAligning && typeof status.alignmentAngle === 'number') {
            this.updateAlignmentDisplay(status.alignmentAngle, status.isAligned || false);
        }
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
        const deviceInfoCard = document.getElementById('deviceInfoCard');
        if (deviceInfoCard) {
            deviceInfoCard.style.display = show ? 'block' : 'none';
        }
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

        // Update the global status bar GSPro indicator
        const statusGSPro = document.getElementById('statusGSPro');
        if (statusGSPro) {
            if (status.connectionStatus === 'connected') {
                statusGSPro.classList.add('connected');
                statusGSPro.classList.remove('disconnected');
            } else {
                statusGSPro.classList.remove('connected');
                statusGSPro.classList.add('disconnected');
            }
        }

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

    async saveAlignment() {
        try {
            this.alignmentExplicitlyStopped = true;

            const response = await fetch('/api/alignment/stop', {
                method: 'POST'
            });

            if (!response.ok) {
                throw new Error('Failed to save alignment');
            }

            console.log('Alignment saved (OK button)');
            this.showToast('Calibration saved successfully', 'success');

            // Reset display
            this.updateAlignmentDisplay(0, false);

            // Switch to device screen
            this.showScreen('device');
        } catch (error) {
            console.error('Error saving alignment:', error);
            this.showToast('Failed to save calibration', 'error');
        }
    }

    async cancelAlignment(skipNavigation = false) {
        try {
            this.alignmentExplicitlyStopped = true;

            const response = await fetch('/api/alignment/cancel', {
                method: 'POST'
            });

            if (!response.ok) {
                throw new Error('Failed to cancel alignment');
            }

            console.log('Alignment cancelled');

            // Only show toast if not part of a handedness toggle (to avoid duplicate toasts)
            if (!skipNavigation) {
                this.showToast('Calibration cancelled', 'info');
            }

            // Reset display
            this.updateAlignmentDisplay(0, false);

            // If explicitly cancelled via button, switch to device screen
            // Skip navigation if this is part of a handedness toggle restart
            if (this.currentScreen === 'alignment' && !skipNavigation) {
                this.showScreen('device');
            }
        } catch (error) {
            console.error('Error cancelling alignment:', error);
            this.showToast('Failed to cancel alignment', 'error');
        }
    }

    updateAlignmentDisplay(angle, isAligned) {
        // Update numeric angle
        const angleElement = document.getElementById('alignmentAngle');
        const directionElement = document.getElementById('alignmentDirection');
        const statusElement = document.getElementById('alignedStatus');
        const pointerElement = document.getElementById('aimPointer');

        if (!angleElement) return; // Not on alignment screen

        // Flip angle sign for left-handed users
        let displayAngle = angle;
        if (this.currentHandedness === 'left') {
            displayAngle = -angle;
        }

        // Format angle
        const formattedAngle = Math.abs(displayAngle).toFixed(1);
        angleElement.textContent = `${formattedAngle}°`;

        // Update direction text
        if (Math.abs(displayAngle) < 0.5) {
            directionElement.textContent = 'Aimed straight';
        } else if (displayAngle > 0) {
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

            // If we're currently in alignment mode, restart it with the new handedness
            if (this.currentScreen === 'alignment') {
                console.log('Restarting alignment with new handedness:', handedness);
                // First cancel the current alignment, then restart with new handedness
                // Pass skipNavigation=true to stay on alignment screen
                await this.cancelAlignment(true);
                // Small delay to ensure cancel completes
                await new Promise(resolve => setTimeout(resolve, 100));
                await this.startAlignment();
            }
        } catch (error) {
            console.error('Error setting handedness:', error);
            this.showToast('Failed to set handedness', 'error');
        }
    }

    updateHandednessDisplay(handedness) {
        const leftBtn = document.getElementById('leftHandedBtn');
        const rightBtn = document.getElementById('rightHandedBtn');

        // Update button states
        if (leftBtn && rightBtn) {
            if (handedness === 'left') {
                leftBtn.classList.add('active');
                rightBtn.classList.remove('active');
            } else {
                rightBtn.classList.add('active');
                leftBtn.classList.remove('active');
            }
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

// Shot Monitor Module
class ShotMonitor {
    constructor() {
        this.initializeEventListeners();
    }

    initializeEventListeners() {
        // Metrics tab switching
        const metricsTabs = document.querySelectorAll('.monitor-metrics-tabs .tab-button');
        metricsTabs.forEach(button => {
            button.addEventListener('click', (e) => {
                const tab = e.target.dataset.tab;
                this.switchMetricsTab(tab);
            });
        });
    }

    updateBallPosition(position, ballDetected, ballReady) {
        // Update the global status bar Ball Ready indicator
        const statusBallReady = document.getElementById('statusBallReady');
        if (statusBallReady) {
            if (ballReady) {
                statusBallReady.classList.add('connected');
                statusBallReady.classList.remove('disconnected');
            } else {
                statusBallReady.classList.remove('connected');
                statusBallReady.classList.add('disconnected');
            }
        }

        const ballDot = document.getElementById('ballDot');
        const svgContainer = document.querySelector('.ball-position-svg');

        // Check if position has valid numeric properties
        const hasValidPosition = position &&
                                typeof position.x === 'number' && !isNaN(position.x) &&
                                typeof position.y === 'number' && !isNaN(position.y) &&
                                typeof position.z === 'number' && !isNaN(position.z);

        if (!hasValidPosition || !ballDetected) {
            // No ball detected or invalid position - show subtle red border on entire SVG container
            ballDot.style.display = 'none';
            svgContainer.style.border = '2px solid rgba(239, 68, 68, 0.3)';
            document.getElementById('coordX').textContent = '--';
            document.getElementById('coordY').textContent = '--';
            document.getElementById('coordZ').textContent = '--';
            return;
        }

        // Reset SVG container border when ball is detected
        svgContainer.style.border = 'none';

        // Show and update ball dot position
        ballDot.style.display = 'block';

        // Convert sensor units to SVG coordinates
        // SVG viewBox: 0 0 300 400, center at 150, 200
        // The sensor reports values in 0.1mm increments (tenths of a millimeter)
        // So we need to divide by 10 to get actual mm, then scale to SVG
        const centerX = 150;
        const centerY = 200;

        // Convert from sensor units (0.1mm) to actual millimeters
        const actualX = position.x / 10;
        const actualY = position.y / 10;
        const actualZ = position.z / 10;

        // SVG visual range (as shown in labels)
        const svgVisualRange = 150; // ±150mm shown on labels

        // Actual expected coordinate range in mm (after conversion)
        // Typical values should be within ±500mm
        const actualRange = 500; // mm

        // Calculate scale to map actual mm coordinates to SVG coordinates
        const scale = svgVisualRange / actualRange;

        // X: positive right, negative left
        // Y: positive back (down in SVG), negative front (up in SVG)
        const svgX = centerX + (actualX * scale);
        const svgY = centerY + (actualY * scale);

        ballDot.setAttribute('cx', svgX);
        ballDot.setAttribute('cy', svgY);

        // Update coordinate display with actual mm values
        document.getElementById('coordX').textContent = `${actualX.toFixed(1)}mm`;
        document.getElementById('coordY').textContent = `${actualY.toFixed(1)}mm`;
        document.getElementById('coordZ').textContent = `${actualZ.toFixed(1)}mm`;

        // Calculate distance from center using actual mm values
        const distance = Math.sqrt(actualX * actualX + actualY * actualY);
        const distanceIndicator = document.getElementById('distanceIndicator');
        const distanceValue = document.getElementById('distanceValue');

        if (distanceIndicator && distanceValue) {
            distanceIndicator.style.display = 'flex';
            distanceValue.textContent = `${distance.toFixed(1)}mm`;

            // Color code based on distance
            distanceValue.classList.remove('excellent', 'good', 'poor');
            if (distance < 30) {
                distanceValue.classList.add('excellent');
            } else if (distance < 60) {
                distanceValue.classList.add('good');
            } else {
                distanceValue.classList.add('poor');
            }
        }

        // Set ball appearance based on ready state
        if (ballReady) {
            // Ball detected and ready - green fill, reset target zone
            ballDot.setAttribute('fill', '#22c55e');
            ballDot.setAttribute('stroke', '#fff');
            targetZone.setAttribute('stroke', '#22c55e');
            targetZone.setAttribute('stroke-width', '2');
            targetZone.setAttribute('stroke-opacity', '1');
        } else {
            // Ball detected but not ready - red outline only
            ballDot.setAttribute('fill', 'none');
            ballDot.setAttribute('stroke', '#ef4444');
            ballDot.setAttribute('stroke-width', '3');
            targetZone.setAttribute('stroke', '#22c55e');
            targetZone.setAttribute('stroke-width', '2');
            targetZone.setAttribute('stroke-opacity', '0.3');
        }
    }

    updateStatus(status) {
        // Update ball position visualization with detection state
        this.updateBallPosition(status.ballPosition, status.ballDetected, status.ballReady);
    }

    updateCurrentShot(ballData, clubData) {
        const placeholder = document.getElementById('monitorShotPlaceholder');
        const shotData = document.getElementById('monitorShotData');

        // Hide placeholder and show actual data
        if (placeholder) placeholder.style.display = 'none';
        if (shotData) shotData.style.display = 'block';

        // Update ball metrics
        const ballMetrics = document.getElementById('monitorBallMetrics');
        if (ballMetrics) ballMetrics.innerHTML = this.formatMetrics(ballData);

        // Update club metrics
        const clubMetrics = document.getElementById('monitorClubMetrics');
        if (clubMetrics) clubMetrics.innerHTML = this.formatMetrics(clubData);
    }

    formatMetrics(data) {
        if (!data || Object.keys(data).length === 0) {
            return '<div class="no-metrics">No data available</div>';
        }

        let html = '<div class="metrics-grid">';
        for (const [key, value] of Object.entries(data)) {
            if (value !== null && value !== undefined) {
                const label = key.replace(/([A-Z])/g, ' $1').trim();
                const displayLabel = label.charAt(0).toUpperCase() + label.slice(1);
                html += `
                    <div class="metric-item">
                        <div class="metric-label">${displayLabel}</div>
                        <div class="metric-value">${value}</div>
                    </div>
                `;
            }
        }
        html += '</div>';
        return html;
    }

    addShotToHistory(ballData, clubData) {
    }

    switchMetricsTab(tabName) {
        // Remove active from all tabs
        document.querySelectorAll('.monitor-metrics-tabs .tab-button').forEach(btn => {
            btn.classList.remove('active');
        });
        document.querySelectorAll('.monitor-metrics-content .metrics-tab').forEach(tab => {
            tab.classList.remove('active');
        });

        // Add active to selected tab
        document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');
        document.getElementById(tabName === 'monitorBall' ? 'monitorBallMetrics' : 'monitorClubMetrics').classList.add('active');
    }
}

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.app = new SquareGolfApp();
    window.shotMonitor = new ShotMonitor();

    // Status bar click-to-expand functionality
    const statusBar = document.getElementById('statusBar');
    if (statusBar) {
        statusBar.addEventListener('click', () => {
            statusBar.classList.toggle('expanded');
        });
    }
});
