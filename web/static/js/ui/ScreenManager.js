// ui/ScreenManager.js
export class ScreenManager {
    #pageTitles = {
        device: 'Device',
        gspro: 'GSPro',
        infiniteTees: 'Infinite Tees',
        settings: 'Settings'
    };

    constructor(eventBus) {
        this.currentScreen = 'device';
        this.eventBus = eventBus;
        this.navButtons = [...document.querySelectorAll('.nav-button')];
        this.screens = [...document.querySelectorAll('.screen')];
        this.pageTitle = document.getElementById('pageTitle');
    }

    show(screenName) {
        if (screenName === this.currentScreen) return;

        this.eventBus.emit('screen:before-change', {
            from: this.currentScreen,
            to: screenName
        });

        for (const button of this.navButtons) {
            button.classList.toggle('active', button.dataset.screen === screenName);
        }

        for (const screen of this.screens) {
            screen.classList.toggle('active', screen.id === `${screenName}Screen`);
        }

        if (this.pageTitle) {
            this.pageTitle.textContent = this.#pageTitles[screenName] ?? screenName;
        }

        this.currentScreen = screenName;
        this.eventBus.emit('screen:changed', screenName);
    }

    getCurrent() {
        return this.currentScreen;
    }
}
