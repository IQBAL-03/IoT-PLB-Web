/**
 * @file Soil moisture gauge and visual status renderer.
 */

import { $ } from '../../utils/dom.js';

export const moistureView = {
  get elements() {
    return {
      gaugeArc:       $('gaugeArc'),
      moistureValue:  $('moistureValue'),
      moistureStatus: $('moistureStatus'),
      moistureBar:    $('moistureBar'),
    };
  },

  /**
   * Updates soil moisture gauge, bar width, and status label according to moisture level.
   * @param {number|null} pct
   */
  renderMoisture(pct) {
    if (pct === null || pct === undefined) return;
    const el = this.elements;
    if (!el.gaugeArc) return;

    const circumference = 314;
    el.gaugeArc.style.strokeDashoffset = circumference - (pct / 100) * circumference;
    el.moistureValue.textContent = pct;
    el.moistureBar.style.width   = `${pct}%`;

    let statusText, barColor, strokeColor;
    if (pct < 30) {
      statusText  = '🔴 Sangat Kering — Perlu Disiram!';
      barColor    = 'linear-gradient(90deg, #e63946, #f4a261)';
      strokeColor = '#e63946';
    } else if (pct < 50) {
      statusText  = '🟡 Cukup Kering — Pantau Terus';
      barColor    = 'linear-gradient(90deg, #f4a261, #ffd166)';
      strokeColor = '#f4a261';
    } else if (pct < 75) {
      statusText  = '🟢 Lembap — Kondisi Baik';
      barColor    = 'linear-gradient(90deg, var(--green-600), var(--green-400))';
      strokeColor = '#52B788';
    } else {
      statusText  = '💧 Sangat Lembap — Tidak Perlu Disiram';
      barColor    = 'linear-gradient(90deg, #0096c7, #52B788)';
      strokeColor = '#0096c7';
    }

    el.moistureStatus.textContent   = statusText;
    el.moistureBar.style.background = barColor;
    el.gaugeArc.style.stroke        = strokeColor;
  },
};

