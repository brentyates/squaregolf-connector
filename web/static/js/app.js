// app.js - Main entry point
import { SquareGolfApp } from './core/SquareGolfApp.js';
import { ShotMonitor } from './features/ShotMonitor.js';

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
