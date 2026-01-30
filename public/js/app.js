/**
 * Swadesh Labs - AI Predictive Maintenance Dashboard
 * Frontend JavaScript - SSE Connection & UI Updates
 */

// ============================================
// Configuration
// ============================================
const CONFIG = {
    SSE_ENDPOINT: '/events',
    CHART_MAX_POINTS: 20,
    COLORS: {
        healthy: '#00FF94',
        unbalance: '#FFC107',
        bearing_fault: '#FF003C'
    },
    CRITICAL_THRESHOLD: 0.85
};

// ============================================
// State Management
// ============================================
let state = {
    connected: false,
    lastPayload: null,
    predictionHistory: [],
    vibrationData: [],
    temperatureData: [],
    workOrderShown: false
};

// ============================================
// Three.js Motor Variables
// ============================================
let scene, camera, renderer, controls;
let motorGroup, shaftGroup, fanGroup;
let motorMaterial, glowMaterial;
let currentColor = new THREE.Color(0x00FF94);
let targetColor = new THREE.Color(0x00FF94);
let rotationSpeed = 0.02;
let vibrationIntensity = 0;

// ============================================
// 3D Motor Initialization
// ============================================
function init3DMotor() {
    const canvas = document.getElementById('motor-canvas');
    const container = canvas.parentElement;
    
    // Scene
    scene = new THREE.Scene();
    scene.background = new THREE.Color(0x12121a);
    
    // Camera
    camera = new THREE.PerspectiveCamera(45, container.clientWidth / container.clientHeight, 0.1, 1000);
    camera.position.set(8, 5, 10);
    
    // Renderer
    renderer = new THREE.WebGLRenderer({ canvas, antialias: true, alpha: true });
    renderer.setSize(container.clientWidth, container.clientHeight);
    renderer.setPixelRatio(window.devicePixelRatio);
    renderer.shadowMap.enabled = true;
    renderer.shadowMap.type = THREE.PCFSoftShadowMap;
    
    // Controls
    controls = new THREE.OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.05;
    controls.minDistance = 8;
    controls.maxDistance = 20;
    controls.maxPolarAngle = Math.PI / 1.8;
    
    // Lighting
    setupLighting();
    
    // Create Motor
    createMotor();
    
    // Grid helper
    const gridHelper = new THREE.GridHelper(20, 20, 0x333344, 0x222233);
    scene.add(gridHelper);
    
    // Animation loop
    animate();
    
    // Handle resize
    window.addEventListener('resize', onWindowResize);
}

function setupLighting() {
    // Ambient light
    const ambientLight = new THREE.AmbientLight(0x404040, 0.5);
    scene.add(ambientLight);
    
    // Main directional light
    const mainLight = new THREE.DirectionalLight(0xffffff, 1);
    mainLight.position.set(10, 15, 10);
    mainLight.castShadow = true;
    mainLight.shadow.mapSize.width = 2048;
    mainLight.shadow.mapSize.height = 2048;
    scene.add(mainLight);
    
    // Fill light
    const fillLight = new THREE.DirectionalLight(0x6366f1, 0.3);
    fillLight.position.set(-10, 5, -10);
    scene.add(fillLight);
    
    // Rim light
    const rimLight = new THREE.DirectionalLight(0x00FF94, 0.2);
    rimLight.position.set(0, -5, -10);
    scene.add(rimLight);
}

