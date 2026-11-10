/**
 * @file Device selector dropdown and claim modal UI rendering.
 */

import { $ } from '../../utils/dom.js';

export const deviceView = {
  get elements() {
    return {
      deviceSelector:     $('deviceSelector'),
      addDeviceBtn:       $('addDeviceBtn'),
      unpairDeviceBtn:    $('unpairDeviceBtn'),
      claimModal:         $('claimModal'),
      claimModalBackdrop: $('claimModalBackdrop'),
      closeClaimModal:    $('closeClaimModal'),
      claimForm:          $('claimForm'),
      claimDeviceId:      $('claimDeviceId'),
      claimPairingCode:   $('claimPairingCode'),
      claimDeviceName:    $('claimDeviceName'),
      claimSubmitBtn:     $('claimSubmitBtn'),
    };
  },

  /**
   * Renders device options into the top selector.
   * @param {Array} devices
   * @param {string} currentDeviceId
   */
  renderSelector(devices, currentDeviceId) {
    const sel = $('deviceSelector');
    if (!sel) return;

    sel.innerHTML = devices.map(d => `
      <option value="${d.device_id}" ${d.device_id === currentDeviceId ? 'selected' : ''}>
        ${d.is_owner ? '[Milik Anda]' : (d.has_owner ? '[Terkunci]' : '[Tersedia]')} ${d.name} (${d.device_id})
      </option>
    `).join('');

    const unpairBtn = $('unpairDeviceBtn');
    if (unpairBtn) {
      const current = devices.find(d => d.device_id === currentDeviceId);
      unpairBtn.style.display = (current && current.is_owner) ? 'inline-flex' : 'none';
    }
  },

  openClaimModal() {
    const modal = $('claimModal');
    if (!modal) return;
    modal.style.display = 'flex';
    modal.setAttribute('aria-hidden', 'false');
  },

  closeClaimModal() {
    const modal = $('claimModal');
    if (!modal) return;
    modal.style.display = 'none';
    modal.setAttribute('aria-hidden', 'true');
  },

  setClaimLoading(loading) {
    const btn = this.elements.claimSubmitBtn;
    if (!btn) return;
    btn.disabled = loading;
    btn.textContent = loading ? 'Menghubungkan...' : 'Hubungkan ke Akun';
  },

  resetClaimForm() {
    const form = this.elements.claimForm;
    if (form) form.reset();
  },
};

