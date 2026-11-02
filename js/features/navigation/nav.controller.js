/**
 * @file Mobile navigation drawer and hamburger toggle controller.
 */

import { $ } from '../../utils/dom.js';

export const navController = {
  /**
   * Initializes mobile hamburger button listeners and outside dismissal.
   */
  init() {
    this.bindEvents();
  },

  bindEvents() {
    const btn = $('hamburgerBtn');
    const nav = $('headerNav');
    if (!btn || !nav) return;

    btn.addEventListener('click', (e) => {
      e.stopPropagation();
      const isOpen = nav.classList.toggle('is-open');
      btn.classList.toggle('is-active', isOpen);
      btn.setAttribute('aria-expanded', isOpen ? 'true' : 'false');
    });

    // Close mobile menu when clicking outside
    document.addEventListener('click', (e) => {
      if (!nav.contains(e.target) && !btn.contains(e.target)) {
        if (nav.classList.contains('is-open')) {
          nav.classList.remove('is-open');
          btn.classList.remove('is-active');
          btn.setAttribute('aria-expanded', 'false');
        }
      }
    });

    // Close mobile menu when pressing Escape key
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape' && nav.classList.contains('is-open')) {
        nav.classList.remove('is-open');
        btn.classList.remove('is-active');
        btn.setAttribute('aria-expanded', 'false');
        btn.focus();
      }
    });

    // Close drawer when interacting with modal-triggering buttons inside
    const closeOnAction = (elementId) => {
      const el = $(elementId);
      if (el) {
        el.addEventListener('click', () => {
          if (nav.classList.contains('is-open')) {
            nav.classList.remove('is-open');
            btn.classList.remove('is-active');
            btn.setAttribute('aria-expanded', 'false');
          }
        });
      }
    };

    closeOnAction('addDeviceBtn');
  },
};