function createMotor() {
    motorGroup = new THREE.Group();
    
    // Materials
    const bodyMaterial = new THREE.MeshStandardMaterial({
        color: 0x2a2a3a,
        metalness: 0.7,
        roughness: 0.3
    });
    
    motorMaterial = new THREE.MeshStandardMaterial({
        color: 0x3a3a4a,
        metalness: 0.8,
        roughness: 0.2,
        emissive: 0x00FF94,
        emissiveIntensity: 0.1
    });
    
    const shaftMaterial = new THREE.MeshStandardMaterial({
        color: 0x888899,
        metalness: 0.9,
        roughness: 0.1
    });
    
    const finMaterial = new THREE.MeshStandardMaterial({
        color: 0x4a4a5a,
        metalness: 0.6,
        roughness: 0.4
    });
    
    // Main motor body (cylindrical)
    const bodyGeometry = new THREE.CylinderGeometry(1.8, 1.8, 4, 32);
    const motorBody = new THREE.Mesh(bodyGeometry, motorMaterial);
    motorBody.rotation.z = Math.PI / 2;
    motorBody.castShadow = true;
    motorBody.receiveShadow = true;
    motorGroup.add(motorBody);
    
    // Cooling fins
    const finCount = 24;
    for (let i = 0; i < finCount; i++) {
        const finGeometry = new THREE.BoxGeometry(3.8, 0.08, 0.3);
        const fin = new THREE.Mesh(finGeometry, finMaterial);
        const angle = (i / finCount) * Math.PI * 2;
        fin.position.y = Math.cos(angle) * 1.85;
        fin.position.z = Math.sin(angle) * 1.85;
        fin.rotation.x = angle;
        fin.castShadow = true;
        motorGroup.add(fin);
    }
    
    // Front end bell
    const endBellGeometry = new THREE.CylinderGeometry(1.9, 1.7, 0.5, 32);
    const frontEndBell = new THREE.Mesh(endBellGeometry, bodyMaterial);
    frontEndBell.rotation.z = Math.PI / 2;
    frontEndBell.position.x = 2.2;
    frontEndBell.castShadow = true;
    motorGroup.add(frontEndBell);
    
    // Rear end bell
    const rearEndBell = new THREE.Mesh(endBellGeometry, bodyMaterial);
    rearEndBell.rotation.z = Math.PI / 2;
    rearEndBell.position.x = -2.2;
    rearEndBell.castShadow = true;
    motorGroup.add(rearEndBell);
    
    // Shaft
    shaftGroup = new THREE.Group();
    const shaftGeometry = new THREE.CylinderGeometry(0.2, 0.2, 5, 16);
    const shaft = new THREE.Mesh(shaftGeometry, shaftMaterial);
    shaft.rotation.z = Math.PI / 2;
    shaft.castShadow = true;
    shaftGroup.add(shaft);
    
    // Shaft keyway
    const keywayGeometry = new THREE.BoxGeometry(1.5, 0.08, 0.15);
    const keyway = new THREE.Mesh(keywayGeometry, shaftMaterial);
    keyway.position.set(1.8, 0.18, 0);
    shaftGroup.add(keyway);
    
    motorGroup.add(shaftGroup);
    
    // Fan housing (rear)
    const fanHousingGeometry = new THREE.CylinderGeometry(1.3, 1.5, 0.8, 32);
    const fanHousing = new THREE.Mesh(fanHousingGeometry, bodyMaterial);
    fanHousing.rotation.z = Math.PI / 2;
    fanHousing.position.x = -2.8;
    fanHousing.castShadow = true;
    motorGroup.add(fanHousing);
    
    // Fan cover (with grille pattern)
    const fanCoverGeometry = new THREE.TorusGeometry(1.1, 0.15, 8, 32);
    const fanCover = new THREE.Mesh(fanCoverGeometry, bodyMaterial);
    fanCover.rotation.y = Math.PI / 2;
    fanCover.position.x = -3.3;
    motorGroup.add(fanCover);
    
    // Fan grille lines
    for (let i = 0; i < 8; i++) {
        const grilleGeometry = new THREE.CylinderGeometry(0.03, 0.03, 2.2, 8);
        const grille = new THREE.Mesh(grilleGeometry, bodyMaterial);
        const angle = (i / 8) * Math.PI;
        grille.position.x = -3.3;
        grille.rotation.x = angle;
        grille.rotation.z = Math.PI / 2;
        motorGroup.add(grille);
    }
    
    // Internal fan (rotating)
    fanGroup = new THREE.Group();
    const fanBladeCount = 6;
    for (let i = 0; i < fanBladeCount; i++) {
        const bladeGeometry = new THREE.BoxGeometry(0.05, 0.8, 0.3);
        const blade = new THREE.Mesh(bladeGeometry, shaftMaterial);
        const angle = (i / fanBladeCount) * Math.PI * 2;
        blade.position.y = Math.cos(angle) * 0.5;
        blade.position.z = Math.sin(angle) * 0.5;
        blade.rotation.x = angle + 0.3;
        fanGroup.add(blade);
    }
    fanGroup.position.x = -2.9;
    motorGroup.add(fanGroup);
    
    // Terminal box
    const terminalBoxGeometry = new THREE.BoxGeometry(1.2, 0.8, 1);
    const terminalBox = new THREE.Mesh(terminalBoxGeometry, bodyMaterial);
    terminalBox.position.set(0, 2.2, 0);
    terminalBox.castShadow = true;
    motorGroup.add(terminalBox);
    
    // Terminal box lid
    const lidGeometry = new THREE.BoxGeometry(1.0, 0.1, 0.8);
    const lid = new THREE.Mesh(lidGeometry, finMaterial);
    lid.position.set(0, 2.65, 0);
    motorGroup.add(lid);
    
    // Conduit fitting
    const conduitGeometry = new THREE.CylinderGeometry(0.15, 0.15, 0.4, 16);
    const conduit = new THREE.Mesh(conduitGeometry, shaftMaterial);
    conduit.position.set(0, 2.8, 0);
    motorGroup.add(conduit);
    
    // Mounting feet
    const footGeometry = new THREE.BoxGeometry(1.2, 0.3, 1.5);
    const leftFoot = new THREE.Mesh(footGeometry, bodyMaterial);
    leftFoot.position.set(-1.3, -2, 0);
    leftFoot.castShadow = true;
    leftFoot.receiveShadow = true;
    motorGroup.add(leftFoot);
    
    const rightFoot = new THREE.Mesh(footGeometry, bodyMaterial);
    rightFoot.position.set(1.3, -2, 0);
    rightFoot.castShadow = true;
    rightFoot.receiveShadow = true;
    motorGroup.add(rightFoot);
    
    // Mounting holes
    const holeGeometry = new THREE.CylinderGeometry(0.12, 0.12, 0.35, 16);
    const holeMaterial = new THREE.MeshStandardMaterial({ color: 0x111111 });
    
    [[-1.3, -0.5], [-1.3, 0.5], [1.3, -0.5], [1.3, 0.5]].forEach(([x, z]) => {
        const hole = new THREE.Mesh(holeGeometry, holeMaterial);
        hole.position.set(x, -2.1, z);
        motorGroup.add(hole);
    });
    
    // Nameplate
    const nameplateGeometry = new THREE.BoxGeometry(1.5, 0.02, 0.8);
    const nameplateMaterial = new THREE.MeshStandardMaterial({ 
        color: 0xcccccc,
        metalness: 0.8,
        roughness: 0.2
    });
    const nameplate = new THREE.Mesh(nameplateGeometry, nameplateMaterial);
    nameplate.position.set(0, 1.81, 0.8);
    nameplate.rotation.x = -0.1;
    motorGroup.add(nameplate);
    
    // Bearing housings (visible rings)
    const bearingGeometry = new THREE.TorusGeometry(0.35, 0.08, 8, 32);
    const bearingMaterial = new THREE.MeshStandardMaterial({
        color: 0x666677,
        metalness: 0.9,
        roughness: 0.1
    });
    
    const frontBearing = new THREE.Mesh(bearingGeometry, bearingMaterial);
    frontBearing.rotation.y = Math.PI / 2;
    frontBearing.position.x = 2.45;
    motorGroup.add(frontBearing);
    
    const rearBearing = new THREE.Mesh(bearingGeometry, bearingMaterial);
    rearBearing.rotation.y = Math.PI / 2;
    rearBearing.position.x = -2.45;
    motorGroup.add(rearBearing);
    
    // Status indicator light
    const indicatorGeometry = new THREE.SphereGeometry(0.1, 16, 16);
    glowMaterial = new THREE.MeshBasicMaterial({ color: 0x00FF94 });
    const indicator = new THREE.Mesh(indicatorGeometry, glowMaterial);
    indicator.position.set(0.4, 2.65, 0.35);
    motorGroup.add(indicator);
    
    // Position the motor group
    motorGroup.position.y = 2;
    scene.add(motorGroup);
    
    // Add base platform
    const platformGeometry = new THREE.BoxGeometry(6, 0.2, 3);
    const platformMaterial = new THREE.MeshStandardMaterial({
        color: 0x1a1a25,
        metalness: 0.5,
        roughness: 0.5
    });
    const platform = new THREE.Mesh(platformGeometry, platformMaterial);
    platform.position.y = -0.1;
    platform.receiveShadow = true;
    scene.add(platform);
}

