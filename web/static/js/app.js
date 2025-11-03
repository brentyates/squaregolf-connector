// app.js - Main entry point
import { SquareGolfApp } from './core/SquareGolfApp.js';

document.addEventListener('DOMContentLoaded', () => {
    window.app = new SquareGolfApp();

    // Status bar click-to-expand functionality
    const statusBar = document.getElementById('statusBar');
    if (statusBar) {
        statusBar.addEventListener('click', () => {
            statusBar.classList.toggle('expanded');
        });
    }
});
