// services/WebSocketService.js
export class WebSocketService {
    #reconnectTimeoutId = null;
    #reconnectDelayMs = 3000;
    #maxReconnectDelayMs = 15000;
    #manuallyDisconnected = false;

    constructor(eventBus) {
        this.eventBus = eventBus;
        this.ws = null;
    }

    connect() {
        if (this.ws && [WebSocket.OPEN, WebSocket.CONNECTING].includes(this.ws.readyState)) {
            return;
        }

        this.#manuallyDisconnected = false;
        this.#clearReconnectTimer();

        const wsUrl = new URL('/ws', window.location.href);
        wsUrl.protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';

        try {
            this.ws = new WebSocket(wsUrl);
            this.ws.addEventListener('open', this.#handleOpen);
            this.ws.addEventListener('message', this.#handleMessage);
            this.ws.addEventListener('close', this.#handleClose);
            this.ws.addEventListener('error', this.#handleError);
        } catch (error) {
            console.error('Failed to connect WebSocket:', error);
            this.#scheduleReconnect();
        }
    }

    disconnect() {
        this.#manuallyDisconnected = true;
        this.#clearReconnectTimer();

        if (!this.ws) return;

        this.ws.removeEventListener('open', this.#handleOpen);
        this.ws.removeEventListener('message', this.#handleMessage);
        this.ws.removeEventListener('close', this.#handleClose);
        this.ws.removeEventListener('error', this.#handleError);
        this.ws.close();
        this.ws = null;
    }

    #handleOpen = () => {
        console.log('WebSocket connected');
        this.#reconnectDelayMs = 3000;
        this.eventBus.emit('ws:connected');
        this.#clearReconnectTimer();
    };

    #handleMessage = ({ data }) => {
        try {
            this.eventBus.emit('ws:message', JSON.parse(data));
        } catch (error) {
            console.error('Error parsing WebSocket message:', error);
        }
    };

    #handleClose = () => {
        console.log('WebSocket disconnected');
        this.eventBus.emit('ws:disconnected');
        this.ws = null;

        if (!this.#manuallyDisconnected) {
            this.#scheduleReconnect();
        }
    };

    #handleError = (error) => {
        console.error('WebSocket error:', error);
        this.eventBus.emit('ws:error', error);
    };

    #scheduleReconnect() {
        if (this.#reconnectTimeoutId || this.#manuallyDisconnected) return;

        this.#reconnectTimeoutId = window.setTimeout(() => {
            this.#reconnectTimeoutId = null;
            console.log('Attempting to reconnect WebSocket...');
            this.connect();
        }, this.#reconnectDelayMs);

        this.#reconnectDelayMs = Math.min(this.#reconnectDelayMs * 2, this.#maxReconnectDelayMs);
    }

    #clearReconnectTimer() {
        if (!this.#reconnectTimeoutId) return;

        window.clearTimeout(this.#reconnectTimeoutId);
        this.#reconnectTimeoutId = null;
    }
}