function animate() {
    requestAnimationFrame(animate);
    
    // Rotate shaft and fan
    if (shaftGroup) {
        shaftGroup.rotation.x += rotationSpeed;
    }
    if (fanGroup) {
        fanGroup.rotation.x += rotationSpeed * 2;
    }
    
    // Apply vibration effect
    if (motorGroup && vibrationIntensity > 0) {
        motorGroup.position.y = 2 + (Math.random() - 0.5) * vibrationIntensity * 0.05;
        motorGroup.position.x = (Math.random() - 0.5) * vibrationIntensity * 0.02;
    }
    
    // Smooth color transition
    currentColor.lerp(targetColor, 0.05);
    if (motorMaterial) {
        motorMaterial.emissive.copy(currentColor);
    }
    if (glowMaterial) {
        glowMaterial.color.copy(currentColor);
    }
    
    controls.update();
    renderer.render(scene, camera);
}

function onWindowResize() {
    const container = document.getElementById('motor-3d');
    camera.aspect = container.clientWidth / container.clientHeight;
    camera.updateProjectionMatrix();
    renderer.setSize(container.clientWidth, container.clientHeight);
}

function updateMotor3D(label, confidence) {
    // Set target color based on label
    switch (label) {
        case 'healthy':
            targetColor.setHex(0x00FF94);
            vibrationIntensity = 0;
            rotationSpeed = 0.02;
            break;
        case 'unbalance':
            targetColor.setHex(0xFFC107);
            vibrationIntensity = 0.3;
            rotationSpeed = 0.015;
            break;
        case 'bearing_fault':
            targetColor.setHex(0xFF003C);
            vibrationIntensity = 0.8 + (confidence * 0.5);
            rotationSpeed = 0.01;
            break;
    }
    
    // Update emissive intensity based on confidence
    if (motorMaterial) {
        motorMaterial.emissiveIntensity = 0.1 + (confidence * 0.3);
    }
    
    // Update status ring
    const statusRing = document.getElementById('status-ring');
    statusRing.classList.remove('healthy', 'unbalance', 'fault');
    
    switch (label) {
        case 'healthy':
            statusRing.classList.add('healthy');
            break;
        case 'unbalance':
            statusRing.classList.add('unbalance');
            break;
        case 'bearing_fault':
            statusRing.classList.add('fault');
            break;
    }
}

