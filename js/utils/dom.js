/**
 * @file DOM helper utilities.
 */

/**
 * Shorthand for document.getElementById
 * @param {string} id
 * @returns {HTMLElement|null}
 */
export const $ = (id) => document.getElementById(id);

/**
 * Shorthand for document.querySelectorAll
 * @param {string} selector
 * @returns {NodeListOf<HTMLElement>}
 */
export const $$ = (selector) => document.querySelectorAll(selector);

/**
 * Configures an eye icon button to toggle password visibility.
 * @param {string} buttonId
 * @param {string} inputId
 */
export function setupPasswordToggle(buttonId, inputId) {
  const btn = $(buttonId);
  const input = $(inputId);
  if (!btn || !input) return;

  btn.addEventListener('click', (e) => {
    e.preventDefault();
    const isPass = input.type === 'password';
    input.type = isPass ? 'text' : 'password';
    const openEye = btn.querySelector('.eye-open');
    const closedEye = btn.querySelector('.eye-closed');
    if (openEye && closedEye) {
      openEye.style.display = isPass ? 'none' : 'block';
      closedEye.style.display = isPass ? 'block' : 'none';
    }
    btn.setAttribute('title', isPass ? 'Sembunyikan password' : 'Lihat password');
  });
}

