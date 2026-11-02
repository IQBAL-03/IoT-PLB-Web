/**
 * @file Controller handling pump triggers and automatic/manual mode switching.
 */

import { state } from '../../core/state.js';
import { eventBus } from '../../core/events.js';
import { showToast } from '../../utils/toast.js';
import { pumpService } from './pump.service.js';
import { pumpView } from './pump.view.js';

export const pumpController = {
  init() {
    pumpView.renderMode(false);
    pumpView.renderPump(false, false, false);
    this.bindEvents();
  },

  bindEvents() {
    const el = pumpView.elements;

    if (el.modeSwitch) {
      el.modeSwitch.addEventListener('change', (e) => {
        el.modeSwitch._userChanged = true;
        this.setMode(e.target.checked);
      });
    }

    if (el.pumpOnBtn)  el.pumpOnBtn.addEventListener('click',  () => this.setPump(true));
    if (el.pumpOffBtn) el.pumpOffBtn.addEventListener('click', () => this.setPump(false));
  },

  async setPump(on) {
    pumpView.setPumpLoading(on, true);

    try {
      const res = await pumpService.setPump(on, state.currentDeviceId);
      state.pumpOn = on;
      pumpView.renderPump(on);
      showToast(res.message || (on ? 'Pompa dinyalakan' : 'Pompa dimatikan'), 'success');
      eventBus.emit('telemetry:refreshRequired');
    } catch (err) {
      showToast(`Gagal: ${err.message}`, 'error');
      console.error('[PlantCare] Pump control error:', err);
    } finally {
      pumpView.setPumpLoading(on, false);
    }
  },

  async setMode(isAuto) {
    const isManual = !isAuto;
    state.isManual = isManual;
    pumpView.renderMode(isManual);

    try {
      await pumpService.setMode(isAuto, state.currentDeviceId);
      showToast(`Mode diubah ke: ${isManual ? 'Manual' : 'Otomatis'}`, 'info');
      eventBus.emit('telemetry:refreshRequired');
    } catch (err) {
      showToast(`Gagal: ${err.message}`, 'error');
      console.error('[PlantCare] Mode change error:', err);
    } finally {
      setTimeout(() => {
        if (pumpView.elements.modeSwitch) {
          pumpView.elements.modeSwitch._userChanged = false;
        }
      }, 1500);
    }
  },
};