// ============================================
// Chart Initialization
// ============================================
let vibrationChart, temperatureChart;

function initCharts() {
    const chartConfig = {
        responsive: true,
        maintainAspectRatio: false,
        animation: { duration: 300 },
        scales: {
            x: {
                display: false
            },
            y: {
                grid: { color: 'rgba(255, 255, 255, 0.05)' },
                ticks: { color: '#a0a0b0', font: { size: 10 } }
            }
        },
        plugins: {
            legend: { display: false }
        }
    };

    // Vibration Chart
    const vibCtx = document.getElementById('vibrationChart').getContext('2d');
    vibrationChart = new Chart(vibCtx, {
        type: 'line',
        data: {
            labels: [],
            datasets: [{
                data: [],
                borderColor: CONFIG.COLORS.healthy,
                backgroundColor: 'rgba(0, 255, 148, 0.1)',
                fill: true,
                tension: 0.4,
                pointRadius: 0
            }]
        },
        options: {
            ...chartConfig,
            scales: {
                ...chartConfig.scales,
                y: {
                    ...chartConfig.scales.y,
                    min: 200,
                    max: 600
                }
            }
        }
    });

    // Temperature Chart
    const tempCtx = document.getElementById('temperatureChart').getContext('2d');
    temperatureChart = new Chart(tempCtx, {
        type: 'line',
        data: {
            labels: [],
            datasets: [{
                data: [],
                borderColor: '#ff6b6b',
                backgroundColor: 'rgba(255, 107, 107, 0.1)',
                fill: true,
                tension: 0.4,
                pointRadius: 0
            }]
        },
        options: {
            ...chartConfig,
            scales: {
                ...chartConfig.scales,
                y: {
                    ...chartConfig.scales.y,
                    min: 30,
                    max: 90
                }
            }
        }
    });
}

