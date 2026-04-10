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

        const centerX = 70;
        const centerY = 85;
        const actualX = position.x / 10;
        const actualY = position.y / 10;
        const svgVisualRange = 70;
        const actualRange = 500;
        const scale = svgVisualRange / actualRange;

        const svgX = this.clamp(centerX + (actualY * scale), 6, 134);
        const svgY = this.clamp(centerY + (actualX * scale), 6, 164);

        ballDot.setAttribute('cx', svgX);
        ballDot.setAttribute('cy', svgY);

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

        const lmChip = document.getElementById('lmStatusBlock');
        if (lmChip) {
            lmChip.classList.toggle('status-done', status.launchMonitorStatus === 'done');
        }

        this._isLeftHanded = status.handedness === 1;

        const isOmni = status.deviceType === 'omni';
        document.querySelectorAll('.omni-only').forEach(el => {
            el.style.display = isOmni ? '' : 'none';
        });
        const grid = document.querySelector('.diagnostics-grid');
        if (grid) grid.classList.toggle('has-impact', isOmni);

        if (status.lastBallMetrics && Object.keys(status.lastBallMetrics).length > 0) {
            this.updateCurrentShot(status.lastBallMetrics, status.lastClubMetrics);
            this.updateDiagrams(status.lastBallMetrics, status.lastClubMetrics);
        }
    }

    updateCurrentShot(ballData, clubData) {
        const ballSpeedMPH = typeof ballData?.speed === 'number' ? ballData.speed * 2.237 : null;
        this.updateMetricValue('metricBallSpeed', ballSpeedMPH, 'mph', false, ballData?.isBallSpeedValid, 'metricItemBallSpeed');
        this.updateMetricValue('metricLaunchAngle', ballData?.launchAngle, '°', false, true, 'metricItemLaunchAngle');
        this.updateMetricValue('metricDirection', ballData?.horizontalAngle, '°', true, true, 'metricItemDirection');
        this.updateMetricValue('metricBackSpin', ballData?.backSpin, 'rpm', false, ballData?.isBackSpinValid, 'metricItemBackSpin');
        this.updateMetricValue('metricSideSpin', ballData?.sideSpin, 'rpm', true, ballData?.isSideSpinValid, 'metricItemSideSpin');
        this.updateMetricValue('metricTotalSpin', ballData?.totalSpin, 'rpm', false, ballData?.isTotalSpinValid, 'metricItemTotalSpin');
        this.updateMetricValue('metricSpinAxis', ballData?.spinAxis, '°', false, ballData?.isSpinAxisValid, 'metricItemSpinAxis');

        this.updateMetricValue('metricAttackAngle', clubData?.attackAngle, '°', true, clubData?.isAttackAngleValid, 'metricItemAttackAngle');
        this.updateMetricValue('metricClubPath', clubData?.path, '°', true, clubData?.isPathValid, 'metricItemClubPath');
        this.updateMetricValue('metricFaceAngle', clubData?.angle, '°', true, clubData?.isFaceAngleValid, 'metricItemFaceAngle');
        this.updateMetricValue('metricDynamicLoft', clubData?.dynamicLoft, '°', false, clubData?.isDynamicLoftValid, 'metricItemDynamicLoft');
        this.updateMetricValue('metricClubSpeed', clubData?.clubSpeed, '', false, clubData?.isClubSpeedValid, 'metricItemClubSpeed');
        this.updateMetricValue('metricSmashFactor', clubData?.smashFactor, '', false, clubData?.isSmashFactorValid, 'metricItemSmashFactor');
        this.updateMetricValue('metricImpactH', clubData?.impactHorizontal, '°', true, clubData?.isImpactHorizontalValid, 'metricItemImpactH');
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

    updateDiagrams(ballData, clubData) {
        this.updateClubPathDiagram(clubData);
        this.updateLaunchDiagram(ballData, clubData);
        this.updateSpinDiagram(ballData);
        this.updateImpactDiagram(clubData);
        this.updateEfficiency(ballData, clubData);
    }

    setDiagValue(id, value, unit, showSign, isValid) {
        const el = document.getElementById(id);
        if (!el) return;

        if (value === null || value === undefined) {
            el.textContent = '-';
            el.className = 'diag-value-number';
            return;
        }

        let display = typeof value === 'number' ? value.toFixed(1) : value;
        if (showSign && typeof value === 'number' && value !== 0) {
            const prefix = value > 0 ? 'R' : 'L';
            display = `${prefix}${Math.abs(value).toFixed(1)}`;
        }
        if (unit) display += unit;

        el.textContent = display;
        el.className = 'diag-value-number';
        if (isValid === false) el.classList.add('invalid');
        else if (isValid === true) el.classList.add('valid');
    }

    setBadge(id, isValid) {
        const el = document.getElementById(id);
        if (!el) return;
        el.className = 'diag-badge';
        if (isValid === true) {
            el.classList.add('valid');
            el.textContent = 'Valid';
        } else if (isValid === false) {
            el.classList.add('invalid');
            el.textContent = 'Unvalidated';
        } else {
            el.textContent = 'Waiting';
        }
    }

    svgColor(isValid) {
        if (isValid === false) return '#c76834';
        if (isValid === true) return '#5f8d4e';
        return 'var(--text-muted)';
    }

    updateClubPathDiagram(clubData) {
        const pathLine = document.getElementById('clubPathLine');
        const faceLine = document.getElementById('clubFaceLine');
        if (!pathLine || !faceLine) return;

        const hasPath = typeof clubData?.path === 'number';
        const hasFace = typeof clubData?.angle === 'number';

        if (!hasPath && !hasFace) {
            this.setBadge('clubDiagBadge', null);
            this.setDiagValue('diagClubPath', null);
            this.setDiagValue('diagFaceAngle', null);
            this.setDiagValue('diagFaceToPath', null);
            pathLine.setAttribute('stroke', 'var(--text-muted)');
            faceLine.setAttribute('stroke', 'var(--text-muted)');
            return;
        }

        const pathValid = clubData?.isPathValid !== false;
        const faceValid = clubData?.isFaceAngleValid !== false;
        this.setBadge('clubDiagBadge', pathValid && faceValid);

        // Club path: rotate from center, clamped to ±20° visual
        const pathAngle = hasPath ? this.clamp(clubData.path, -20, 20) : 0;
        const pathRad = (pathAngle * Math.PI) / 180;
        const len = 70;
        pathLine.setAttribute('x1', 100 - Math.sin(pathRad) * len);
        pathLine.setAttribute('y1', 100 + Math.cos(pathRad) * len);
        pathLine.setAttribute('x2', 100 + Math.sin(pathRad) * len);
        pathLine.setAttribute('y2', 100 - Math.cos(pathRad) * len);
        pathLine.setAttribute('stroke', this.svgColor(hasPath ? clubData.isPathValid : null));

        // Face angle: short line perpendicular to face direction
        const faceAngle = hasFace ? this.clamp(clubData.angle, -20, 20) : 0;
        const faceRad = (faceAngle * Math.PI) / 180;
        const faceLen = 22;
        faceLine.setAttribute('x1', 100 - Math.cos(faceRad) * faceLen);
        faceLine.setAttribute('y1', 100 - Math.sin(faceRad) * faceLen);
        faceLine.setAttribute('x2', 100 + Math.cos(faceRad) * faceLen);
        faceLine.setAttribute('y2', 100 + Math.sin(faceRad) * faceLen);
        faceLine.setAttribute('stroke', this.svgColor(hasFace ? clubData.isFaceAngleValid : null));

        this.setDiagValue('diagClubPath', clubData.path, '°', true, clubData.isPathValid);
        this.setDiagValue('diagFaceAngle', clubData.angle, '°', true, clubData.isFaceAngleValid);

        if (hasPath && hasFace) {
            const ftp = clubData.angle - clubData.path;
            const ftpLabel = Math.abs(ftp) < 0.1 ? 'Square' : (ftp > 0 ? 'Open' : 'Closed');
            const el = document.getElementById('diagFaceToPath');
            if (el) {
                el.textContent = `${Math.abs(ftp).toFixed(1)}° ${ftpLabel}`;
                el.className = 'diag-value-number';
            }
        } else {
            this.setDiagValue('diagFaceToPath', null);
        }
    }

    updateLaunchDiagram(ballData, clubData) {
        const launchLine = document.getElementById('launchAngleLine');
        const attackLine = document.getElementById('attackAngleLine');
        const trajectoryArc = document.getElementById('trajectoryArc');
        const launchArc = document.getElementById('launchArc');
        if (!launchLine) return;

        const hasLaunch = typeof ballData?.launchAngle === 'number';
        const hasAttack = typeof clubData?.attackAngle === 'number';

        if (!hasLaunch && !hasAttack) {
            this.setBadge('launchDiagBadge', null);
            this.setDiagValue('diagLaunchAngle', null);
            this.setDiagValue('diagAttackAngle', null);
            this.setDiagValue('diagDynamicLoft', null);
            launchLine.setAttribute('stroke', 'var(--text-muted)');
            attackLine.setAttribute('stroke', 'var(--text-muted)');
            trajectoryArc.setAttribute('d', '');
            if (launchArc) launchArc.setAttribute('d', '');
            return;
        }

        this.setBadge('launchDiagBadge', true);

        const ox = 40, oy = 130;

        // Launch angle line
        if (hasLaunch) {
            const la = this.clamp(ballData.launchAngle, 0, 45);
            const laRad = (la * Math.PI) / 180;
            const lineLen = 120;
            launchLine.setAttribute('x2', ox + Math.cos(laRad) * lineLen);
            launchLine.setAttribute('y2', oy - Math.sin(laRad) * lineLen);
            launchLine.setAttribute('stroke', '#5f8d4e');

            // Trajectory arc
            const peakX = ox + 140;
            const peakY = oy - Math.sin(laRad) * 100;
            const endX = ox + 240;
            const endY = oy - 10;
            trajectoryArc.setAttribute('d', `M ${ox} ${oy} Q ${peakX} ${peakY} ${endX} ${endY}`);
            trajectoryArc.setAttribute('stroke', '#5f8d4e');

            // Small arc to show the angle
            const arcR = 30;
            const arcEndX = ox + Math.cos(laRad) * arcR;
            const arcEndY = oy - Math.sin(laRad) * arcR;
            if (launchArc) {
                launchArc.setAttribute('d', `M ${ox + arcR} ${oy} A ${arcR} ${arcR} 0 0 0 ${arcEndX} ${arcEndY}`);
                launchArc.setAttribute('stroke', '#5f8d4e');
            }
        }

        // Attack angle line
        if (hasAttack) {
            const aa = this.clamp(clubData.attackAngle, -15, 15);
            const aaRad = (aa * Math.PI) / 180;
            const aaLen = 40;
            attackLine.setAttribute('x1', ox - Math.cos(aaRad) * aaLen);
            attackLine.setAttribute('y1', oy + Math.sin(aaRad) * aaLen);
            attackLine.setAttribute('x2', ox);
            attackLine.setAttribute('y2', oy);
            attackLine.setAttribute('stroke', this.svgColor(clubData.isAttackAngleValid));
        }

        this.setDiagValue('diagLaunchAngle', ballData?.launchAngle, '°', false, true);
        this.setDiagValue('diagAttackAngle', clubData?.attackAngle, '°', true, clubData?.isAttackAngleValid);
        this.setDiagValue('diagDynamicLoft', clubData?.dynamicLoft, '°', false, clubData?.isDynamicLoftValid);
    }

    updateSpinDiagram(ballData) {
        const spinLine = document.getElementById('spinAxisLine');
        if (!spinLine) return;

        const hasAxis = typeof ballData?.spinAxis === 'number';
        const hasSpin = typeof ballData?.totalSpin === 'number';

        if (!hasAxis && !hasSpin) {
            this.setBadge('spinDiagBadge', null);
            this.setDiagValue('diagSpinAxis', null);
            this.setDiagValue('diagTotalSpin', null);
            this.setDiagValue('diagBackSideSpin', null);
            spinLine.setAttribute('stroke', 'var(--text-muted)');
            return;
        }

        const axisValid = ballData?.isSpinAxisValid !== false;
        const spinValid = ballData?.isTotalSpinValid !== false;
        this.setBadge('spinDiagBadge', axisValid && spinValid);

        // Rotate spin axis line - spin axis is in degrees, positive = right tilt
        if (hasAxis) {
            const angle = this.clamp(ballData.spinAxis, -45, 45);
            const rad = (angle * Math.PI) / 180;
            const len = 75;
            spinLine.setAttribute('x1', 100 - Math.sin(rad) * len);
            spinLine.setAttribute('y1', 100 - Math.cos(rad) * len);
            spinLine.setAttribute('x2', 100 + Math.sin(rad) * len);
            spinLine.setAttribute('y2', 100 + Math.cos(rad) * len);
            spinLine.setAttribute('stroke', this.svgColor(ballData.isSpinAxisValid));
        }

        this.setDiagValue('diagSpinAxis', ballData?.spinAxis, '°', true, ballData?.isSpinAxisValid);

        if (hasSpin) {
            const el = document.getElementById('diagTotalSpin');
            if (el) {
                el.textContent = `${Math.round(ballData.totalSpin)} rpm`;
                el.className = 'diag-value-number';
                if (spinValid) el.classList.add('valid');
                else el.classList.add('invalid');
            }
        } else {
            this.setDiagValue('diagTotalSpin', null);
        }

        const hasBack = typeof ballData?.backSpin === 'number';
        const hasSide = typeof ballData?.sideSpin === 'number';
        if (hasBack || hasSide) {
            const el = document.getElementById('diagBackSideSpin');
            if (el) {
                const back = hasBack ? Math.round(ballData.backSpin) : '-';
                const side = hasSide ? Math.round(ballData.sideSpin) : '-';
                el.textContent = `${back} / ${side}`;
                el.className = 'diag-value-number';
            }
        } else {
            this.setDiagValue('diagBackSideSpin', null);
        }
    }

    updateImpactDiagram(clubData) {
        const dot = document.getElementById('impactDot');
        const ring = document.getElementById('impactRing');
        if (!dot || !ring) return;

        const hasH = typeof clubData?.impactHorizontal === 'number';
        const hasV = typeof clubData?.impactVertical === 'number';

        if (!hasH && !hasV) {
            this.setBadge('impactDiagBadge', null);
            this.setDiagValue('diagImpactH', null);
            this.setDiagValue('diagImpactV', null);
            dot.setAttribute('fill', 'var(--text-muted)');
            dot.setAttribute('cx', 100);
            dot.setAttribute('cy', 80);
            ring.setAttribute('cx', 100);
            ring.setAttribute('cy', 80);
            ring.setAttribute('stroke', 'var(--text-muted)');
            return;
        }

        const hValid = clubData?.isImpactHorizontalValid !== false;
        const vValid = clubData?.isImpactVerticalValid !== false;
        this.setBadge('impactDiagBadge', hValid && vValid);

        // Landscape face: x=20..180 (center 100), y=20..140 (center 80)
        const hVal = hasH ? this.clamp(clubData.impactHorizontal, -20, 20) : 0;
        const vVal = hasV ? this.clamp(clubData.impactVertical, -20, 20) : 0;

        // RH: heel on left, toe on right. LH: flipped.
        const hSign = this._isLeftHanded ? -1 : 1;
        const cx = 100 + (hVal / 20) * 65 * hSign;
        const cy = 80 - (vVal / 20) * 50;

        const color = this.svgColor(hValid && vValid);
        dot.setAttribute('cx', cx);
        dot.setAttribute('cy', cy);
        dot.setAttribute('fill', color);
        ring.setAttribute('cx', cx);
        ring.setAttribute('cy', cy);
        ring.setAttribute('stroke', color);

        // Update toe/heel labels based on handedness
        const toeLabel = document.querySelector('#faceImpactSvg text:nth-of-type(1)');
        const heelLabel = document.querySelector('#faceImpactSvg text:nth-of-type(2)');
        if (toeLabel && heelLabel) {
            if (this._isLeftHanded) {
                toeLabel.textContent = 'TOE';
                heelLabel.textContent = 'HEEL';
            } else {
                toeLabel.textContent = 'HEEL';
                heelLabel.textContent = 'TOE';
            }
        }

        this.setDiagValue('diagImpactH', clubData.impactHorizontal, '°', true, clubData.isImpactHorizontalValid);
        this.setDiagValue('diagImpactV', clubData.impactVertical, '°', false, clubData.isImpactVerticalValid);
    }

    updateEfficiency(ballData, clubData) {
        const section = document.getElementById('diagEfficiency');
        if (!section) return;

        const hasClubSpeed = typeof clubData?.clubSpeed === 'number';
        const hasSmash = typeof clubData?.smashFactor === 'number';
        const hasBallSpeed = typeof ballData?.speed === 'number';

        const clubSpeedEl = document.getElementById('diagClubSpeed');
        const smashEl = document.getElementById('diagSmashFactor');
        const ballSpeedEl = document.getElementById('diagBallSpeed');

        if (clubSpeedEl) {
            clubSpeedEl.textContent = hasClubSpeed ? `${clubData.clubSpeed.toFixed(1)}` : '-';
        }
        if (smashEl) {
            smashEl.textContent = hasSmash ? `×${clubData.smashFactor.toFixed(2)}` : '-';
        }
        if (ballSpeedEl) {
            const mph = hasBallSpeed ? (ballData.speed * 2.237).toFixed(1) : '-';
            ballSpeedEl.textContent = mph;
        }
    }
}
