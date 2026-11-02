/**
 * @file IoT hardware webhook simulation service.
 */

import { apiFetch } from '../../core/api.js';

export const simulatorService = {
  /**
   * Posts simulated telemetry directly to the webhook endpoint.
   * @param {object} payload
   */
  async sendTelemetry(payload) {
    return apiFetch('/webhook', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },
};

