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
    CRITICAL_THRESHOLD: 0.8
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
    workOrderShown: false,
    currentAlertId: null
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

// CRITICAL FAULT CHECK - MUST BE CALLED ON EVERY PAYLOAD
function checkCriticalFault(payload) {
    console.log('[CRITICAL CHECK] Function called with:', payload);

    if (payload.ml_label === 'bearing_fault' && payload.confidence > CONFIG.CRITICAL_THRESHOLD && !state.workOrderShown) {
        console.log('[ALERT] SHOWING MODAL NOW!');
        showWorkOrderModal(payload);
    }
}

// Make it global for testing
window.checkCriticalFault = checkCriticalFault;
