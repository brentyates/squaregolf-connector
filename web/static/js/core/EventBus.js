// core/EventBus.js
export class EventBus {
    #events = new Map();

    on(event, callback) {
        const listeners = this.#events.get(event) ?? new Set();
        listeners.add(callback);
        this.#events.set(event, listeners);

        return () => this.off(event, callback);
    }

    off(event, callback) {
        const listeners = this.#events.get(event);
        if (!listeners) return;

        listeners.delete(callback);
        if (listeners.size === 0) {
            this.#events.delete(event);
        }
    }

    emit(event, data) {
        const listeners = this.#events.get(event);
        if (!listeners) return;

        for (const callback of [...listeners]) {
            try {
                callback(data);
            } catch (error) {
                console.error(`Error in event handler for ${event}:`, error);
            }
        }
    }

    once(event, callback) {
        const unsubscribe = this.on(event, (data) => {
            unsubscribe();
            callback(data);
        });

        return unsubscribe;
    }
}
