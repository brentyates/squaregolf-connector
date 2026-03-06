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
import { ScreenManager } from '../ui/ScreenManager.js';

export class SquareGolfApp {
    constructor() {
        // Core infrastructure
        this.eventBus = new EventBus();
        this.api = new ApiClient();

        // UI managers
        this.toast = new ToastManager();
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
        this.alignmentInFlight = false;
        this.pendingDeviceAction = null;
        this.alignmentCloseDelayMs = 300;

        this.init();
    }

    $(id) {
        return document.getElementById(id);
    }

    $$(selector) {
        return [...document.querySelectorAll(selector)];
    }

    bind(id, eventName, handler) {
        this.$(id)?.addEventListener(eventName, handler);
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
            this.pendingDeviceAction = 'manual-connect';
        });
        this.eventBus.on('device:disconnecting', () => {
            this.toast.info('Disconnection initiated...');
            this.pendingDeviceAction = 'manual-disconnect';
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
            this.setAlignmentBusy(false);
            this.toast.success('Calibration saved');
            this.setAlignmentError('');
            this.updateAlignmentDisplay(0, false);
            this.closeAlignmentPanel();
        });
        this.eventBus.on('alignment:cancelled', () => {
            this.setAlignmentBusy(false);
            this.toast.info('Calibration cancelled');
            this.setAlignmentError('');
            this.updateAlignmentDisplay(0, false);
            this.closeAlignmentPanel();
        });
        this.eventBus.on('alignment:error', (msg) => {
            this.setAlignmentBusy(false);
            this.setAlignmentError(msg);
            this.toast.error(msg);
        });
        this.eventBus.on('alignment:update', ({ angle, isAligned }) => {
            this.updateAlignmentDisplay(angle, isAligned);
        });
        this.eventBus.on('alignment:started', () => this.setAlignmentBusy(false));
        this.eventBus.on('alignment:stopped', () => this.setAlignmentBusy(false));
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
        this.screen.navButtons.forEach((button) => {
            button.addEventListener('click', ({ currentTarget }) => {
                this.screen.show(currentTarget.dataset.screen);
            });
        });

        // Status bar navigation
        this.bind('statusDevice', 'click', () => this.screen.show('device'));
        this.bind('statusGSPro', 'click', () => this.screen.show('gspro'));
        this.bind('statusInfiniteTees', 'click', () => this.screen.show('infiniteTees'));
        this.bind('statusBallReady', 'click', () => this.screen.show('device'));

        // Alignment panel controls
        this.bind('calibrateBtn', 'click', () => this.openAlignmentPanel());
        this.bind('closeAlignmentBtn', 'click', () => this.closeAlignmentPanel());
        this.bind('retryAlignmentBtn', 'click', () => this.retryAlignment());

        // Device controls
        this.bind('connectBtn', 'click', () => {
            this.pendingDeviceAction = 'manual-connect';
            this.deviceService.connect('');
        });
        this.bind('disconnectBtn', 'click', () => {
            this.pendingDeviceAction = 'manual-disconnect';
            this.deviceService.disconnect();
        });

        // GSPro controls
        this.bind('gsproConnectBtn', 'click', () => {
            const config = this.getConnectionConfig('gspro', true);
            if (!config) return;
            const { ip, port } = config;
            this.gsproService.connect(ip, port);
        });
        this.bind('gsproDisconnectBtn', 'click', () => {
            this.gsproService.disconnect();
        });

        // GSPro settings
        this.bind('gsproIP', 'change', () => this.saveGSProConfig());
        this.bind('gsproPort', 'change', () => this.saveGSProConfig());
        this.bind('gsproAutoConnect', 'change', () => this.saveGSProConfig());
        this.bind('gsproIP', 'input', () => this.clearFieldError('gsproIP'));
        this.bind('gsproPort', 'input', () => this.clearFieldError('gsproPort'));

        // Infinite Tees controls
        this.bind('infiniteTeesConnectBtn', 'click', () => {
            const config = this.getConnectionConfig('infiniteTees', true);
            if (!config) return;
            const { ip, port } = config;
            this.infiniteTeesService.connect(ip, port);
        });
        this.bind('infiniteTeesDisconnectBtn', 'click', () => {
            this.infiniteTeesService.disconnect();
        });

        // Infinite Tees settings
        this.bind('infiniteTeesIP', 'change', () => this.saveInfiniteTeesConfig());
        this.bind('infiniteTeesPort', 'change', () => this.saveInfiniteTeesConfig());
        this.bind('infiniteTeesAutoConnect', 'change', () => this.saveInfiniteTeesConfig());
        this.bind('infiniteTeesIP', 'input', () => this.clearFieldError('infiniteTeesIP'));
        this.bind('infiniteTeesPort', 'input', () => this.clearFieldError('infiniteTeesPort'));

        // Camera controls
        this.bind('cameraSaveBtn', 'click', () => this.cameraManager.save());

        // Alignment controls
        this.bind('leftHandedBtn', 'click', () => this.handleHandednessChange('left'));
        this.bind('rightHandedBtn', 'click', () => this.handleHandednessChange('right'));
        this.bind('saveAlignmentBtn', 'click', () => {
            this.alignmentExplicitlyStopped = true;
            this.alignmentManager.save();
        });
        this.bind('cancelAlignmentBtn', 'click', () => {
            this.alignmentExplicitlyStopped = true;
            this.alignmentManager.cancel();
        });

        // Settings controls
        this.$$('input[name="spinMode"]').forEach((radio) => {
            radio.addEventListener('change', () => this.saveSettings());
        });
    }

    async handleHandednessChange(handedness) {
        const result = await this.alignmentManager.setHandedness(handedness);

        if (result.success && this.$('alignmentPanel')?.classList.contains('open')) {
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
        this.updateBinaryIndicator('statusWebSocket', connected);
    }

    updateDeviceConnectionIndicator(deviceStatus) {
        this.updateBinaryIndicator('statusDevice', deviceStatus === 'connected');
    }

    setHidden(element, shouldHide) {
        if (!element) return;
        element.classList.toggle('hidden', shouldHide);
    }

    updateBinaryIndicator(elementId, isConnected) {
        const element = this.$(elementId);
        if (!element) return;

        element.classList.toggle('connected', isConnected);
        element.classList.toggle('disconnected', !isConnected);
    }

    updateGlobalConnectionIndicator(elementId, connectionStatus) {
        this.updateBinaryIndicator(elementId, connectionStatus === 'connected');
    }

    updateConnectionPanel({
        status,
        statusElementId,
        errorElementId,
        connectBtnId,
        disconnectBtnId,
        ipFieldId,
        portFieldId
    }) {
        const statusElement = this.$(statusElementId);
        const errorElement = this.$(errorElementId);
        const connectBtn = this.$(connectBtnId);
        const disconnectBtn = this.$(disconnectBtnId);
        const ipField = this.$(ipFieldId);
        const portField = this.$(portFieldId);

        if (statusElement) {
            statusElement.className = 'status-value';
            statusElement.classList.add(status.connectionStatus);
        }

        const isConnected = status.connectionStatus === 'connected';
        const isConnecting = status.connectionStatus === 'connecting';
        const isDisconnected = status.connectionStatus === 'disconnected';
        const isError = status.connectionStatus === 'error';
        const statusText = {
            connected: 'Connected',
            connecting: 'Connecting...',
            disconnected: 'Disconnected',
            error: 'Error'
        };

        if (statusElement) {
            statusElement.textContent = statusText[status.connectionStatus] || 'Disconnected';
        }

        if (connectBtn) connectBtn.disabled = isConnected || isConnecting;
        if (disconnectBtn) disconnectBtn.disabled = !isConnected;
        if (ipField) ipField.disabled = isConnected || isConnecting;
        if (portField) portField.disabled = isConnected || isConnecting;

        if (errorElement) {
            if (isError && status.lastError) {
                errorElement.textContent = status.lastError;
                this.setHidden(errorElement, false);
            } else {
                this.setHidden(errorElement, true);
            }
        }
    }

    updateDeviceControls({ canConnect, canDisconnect, showCalibrate, showDeviceInfo, errorMessage = '' }) {
        const connectBtn = this.$('connectBtn');
        const disconnectBtn = this.$('disconnectBtn');
        const calibrateBtn = this.$('calibrateBtn');
        const deviceInfoInline = this.$('deviceInfoInline');
        const errorElement = this.$('deviceError');

        if (connectBtn) connectBtn.disabled = !canConnect;
        if (disconnectBtn) disconnectBtn.disabled = !canDisconnect;

        this.setHidden(calibrateBtn, !showCalibrate);
        this.setHidden(deviceInfoInline, !showDeviceInfo);

        if (errorElement) {
            if (errorMessage) {
                errorElement.textContent = errorMessage;
                this.setHidden(errorElement, false);
            } else {
                this.setHidden(errorElement, true);
            }
        }
    }

    showSlidingPanel(panelId, overlayId) {
        const panel = this.$(panelId);
        const overlay = this.$(overlayId);

        if (panel) {
            panel.classList.remove('hidden');
        }

        if (overlay) {
            overlay.classList.remove('hidden');
        }

        requestAnimationFrame(() => {
            if (panel) {
                panel.classList.add('open');
            }
            if (overlay) {
                overlay.classList.add('open');
            }
        });

        return { panel, overlay };
    }

    hideSlidingPanel(panelId, overlayId, delayMs) {
        const panel = this.$(panelId);
        const overlay = this.$(overlayId);
        const finalizeHide = (element) => {
            this.setHidden(element, true);
        };
        const hideAfterTransition = (element) => {
            if (!element) return;

            let finalized = false;
            const complete = () => {
                if (finalized) return;
                finalized = true;
                element.removeEventListener('transitionend', onTransitionEnd);
                finalizeHide(element);
            };
            const onTransitionEnd = (event) => {
                if (event.target !== element) return;
                complete();
            };

            element.addEventListener('transitionend', onTransitionEnd);
            window.setTimeout(complete, delayMs + 50);
        };

        if (panel) {
            panel.classList.remove('open');
            hideAfterTransition(panel);
        }

        if (overlay) {
            overlay.classList.remove('open');
            hideAfterTransition(overlay);
        }
    }

    setTextContent(elementId, value, fallback = '-') {
        const element = this.$(elementId);
        if (element) {
            element.textContent = value ?? fallback;
        }
    }

    updateOptionalDeviceInfo({ itemId, valueId, value, formatter = (entry) => entry }) {
        const itemElement = this.$(itemId);
        const valueElement = this.$(valueId);
        const hasValue = value !== null && value !== undefined && value !== '';

        this.setHidden(itemElement, !hasValue);
        if (valueElement) {
            valueElement.textContent = hasValue ? formatter(value) : '-';
        }
    }

    updateBatteryDisplay(level) {
        const batteryElement = this.$('batteryLevel');
        const batteryIconElement = this.$('batteryIcon');
        if (!batteryElement || !batteryIconElement) return;

        if (typeof level !== 'number') {
            batteryIconElement.textContent = '—';
            batteryIconElement.className = 'battery-icon';
            batteryElement.textContent = '—';
            return;
        }

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

        batteryIconElement.textContent = icon;
        batteryIconElement.className = `battery-icon ${className}`;
        batteryElement.textContent = `${level}%`;
    }

    updateVersionDisplay(status) {
        this.setTextContent('firmwareVersion', status.firmwareVersion !== null ? status.firmwareVersion : null);
        this.setTextContent('launcherVersion', status.launcherVersion !== null ? status.launcherVersion : null);
        this.setTextContent('mmiVersion', status.mmiVersion !== null ? status.mmiVersion : null);
    }

    updateDeviceStatus(status) {
        // Update the main navigation device connection indicator
        this.updateDeviceConnectionIndicator(status.connectionStatus);
        this.updateDeviceHeaderStatus(status);

        switch (status.connectionStatus) {
            case 'connected':
                this.updateDeviceControls({
                    canConnect: false,
                    canDisconnect: true,
                    showCalibrate: true,
                    showDeviceInfo: true
                });
                this.pendingDeviceAction = null;
                break;
            case 'scanning':
                this.updateDeviceControls({
                    canConnect: false,
                    canDisconnect: true,
                    showCalibrate: false,
                    showDeviceInfo: false
                });
                break;
            case 'connecting':
                this.updateDeviceControls({
                    canConnect: false,
                    canDisconnect: true,
                    showCalibrate: false,
                    showDeviceInfo: false
                });
                break;
            case 'disconnected':
                this.updateDeviceControls({
                    canConnect: true,
                    canDisconnect: false,
                    showCalibrate: false,
                    showDeviceInfo: false
                });
                this.pendingDeviceAction = null;
                break;
            case 'error':
                this.updateDeviceControls({
                    canConnect: true,
                    canDisconnect: true,
                    showCalibrate: false,
                    showDeviceInfo: false,
                    errorMessage: status.lastError || ''
                });
                this.pendingDeviceAction = null;
                break;
        }

        this.setTextContent('connectedDeviceName', status.deviceName || 'SquareGolf');
        this.updateBatteryDisplay(status.batteryLevel);
        this.updateVersionDisplay(status);
        this.updateOptionalDeviceInfo({
            itemId: 'clubItem',
            valueId: 'clubValue',
            value: status.club,
            formatter: (club) => club.regularCode || club.name
        });

        const handedness = status.handedness === null ? null : (status.handedness === 0 ? 'Right' : 'Left');
        this.updateOptionalDeviceInfo({
            itemId: 'handednessItem',
            valueId: 'handednessValue',
            value: handedness
        });

        if (handedness) {
            this.currentHandedness = handedness.toLowerCase();
            this.updateHandednessDisplay(this.currentHandedness);
        }

        // Update Shot Monitor
        this.shotMonitor.updateStatus(status);

        // If we have new shot data, update current shot
        if (status.lastBallMetrics && Object.keys(status.lastBallMetrics).length > 0) {
            this.shotMonitor.updateCurrentShot(status.lastBallMetrics, status.lastClubMetrics);
        }

        // Update alignment display if alignment data is present
        if (status.isAligning && typeof status.alignmentAngle === 'number') {
            this.updateAlignmentDisplay(status.alignmentAngle, status.isAligned || false);
        }
    }

    openAlignmentPanel() {
        if (this.alignmentInFlight || this.$('alignmentPanel')?.classList.contains('open')) {
            return;
        }

        this.setAlignmentError('');
        const { overlay } = this.showSlidingPanel('alignmentPanel', 'alignmentOverlay');

        if (overlay) {
            overlay.addEventListener('click', () => this.closeAlignmentPanel(), { once: true });
        }

        this.setAlignmentBusy(true);
        this.alignmentManager.start();
    }

    closeAlignmentPanel() {
        if (this.alignmentPanelClosing) return;
        this.alignmentPanelClosing = true;
        this.hideSlidingPanel('alignmentPanel', 'alignmentOverlay', this.alignmentCloseDelayMs);

        // Only stop alignment (no toast) when panel is closed via X or overlay
        // Cancel button handles its own toast via the cancelled event
        if (!this.alignmentExplicitlyStopped) {
            this.setAlignmentBusy(true);
            this.alignmentManager.stop();
        }
        this.alignmentExplicitlyStopped = false;

        setTimeout(() => {
            this.alignmentPanelClosing = false;
        }, this.alignmentCloseDelayMs + 50);
    }

    updateDeviceHeaderStatus(status) {
        const container = this.$('deviceConnectionStatus');
        const icon = container?.querySelector('.material-icons');
        const text = this.$('deviceConnectionText');
        const hint = this.$('deviceConnectionHint');

        if (!container || !icon || !text || !hint) return;

        container.classList.remove('connected', 'connecting', 'disconnected', 'error');

        const isManualAction = this.pendingDeviceAction === 'manual-connect' || this.pendingDeviceAction === 'manual-disconnect';
        const stateCopy = {
            connected: {
                stateClass: 'connected',
                icon: 'bluetooth_connected',
                text: 'Connected',
                hint: status.deviceName ? `Connected to ${status.deviceName}.` : 'Device ready.'
            },
            scanning: {
                stateClass: 'connecting',
                icon: 'bluetooth_searching',
                text: 'Scanning',
                hint: isManualAction ? 'Looking for a launch monitor...' : 'Auto-connect is looking for a launch monitor in the background.'
            },
            connecting: {
                stateClass: 'connecting',
                icon: 'sync',
                text: 'Connecting',
                hint: isManualAction ? 'Opening a device connection...' : 'Auto-connect is opening the device connection in the background.'
            },
            error: {
                stateClass: 'error',
                icon: 'error',
                text: 'Connection error',
                hint: status.lastError || 'The last connection attempt failed.'
            },
            disconnected: {
                stateClass: 'disconnected',
                icon: 'bluetooth_disabled',
                text: 'Disconnected',
                hint: 'Auto-connect runs in the background.'
            }
        };
        const copy = stateCopy[status.connectionStatus] || stateCopy.disconnected;

        container.classList.add(copy.stateClass);
        icon.textContent = copy.icon;
        text.textContent = copy.text;
        hint.textContent = copy.hint;
    }

    setAlignmentBusy(isBusy) {
        this.alignmentInFlight = isBusy;

        ['saveAlignmentBtn', 'cancelAlignmentBtn', 'leftHandedBtn', 'rightHandedBtn', 'retryAlignmentBtn'].forEach(id => {
            const element = this.$(id);
            if (element) {
                element.disabled = isBusy;
            }
        });
    }

    setAlignmentError(message) {
        const errorElement = this.$('alignmentError');
        const textElement = errorElement?.querySelector('.error-text');

        if (!errorElement || !textElement) return;

        if (message) {
            textElement.textContent = message;
            this.setHidden(errorElement, false);
        } else {
            textElement.textContent = '';
            this.setHidden(errorElement, true);
        }
    }

    retryAlignment() {
        if (this.alignmentInFlight) return;
        this.setAlignmentError('');
        this.setAlignmentBusy(true);
        this.alignmentManager.start();
    }

    updateGSProStatus(status) {
        this.updateGlobalConnectionIndicator('statusGSPro', status.connectionStatus);
        this.updateConnectionPanel({
            status,
            statusElementId: 'gsproStatus',
            errorElementId: 'gsproError',
            connectBtnId: 'gsproConnectBtn',
            disconnectBtnId: 'gsproDisconnectBtn',
            ipFieldId: 'gsproIP',
            portFieldId: 'gsproPort'
        });
    }

    async saveGSProConfig() {
        const config = this.getConnectionConfig('gspro', false);
        if (!config) return;

        const { ip, port } = config;
        const autoConnect = this.$('gsproAutoConnect')?.checked;

        await this.gsproService.saveConfig(ip, port, autoConnect);
    }

    updateInfiniteTeesStatus(status) {
        this.updateGlobalConnectionIndicator('statusInfiniteTees', status.connectionStatus);
        this.updateConnectionPanel({
            status,
            statusElementId: 'infiniteTeesStatus',
            errorElementId: 'infiniteTeesError',
            connectBtnId: 'infiniteTeesConnectBtn',
            disconnectBtnId: 'infiniteTeesDisconnectBtn',
            ipFieldId: 'infiniteTeesIP',
            portFieldId: 'infiniteTeesPort'
        });
    }

    async saveInfiniteTeesConfig() {
        const config = this.getConnectionConfig('infiniteTees', false);
        if (!config) return;

        const { ip, port } = config;
        const autoConnect = this.$('infiniteTeesAutoConnect')?.checked;

        await this.infiniteTeesService.saveConfig(ip, port, autoConnect);
    }

    getConnectionConfig(prefix, notifyOnError) {
        const ipField = this.$(`${prefix}IP`);
        const portField = this.$(`${prefix}Port`);

        if (!ipField || !portField) return null;

        const ip = ipField.value.trim();
        const port = Number.parseInt(portField.value, 10);
        let valid = true;

        if (!ip) {
            this.setFieldError(ipField.id);
            valid = false;
        } else {
            this.clearFieldError(ipField.id);
        }

        if (!Number.isInteger(port) || port < 1 || port > 65535) {
            this.setFieldError(portField.id);
            valid = false;
        } else {
            this.clearFieldError(portField.id);
        }

        if (!valid) {
            if (notifyOnError) {
                this.toast.error('Enter a valid host and port before connecting.');
            }
            return null;
        }

        return { ip, port };
    }

    setFieldError(fieldId) {
        const field = this.$(fieldId);
        if (field) {
            field.classList.add('error');
        }
    }

    clearFieldError(fieldId) {
        const field = this.$(fieldId);
        if (field) {
            field.classList.remove('error');
        }
    }

    updateAlignmentDisplay(angle, isAligned) {
        const angleElement = this.$('alignmentAngle');
        const directionElement = this.$('alignmentDirection');
        const statusElement = this.$('alignedStatus');
        const pointerElement = this.$('aimPointer');

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
        const leftBtn = this.$('leftHandedBtn');
        const rightBtn = this.$('rightHandedBtn');

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
        const cameraCard = this.$('cameraSettingsCard');
        const cameraSaveBtn = this.$('cameraSaveBtn');
        const cameraURL = this.$('cameraURL');
        const cameraEnabled = this.$('cameraEnabled');
        const cameraSupported = Boolean(this.features.externalCamera);

        this.setHidden(cameraCard, !cameraSupported);

        if (cameraSaveBtn) cameraSaveBtn.disabled = !cameraSupported;
        if (cameraURL) cameraURL.disabled = !cameraSupported;
        if (cameraEnabled) cameraEnabled.disabled = !cameraSupported;
    }

    applySettings(settings) {
        const spinMode = settings.spinMode || 'advanced';
        const spinModeRadio = document.querySelector(`input[name="spinMode"][value="${spinMode}"]`);
        if (spinModeRadio) spinModeRadio.checked = true;

        const gsproIP = this.$('gsproIP');
        const gsproPort = this.$('gsproPort');
        const gsproAutoConnect = this.$('gsproAutoConnect');
        if (gsproIP) gsproIP.value = settings.gsproIP || '127.0.0.1';
        if (gsproPort) gsproPort.value = settings.gsproPort || 921;
        if (gsproAutoConnect) gsproAutoConnect.checked = settings.gsproAutoConnect || false;

        const itIP = this.$('infiniteTeesIP');
        const itPort = this.$('infiniteTeesPort');
        const itAutoConnect = this.$('infiniteTeesAutoConnect');
        if (itIP) itIP.value = settings.infiniteTeesIP || '127.0.0.1';
        if (itPort) itPort.value = settings.infiniteTeesPort || 999;
        if (itAutoConnect) itAutoConnect.checked = settings.infiniteTeesAutoConnect || false;
    }

    async saveSettings() {
        const spinMode = document.querySelector('input[name="spinMode"]:checked')?.value;
        await this.settingsManager.save({
            ...this.settingsManager.getAll(),
            spinMode
        });
    }
}
