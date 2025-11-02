// features/ShotMonitor.js
export class ShotMonitor {
    constructor() {
        this.initializeEventListeners();
    }

    initializeEventListeners() {
        // Metrics tab switching
        const metricsTabs = document.querySelectorAll('.monitor-metrics-tabs .tab-button');
        metricsTabs.forEach(button => {
            button.addEventListener('click', (e) => {
                const tab = e.target.dataset.tab;
                this.switchMetricsTab(tab);
            });
        });
    }

    updateBallPosition(position, ballDetected, ballReady) {
        // Update the global status bar Ball Ready indicator
        const statusBallReady = document.getElementById('statusBallReady');
        if (statusBallReady) {
            if (ballReady) {
                statusBallReady.classList.add('connected');
                statusBallReady.classList.remove('disconnected');
            } else {
                statusBallReady.classList.remove('connected');
                statusBallReady.classList.add('disconnected');
            }
        }

        const ballDot = document.getElementById('ballDot');
        const svgContainer = document.querySelector('.ball-position-svg');

        // Check if position has valid numeric properties
        const hasValidPosition = position &&
                                typeof position.x === 'number' && !isNaN(position.x) &&
                                typeof position.y === 'number' && !isNaN(position.y) &&
                                typeof position.z === 'number' && !isNaN(position.z);

        if (!hasValidPosition || !ballDetected) {
            // No ball detected or invalid position - show subtle red border on entire SVG container
            ballDot.style.display = 'none';
            svgContainer.style.border = '2px solid rgba(239, 68, 68, 0.3)';
            document.getElementById('coordX').textContent = '--';
            document.getElementById('coordY').textContent = '--';
            document.getElementById('coordZ').textContent = '--';
            return;
        }

        // Reset SVG container border when ball is detected
        svgContainer.style.border = 'none';

        // Show and update ball dot position
        ballDot.style.display = 'block';

        // Convert sensor units to SVG coordinates
        // SVG viewBox: 0 0 300 400, center at 150, 200
        // The sensor reports values in 0.1mm increments (tenths of a millimeter)
        // So we need to divide by 10 to get actual mm, then scale to SVG
        const centerX = 150;
        const centerY = 200;

        // Convert from sensor units (0.1mm) to actual millimeters
        const actualX = position.x / 10;
        const actualY = position.y / 10;
        const actualZ = position.z / 10;

        // SVG visual range (as shown in labels)
        const svgVisualRange = 150; // ±150mm shown on labels

        // Actual expected coordinate range in mm (after conversion)
        // Typical values should be within ±500mm
        const actualRange = 500; // mm

        // Calculate scale to map actual mm coordinates to SVG coordinates
        const scale = svgVisualRange / actualRange;

        // X: positive right, negative left
        // Y: positive back (down in SVG), negative front (up in SVG)
        const svgX = centerX + (actualX * scale);
        const svgY = centerY + (actualY * scale);

        ballDot.setAttribute('cx', svgX);
        ballDot.setAttribute('cy', svgY);

        // Update coordinate display with actual mm values
        document.getElementById('coordX').textContent = `${actualX.toFixed(1)}mm`;
        document.getElementById('coordY').textContent = `${actualY.toFixed(1)}mm`;
        document.getElementById('coordZ').textContent = `${actualZ.toFixed(1)}mm`;

        // Calculate distance from center using actual mm values
        const distance = Math.sqrt(actualX * actualX + actualY * actualY);
        const distanceIndicator = document.getElementById('distanceIndicator');
        const distanceValue = document.getElementById('distanceValue');

        if (distanceIndicator && distanceValue) {
            distanceIndicator.style.display = 'flex';
            distanceValue.textContent = `${distance.toFixed(1)}mm`;

            // Color code based on distance
            distanceValue.classList.remove('excellent', 'good', 'poor');
            if (distance < 30) {
                distanceValue.classList.add('excellent');
            } else if (distance < 60) {
                distanceValue.classList.add('good');
            } else {
                distanceValue.classList.add('poor');
            }
        }

        // Set ball appearance based on ready state
        const targetZone = document.getElementById('targetZone');
        if (ballReady) {
            // Ball detected and ready - green fill, reset target zone
            ballDot.setAttribute('fill', '#22c55e');
            ballDot.setAttribute('stroke', '#fff');
            if (targetZone) {
                targetZone.setAttribute('stroke', '#22c55e');
                targetZone.setAttribute('stroke-width', '2');
                targetZone.setAttribute('stroke-opacity', '1');
            }
        } else {
            // Ball detected but not ready - red outline only
            ballDot.setAttribute('fill', 'none');
            ballDot.setAttribute('stroke', '#ef4444');
            ballDot.setAttribute('stroke-width', '3');
            if (targetZone) {
                targetZone.setAttribute('stroke', '#22c55e');
                targetZone.setAttribute('stroke-width', '2');
                targetZone.setAttribute('stroke-opacity', '0.3');
            }
        }
    }

    updateStatus(status) {
        // Update ball position visualization with detection state
        this.updateBallPosition(status.ballPosition, status.ballDetected, status.ballReady);
    }

    updateCurrentShot(ballData, clubData) {
        const placeholder = document.getElementById('monitorShotPlaceholder');
        const shotData = document.getElementById('monitorShotData');

        // Hide placeholder and show actual data
        if (placeholder) placeholder.style.display = 'none';
        if (shotData) shotData.style.display = 'block';

        // Update ball metrics
        const ballMetrics = document.getElementById('monitorBallMetrics');
        if (ballMetrics) ballMetrics.innerHTML = this.formatMetrics(ballData);

        // Update club metrics
        const clubMetrics = document.getElementById('monitorClubMetrics');
        if (clubMetrics) clubMetrics.innerHTML = this.formatMetrics(clubData);
    }

    formatMetrics(data) {
        if (!data || Object.keys(data).length === 0) {
            return '<div class="no-metrics">No data available</div>';
        }

        let html = '<div class="metrics-grid">';
        for (const [key, value] of Object.entries(data)) {
            if (value !== null && value !== undefined) {
                const label = key.replace(/([A-Z])/g, ' $1').trim();
                const displayLabel = label.charAt(0).toUpperCase() + label.slice(1);
                html += `
                    <div class="metric-item">
                        <div class="metric-label">${displayLabel}</div>
                        <div class="metric-value">${value}</div>
                    </div>
                `;
            }
        }
        html += '</div>';
        return html;
    }

    addShotToHistory(ballData, clubData) {
        // Placeholder for shot history functionality
    }

    switchMetricsTab(tabName) {
        // Remove active from all tabs
        document.querySelectorAll('.monitor-metrics-tabs .tab-button').forEach(btn => {
            btn.classList.remove('active');
        });
        document.querySelectorAll('.monitor-metrics-content .metrics-tab').forEach(tab => {
            tab.classList.remove('active');
        });

        // Add active to selected tab
        const tabButton = document.querySelector(`[data-tab="${tabName}"]`);
        if (tabButton) {
            tabButton.classList.add('active');
        }

        const metricsTab = document.getElementById(tabName === 'monitorBall' ? 'monitorBallMetrics' : 'monitorClubMetrics');
        if (metricsTab) {
            metricsTab.classList.add('active');
        }
    }
}
