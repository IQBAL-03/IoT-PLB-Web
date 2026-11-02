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
      modeAutoBtn:       $('modeAutoBtn'),
      modeManualBtn:     $('modeManualBtn'),
      modeBadge:         $('modeBadge'),
      modeInfoBox:       $('modeInfoBox'),
      modeCurrent:       $('modeCurrent'),
      modeDesc:          $('modeDesc'),
      modeSwitch:        $('modeSwitch'),
      manualBadgeStatus: $('manualBadgeStatus'),
      controlPanel:      $('controlPanel'),
      controlInfo:       $('controlInfo'),
      pumpOnBtn:         $('pumpOnBtn'),
      pumpOffBtn:        $('pumpOffBtn'),
    };
  },

  /**
   * Renders pump card state and active glowing indicators.
   * If unit is unlinked or IoT is offline, displays default OFF standby state.
   * @param {boolean} on
   * @param {boolean} [isOnline=true]
   * @param {boolean} [isLinked=true]
   */
  renderPump(on, isOnline = true, isLinked = true) {
    const el = this.elements;
    if (!el.pumpCard) return;

    const isPumpActive = Boolean(on && isOnline && isLinked);

    el.pumpCard.classList.remove('pump-on', 'pump-off');
    el.pumpCard.classList.add(isPumpActive ? 'pump-on' : 'pump-off');
    el.pumpIndicator.classList.remove('pump-on', 'pump-off');
    el.pumpIndicator.classList.add(isPumpActive ? 'pump-on' : 'pump-off');

    if (!isLinked) {
      el.pumpState.textContent = 'MATI';
      el.pumpDetail.textContent = 'Unit belum dihubungkan ke akun';
    } else if (!isOnline) {
      el.pumpState.textContent = 'MATI';
      el.pumpDetail.textContent = 'Perangkat IoT tidak terdeteksi (Offline)';
    } else {
      el.pumpState.textContent  = isPumpActive ? 'AKTIF' : 'MATI';
      el.pumpDetail.textContent = isPumpActive
        ? 'Pompa sedang menyiram tanaman'
        : 'Pompa dalam keadaan standby';
    }
  },

  /**
   * Updates segmented control buttons, header badge, and explanation box according to active mode.
   * @param {boolean} isManual
   */
  renderMode(isManual) {
    const el = this.elements;

    if (el.modeSwitch) {
      el.modeSwitch.checked = !isManual;
    }

    if (el.modeAutoBtn && el.modeManualBtn) {
      el.modeAutoBtn.classList.toggle('segmented-btn--active', !isManual);
      el.modeAutoBtn.setAttribute('aria-pressed', !isManual ? 'true' : 'false');
      el.modeManualBtn.classList.toggle('segmented-btn--active', isManual);
      el.modeManualBtn.setAttribute('aria-pressed', isManual ? 'true' : 'false');
    }

    if (el.modeBadge) {
      el.modeBadge.textContent = isManual ? 'Manual' : 'Otomatis';
      el.modeBadge.className = `mode-badge ${isManual ? 'mode-badge--manual' : 'mode-badge--auto'}`;
    }

    if (el.modeInfoBox) {
      el.modeInfoBox.classList.toggle('mode-info--manual', isManual);
    }

    if (el.modeCurrent) {
      el.modeCurrent.textContent = isManual ? 'Mode Aktif: Manual' : 'Mode Aktif: Otomatis';
    }

    if (el.modeDesc) {
      el.modeDesc.textContent = isManual
        ? 'Mode manual aktif. Pompa dapat dinyalakan atau dimatikan secara langsung melalui tombol kontrol di bawah.'
        : 'Mode otomatis aktif. Pompa menyiram secara mandiri saat sensor mendeteksi kelembapan tanah di bawah batas.';
    }

    if (el.pumpOnBtn)  el.pumpOnBtn.disabled  = !isManual;
    if (el.pumpOffBtn) el.pumpOffBtn.disabled = !isManual;
    if (el.controlPanel) el.controlPanel.classList.toggle('is-active', isManual);
    if (el.controlInfo) {
      el.controlInfo.textContent = isManual
        ? 'Mode manual aktif. Gunakan tombol di bawah untuk mengontrol pompa.'
        : 'Aktifkan mode manual untuk mengontrol pompa secara langsung.';
    }

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
      btn.disabled = $('modeSwitch')?.checked === true;
    }
  },
};
