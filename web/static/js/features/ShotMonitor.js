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
        this.updateTelemetry(status);
    }

    updateCurrentShot(ballData, clubData) {
        // Update ball metrics in the metrics bar
        // Backend field names: speed (m/s), launchAngle, horizontalAngle, totalSpin, spinAxis, backSpin, sideSpin
        const ballSpeedMPH = typeof ballData?.speed === 'number' ? ballData.speed * 2.237 : null;
        this.updateMetricValue('metricBallSpeed', ballSpeedMPH, 'mph', false, ballData?.isBallSpeedValid, 'metricItemBallSpeed');
        this.updateMetricValue('metricLaunchAngle', ballData?.launchAngle, '°', false, true, 'metricItemLaunchAngle');
        this.updateMetricValue('metricDirection', ballData?.horizontalAngle, '°', true, true, 'metricItemDirection');
        this.updateMetricValue('metricBackSpin', ballData?.backSpin, 'rpm', false, ballData?.isBackSpinValid, 'metricItemBackSpin');
        this.updateMetricValue('metricSideSpin', ballData?.sideSpin, 'rpm', true, ballData?.isSideSpinValid, 'metricItemSideSpin');
        this.updateMetricValue('metricTotalSpin', ballData?.totalSpin, 'rpm', false, ballData?.isTotalSpinValid, 'metricItemTotalSpin');
        this.updateMetricValue('metricSpinAxis', ballData?.spinAxis, '°', false, ballData?.isSpinAxisValid, 'metricItemSpinAxis');

        // Update club metrics in the metrics bar
        // Backend field names: path, angle, attackAngle, dynamicLoft
        this.updateMetricValue('metricAttackAngle', clubData?.attackAngle, '°', true, clubData?.isAttackAngleValid, 'metricItemAttackAngle');
        this.updateMetricValue('metricClubPath', clubData?.path, '°', true, clubData?.isPathValid, 'metricItemClubPath');
        this.updateMetricValue('metricFaceAngle', clubData?.angle, '°', true, clubData?.isFaceAngleValid, 'metricItemFaceAngle');
        this.updateMetricValue('metricDynamicLoft', clubData?.dynamicLoft, '°', false, clubData?.isDynamicLoftValid, 'metricItemDynamicLoft');
    }

    updateMetricValue(elementId, value, unit, showSign = false, isValid = true, containerId = null) {
        const element = document.getElementById(elementId);
        const container = containerId ? document.getElementById(containerId) : null;
        if (!element) return;

        if (value === null || value === undefined) {
            element.textContent = '-';
            element.removeAttribute('title');
            if (container) {
                container.classList.remove('metric-valid', 'metric-invalid');
            }
            return;
        }

        let displayValue = typeof value === 'number' ? value.toFixed(1) : value;

        // Add sign prefix for directional values
        if (showSign && typeof value === 'number' && value !== 0) {
            const prefix = value > 0 ? 'R' : 'L';
            displayValue = `${prefix}${Math.abs(value).toFixed(1)}`;
        }

        element.textContent = displayValue;
        if (container) {
            container.classList.toggle('metric-valid', isValid !== false);
            container.classList.toggle('metric-invalid', isValid === false);
        }
        if (unit) {
            const quality = isValid === false ? 'LM did not validate this metric' : 'LM validated this metric';
            element.title = `${displayValue} ${unit} • ${quality}`;
        } else {
            element.removeAttribute('title');
        }
    }

    updateTelemetry(status) {
        this.updateStatusLabel('launchMonitorStatus', this.formatStatus(status.launchMonitorStatus));
        this.updateTelemetrySection({
            summaryId: 'ballTelemetrySummary',
            rawId: 'telemetryBallRaw',
            metrics: [
                { id: 'telemetryBallSpeed', value: status.lastBallMetrics?.speed, unit: 'm/s', valid: status.lastBallMetrics?.isBallSpeedValid, label: 'Ball speed' },
                { id: 'telemetryLaunchAngle', value: status.lastBallMetrics?.launchAngle, unit: 'deg', valid: true, label: 'Launch angle' },
                { id: 'telemetryDirection', value: status.lastBallMetrics?.horizontalAngle, unit: 'deg', valid: true, label: 'Horizontal angle' },
                { id: 'telemetryTotalSpin', value: status.lastBallMetrics?.totalSpin, unit: 'rpm', valid: status.lastBallMetrics?.isTotalSpinValid, label: 'Total spin' },
                { id: 'telemetrySpinAxis', value: status.lastBallMetrics?.spinAxis, unit: 'deg', valid: status.lastBallMetrics?.isSpinAxisValid, label: 'Spin axis' },
                { id: 'telemetryBackSpin', value: status.lastBallMetrics?.backSpin, unit: 'rpm', valid: status.lastBallMetrics?.isBackSpinValid, label: 'Backspin' },
                { id: 'telemetrySideSpin', value: status.lastBallMetrics?.sideSpin, unit: 'rpm', valid: status.lastBallMetrics?.isSideSpinValid, label: 'Sidespin' },
            ],
            rawData: status.lastBallMetrics?.rawData,
            waitingText: 'Waiting for shot'
        });
        this.updateTelemetrySection({
            summaryId: 'clubTelemetrySummary',
            rawId: 'telemetryClubRaw',
            metrics: [
                { id: 'telemetryAttackAngle', value: status.lastClubMetrics?.attackAngle, unit: 'deg', valid: status.lastClubMetrics?.isAttackAngleValid, label: 'Attack angle' },
                { id: 'telemetryClubPath', value: status.lastClubMetrics?.path, unit: 'deg', valid: status.lastClubMetrics?.isPathValid, label: 'Club path' },
                { id: 'telemetryFaceAngle', value: status.lastClubMetrics?.angle, unit: 'deg', valid: status.lastClubMetrics?.isFaceAngleValid, label: 'Face angle' },
                { id: 'telemetryDynamicLoft', value: status.lastClubMetrics?.dynamicLoft, unit: 'deg', valid: status.lastClubMetrics?.isDynamicLoftValid, label: 'Dynamic loft' },
            ],
            rawData: status.lastClubMetrics?.rawData,
            waitingText: 'Waiting for club data'
        });
    }

    updateTelemetrySection({ summaryId, rawId, metrics, rawData, waitingText }) {
        const hasPacket = Array.isArray(rawData) && rawData.length > 0;
        let invalidCount = 0;

        metrics.forEach(({ id, value, unit, valid, label }) => {
            const element = document.getElementById(id);
            if (!element) return;

            if (value === null || value === undefined) {
                element.textContent = '-';
                element.className = 'telemetry-value telemetry-waiting';
                element.title = `${label}: not present`;
                return;
            }

            const formatted = typeof value === 'number' ? `${value.toFixed(1)}${unit ? ` ${unit}` : ''}` : `${value}`;
            element.textContent = formatted;
            const isExplicitlyInvalid = valid === false;
            element.className = `telemetry-value ${isExplicitlyInvalid ? 'telemetry-invalid' : 'telemetry-valid'}`;
            element.title = isExplicitlyInvalid
                ? `${label}: present, but the launch monitor did not validate it. This may be estimated or unreliable.`
                : `${label}: validated by the launch monitor.`;
            if (isExplicitlyInvalid) invalidCount += 1;
        });

        this.updateStatusLabel(summaryId, hasPacket ? (invalidCount > 0 ? `${invalidCount} metric${invalidCount === 1 ? '' : 's'} not validated by LM` : 'All visible metrics validated by LM') : waitingText);
        this.updateStatusLabel(rawId, hasPacket ? rawData.join(' ') : '-');
    }

    updateStatusLabel(elementId, value) {
        const element = document.getElementById(elementId);
        if (!element) return;
        element.textContent = value ?? '-';
    }

    formatStatus(status) {
        if (!status) return '-';
        return `${status.charAt(0).toUpperCase()}${status.slice(1)}`;
    }
}
