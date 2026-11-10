/**
 * @file Centralized HTTP API client with Bearer authentication and mock support.
 */

import { CONFIG } from '../config/constants.js';
import { state } from './state.js';

const delay = (ms) => new Promise(r => setTimeout(r, ms));

const MOCK = {
  status: () => ({
    device_id: state.currentDeviceId,
    pump_on:   state.pumpOn,
    moisture:  state.moisture ?? 45,
    mode:      state.isManual ? 'manual' : 'auto',
    threshold: 50,
    timestamp: new Date().toISOString(),
  }),
  history: [],
};

async function mockFetch(endpoint, options) {
  await delay(150);
  if (endpoint.startsWith('/status'))  return MOCK.status();
  if (endpoint.startsWith('/history')) return MOCK.history;
  if (endpoint.startsWith('/devices')) return [{ device_id: 'DEV-PLB-001', name: 'Unit Utama', is_owner: true }];
  if (endpoint === '/ping')            return { status: 'ok' };
  if (endpoint === '/pump/on')         { state.pumpOn = true;  return { ok: true, message: 'Pompa dinyalakan' }; }
  if (endpoint === '/pump/off')        { state.pumpOn = false; return { ok: true, message: 'Pompa dimatikan' }; }
  if (endpoint === '/mode')            return { ok: true };
  throw new Error('Unknown endpoint');
}

/**
 * Performs an authenticated HTTP request to the API backend.
 * @param {string} endpoint - API route (e.g. '/status')
 * @param {RequestInit} [options] - Fetch options
 * @returns {Promise<any>}
 */
export async function apiFetch(endpoint, options = {}) {
  if (CONFIG.MOCK_MODE) return mockFetch(endpoint, options);

  const headers = {
    'Content-Type': 'application/json',
    'Accept-Encoding': 'gzip',
    ...(options.headers || {}),
  };

  if (state.token) {
    headers['Authorization'] = `Bearer ${state.token}`;
  }

  const res = await fetch(`${CONFIG.API_BASE}${endpoint}`, {
    signal: AbortSignal.timeout(8000),
    ...options,
    headers,
  });

  if (!res.ok) {
    const errorBody = await res.json().catch(() => ({}));
    const message = errorBody.message || errorBody.error || `HTTP ${res.status}`;
    const err = new Error(message);
    err.status = res.status;
    err.data = errorBody;
    throw err;
  }

  return res.json();
}

