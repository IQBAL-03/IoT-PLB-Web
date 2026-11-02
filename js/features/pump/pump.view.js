/**
 * @file Pump and mode card presentation view.
 */

import { $ } from '../../utils/dom.js';

export const pumpView = {
  get elements() {
    return {
      pumpCard:          $('pumpCard'),
      pumpIndicator:     $('pumpIndicator'),
      pumpDot:           $('pumpDot'),
      pumpState:         $('pumpState'),
      pumpDetail:        $('pumpDetail'),
      modeSwitch:        $('modeSwitch'),
      modeCurrent:       $('modeCurrent'),
      modeDesc:          $('modeDesc'),
      modeManualLabel:   $('modeManualLabel'),
      modeAutoLabel:     $('modeAutoLabel'),
      manualBadgeStatus: $('manualBadgeStatus'),
      controlPanel:      $('controlPanel'),
      controlInfo:       $('controlInfo'),
      pumpOnBtn:         $('pumpOnBtn'),
      pumpOffBtn:        $('pumpOffBtn'),
    };
  },

  renderPump(on) {
    const el = this.elements;
    if (!el.pumpCard) return;

    el.pumpCard.classList.remove('pump-on', 'pump-off');
    el.pumpCard.classList.add(on ? 'pump-on' : 'pump-off');
    el.pumpIndicator.classList.remove('pump-on', 'pump-off');
    el.pumpIndicator.classList.add(on ? 'pump-on' : 'pump-off');
    el.pumpState.textContent  = on ? 'AKTIF' : 'MATI';
    el.pumpDetail.textContent = on ? 'Pompa sedang menyiram tanaman' : 'Pompa dalam keadaan standby';
  },

  renderMode(isManual) {
    const el = this.elements;
    if (!el.modeSwitch) return;

    el.modeManualLabel.classList.toggle('mode-label--active', isManual);
    el.modeAutoLabel.classList.toggle('mode-label--active', !isManual);
    el.modeCurrent.textContent = isManual ? 'Mode: Manual' : 'Mode: Otomatis';
    el.modeDesc.textContent    = isManual
      ? 'Pompa dikendalikan secara manual melalui tombol di bawah.'
      : 'Pompa bekerja otomatis berdasarkan pembacaan sensor kelembapan.';

    el.pumpOnBtn.disabled  = !isManual;
    el.pumpOffBtn.disabled = !isManual;
    el.controlPanel.classList.toggle('is-active', isManual);
    el.controlInfo.textContent = isManual
      ? 'Mode manual aktif. Gunakan tombol di bawah untuk mengontrol pompa.'
      : 'Aktifkan mode manual untuk mengontrol pompa secara langsung.';

    if (el.manualBadgeStatus) {
      el.manualBadgeStatus.textContent = isManual ? 'Mode Manual Aktif' : 'Mode Otomatis Aktif';
      el.manualBadgeStatus.className   = `section__badge${isManual ? ' manual' : ''}`;
    }
  },

  setPumpLoading(on, loading) {
    const el = this.elements;
    const btn = on ? el.pumpOnBtn : el.pumpOffBtn;
    if (!btn) return;

    if (loading) {
      btn._originalHTML = btn.innerHTML;
      btn.disabled = true;
      btn.innerHTML = `
        <svg class="btn__icon spinning" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <polyline points="23 4 23 10 17 10"></polyline>
          <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10"></path>
        </svg>
        ${on ? 'Menyalakan...' : 'Mematikan...'}
      `;
    } else {
      if (btn._originalHTML) btn.innerHTML = btn._originalHTML;
      btn.disabled = !$('modeSwitch')?.checked === false; // disabled if auto
    }
  },
};

