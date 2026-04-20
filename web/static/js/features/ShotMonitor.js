// features/ShotMonitor.js
export class ShotMonitor {
    constructor(apiClient, eventBus) {
        this.api = apiClient;
        this.eventBus = eventBus;
        this._shotHistory = [];
        this._maxHistory = 20;
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
        const ballSpeedMPH = typeof ballData?.speed === 'number' ? ballData.speed * 2.23694 : null;
        this.updateMetricValue('metricBallSpeed', ballSpeedMPH, 'mph', false, ballData?.isBallSpeedValid, 'metricItemBallSpeed');
        this.updateMetricValue('metricLaunchAngle', ballData?.launchAngle, '°', false, true, 'metricItemLaunchAngle');
        this.updateMetricValue('metricDirection', ballData?.horizontalAngle, '°', 'lr', true, 'metricItemDirection');
        const isTopspin = typeof ballData?.backSpin === 'number' && ballData.backSpin < 0;
        this.updateMetricValue('metricBackSpin', ballData?.backSpin, 'rpm', isTopspin ? 'topspin' : null, ballData?.isBackSpinValid, 'metricItemBackSpin');
        this.updateMetricValue('metricSideSpin', ballData?.sideSpin, 'rpm', 'spin', ballData?.isSideSpinValid, 'metricItemSideSpin');
        this.updateMetricValue('metricTotalSpin', ballData?.totalSpin, 'rpm', isTopspin ? 'topspin' : null, ballData?.isTotalSpinValid, 'metricItemTotalSpin');
        this.updateMetricValue('metricSpinAxis', ballData?.spinAxis, '°', false, ballData?.isSpinAxisValid, 'metricItemSpinAxis');

        this.updateMetricValue('metricAttackAngle', clubData?.attackAngle, '°', 'attack', clubData?.isAttackAngleValid, 'metricItemAttackAngle');
        this.updateMetricValue('metricClubPath', clubData?.path, '°', 'path', clubData?.isPathValid, 'metricItemClubPath');
        this.updateMetricValue('metricFaceAngle', clubData?.angle, '°', 'face', clubData?.isFaceAngleValid, 'metricItemFaceAngle');
        this.updateMetricValue('metricDynamicLoft', clubData?.dynamicLoft, '°', false, clubData?.isDynamicLoftValid, 'metricItemDynamicLoft');
        const clubSpeedMPH = typeof clubData?.clubSpeed === 'number' ? clubData.clubSpeed * 2.23694 : null;
        this.updateMetricValue('metricClubSpeed', clubSpeedMPH, 'mph', false, clubData?.isClubSpeedValid, 'metricItemClubSpeed');
        this.updateMetricValue('metricSmashFactor', clubData?.smashFactor, '', false, clubData?.isSmashFactorValid, 'metricItemSmashFactor');
        this.updateMetricValue('metricImpactH', clubData?.impactHorizontal, 'mm', 'impact', clubData?.isImpactHorizontalValid, 'metricItemImpactH');
    }

