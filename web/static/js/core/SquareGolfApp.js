// core/SquareGolfApp.js
import { EventBus } from './EventBus.js';
import { WebSocketService } from '../services/WebSocketService.js';
import { DeviceService } from '../services/DeviceService.js';
import { GSProService } from '../services/GSProService.js';
import { InfiniteTeesService } from '../services/InfiniteTeesService.js';
import { ApiClient } from '../services/ApiClient.js';
import { AlignmentManager } from '../features/AlignmentManager.js';
import { SettingsManager } from '../features/SettingsManager.js';
import { CameraManager } from '../features/CameraManager.js';
import { ShotMonitor } from '../features/ShotMonitor.js';
import { ToastManager } from '../ui/ToastManager.js';
import { LoadingManager } from '../ui/LoadingManager.js';
import { ScreenManager } from '../ui/ScreenManager.js';

export class SquareGolfApp {
    constructor() {
        // Core infrastructure
        this.eventBus = new EventBus();
        this.api = new ApiClient();

        // UI managers
        this.toast = new ToastManager();
        this.loading = new LoadingManager();
        this.screen = new ScreenManager(this.eventBus);

        // Services
        this.ws = new WebSocketService(this.eventBus);
        this.deviceService = new DeviceService(this.api, this.eventBus);
        this.gsproService = new GSProService(this.api, this.eventBus);
        this.infiniteTeesService = new InfiniteTeesService(this.api, this.eventBus);

        // Features
        this.alignmentManager = new AlignmentManager(this.api, this.eventBus);
        this.settingsManager = new SettingsManager(this.api, this.eventBus);
        this.cameraManager = new CameraManager(this.api, this.eventBus);
        this.shotMonitor = new ShotMonitor(this.api, this.eventBus);

        // Local state
        this.features = {};
        this.currentHandedness = 'right';
        this.alignmentExplicitlyStopped = false;
        this.alignmentPanelClosing = false;

        this.init();
    }

    init() {
        this.loadFeatures().then(() => {
            this.setupEventListeners();
            this.setupEventBusListeners();
            this.ws.connect();
            this.settingsManager.load();
        });
    }

    setupEventBusListeners() {
        // WebSocket events
        this.eventBus.on('ws:connected', () => this.updateConnectionIndicator(true));
        this.eventBus.on('ws:disconnected', () => this.updateConnectionIndicator(false));
        this.eventBus.on('ws:error', () => this.updateConnectionIndicator(false));
        this.eventBus.on('ws:message', (msg) => this.handleWebSocketMessage(msg));

        // Device events
        this.eventBus.on('device:connecting', () => {
            this.toast.info('Connection initiated...');
            this.loading.show('Connecting to device...');
        });
        this.eventBus.on('device:disconnecting', () => {
            this.toast.info('Disconnection initiated...');
        });
        this.eventBus.on('device:error', (msg) => this.toast.error(`Connection failed: ${msg}`));
        this.eventBus.on('device:status', (status) => this.updateDeviceStatus(status));

        // GSPro events
        this.eventBus.on('gspro:connecting', () => {
            this.toast.info('GSPro connection initiated...');
        });
        this.eventBus.on('gspro:disconnecting', () => {
            this.toast.info('GSPro disconnection initiated...');
        });
        this.eventBus.on('gspro:error', (msg) => this.toast.error(`GSPro: ${msg}`));
        this.eventBus.on('gspro:status', (status) => this.updateGSProStatus(status));

        // Infinite Tees events
        this.eventBus.on('infinitetees:connecting', () => {
            this.toast.info('Infinite Tees connection initiated...');
        });
        this.eventBus.on('infinitetees:disconnecting', () => {
            this.toast.info('Infinite Tees disconnection initiated...');
        });
        this.eventBus.on('infinitetees:error', (msg) => this.toast.error(`Infinite Tees: ${msg}`));
        this.eventBus.on('infinitetees:status', (status) => this.updateInfiniteTeesStatus(status));

        // Alignment events
        this.eventBus.on('alignment:saved', () => {
            this.toast.success('Calibration saved');
            this.updateAlignmentDisplay(0, false);
            this.closeAlignmentPanel();
        });
        this.eventBus.on('alignment:cancelled', () => {
            this.toast.info('Calibration cancelled');
            this.updateAlignmentDisplay(0, false);
            this.closeAlignmentPanel();
        });
        this.eventBus.on('alignment:error', (msg) => this.toast.error(msg));
        this.eventBus.on('alignment:update', ({ angle, isAligned }) => {
            this.updateAlignmentDisplay(angle, isAligned);
        });
        this.eventBus.on('alignment:handedness-changed', (handedness) => {
            this.currentHandedness = handedness;
            this.updateHandednessDisplay(handedness);
        });

        // Screen navigation events
        this.eventBus.on('screen:before-change', ({ from, to }) => {
            // Close alignment panel if leaving device screen
            if (from === 'device' && to !== 'device') {
                this.closeAlignmentPanel();
            }
        });
        this.eventBus.on('screen:changed', () => {
            // No auto-start logic needed - alignment is now triggered by button
        });

        // Settings events
        this.eventBus.on('settings:loaded', (settings) => this.applySettings(settings));
        this.eventBus.on('settings:error', (msg) => this.toast.error(`Failed to save settings: ${msg}`));

        // Camera events
        this.eventBus.on('camera:saved', () => this.toast.success('Camera settings saved successfully'));
        this.eventBus.on('camera:error', (msg) => this.toast.error(`Failed to save camera config: ${msg}`));
    }

