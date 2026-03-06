// features/ShotMonitor.js
export class ShotMonitor {
    constructor(apiClient, eventBus) {
        this.api = apiClient;
        this.eventBus = eventBus;
    }

    updateBinaryIndicator(elementId, isConnected) {
        const element = document.getElementById(elementId);
        if (!element) return;

        element.classList.toggle('connected', isConnected);
        element.classList.toggle('disconnected', !isConnected);
    }

    clamp(value, min, max) {
        return Math.min(Math.max(value, min), max);
    }

    updateBallPosition(position, ballDetected, ballReady) {
        this.updateBinaryIndicator('statusBallReady', ballReady);

        const ballDot = document.getElementById('ballDot');
        const ballOverlay = document.getElementById('ballOverlay');

        if (!ballDot) return;

        const hasValidPosition = position &&
            typeof position.x === 'number' && !isNaN(position.x) &&
            typeof position.y === 'number' && !isNaN(position.y) &&
            typeof position.z === 'number' && !isNaN(position.z);

        if (!hasValidPosition || !ballDetected) {
            ballDot.classList.add('hidden');
            if (ballOverlay) ballOverlay.classList.add('no-ball');
            return;
        }

        if (ballOverlay) ballOverlay.classList.remove('no-ball');
        ballDot.classList.remove('hidden');

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
        const svgX = this.clamp(centerX + (actualY * scale), 6, 134);
        const svgY = this.clamp(centerY + (actualX * scale), 6, 164);

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
        // Backend field names: speed (m/s), launchAngle, horizontalAngle, totalSpin, spinAxis, backSpin, sideSpin
        const ballSpeedMPH = typeof ballData?.speed === 'number' ? ballData.speed * 2.237 : null;
        this.updateMetricValue('metricBallSpeed', ballSpeedMPH, 'mph');
        this.updateMetricValue('metricLaunchAngle', ballData?.launchAngle, '°');
        this.updateMetricValue('metricDirection', ballData?.horizontalAngle, '°', true);
        this.updateMetricValue('metricBackSpin', ballData?.backSpin, 'rpm');
        this.updateMetricValue('metricSideSpin', ballData?.sideSpin, 'rpm', true);
        this.updateMetricValue('metricTotalSpin', ballData?.totalSpin, 'rpm');
        this.updateMetricValue('metricSpinAxis', ballData?.spinAxis, '°');

        // Update club metrics in the metrics bar
        // Backend field names: path, angle, attackAngle, dynamicLoft
        this.updateMetricValue('metricAttackAngle', clubData?.attackAngle, '°', true);
        this.updateMetricValue('metricClubPath', clubData?.path, '°', true);
        this.updateMetricValue('metricFaceAngle', clubData?.angle, '°', true);
        this.updateMetricValue('metricDynamicLoft', clubData?.dynamicLoft, '°');
    }

    updateMetricValue(elementId, value, unit, showSign = false) {
        const element = document.getElementById(elementId);
        if (!element) return;

        if (value === null || value === undefined) {
            element.textContent = '-';
            element.removeAttribute('title');
            return;
        }

        let displayValue = typeof value === 'number' ? value.toFixed(1) : value;

        // Add sign prefix for directional values
        if (showSign && typeof value === 'number' && value !== 0) {
            const prefix = value > 0 ? 'R' : 'L';
            displayValue = `${prefix}${Math.abs(value).toFixed(1)}`;
        }

        element.textContent = displayValue;
        if (unit) {
            element.title = `${displayValue} ${unit}`;
        } else {
            element.removeAttribute('title');
        }
    }
}
