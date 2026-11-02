/**
 * @file Simulator controller for testing telemetry and pump webhooks.
 */

import { state } from '../../core/state.js';
import { eventBus } from '../../core/events.js';
import { showToast } from '../../utils/toast.js';
import { simulatorService } from './simulator.service.js';
import { simulatorView } from './simulator.view.js';

export const simulatorController = {
  init() {
    const el = simulatorView.elements;
    if (el.testForm) {
      el.testForm.addEventListener('submit', this.handleSend.bind(this));
    }
  },

  async handleSend(e) {
    e.preventDefault();
    const el = simulatorView.elements;
    const moisture = parseFloat(el.testMoisture.value) || 0;
    const action = el.testPump.value;
    const pumpOn = action === 'ON';
    const mode = el.testMode.value;

    simulatorView.setLoading(true);

    try {
      await simulatorService.sendTelemetry({
        device_id: state.currentDeviceId,
        moisture: moisture,
        pump_on: pumpOn,
        action: action,
        mode: mode,
        force_log: true,
      });

      showToast(`Data unit ${state.currentDeviceId} tersimpan! Kelembapan: ${moisture}%, Pompa: ${action}`, 'success');
      eventBus.emit('telemetry:refreshRequired');
    } catch (err) {
      showToast(`Gagal: ${err.message}`, 'error');
      console.error('[PlantCare] Test send error:', err);
    } finally {
      simulatorView.setLoading(false);
    }
  },
};

