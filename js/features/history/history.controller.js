/**
 * @file History controller managing audit logs, filtering, and purges.
 */

import { state } from '../../core/state.js';
import { eventBus } from '../../core/events.js';
import { showToast } from '../../utils/toast.js';
import { historyService } from './history.service.js';
import { historyView } from './history.view.js';

export const historyController = {
  init() {
    this.bindEvents();

    eventBus.on('device:changed', () => this.fetchHistory());
    eventBus.on('telemetry:historySynced', () => this.fetchHistory());
  },

  bindEvents() {
    const el = historyView.elements;

    if (el.historyFilter) {
      el.historyFilter.addEventListener('change', (e) => {
        state.historyFilter = e.target.value;
        historyView.renderRows(state.history, state.historyFilter);
      });
    }

    if (el.clearHistoryBtn) {
      el.clearHistoryBtn.addEventListener('click', this.handleClearHistory.bind(this));
    }
  },

  async fetchHistory() {
    try {
      const data = await historyService.getHistory(state.currentDeviceId);
      state.history = data || [];
      historyView.renderRows(state.history, state.historyFilter);
    } catch (err) {
      historyView.setSkeletonVisible(false);
      historyView.renderRows([], state.historyFilter);
      console.error('[PlantCare] History fetch error:', err);
    }
  },

  async handleClearHistory() {
    if (!confirm(`Hapus seluruh riwayat untuk unit ${state.currentDeviceId}?`)) return;

    try {
      await historyService.clearHistory(state.currentDeviceId);
      showToast('Riwayat unit berhasil dibersihkan', 'info');
      await this.fetchHistory();
    } catch (err) {
      showToast(`Gagal: ${err.message}`, 'error');
      console.error('[PlantCare] Clear history error:', err);
    }
  },
};

