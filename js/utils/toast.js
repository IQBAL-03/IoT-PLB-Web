/**
 * @file Toast notification utility service.
 */

import { $ } from './dom.js';

const badges = {
  success: '[OK]',
  error:   '[ERR]',
  info:    '[INFO]',
  warning: '[WARN]',
};

/**
 * Displays a toast notification in the UI.
 * @param {string} message
 * @param {'success'|'error'|'info'|'warning'} [type='info']
 */
export function showToast(message, type = 'info') {
  const container = $('toastContainer');
  if (!container) return;

  const toast = document.createElement('div');
  toast.className = `toast toast--${type}`;
  toast.setAttribute('role', 'alert');
  toast.innerHTML = `<strong>${badges[type] || '[INFO]'}</strong> <span>${message}</span>`;
  container.appendChild(toast);

  setTimeout(() => {
    toast.classList.add('toast--out');
    toast.addEventListener('animationend', () => toast.remove(), { once: true });
  }, 3500);
}
