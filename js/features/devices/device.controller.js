/**
 * @file Device controller orchestrating device selection and claims.
 */

import { state } from '../../core/state.js';
import { eventBus } from '../../core/events.js';
import { showToast } from '../../utils/toast.js';
import { deviceService } from './device.service.js';
import { deviceView } from './device.view.js';

export const deviceController = {
  async init() {
    this.bindEvents();
    await this.fetchDevices();

    eventBus.on('auth:changed', async () => {
      await this.fetchDevices();
    });
  },

  bindEvents() {
    const el = deviceView.elements;

    if (el.deviceSelector) {
      el.deviceSelector.addEventListener('change', this.handleDeviceChange.bind(this));
    }

    if (el.addDeviceBtn) {
      el.addDeviceBtn.addEventListener('click', () => {
        if (!state.token) {
          showToast('Silakan login terlebih dahulu untuk mendaftarkan unit.', 'info');
          eventBus.emit('auth:openModal');
          return;
        }
        deviceView.openClaimModal();
      });
    }

    if (el.closeClaimModal) el.closeClaimModal.addEventListener('click', deviceView.closeClaimModal);
    if (el.claimModalBackdrop) el.claimModalBackdrop.addEventListener('click', deviceView.closeClaimModal);

    if (el.unpairDeviceBtn) {
      el.unpairDeviceBtn.addEventListener('click', this.handleUnpairDevice.bind(this));
    }

    if (el.claimForm) {
      el.claimForm.addEventListener('submit', this.handleClaimDevice.bind(this));
    }
  },

  async fetchDevices() {
    try {
      const list = await deviceService.getDevices();
      state.devices = list || [];

      // Validate selected device ID
      const exists = state.devices.some(d => d.device_id === state.currentDeviceId);
      if (!exists && state.devices.length > 0) {
        state.setCurrentDevice(state.devices[0].device_id);
      }

      deviceView.renderSelector(state.devices, state.currentDeviceId);
    } catch (err) {
      console.warn('[PlantCare] Could not fetch devices:', err);
    }
  },

  async handleDeviceChange(e) {
    const newId = e.target.value;
    state.setCurrentDevice(newId);
    showToast(`Beralih ke unit: ${newId}`, 'info');
    deviceView.renderSelector(state.devices, state.currentDeviceId);
    eventBus.emit('device:changed', newId);
  },

  async handleUnpairDevice() {
    const currentId = state.currentDeviceId;
    if (!currentId) return;

    if (!confirm(`Apakah Anda yakin ingin melepas tautan unit ${currentId} dari akun Anda?`)) {
      return;
    }

    try {
      const res = await deviceService.unpairDevice(currentId);
      showToast(res.message || 'Unit berhasil dilepas dari akun Anda.', 'success');
      await this.fetchDevices();
      eventBus.emit('device:changed', state.currentDeviceId);
    } catch (err) {
      showToast(err.message, 'error');
    }
  },

  async handleClaimDevice(e) {
    e.preventDefault();
    const el = deviceView.elements;
    deviceView.setClaimLoading(true);

    try {
      const res = await deviceService.claimDevice(
        el.claimDeviceId.value,
        el.claimPairingCode.value,
        el.claimDeviceName.value
      );

      showToast(res.message || 'Unit berhasil dihubungkan!', 'success');
      deviceView.closeClaimModal();
      deviceView.resetClaimForm();

      state.setCurrentDevice(res.device_id);
      await this.fetchDevices();
      eventBus.emit('device:changed', res.device_id);
    } catch (err) {
      showToast(err.message, 'error');
    } finally {
      deviceView.setClaimLoading(false);
    }
  },
};