// ============================================
// SSE Connection
// ============================================
function connectSSE() {
    console.log('[SSE] Connecting to', CONFIG.SSE_ENDPOINT);
    
    const eventSource = new EventSource(CONFIG.SSE_ENDPOINT);
    
    eventSource.onopen = () => {
        console.log('[SSE] Connected');
        updateConnectionStatus(true);
    };
    
    eventSource.onmessage = (event) => {
        try {
            const data = JSON.parse(event.data);
            
            // Skip connection acknowledgment
            if (data.type === 'connected') {
                console.log('[SSE] Server acknowledged connection');
                return;
            }
            
            // Process ML payload
            handleMLPayload(data);
        } catch (err) {
            console.error('[SSE] Parse error:', err);
        }
    };
    
    eventSource.onerror = (err) => {
        console.error('[SSE] Connection error:', err);
        updateConnectionStatus(false);
        
        // Reconnect after 3 seconds
        setTimeout(() => {
            console.log('[SSE] Attempting reconnection...');
            eventSource.close();
            connectSSE();
        }, 3000);
    };
}

function updateConnectionStatus(connected) {
    state.connected = connected;
    const badge = document.getElementById('connection-status');
    const statusText = badge.querySelector('.status-text');
    const aiIndicator = document.getElementById('ai-indicator');
    
    if (connected) {
        badge.classList.add('connected');
        statusText.textContent = 'Online';
        if (aiIndicator) aiIndicator.classList.add('active');
    } else {
        badge.classList.remove('connected');
        statusText.textContent = 'Reconnecting';
        if (aiIndicator) aiIndicator.classList.remove('active');
    }
}

// System time display
function updateSystemTime() {
    const timeEl = document.getElementById('system-time');
    if (timeEl) {
        const now = new Date();
        timeEl.textContent = now.toLocaleTimeString('en-US', { 
            hour12: false, 
            hour: '2-digit', 
            minute: '2-digit', 
            second: '2-digit' 
        });
    }
}
setInterval(updateSystemTime, 1000);

// ============================================
// ML Payload Handler
// ============================================
function handleMLPayload(payload) {
    console.log('[ML] Received:', payload);
    state.lastPayload = payload;
    
    // Update all UI components
    updateConfidenceMeter(payload.confidence);
    updateMotorStatus(payload.ml_label);
    updateAIStatusCard(payload);
    updateTelemetry(payload.telemetry);
    updateCharts(payload.telemetry);
    updateAnomalyScore(payload.anomaly_score);
    addToPredictionHistory(payload);
    
    // Check for critical fault
    checkCriticalFault(payload);
}

// ============================================
// UI Update Functions
// ============================================

// Confidence Meter
function updateConfidenceMeter(confidence) {
    const percentage = Math.round(confidence * 100);
    const bar = document.getElementById('confidence-bar');
    const value = document.getElementById('confidence-value');
    
    bar.style.width = `${percentage}%`;
    value.textContent = percentage;
    
    // Update bar color based on confidence
    bar.classList.remove('warning', 'danger');
    if (percentage > 95) {
        bar.classList.add('danger');
    } else if (percentage > 85) {
        bar.classList.add('warning');
    }
}

// Motor Status & 3D Color
function updateMotorStatus(label) {
    const statusLabel = document.querySelector('.status-label');
    
    // Remove all classes
    statusLabel.classList.remove('healthy', 'unbalance', 'fault');
    
    // Apply new class based on label
    let displayLabel = 'UNKNOWN';
    let className = '';
    
    switch (label) {
        case 'healthy':
            displayLabel = 'HEALTHY';
            className = 'healthy';
            break;
        case 'unbalance':
            displayLabel = 'UNBALANCE';
            className = 'unbalance';
            break;
        case 'bearing_fault':
            displayLabel = 'BEARING FAULT';
            className = 'fault';
            break;
    }
    
    statusLabel.classList.add(className);
    statusLabel.textContent = displayLabel;
    
    // Update 3D motor
    if (state.lastPayload) {
        updateMotor3D(label, state.lastPayload.confidence);
    }
    
    // Update RPM display based on status
    const rpmDisplay = document.getElementById('motor-rpm');
    if (rpmDisplay) {
        switch (label) {
            case 'healthy':
                rpmDisplay.textContent = '1450';
                break;
            case 'unbalance':
                rpmDisplay.textContent = '1380';
                break;
            case 'bearing_fault':
                rpmDisplay.textContent = '1250';
                break;
        }
    }
}

