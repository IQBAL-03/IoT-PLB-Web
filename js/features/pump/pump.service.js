/**
 * @file Pump control and operational mode API service.
 */

import { apiFetch } from '../../core/api.js';

export const pumpService = {
  /**
   * Turns the pump ON or OFF.
   * @param {boolean} on
   * @param {string} deviceId
   */
  async setPump(on, deviceId) {
    return apiFetch(on ? '/pump/on' : '/pump/off', {
      method: 'POST',
      body: JSON.stringify({ device_id: deviceId }),
    });
  },

  /**
   * Sets the operational mode to automatic or manual.
   * @param {boolean} isAuto
   * @param {string} deviceId
   */
  async setMode(isAuto, deviceId) {
    return apiFetch('/mode', {
      method: 'POST',
      body: JSON.stringify({
        device_id: deviceId,
        mode: isAuto ? 'auto' : 'manual',
      }),
    });
  },
};

