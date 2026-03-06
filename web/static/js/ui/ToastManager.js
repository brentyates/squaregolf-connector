// ui/ToastManager.js
export class ToastManager {
    #maxVisibleToasts = 4;
    #toastLifetimeMs = 5000;

    constructor() {
        this.container = document.getElementById('toastContainer');
    }

    show(message, type = 'info') {
        if (!this.container) return;

        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.textContent = message;

        if (this.container.childElementCount >= this.#maxVisibleToasts) {
            this.container.firstElementChild?.remove();
        }

        this.container.append(toast);
        window.setTimeout(() => toast.remove(), this.#toastLifetimeMs);
    }

    success(message) {
        this.show(message, 'success');
    }

    error(message) {
        this.show(message, 'error');
    }

    warning(message) {
        this.show(message, 'warning');
    }

    info(message) {
        this.show(message, 'info');
    }
}
