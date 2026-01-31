function checkCriticalFault(payload) {
    console.log('[CHECK] Entering checkCriticalFault, label:', payload.ml_label, 'conf:', payload.confidence);

    // Check if it's a bearing fault with high confidence and not already shown
    if (payload.ml_label === 'bearing_fault' &&
        payload.confidence > CONFIG.CRITICAL_THRESHOLD &&
        !state.workOrderShown) {

        console.log('[ALERT] ✅ SHOWING WORK ORDER POPUP!');
        console.log('[ALERT] Confidence:', payload.confidence, '> Threshold:', CONFIG.CRITICAL_THRESHOLD);

        // Show modal immediately
        showWorkOrderModal(payload);

    } else {
        // Log why it didn't trigger
        if (payload.ml_label !== 'bearing_fault') {
            console.log('[CHECK] ❌ Not bearing_fault, is:', payload.ml_label);
        } else if (payload.confidence <= CONFIG.CRITICAL_THRESHOLD) {
            console.log('[CHECK] ❌ Confidence too low:', payload.confidence, '<=', CONFIG.CRITICAL_THRESHOLD);
        } else if (state.workOrderShown) {
            console.log('[CHECK] ❌ Already shown this session');
        }
    }
}