    setupEventListeners() {
        // Navigation
        document.querySelectorAll('.nav-button').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const screen = e.target.dataset.screen || e.target.closest('[data-screen]').dataset.screen;
                this.screen.show(screen);
            });
        });

        // Status bar navigation
        document.getElementById('statusDevice')?.addEventListener('click', () => this.screen.show('device'));
        document.getElementById('statusGSPro')?.addEventListener('click', () => this.screen.show('gspro'));
        document.getElementById('statusInfiniteTees')?.addEventListener('click', () => this.screen.show('infiniteTees'));
        document.getElementById('statusBallReady')?.addEventListener('click', () => this.screen.show('device'));

        // Alignment panel controls
        document.getElementById('calibrateBtn')?.addEventListener('click', () => this.openAlignmentPanel());
        document.getElementById('closeAlignmentBtn')?.addEventListener('click', () => this.closeAlignmentPanel());

        // Device controls
        document.getElementById('connectBtn')?.addEventListener('click', () => {
            this.deviceService.connect('');
        });
        document.getElementById('disconnectBtn')?.addEventListener('click', () => {
            this.deviceService.disconnect();
        });

        // GSPro controls
        document.getElementById('gsproConnectBtn')?.addEventListener('click', () => {
            const ip = document.getElementById('gsproIP').value.trim();
            const port = parseInt(document.getElementById('gsproPort').value);
            this.gsproService.connect(ip, port);
        });
        document.getElementById('gsproDisconnectBtn')?.addEventListener('click', () => {
            this.gsproService.disconnect();
        });

        // GSPro settings
        document.getElementById('gsproIP')?.addEventListener('change', () => this.saveGSProConfig());
        document.getElementById('gsproPort')?.addEventListener('change', () => this.saveGSProConfig());
        document.getElementById('gsproAutoConnect')?.addEventListener('change', () => this.saveGSProConfig());

        // Infinite Tees controls
        document.getElementById('infiniteTeesConnectBtn')?.addEventListener('click', () => {
            const ip = document.getElementById('infiniteTeesIP').value.trim();
            const port = parseInt(document.getElementById('infiniteTeesPort').value);
            this.infiniteTeesService.connect(ip, port);
        });
        document.getElementById('infiniteTeesDisconnectBtn')?.addEventListener('click', () => {
            this.infiniteTeesService.disconnect();
        });

        // Infinite Tees settings
        document.getElementById('infiniteTeesIP')?.addEventListener('change', () => this.saveInfiniteTeesConfig());
        document.getElementById('infiniteTeesPort')?.addEventListener('change', () => this.saveInfiniteTeesConfig());
        document.getElementById('infiniteTeesAutoConnect')?.addEventListener('change', () => this.saveInfiniteTeesConfig());

        // Camera controls
        document.getElementById('cameraSaveBtn')?.addEventListener('click', () => this.cameraManager.save());

        // Alignment controls
        document.getElementById('leftHandedBtn')?.addEventListener('click', () => this.handleHandednessChange('left'));
        document.getElementById('rightHandedBtn')?.addEventListener('click', () => this.handleHandednessChange('right'));
        document.getElementById('saveAlignmentBtn')?.addEventListener('click', () => {
            this.alignmentExplicitlyStopped = true;
            this.alignmentManager.save();
        });
        document.getElementById('cancelAlignmentBtn')?.addEventListener('click', () => {
            this.alignmentExplicitlyStopped = true;
            this.alignmentManager.cancel();
        });

        // Settings controls
        document.querySelectorAll('input[name="spinMode"]').forEach(radio => {
            radio.addEventListener('change', () => this.saveSettings());
        });
    }

    async handleHandednessChange(handedness) {
        const result = await this.alignmentManager.setHandedness(handedness);

        if (result.success && document.getElementById('alignmentPanel')?.classList.contains('open')) {
            // Restart alignment with new handedness
            await this.alignmentManager.stop();
            await new Promise(resolve => setTimeout(resolve, 100));
            await this.alignmentManager.start();
        }
    }

    handleWebSocketMessage(message) {
        switch (message.type) {
            case 'deviceStatus':
                this.deviceService.updateStatus(message.data);
                break;
            case 'gsproStatus':
                this.gsproService.updateStatus(message.data);
                break;
            case 'infiniteTeesStatus':
                this.infiniteTeesService.updateStatus(message.data);
                break;
            case 'cameraConfig':
                this.cameraManager.updateConfig(message.data);
                break;
            case 'alignmentData':
                if (message.data) {
                    this.alignmentManager.updateDisplay(
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
        const statusWebSocket = document.getElementById('statusWebSocket');
        if (statusWebSocket) {
            if (connected) {
                statusWebSocket.classList.add('connected');
                statusWebSocket.classList.remove('disconnected');
            } else {
                statusWebSocket.classList.remove('connected');
                statusWebSocket.classList.add('disconnected');
            }
        }
    }

    updateDeviceConnectionIndicator(deviceStatus) {
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

    updateDeviceStatus(status) {
        // Update the main navigation device connection indicator
        this.updateDeviceConnectionIndicator(status.connectionStatus);

        // Update connection status display
        const statusElement = document.getElementById('deviceStatus');
        const connectionStatusEl = document.getElementById('deviceConnectionStatus');
        const errorElement = document.getElementById('deviceError');
        const connectBtn = document.getElementById('connectBtn');
        const disconnectBtn = document.getElementById('disconnectBtn');
        const calibrateBtn = document.getElementById('calibrateBtn');
        const deviceInfoInline = document.getElementById('deviceInfoInline');

        // Update icon and status in the new header
        if (connectionStatusEl) {
            connectionStatusEl.className = 'device-connection-status ' + status.connectionStatus;
            const icon = connectionStatusEl.querySelector('.material-icons');
            if (icon) {
                switch (status.connectionStatus) {
                    case 'connected':
                        icon.textContent = 'bluetooth_connected';
                        break;
                    case 'scanning':
                        icon.textContent = 'bluetooth_searching';
                        break;
                    case 'connecting':
                        icon.textContent = 'bluetooth_searching';
                        break;
                    default:
                        icon.textContent = 'bluetooth_disabled';
                }
            }
        }

        if (statusElement) {
            switch (status.connectionStatus) {
                case 'connected':
                    statusElement.textContent = 'Connected';
                    if (connectBtn) connectBtn.disabled = true;
                    if (disconnectBtn) disconnectBtn.disabled = false;
                    if (calibrateBtn) calibrateBtn.style.display = 'flex';
                    if (deviceInfoInline) deviceInfoInline.style.display = 'flex';
                    if (errorElement) errorElement.style.display = 'none';
                    this.loading.hide();
                    break;
                case 'scanning':
                    statusElement.textContent = 'Scanning...';
                    if (connectBtn) connectBtn.disabled = true;
                    if (disconnectBtn) disconnectBtn.disabled = false;
                    if (calibrateBtn) calibrateBtn.style.display = 'none';
                    if (deviceInfoInline) deviceInfoInline.style.display = 'none';
                    if (errorElement) errorElement.style.display = 'none';
                    this.loading.show('Scanning for device...');
                    break;
                case 'connecting':
                    statusElement.textContent = 'Connecting...';
                    if (connectBtn) connectBtn.disabled = true;
                    if (disconnectBtn) disconnectBtn.disabled = false;
                    if (calibrateBtn) calibrateBtn.style.display = 'none';
                    if (deviceInfoInline) deviceInfoInline.style.display = 'none';
                    if (errorElement) errorElement.style.display = 'none';
                    this.loading.show('Connecting to device...');
                    break;
                case 'disconnected':
                    statusElement.textContent = 'Disconnected';
                    if (connectBtn) connectBtn.disabled = false;
                    if (disconnectBtn) disconnectBtn.disabled = true;
                    if (calibrateBtn) calibrateBtn.style.display = 'none';
                    if (deviceInfoInline) deviceInfoInline.style.display = 'none';
                    if (errorElement) errorElement.style.display = 'none';
                    this.loading.hide();
                    break;
                case 'error':
                    statusElement.textContent = 'Error';
                    if (connectBtn) connectBtn.disabled = false;
                    if (disconnectBtn) disconnectBtn.disabled = false;
                    if (calibrateBtn) calibrateBtn.style.display = 'none';
                    if (deviceInfoInline) deviceInfoInline.style.display = 'none';
                    if (status.lastError && errorElement) {
                        errorElement.textContent = status.lastError;
                        errorElement.style.display = 'block';
                    }
                    this.loading.hide();
                    break;
            }
        }

        // Update device information
        if (status.deviceName) {
            const nameElement = document.getElementById('connectedDeviceName');
            if (nameElement) nameElement.textContent = status.deviceName;
        }

        if (status.batteryLevel !== null) {
            const batteryElement = document.getElementById('batteryLevel');
            if (batteryElement) {
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
        }

        // Update version information
        const firmwareElement = document.getElementById('firmwareVersion');
        if (firmwareElement) {
            firmwareElement.textContent = status.firmwareVersion !== null ? status.firmwareVersion : '-';
        }

        const launcherElement = document.getElementById('launcherVersion');
        if (launcherElement) {
            launcherElement.textContent = status.launcherVersion !== null ? status.launcherVersion : '-';
        }

        const mmiElement = document.getElementById('mmiVersion');
        if (mmiElement) {
            mmiElement.textContent = status.mmiVersion !== null ? status.mmiVersion : '-';
        }

        // Update club info
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

        // Update handedness
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

        // Update Shot Monitor
        this.shotMonitor.updateStatus(status);

        // If we have new shot data, update current shot and add to history
        if (status.lastBallMetrics && Object.keys(status.lastBallMetrics).length > 0) {
            this.shotMonitor.updateCurrentShot(status.lastBallMetrics, status.lastClubMetrics);
            this.shotMonitor.addShotToHistory(status.lastBallMetrics, status.lastClubMetrics || {});
        }

        // Update alignment display if alignment data is present
        if (status.isAligning && typeof status.alignmentAngle === 'number') {
            this.updateAlignmentDisplay(status.alignmentAngle, status.isAligned || false);
        }
    }

    openAlignmentPanel() {
        const panel = document.getElementById('alignmentPanel');
        const overlay = document.getElementById('alignmentOverlay');

        if (panel) {
            panel.style.display = 'flex';
            requestAnimationFrame(() => {
                panel.classList.add('open');
            });
        }
        if (overlay) {
            overlay.style.display = 'block';
            requestAnimationFrame(() => {
                overlay.classList.add('open');
            });
            overlay.addEventListener('click', () => this.closeAlignmentPanel(), { once: true });
        }

        this.alignmentManager.start();
    }

    closeAlignmentPanel() {
        if (this.alignmentPanelClosing) return;
        this.alignmentPanelClosing = true;

        const panel = document.getElementById('alignmentPanel');
        const overlay = document.getElementById('alignmentOverlay');

        if (panel) {
            panel.classList.remove('open');
            setTimeout(() => {
                panel.style.display = 'none';
            }, 300);
        }
        if (overlay) {
            overlay.classList.remove('open');
            setTimeout(() => {
                overlay.style.display = 'none';
            }, 300);
        }

        // Only stop alignment (no toast) when panel is closed via X or overlay
        // Cancel button handles its own toast via the cancelled event
        if (!this.alignmentExplicitlyStopped) {
            this.alignmentManager.stop();
        }
        this.alignmentExplicitlyStopped = false;

        setTimeout(() => {
            this.alignmentPanelClosing = false;
        }, 350);
    }

    updateGSProStatus(status) {
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

        if (statusElement) {
            statusElement.className = 'status-value';
            statusElement.classList.add(status.connectionStatus);

            switch (status.connectionStatus) {
                case 'connected':
                    statusElement.textContent = 'Connected';
                    if (connectBtn) connectBtn.disabled = true;
                    if (disconnectBtn) disconnectBtn.disabled = false;
                    if (ipField) ipField.disabled = true;
                    if (portField) portField.disabled = true;
                    if (errorElement) errorElement.style.display = 'none';
                    break;
                case 'connecting':
                    statusElement.textContent = 'Connecting...';
                    if (connectBtn) connectBtn.disabled = true;
                    if (disconnectBtn) disconnectBtn.disabled = true;
                    if (ipField) ipField.disabled = true;
                    if (portField) portField.disabled = true;
                    if (errorElement) errorElement.style.display = 'none';
                    break;
                case 'disconnected':
                    statusElement.textContent = 'Disconnected';
                    if (connectBtn) connectBtn.disabled = false;
                    if (disconnectBtn) disconnectBtn.disabled = true;
                    if (ipField) ipField.disabled = false;
                    if (portField) portField.disabled = false;
                    if (errorElement) errorElement.style.display = 'none';
                    break;
                case 'error':
                    statusElement.textContent = 'Error';
                    if (connectBtn) connectBtn.disabled = false;
                    if (disconnectBtn) disconnectBtn.disabled = true;
                    if (ipField) ipField.disabled = false;
                    if (portField) portField.disabled = false;
                    if (status.lastError && errorElement) {
                        errorElement.textContent = status.lastError;
                        errorElement.style.display = 'block';
                    }
                    break;
            }
        }
    }

    async saveGSProConfig() {
        const ip = document.getElementById('gsproIP')?.value.trim();
        const port = parseInt(document.getElementById('gsproPort')?.value);
        const autoConnect = document.getElementById('gsproAutoConnect')?.checked;

        await this.gsproService.saveConfig(ip, port, autoConnect);
    }

    updateInfiniteTeesStatus(status) {
        // Update the global status bar IT indicator
        const statusIT = document.getElementById('statusInfiniteTees');
        if (statusIT) {
            if (status.connectionStatus === 'connected') {
                statusIT.classList.add('connected');
                statusIT.classList.remove('disconnected');
            } else {
                statusIT.classList.remove('connected');
                statusIT.classList.add('disconnected');
            }
        }

        const statusElement = document.getElementById('infiniteTeesStatus');
        const errorElement = document.getElementById('infiniteTeesError');
        const connectBtn = document.getElementById('infiniteTeesConnectBtn');
        const disconnectBtn = document.getElementById('infiniteTeesDisconnectBtn');
        const ipField = document.getElementById('infiniteTeesIP');
        const portField = document.getElementById('infiniteTeesPort');

        if (statusElement) {
            statusElement.className = 'status-value';
            statusElement.classList.add(status.connectionStatus);

            switch (status.connectionStatus) {
                case 'connected':
                    statusElement.textContent = 'Connected';
                    if (connectBtn) connectBtn.disabled = true;
                    if (disconnectBtn) disconnectBtn.disabled = false;
                    if (ipField) ipField.disabled = true;
                    if (portField) portField.disabled = true;
                    if (errorElement) errorElement.style.display = 'none';
                    break;
                case 'connecting':
                    statusElement.textContent = 'Connecting...';
                    if (connectBtn) connectBtn.disabled = true;
                    if (disconnectBtn) disconnectBtn.disabled = true;
                    if (ipField) ipField.disabled = true;
                    if (portField) portField.disabled = true;
                    if (errorElement) errorElement.style.display = 'none';
                    break;
                case 'disconnected':
                    statusElement.textContent = 'Disconnected';
                    if (connectBtn) connectBtn.disabled = false;
                    if (disconnectBtn) disconnectBtn.disabled = true;
                    if (ipField) ipField.disabled = false;
                    if (portField) portField.disabled = false;
                    if (errorElement) errorElement.style.display = 'none';
                    break;
                case 'error':
                    statusElement.textContent = 'Error';
                    if (connectBtn) connectBtn.disabled = false;
                    if (disconnectBtn) disconnectBtn.disabled = true;
                    if (ipField) ipField.disabled = false;
                    if (portField) portField.disabled = false;
                    if (status.lastError && errorElement) {
                        errorElement.textContent = status.lastError;
                        errorElement.style.display = 'block';
                    }
                    break;
            }
        }
    }

    async saveInfiniteTeesConfig() {
        const ip = document.getElementById('infiniteTeesIP')?.value.trim();
        const port = parseInt(document.getElementById('infiniteTeesPort')?.value);
        const autoConnect = document.getElementById('infiniteTeesAutoConnect')?.checked;

        await this.infiniteTeesService.saveConfig(ip, port, autoConnect);
    }

    updateAlignmentDisplay(angle, isAligned) {
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
            if (iconElement) iconElement.textContent = '✅';
            if (textElement) textElement.textContent = 'Aligned!';
        } else {
            statusElement.classList.add('not-aligned');
            if (iconElement) iconElement.textContent = '⚠️';
            if (textElement) textElement.textContent = 'Not aligned';
        }

        // Rotate compass pointer
        if (pointerElement) {
            pointerElement.setAttribute('transform', `rotate(${angle} 100 100)`);
        }
    }

    updateHandednessDisplay(handedness) {
        const leftBtn = document.getElementById('leftHandedBtn');
        const rightBtn = document.getElementById('rightHandedBtn');

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

    async loadFeatures() {
        try {
            const response = await this.api.get('/api/features');
            if (response.ok) {
                this.features = await response.json();
                this.applyFeatures();
            }
        } catch (error) {
            console.error('Failed to load features:', error);
        }
    }

    applyFeatures() {
        // Feature flags applied here if needed
    }

    applySettings(settings) {
        const spinMode = settings.spinMode || 'advanced';
        const spinModeRadio = document.querySelector(`input[name="spinMode"][value="${spinMode}"]`);
        if (spinModeRadio) spinModeRadio.checked = true;

        const gsproIP = document.getElementById('gsproIP');
        const gsproPort = document.getElementById('gsproPort');
        const gsproAutoConnect = document.getElementById('gsproAutoConnect');
        if (gsproIP) gsproIP.value = settings.gsproIP || '127.0.0.1';
        if (gsproPort) gsproPort.value = settings.gsproPort || 921;
        if (gsproAutoConnect) gsproAutoConnect.checked = settings.gsproAutoConnect || false;

        const itIP = document.getElementById('infiniteTeesIP');
        const itPort = document.getElementById('infiniteTeesPort');
        const itAutoConnect = document.getElementById('infiniteTeesAutoConnect');
        if (itIP) itIP.value = settings.infiniteTeesIP || '127.0.0.1';
        if (itPort) itPort.value = settings.infiniteTeesPort || 999;
        if (itAutoConnect) itAutoConnect.checked = settings.infiniteTeesAutoConnect || false;
    }

    async saveSettings() {
        const spinMode = document.querySelector('input[name="spinMode"]:checked')?.value;
        await this.settingsManager.save({ spinMode });
    }
}
