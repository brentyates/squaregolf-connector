// features/ShotMonitor.js
export class ShotMonitor {
    constructor(apiClient, eventBus) {
        this.api = apiClient;
        this.eventBus = eventBus;
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
        const ballOverlay = document.getElementById('ballOverlay');

        if (!ballDot) return;

        const hasValidPosition = position &&
                                typeof position.x === 'number' && !isNaN(position.x) &&
                                typeof position.y === 'number' && !isNaN(position.y) &&
                                typeof position.z === 'number' && !isNaN(position.z);

        if (!hasValidPosition || !ballDetected) {
            ballDot.style.display = 'none';
            if (ballOverlay) ballOverlay.classList.add('no-ball');
            return;
        }

        if (ballOverlay) ballOverlay.classList.remove('no-ball');
        ballDot.style.display = 'block';

        // Convert sensor units to SVG coordinates
        // New SVG viewBox: 0 0 140 170, center at 70, 85
        const centerX = 70;
        const centerY = 85;

        // Convert from sensor units (0.1mm) to actual millimeters
        const actualX = position.x / 10;
        const actualY = position.y / 10;

        // SVG visual range and scale
        const svgVisualRange = 70;
        const actualRange = 500;
        const scale = svgVisualRange / actualRange;

        // Transform coordinates
        const svgX = centerX + (actualY * scale);
        const svgY = centerY + (actualX * scale);

        ballDot.setAttribute('cx', svgX);
        ballDot.setAttribute('cy', svgY);

        // Set ball appearance based on ready state
        if (ballReady) {
            ballDot.setAttribute('fill', '#22c55e');
            ballDot.setAttribute('stroke', '#fff');
            ballDot.setAttribute('stroke-width', '2');
        } else {
            ballDot.setAttribute('fill', 'none');
            ballDot.setAttribute('stroke', '#ef4444');
            ballDot.setAttribute('stroke-width', '3');
        }
    }

    updateStatus(status) {
        this.updateBallPosition(status.ballPosition, status.ballDetected, status.ballReady);
    }

    updateCurrentShot(ballData, clubData) {
        // Update ball metrics in the metrics bar
        this.updateMetricValue('metricBallSpeed', ballData?.ballSpeed, 'mph');
        this.updateMetricValue('metricLaunchAngle', ballData?.launchAngle, '°');
        this.updateMetricValue('metricDirection', ballData?.direction, '°', true);
        this.updateMetricValue('metricBackSpin', ballData?.backSpin, 'rpm');
        this.updateMetricValue('metricSideSpin', ballData?.sideSpin, 'rpm', true);
        this.updateMetricValue('metricCarry', ballData?.carry, 'yd');
        this.updateMetricValue('metricTotal', ballData?.totalDistance, 'yd');

        // Update club metrics in the metrics bar
        this.updateMetricValue('metricClubSpeed', clubData?.clubSpeed, 'mph');
        this.updateMetricValue('metricAttackAngle', clubData?.attackAngle, '°', true);
        this.updateMetricValue('metricClubPath', clubData?.clubPath, '°', true);
        this.updateMetricValue('metricFaceAngle', clubData?.faceAngle, '°', true);
        this.updateMetricValue('metricSmashFactor', clubData?.smashFactor, '');
    }

    updateMetricValue(elementId, value, unit, showSign = false) {
        const element = document.getElementById(elementId);
        if (!element) return;

        if (value === null || value === undefined) {
            element.textContent = '-';
            return;
        }

        let displayValue = typeof value === 'number' ? value.toFixed(1) : value;

        // Add sign prefix for directional values
        if (showSign && typeof value === 'number' && value !== 0) {
            const prefix = value > 0 ? 'R' : 'L';
            displayValue = `${prefix}${Math.abs(value).toFixed(1)}`;
        }

        element.textContent = unit ? `${displayValue}` : displayValue;
    }

    addShotToHistory(ballData, clubData) {
        // Shot history could be implemented if needed
    }
}
