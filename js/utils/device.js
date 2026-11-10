/**
 * @file Device identification and browser fingerprint helper.
 */

const DEVICE_STORAGE_KEY = 'plantcare_device_token';

/**
 * Returns or generates a persistent unique device token stored in localStorage.
 * @returns {string}
 */
export function getDeviceToken() {
  let token = localStorage.getItem(DEVICE_STORAGE_KEY);
  if (!token) {
    if (typeof crypto !== 'undefined' && crypto.randomUUID) {
      token = crypto.randomUUID();
    } else {
      token = 'dev_' + Date.now().toString(36) + '_' + Math.random().toString(36).substring(2, 10);
    }
    localStorage.setItem(DEVICE_STORAGE_KEY, token);
  }
  return token;
}

/**
 * Resolves a human-readable name for the client platform and browser.
 * @returns {string}
 */
export function getDeviceName() {
  const ua = navigator.userAgent || '';
  let os = 'Perangkat';
  if (ua.includes('Win')) os = 'Windows';
  else if (ua.includes('Mac')) os = 'macOS';
  else if (ua.includes('Linux')) os = 'Linux';
  else if (ua.includes('Android')) os = 'Android';
  else if (ua.includes('iPhone') || ua.includes('iPad')) os = 'iOS';

  let browser = 'Browser';
  if (ua.includes('Edg/')) browser = 'Edge';
  else if (ua.includes('Chrome/')) browser = 'Chrome';
  else if (ua.includes('Safari/') && !ua.includes('Chrome')) browser = 'Safari';
  else if (ua.includes('Firefox/')) browser = 'Firefox';

  return `${os} (${browser})`;
}