// AI Status Card
function updateAIStatusCard(payload) {
    const predictionLabel = document.getElementById('prediction-label');
    const sublabel = document.querySelector('.prediction-sublabel');
    const explanation = document.getElementById('ai-explanation');
    const confidenceBadge = document.getElementById('confidence-badge');
    const anomalyBadge = document.getElementById('anomaly-badge');
    
    // Remove all classes
    predictionLabel.classList.remove('success', 'warning', 'danger');
    
    let displayLabel = payload.ml_label.toUpperCase().replace('_', ' ');
    let className = '';
    let explanationText = '';
    let badgeText = '';
    
    switch (payload.ml_label) {
        case 'healthy':
            className = 'success';
            explanationText = 'All monitored parameters within acceptable thresholds. No anomalies detected in vibration spectrum or thermal signature.';
            badgeText = 'NORMAL';
            break;
        case 'unbalance':
            className = 'warning';
            explanationText = 'Rotational imbalance detected. Recommend inspection of mounting bolts, shaft alignment, and coupling condition.';
            badgeText = 'ATTENTION';
            break;
        case 'bearing_fault':
            className = 'danger';
            explanationText = 'Bearing defect signature detected (BPFI frequency). Inner race degradation pattern identified. Immediate maintenance recommended.';
            badgeText = 'CRITICAL';
            break;
    }
    
    predictionLabel.textContent = displayLabel;
    predictionLabel.classList.add(className);
    sublabel.textContent = `Confidence: ${Math.round(payload.confidence * 100)}%`;
    explanation.innerHTML = `<p>${explanationText}</p>`;
    
    if (confidenceBadge) {
        confidenceBadge.textContent = badgeText;
        confidenceBadge.className = 'card-badge ' + (className === 'danger' ? 'critical' : className);
    }
    
    // Update anomaly badge
    if (anomalyBadge) {
        const anomaly = payload.anomaly_score;
        if (anomaly > 0.5) {
            anomalyBadge.textContent = 'HIGH';
            anomalyBadge.className = 'card-badge critical';
        } else if (anomaly > 0.3) {
            anomalyBadge.textContent = 'MEDIUM';
            anomalyBadge.className = 'card-badge warning';
        } else {
            anomalyBadge.textContent = 'LOW';
            anomalyBadge.className = 'card-badge';
        }
    }
}

// Telemetry Data
function updateTelemetry(telemetry) {
    document.getElementById('telemetry-vibration').textContent = telemetry.vibration_peak.toFixed(0);
    document.getElementById('telemetry-current').textContent = telemetry.current_amps.toFixed(2);
    document.getElementById('telemetry-temp').textContent = telemetry.temperature_c.toFixed(1);
}

// Charts
function updateCharts(telemetry) {
    const timestamp = new Date().toLocaleTimeString();
    
    // Update vibration data
    state.vibrationData.push(telemetry.vibration_peak);
    if (state.vibrationData.length > CONFIG.CHART_MAX_POINTS) {
        state.vibrationData.shift();
    }
    
    // Update temperature data
    state.temperatureData.push(telemetry.temperature_c);
    if (state.temperatureData.length > CONFIG.CHART_MAX_POINTS) {
        state.temperatureData.shift();
    }
    
    // Update chart labels
    const labels = state.vibrationData.map((_, i) => '');
    
    // Update vibration chart
    vibrationChart.data.labels = labels;
    vibrationChart.data.datasets[0].data = state.vibrationData;
    
    // Change color based on current state
    const color = CONFIG.COLORS[state.lastPayload?.ml_label] || CONFIG.COLORS.healthy;
    vibrationChart.data.datasets[0].borderColor = color;
    vibrationChart.data.datasets[0].backgroundColor = color.replace(')', ', 0.1)').replace('rgb', 'rgba');
    
    vibrationChart.update('none');
    
    // Update temperature chart
    temperatureChart.data.labels = labels;
    temperatureChart.data.datasets[0].data = state.temperatureData;
    temperatureChart.update('none');
}

// Anomaly Score
function updateAnomalyScore(score) {
    const percentage = Math.round(score * 100);
    document.getElementById('anomaly-bar').style.width = `${percentage}%`;
    document.getElementById('anomaly-value').textContent = score.toFixed(2);
}

// Prediction History
let eventCounter = 0;

