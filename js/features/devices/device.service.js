/**
 * @file Multi-device management API service.
 */

import { apiFetch } from '../../core/api.js';

export const deviceService = {
  /**
   * Fetches devices accessible to the current user.
   */
  async getDevices() {
    return apiFetch('/devices');
  },

  /**
   * Claims a new hardware device.
   * @param {string} deviceId
   * @param {string} pairingCode
   * @param {string} name
   */
  async claimDevice(deviceId, pairingCode, name) {
    return apiFetch('/devices/claim', {
      method: 'POST',
      body: JSON.stringify({
        device_id: deviceId,
        pairing_code: pairingCode,
        name: name,
      }),
    });
  },

  /**
   * Unpairs/releases an owned device.
   * @param {string} deviceId
   */
  async unpairDevice(deviceId) {
    return apiFetch('/devices/unpair', {
      method: 'POST',
      body: JSON.stringify({ device_id: deviceId }),
    });
  },
};

