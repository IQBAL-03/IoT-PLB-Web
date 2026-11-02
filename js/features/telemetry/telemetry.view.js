/**
 * @file Top bar connection indicator and system clock view.
 */

import { $ } from '../../utils/dom.js';

export const telemetryView = {
  get elements() {
    return {
      connectionBadge: $('connectionBadge'),
      connectionText:  $('connectionText'),
      currentTime:     $('currentTime'),
      refreshBtn:      $('refreshBtn'),
      refreshIcon:     $('refreshIcon'),
    };
  },

  updateClock() {
    const timeEl = $('currentTime');
    if (timeEl) {
      timeEl.textContent = new Date().toLocaleTimeString('id-ID', { hour12: false });
    }
  },

  setConnected(ok) {
    const badge = $('connectionBadge');
    const text = $('connectionText');
    if (!badge || !text) return;

    badge.className = `connection-badge ${ok ? 'connected' : 'error'}`;
    text.textContent = ok ? 'Terhubung' : 'Terputus';
  },

  setRefreshLoading(loading) {
    const btn = $('refreshBtn');
    const icon = $('refreshIcon');
    if (btn) btn.disabled = loading;
    if (icon) {
      icon.classList.toggle('spinning', loading);
    }
  },
};

