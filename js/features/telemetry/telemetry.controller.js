/**
 * @file Telemetry controller handling polling, cross-device sync, and connection health.
 */

import { CONFIG } from '../../config/constants.js';
import { state } from '../../core/state.js';
import { eventBus } from '../../core/events.js';
import { showToast } from '../../utils/toast.js';
import { pumpView } from '../pump/pump.view.js';
import { moistureView } from '../moisture/moisture.view.js';
import { telemetryService } from './telemetry.service.js';
import { telemetryView } from './telemetry.view.js';

export const telemetryController = {
  init() {
    telemetryView.updateClock();
    setInterval(() => telemetryView.updateClock(), 1000);

    this.bindEvents();

    eventBus.on('device:changed', () => this.fetchStatus());
    eventBus.on('telemetry:refreshRequired', () => this.fetchStatus());
  },

  bindEvents() {
    const el = telemetryView.elements;
    if (el.refreshBtn) {
      el.refreshBtn.addEventListener('click', this.handleRefresh.bind(this));
    }

    const syncAll = () => {
      this.fetchStatus();
      eventBus.emit('telemetry:historySynced');
    };

    window.addEventListener('focus', syncAll);
    document.addEventListener('visibilitychange', () => {
      if (!document.hidden) syncAll();
    });
  },

  async fetchStatus() {
    const now = Date.now();
    if (document.hidden && now - state.lastFetch < 10000) return;
    state.lastFetch = now;

    try {
      const data = await telemetryService.getStatus(state.currentDeviceId);
      telemetryView.setConnected(true);
      state.pumpOn = data.pump_on;
      state.moisture = data.moisture;

      pumpView.renderPump(data.pump_on);
      moistureView.renderMoisture(data.moisture);

      // Sync mode switch state if user isn't toggling
      const modeSwitch = pumpView.elements.modeSwitch;
      if (data.mode && modeSwitch && !modeSwitch._userChanged) {
        const isManual = data.mode === 'manual';
        if (modeSwitch.checked !== !isManual) {
          modeSwitch.checked = !isManual;
          state.isManual = isManual;
          pumpView.renderMode(isManual);
        }
      }

      // Cross-device sync: trigger history fetch if server timestamp changed
      if (data.timestamp && data.timestamp !== state.lastTimestamp) {
        state.lastTimestamp = data.timestamp;
        eventBus.emit('telemetry:historySynced');
      }
    } catch (err) {
      telemetryView.setConnected(false);
      console.error('[PlantCare] Status fetch error:', err);
    }
  },

  async keepAlive() {
    if (CONFIG.MOCK_MODE) return;
    try {
      await telemetryService.ping();
    } catch {}
  },

  async handleRefresh() {
    telemetryView.setRefreshLoading(true);
    try {
      await Promise.all([
        this.fetchStatus(),
        new Promise(r => {
          eventBus.emit('telemetry:historySynced');
          eventBus.emit('auth:changed');
          setTimeout(r, 400);
        }),
      ]);
      showToast('Data diperbarui', 'success');
    } finally {
      telemetryView.setRefreshLoading(false);
    }
  },
};

