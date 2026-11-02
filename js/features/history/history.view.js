/**
 * @file History table and filter rendering view.
 */

import { $ } from '../../utils/dom.js';
import { formatTimestamp } from '../../utils/formatters.js';

export const historyView = {
  get elements() {
    return {
      historySkeleton:  $('historySkeleton'),
      historyTableWrap: $('historyTableWrap'),
      historyBody:      $('historyBody'),
      historyEmpty:     $('historyEmpty'),
      historyFilter:    $('historyFilter'),
      clearHistoryBtn:  $('clearHistoryBtn'),
    };
  },

  setSkeletonVisible(visible) {
    const el = this.elements;
    if (el.historySkeleton) {
      el.historySkeleton.style.display = visible ? 'block' : 'none';
    }
  },

  /**
   * Renders rows into the history table.
   * @param {Array} data
   * @param {string} filter - 'all' | 'auto' | 'manual'
   */
  renderRows(data, filter = 'all') {
    const el = this.elements;
    if (!el.historyBody) return;

    this.setSkeletonVisible(false);

    const filtered = filter === 'all'
      ? data
      : data.filter(r => r.mode === filter);

    if (!filtered || !filtered.length) {
      if (el.historyTableWrap) el.historyTableWrap.style.display = 'none';
      if (el.historyEmpty)     el.historyEmpty.style.display     = 'flex';
      return;
    }

    if (el.historyEmpty)     el.historyEmpty.style.display     = 'none';
    if (el.historyTableWrap) el.historyTableWrap.style.display = 'block';

    el.historyBody.innerHTML = filtered.map(r => `
      <tr>
        <td>${formatTimestamp(r.timestamp)}</td>
        <td><span class="badge badge--${(r.action || '').toLowerCase()}">${r.action}</span></td>
        <td><span class="badge badge--${r.mode}">${r.mode === 'auto' ? 'Otomatis' : 'Manual'}</span></td>
        <td>${r.moisture}%</td>
        <td>${r.duration || '—'}</td>
      </tr>
    `).join('');
  },
};

