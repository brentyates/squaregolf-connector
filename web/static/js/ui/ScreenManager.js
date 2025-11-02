// ui/ScreenManager.js
export class ScreenManager {
    constructor(eventBus) {
        this.currentScreen = 'device';
        this.eventBus = eventBus;
        this.pageTitles = {
            device: 'Device Connection',
            shotMonitor: 'Shot Monitor',
            gspro: 'GSPro Connection',
            camera: 'Swing Camera',
            alignment: 'Device Alignment',
            settings: 'Settings'
        };
    }

    show(screenName) {
        // Emit event before navigation
        this.eventBus.emit('screen:before-change', {
            from: this.currentScreen,
            to: screenName
        });

        // Update navigation buttons
        document.querySelectorAll('.nav-button').forEach(btn => {
            btn.classList.remove('active');
        });
        const navButton = document.querySelector(`[data-screen="${screenName}"]`);
        if (navButton) {
            navButton.classList.add('active');
        }

        // Update screens
        document.querySelectorAll('.screen').forEach(screen => {
            screen.classList.remove('active');
        });
        const screenElement = document.getElementById(`${screenName}Screen`);
        if (screenElement) {
            screenElement.classList.add('active');
        }

        // Update page title
        const pageTitle = document.getElementById('pageTitle');
        if (pageTitle) {
            pageTitle.textContent = this.pageTitles[screenName] || screenName;
        }

        this.currentScreen = screenName;

        // Emit event after navigation
        this.eventBus.emit('screen:changed', screenName);
    }

    getCurrent() {
        return this.currentScreen;
    }
}