function addToPredictionHistory(payload) {
    const history = document.getElementById('prediction-history');
    const eventCountEl = document.getElementById('event-count');
    
    // Remove placeholder if exists
    const placeholder = history.querySelector('.placeholder');
    if (placeholder) {
        placeholder.remove();
    }
    
    // Increment counter
    eventCounter++;
    if (eventCountEl) {
        eventCountEl.textContent = eventCounter;
    }
    
    // Create history item
    const item = document.createElement('div');
    item.className = 'history-item';
    
    let labelClass = '';
    switch (payload.ml_label) {
        case 'healthy': labelClass = 'healthy'; break;
        case 'unbalance': labelClass = 'unbalance'; break;
        case 'bearing_fault': labelClass = 'fault'; break;
    }
    
    const time = new Date().toLocaleTimeString('en-US', { 
        hour12: false, 
        hour: '2-digit', 
        minute: '2-digit', 
        second: '2-digit' 
    });
    
    item.innerHTML = `
        <span class="label ${labelClass}">${payload.ml_label.replace('_', ' ')}</span>
        <span class="confidence">${Math.round(payload.confidence * 100)}% | ${time}</span>
    `;
    
    // Add to beginning
    history.insertBefore(item, history.firstChild);
    
    // Keep only last 10
    while (history.children.length > 10) {
        history.removeChild(history.lastChild);
    }
    
    // Store in state
    state.predictionHistory.unshift(payload);
    if (state.predictionHistory.length > 50) {
        state.predictionHistory.pop();
    }
}

// ============================================
// Critical Fault Detection & Work Order Modal
// ============================================
function checkCriticalFault(payload) {
    // Only show if bearing_fault AND confidence > 85% AND not already shown
    if (payload.ml_label === 'bearing_fault' && 
        payload.confidence > CONFIG.CRITICAL_THRESHOLD && 
        !state.workOrderShown) {
        
        console.log('[ALERT] Critical fault detected! Opening work order...');
        showWorkOrderModal(payload);
    }
}

function showWorkOrderModal(payload) {
    state.workOrderShown = true;
    
    const modal = document.getElementById('work-order-modal');
    
    // Update modal content
    document.getElementById('modal-label').textContent = payload.ml_label.toUpperCase().replace('_', ' ');
    document.getElementById('modal-confidence').textContent = `${Math.round(payload.confidence * 100)}%`;
    document.getElementById('modal-timestamp').textContent = new Date().toLocaleString();
    
    // Show modal
    modal.classList.add('active');
    
    // Play alert sound (optional)
    playAlertSound();
}

function closeWorkOrderModal() {
    const modal = document.getElementById('work-order-modal');
    modal.classList.remove('active');
    
    // Reset after 30 seconds to allow re-triggering
    setTimeout(() => {
        state.workOrderShown = false;
    }, 30000);
}

function acknowledgeWorkOrder() {
    console.log('[WORK ORDER] Acknowledged by operator');
    closeWorkOrderModal();
    
    // Here you could make an API call to create a ticket
    // fetch('/api/work-orders', { method: 'POST', ... })
}

function playAlertSound() {
    // Create a simple beep sound using Web Audio API
    try {
        const audioContext = new (window.AudioContext || window.webkitAudioContext)();
        const oscillator = audioContext.createOscillator();
        const gainNode = audioContext.createGain();
        
        oscillator.connect(gainNode);
        gainNode.connect(audioContext.destination);
        
        oscillator.frequency.value = 800;
        oscillator.type = 'sine';
        gainNode.gain.value = 0.3;
        
        oscillator.start();
        oscillator.stop(audioContext.currentTime + 0.3);
    } catch (e) {
        // Audio not supported or blocked
    }
}

// Make modal functions globally available
window.closeWorkOrderModal = closeWorkOrderModal;
window.acknowledgeWorkOrder = acknowledgeWorkOrder;

// ============================================
// Initialization
// ============================================
document.addEventListener('DOMContentLoaded', () => {
    console.log('[Swadesh Labs] Initializing dashboard...');
    
    // Initialize system time
    updateSystemTime();
    
    // Initialize 3D Motor
    init3DMotor();
    
    // Initialize charts
    initCharts();
    
    // Connect to SSE
    connectSSE();
    
    console.log('[Swadesh Labs] Dashboard ready');
});