    updateMetricValue(elementId, value, unit, labelFormat = null, isValid = true, containerId = null) {
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

        const decimals = (unit === 'rpm') ? 0 : 1;
        let displayValue = typeof value === 'number' ? value.toFixed(decimals) : value;

        if (labelFormat && typeof value === 'number' && value !== 0) {
            const label = this.metricLabel(value, labelFormat);
            if (label) {
                displayValue = `${label} ${Math.abs(value).toFixed(decimals)}`;
            }
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
        const shotEntry = { ballData, clubData };
        const lastEntry = this._shotHistory[this._shotHistory.length - 1];
        const isDuplicate = lastEntry &&
            lastEntry.ballData?.speed === ballData?.speed &&
            lastEntry.ballData?.launchAngle === ballData?.launchAngle &&
            lastEntry.ballData?.totalSpin === ballData?.totalSpin;
        if (!isDuplicate) {
            this._shotHistory.push(shotEntry);
            if (this._shotHistory.length > this._maxHistory) this._shotHistory.shift();
        }

        this.updateClubPathDiagram(clubData);
        this.updateLaunchDiagram(ballData, clubData);
        this.updateSpinDiagram(ballData);
        this.updateImpactDiagram(clubData);
        this.updateEfficiency(ballData, clubData);
    }

    setDiagValue(id, value, unit, labelFormat, isValid) {
        const el = document.getElementById(id);
        if (!el) return;

        if (value === null || value === undefined) {
            el.textContent = '-';
            el.className = 'diag-value-number';
            return;
        }

        let display = typeof value === 'number' ? value.toFixed(1) : value;
        if (labelFormat && typeof value === 'number' && value !== 0) {
            const label = this.metricLabel(value, labelFormat);
            if (label) {
                display = `${label} ${Math.abs(value).toFixed(1)}`;
            }
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

    metricLabel(value, format) {
        if (!format || value === 0) return null;
        switch (format) {
            case 'lr': return value > 0 ? 'R' : 'L';
            case 'spin': return value > 0 ? 'L' : 'R';
            case 'path': {
                const inOut = this._isLeftHanded ? (value < 0) : (value > 0);
                return inOut ? 'In-Out' : 'Out-In';
            }
            case 'face': {
                const open = this._isLeftHanded ? (value < 0) : (value > 0);
                return open ? 'Open' : 'Closed';
            }
            case 'attack': return value > 0 ? 'Up' : 'Down';
            case 'impact': return value > 0 ? 'Toe' : 'Heel';
            case 'topspin': return 'T';
            default: return null;
        }
    }

    sectorArc(cx, cy, radius, startAngleDeg, endAngleDeg) {
        const toRad = d => (d * Math.PI) / 180;
        const s = toRad(startAngleDeg - 90);
        const e = toRad(endAngleDeg - 90);
        const x1 = cx + Math.cos(s) * radius;
        const y1 = cy + Math.sin(s) * radius;
        const x2 = cx + Math.cos(e) * radius;
        const y2 = cy + Math.sin(e) * radius;
        const largeArc = Math.abs(endAngleDeg - startAngleDeg) > 180 ? 1 : 0;
        const sweep = endAngleDeg > startAngleDeg ? 1 : 0;
        return `M ${cx} ${cy} L ${x1} ${y1} A ${radius} ${radius} 0 ${largeArc} ${sweep} ${x2} ${y2} Z`;
    }

    updateClubPathDiagram(clubData) {
        const pathLine = document.getElementById('clubPathLine');
        const faceLine = document.getElementById('clubFaceLine');
        const pathArc = document.getElementById('clubPathArc');
        const faceArc = document.getElementById('clubFaceArc');
        const pathArrow = document.getElementById('clubPathArrow');
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
            if (pathArc) pathArc.setAttribute('d', '');
            if (faceArc) faceArc.setAttribute('d', '');
            if (pathArrow) pathArrow.setAttribute('points', '');
            this.updateDispersionDots();
            return;
        }

        const pathValid = clubData?.isPathValid !== false;
        const faceValid = clubData?.isFaceAngleValid !== false;
        this.setBadge('clubDiagBadge', pathValid && faceValid);

        const pathAngle = hasPath ? this.clamp(clubData.path, -20, 20) : 0;
        const pathRad = (pathAngle * Math.PI) / 180;
        const len = 70;
        const x1 = 100 - Math.sin(pathRad) * len;
        const y1 = 100 + Math.cos(pathRad) * len;
        const x2 = 100 + Math.sin(pathRad) * len;
        const y2 = 100 - Math.cos(pathRad) * len;
        pathLine.setAttribute('x1', x1);
        pathLine.setAttribute('y1', y1);
        pathLine.setAttribute('x2', x2);
        pathLine.setAttribute('y2', y2);
        const pathColor = this.svgColor(hasPath ? clubData.isPathValid : null);
        pathLine.setAttribute('stroke', pathColor);

        if (pathArc && hasPath && Math.abs(pathAngle) > 0.5) {
            const arcRadius = 55;
            const start = Math.min(0, pathAngle);
            const end = Math.max(0, pathAngle);
            pathArc.setAttribute('d', this.sectorArc(100, 100, arcRadius, start, end));
            pathArc.setAttribute('fill', pathValid ? 'rgba(95,141,78,0.15)' : 'rgba(199,104,52,0.10)');
        } else if (pathArc) {
            pathArc.setAttribute('d', '');
        }

        if (pathArrow && hasPath) {
            const arrowSize = 6;
            const tipX = x2;
            const tipY = y2;
            const perpX = Math.cos(pathRad) * arrowSize;
            const perpY = Math.sin(pathRad) * arrowSize;
            const backX = Math.sin(pathRad) * arrowSize * 1.5;
            const backY = -Math.cos(pathRad) * arrowSize * 1.5;
            pathArrow.setAttribute('points',
                `${tipX},${tipY} ${tipX - backX + perpX},${tipY - backY + perpY} ${tipX - backX - perpX},${tipY - backY - perpY}`);
            pathArrow.setAttribute('fill', pathColor);
        } else if (pathArrow) {
            pathArrow.setAttribute('points', '');
        }

        const faceAngle = hasFace ? this.clamp(clubData.angle, -20, 20) : 0;
        const faceRad = (faceAngle * Math.PI) / 180;
        const faceLen = 22;
        faceLine.setAttribute('x1', 100 - Math.cos(faceRad) * faceLen);
        faceLine.setAttribute('y1', 100 - Math.sin(faceRad) * faceLen);
        faceLine.setAttribute('x2', 100 + Math.cos(faceRad) * faceLen);
        faceLine.setAttribute('y2', 100 + Math.sin(faceRad) * faceLen);
        faceLine.setAttribute('stroke', this.svgColor(hasFace ? clubData.isFaceAngleValid : null));

        if (faceArc && hasFace && Math.abs(faceAngle) > 0.5) {
            const arcRadius = 40;
            const start = Math.min(0, faceAngle);
            const end = Math.max(0, faceAngle);
            faceArc.setAttribute('d', this.sectorArc(100, 100, arcRadius, start, end));
            faceArc.setAttribute('fill', faceValid ? 'rgba(59,130,246,0.12)' : 'rgba(199,104,52,0.08)');
        } else if (faceArc) {
            faceArc.setAttribute('d', '');
        }

        this.setDiagValue('diagClubPath', clubData.path, '°', 'path', clubData.isPathValid);
        this.setDiagValue('diagFaceAngle', clubData.angle, '°', 'face', clubData.isFaceAngleValid);

        if (hasPath && hasFace) {
            const ftp = clubData.angle - clubData.path;
            const isOpen = this._isLeftHanded ? (ftp < 0) : (ftp > 0);
            const ftpLabel = Math.abs(ftp) < 0.1 ? 'Square' : (isOpen ? 'Open' : 'Closed');
            const el = document.getElementById('diagFaceToPath');
            if (el) {
                el.textContent = `${Math.abs(ftp).toFixed(1)}° ${ftpLabel}`;
                el.className = 'diag-value-number';
            }
        } else {
            this.setDiagValue('diagFaceToPath', null);
        }

        this.updateDispersionDots();
    }

    updateDispersionDots() {
        const container = document.getElementById('dispersionDots');
        if (!container) return;
        container.innerHTML = '';

        const history = this._shotHistory;
        if (history.length < 2) return;

        for (let i = 0; i < history.length; i++) {
            const dir = history[i].ballData?.horizontalAngle;
            if (typeof dir !== 'number') continue;

            const isLatest = (i === history.length - 1);
            const angle = this.clamp(dir, -20, 20);
            const rad = (angle * Math.PI) / 180;
            const dist = 50 + ((i * 17) % 30);
            const dotX = 100 + Math.sin(rad) * dist;
            const dotY = 100 - Math.cos(rad) * dist;

            const circle = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
            circle.setAttribute('cx', dotX);
            circle.setAttribute('cy', dotY);
            circle.setAttribute('r', isLatest ? '3' : '2');
            circle.setAttribute('fill', isLatest ? '#ef4444' : 'rgba(255,255,255,0.5)');
            circle.setAttribute('opacity', isLatest ? '0.9' : '0.4');
            container.appendChild(circle);
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

        if (hasLaunch) {
            const la = this.clamp(ballData.launchAngle, 0, 45);
            const laRad = (la * Math.PI) / 180;
            const lineLen = 120;
            const launchEndX = ox + Math.cos(laRad) * lineLen;
            const launchEndY = oy - Math.sin(laRad) * lineLen;
            launchLine.setAttribute('x2', launchEndX);
            launchLine.setAttribute('y2', launchEndY);
            launchLine.setAttribute('stroke', '#5f8d4e');

            const peakHeight = Math.sin(laRad) * 110;
            const peakX = ox + 130;
            const peakY = oy - peakHeight;
            const endX = 280;
            const endY = oy - 5;
            const cp1x = ox + 60;
            const cp1y = oy - peakHeight * 1.1;
            const cp2x = ox + 200;
            const cp2y = oy - peakHeight * 0.7;
            trajectoryArc.setAttribute('d', `M ${ox} ${oy} C ${cp1x} ${cp1y} ${cp2x} ${cp2y} ${endX} ${endY}`);
            trajectoryArc.setAttribute('stroke', '#5f8d4e');

            const arcR = 30;
            const arcEndX = ox + Math.cos(laRad) * arcR;
            const arcEndY = oy - Math.sin(laRad) * arcR;
            if (launchArc) {
                launchArc.setAttribute('d', `M ${ox} ${oy} L ${ox + arcR} ${oy} A ${arcR} ${arcR} 0 0 0 ${arcEndX} ${arcEndY} Z`);
                launchArc.setAttribute('stroke', '#5f8d4e');
                launchArc.setAttribute('fill', 'rgba(95,141,78,0.1)');
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
        this.setDiagValue('diagAttackAngle', clubData?.attackAngle, '°', 'attack', clubData?.isAttackAngleValid);
        this.setDiagValue('diagDynamicLoft', clubData?.dynamicLoft, '°', false, clubData?.isDynamicLoftValid);
    }

    updateSpinDiagram(ballData) {
        const spinLine = document.getElementById('spinAxisLine');
        const shade = document.getElementById('spinBallShade');
        const curveArrow = document.getElementById('spinCurveArrow');
        const curveHead = document.getElementById('spinCurveHead');
        if (!spinLine) return;

        const hasAxis = typeof ballData?.spinAxis === 'number';
        const hasSpin = typeof ballData?.totalSpin === 'number';

        if (!hasAxis && !hasSpin) {
            this.setBadge('spinDiagBadge', null);
            this.setDiagValue('diagSpinAxis', null);
            this.setDiagValue('diagTotalSpin', null);
            this.setDiagValue('diagBackSideSpin', null);
            spinLine.setAttribute('stroke', 'var(--text-muted)');
            if (shade) shade.setAttribute('transform', '');
            if (curveArrow) curveArrow.setAttribute('d', '');
            if (curveHead) curveHead.setAttribute('points', '');
            return;
        }

        const axisValid = ballData?.isSpinAxisValid !== false;
        const spinValid = ballData?.isTotalSpinValid !== false;
        this.setBadge('spinDiagBadge', axisValid && spinValid);

        if (hasAxis) {
            const angle = this.clamp(ballData.spinAxis, -45, 45);
            const rad = (angle * Math.PI) / 180;
            const len = 75;
            spinLine.setAttribute('x1', 100 - Math.sin(rad) * len);
            spinLine.setAttribute('y1', 100 - Math.cos(rad) * len);
            spinLine.setAttribute('x2', 100 + Math.sin(rad) * len);
            spinLine.setAttribute('y2', 100 + Math.cos(rad) * len);
            spinLine.setAttribute('stroke', this.svgColor(ballData.isSpinAxisValid));

            if (shade) {
                shade.setAttribute('transform', `rotate(${angle} 100 100)`);
                const color = angle > 0 ? 'rgba(59,130,246,0.12)' : 'rgba(199,104,52,0.12)';
                shade.setAttribute('fill', Math.abs(angle) < 1 ? 'rgba(148,163,184,0.06)' : color);
            }

            if (curveArrow && curveHead && Math.abs(angle) > 2) {
                const dir = angle > 0 ? 1 : -1;
                const r = 52;
                const startA = -30 * dir;
                const endA = 30 * dir;
                const s = ((startA - 90) * Math.PI) / 180;
                const e = ((endA - 90) * Math.PI) / 180;
                const sx = 100 + Math.cos(s) * r;
                const sy = 100 + Math.sin(s) * r;
                const ex = 100 + Math.cos(e) * r;
                const ey = 100 + Math.sin(e) * r;
                const sweep = dir > 0 ? 1 : 0;
                curveArrow.setAttribute('d', `M ${sx} ${sy} A ${r} ${r} 0 0 ${sweep} ${ex} ${ey}`);
                curveArrow.setAttribute('stroke', this.svgColor(ballData.isSpinAxisValid));

                const tipX = ex;
                const tipY = ey;
                const tangentX = -Math.sin(e) * dir * 5;
                const tangentY = Math.cos(e) * dir * 5;
                const normalX = Math.cos(e) * 4;
                const normalY = Math.sin(e) * 4;
                curveHead.setAttribute('points',
                    `${tipX},${tipY} ${tipX - tangentX + normalX},${tipY - tangentY + normalY} ${tipX - tangentX - normalX},${tipY - tangentY - normalY}`);
                curveHead.setAttribute('fill', this.svgColor(ballData.isSpinAxisValid));
            } else {
                if (curveArrow) curveArrow.setAttribute('d', '');
                if (curveHead) curveHead.setAttribute('points', '');
            }
        }

        this.setDiagValue('diagSpinAxis', ballData?.spinAxis, '°', 'spin', ballData?.isSpinAxisValid);

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
                const isTopspin = hasBack && ballData.backSpin < 0;
                const back = hasBack ? `${isTopspin ? 'T ' : ''}${Math.abs(Math.round(ballData.backSpin))}` : '-';
                const sideVal = hasSide ? Math.round(ballData.sideSpin) : null;
                const side = sideVal !== null ? `${sideVal > 0 ? 'L ' : sideVal < 0 ? 'R ' : ''}${Math.abs(sideVal)}` : '-';
                el.textContent = `${back} / ${side}`;
                el.className = 'diag-value-number';
            }
        } else {
            this.setDiagValue('diagBackSideSpin', null);
        }
    }

    impactMmToSvg(valueMm) {
        const CLUB_SIZE_MM = 44.88;
        const CLUB_PIXELS = 80;
        const mmPerPixel = CLUB_SIZE_MM / CLUB_PIXELS;
        return valueMm / mmPerPixel;
    }

    updateImpactDiagram(clubData) {
        const dot = document.getElementById('impactDot');
        const ring = document.getElementById('impactRing');
        const historyGroup = document.getElementById('impactDotHistory');
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

        if (historyGroup) {
            historyGroup.innerHTML = '';
            const hSign = this._isLeftHanded ? -1 : 1;
            for (let i = 0; i < this._shotHistory.length - 1; i++) {
                const prev = this._shotHistory[i].clubData;
                const pH = prev?.impactHorizontal;
                const pV = prev?.impactVertical;
                if (typeof pH !== 'number' || typeof pV !== 'number') continue;

                const hClamped = this.clamp(pH, -22, 22);
                const vClamped = this.clamp(pV, -22, 22);
                const histCx = 100 + this.impactMmToSvg(hClamped) * hSign;
                const histCy = 80 - this.impactMmToSvg(vClamped);

                const circle = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
                circle.setAttribute('cx', histCx);
                circle.setAttribute('cy', histCy);
                circle.setAttribute('r', '4');
                circle.setAttribute('fill', 'rgba(255,255,255,0.45)');
                circle.setAttribute('stroke', 'rgba(255,255,255,0.2)');
                circle.setAttribute('stroke-width', '1');
                historyGroup.appendChild(circle);
            }
        }

        const hMm = hasH ? this.clamp(clubData.impactHorizontal, -22, 22) : 0;
        const vMm = hasV ? this.clamp(clubData.impactVertical, -22, 22) : 0;

        const hSign = this._isLeftHanded ? -1 : 1;
        const cx = 100 + this.impactMmToSvg(hMm) * hSign;
        const cy = 80 - this.impactMmToSvg(vMm);

        dot.setAttribute('cx', cx);
        dot.setAttribute('cy', cy);
        dot.setAttribute('fill', '#ef4444');
        ring.setAttribute('cx', cx);
        ring.setAttribute('cy', cy);
        ring.setAttribute('stroke', '#ef4444');
        ring.setAttribute('opacity', '0.5');

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

        this.setDiagValue('diagImpactH', clubData.impactHorizontal, 'mm', 'impact', clubData.isImpactHorizontalValid);
        this.setDiagValue('diagImpactV', clubData.impactVertical, 'mm', false, clubData.isImpactVerticalValid);
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
            clubSpeedEl.textContent = hasClubSpeed ? `${(clubData.clubSpeed * 2.23694).toFixed(1)}` : '-';
        }
        if (smashEl) {
            smashEl.textContent = hasSmash ? `×${clubData.smashFactor.toFixed(2)}` : '-';
        }
        if (ballSpeedEl) {
            const mph = hasBallSpeed ? (ballData.speed * 2.23694).toFixed(1) : '-';
            ballSpeedEl.textContent = mph;
        }
    }
}
