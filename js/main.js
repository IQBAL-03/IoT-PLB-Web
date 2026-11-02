/**
 * @file Main application entry point & orchestrator.
 */

import { CONFIG } from './config/constants.js';
import { eventBus } from './core/events.js';
import { authController } from './features/auth/auth.controller.js';
import { authView } from './features/auth/auth.view.js';
import { deviceController } from './features/devices/device.controller.js';
import { pumpController } from './features/pump/pump.controller.js';
import { moistureView } from './features/moisture/moisture.view.js';
import { historyController } from './features/history/history.controller.js';
import { simulatorController } from './features/simulator/simulator.controller.js';
import { telemetryController } from './features/telemetry/telemetry.controller.js';

async function bootstrap() {
  // Listen for global auth modal open requests
  eventBus.on('auth:openModal', () => {
    authView.openModal();
  });

  // Set default uninitialized presentation state
  moistureView.renderMoisture(null, false, false);

  // Initialize all feature controllers
  authController.init();
  pumpController.init();
  historyController.init();
  simulatorController.init();
  telemetryController.init();

  // Load initial dataset
  await deviceController.init();
  await telemetryController.fetchStatus();
  await historyController.fetchHistory();

  // Background polling intervals
  setInterval(() => telemetryController.fetchStatus(), CONFIG.POLL_INTERVAL);
  setInterval(() => historyController.fetchHistory(), CONFIG.HISTORY_POLL_INTERVAL);
  setInterval(() => telemetryController.keepAlive(), CONFIG.KEEPALIVE_INTERVAL);
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', bootstrap);
} else {
  bootstrap();
}

