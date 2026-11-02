/**
 * @file Centralized application state store with persistence helpers.
 */

let user = null;
try {
  user = JSON.parse(localStorage.getItem('plantcare_user') || 'null');
} catch {
  user = null;
}

export const state = {
  isManual: false,
  pumpOn: false,
  moisture: null,
  connected: false,
  history: [],
  historyFilter: 'all',
  lastFetch: 0,
  lastTimestamp: null,
  currentDeviceId: localStorage.getItem('plantcare_device') || 'DEV-PLB-001',
  devices: [],
  token: localStorage.getItem('plantcare_token') || null,
  user: user,

  /**
   * Updates authentication state and persists to local storage.
   * @param {string|null} token
   * @param {object|null} userObj
   */
  setAuth(token, userObj) {
    this.token = token;
    this.user = userObj;
    if (token) {
      localStorage.setItem('plantcare_token', token);
      localStorage.setItem('plantcare_user', JSON.stringify(userObj));
    } else {
      localStorage.removeItem('plantcare_token');
      localStorage.removeItem('plantcare_user');
    }
  },

  /**
   * Sets active IoT device ID and persists preference.
   * @param {string} deviceId
   */
  setCurrentDevice(deviceId) {
    this.currentDeviceId = deviceId;
    localStorage.setItem('plantcare_device', deviceId);
  },
};

