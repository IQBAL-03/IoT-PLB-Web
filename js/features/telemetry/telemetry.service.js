/**
 * @file Telemetry and system health polling API service.
 */

import { apiFetch } from '../../core/api.js';

export const telemetryService = {
  /**
   * Reads real-time status of the specified device.
   * @param {string} deviceId
   */
  async getStatus(deviceId) {
    return apiFetch(`/status?device_id=${encodeURIComponent(deviceId)}`);
  },

  /**
   * Health-check ping to keep the serverless container and database warm.
   */
  async ping() {
    return apiFetch('/ping');
  },
};

