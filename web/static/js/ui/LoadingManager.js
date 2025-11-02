// ui/LoadingManager.js
export class LoadingManager {
    constructor() {
        this.overlay = document.getElementById('loadingOverlay');
        this.textElement = this.overlay.querySelector('.loading-text');
    }

    show(text = 'Loading...') {
        this.textElement.textContent = text;
        this.overlay.classList.add('show');
    }

    hide() {
        this.overlay.classList.remove('show');
    }
}
