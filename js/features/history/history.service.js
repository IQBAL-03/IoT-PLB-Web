/**
 * @file Pump history and audit log API service.
 */

import { apiFetch } from '../../core/api.js';

export const historyService = {
  /**
   * Retrieves pump activity history for the specified device.
   * @param {string} deviceId
   */
  async getHistory(deviceId) {
    return apiFetch(`/history?device_id=${encodeURIComponent(deviceId)}`);
  },

  /**
   * Purges all historical logs for the specified device.
   * @param {string} deviceId
   */
  async clearHistory(deviceId) {
    return apiFetch(`/history/clear?device_id=${encodeURIComponent(deviceId)}`, {
      method: 'POST',
    });
  },
};

