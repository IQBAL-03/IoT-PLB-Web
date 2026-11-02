/**
 * @file Authentication API service.
 */

import { apiFetch } from '../../core/api.js';

export const authService = {
  /**
   * Logs in a user with email and password.
   * @param {string} email
   * @param {string} password
   */
  async login(email, password) {
    return apiFetch('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
  },

  /**
   * Registers a new account.
   * @param {string} name
   * @param {string} email
   * @param {string} password
   */
  async register(name, email, password) {
    return apiFetch('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ name, email, password }),
    });
  },

  /**
   * Authenticates using Google profile data.
   * @param {string} email
   * @param {string} name
   */
  async googleAuth(email, name) {
    return apiFetch('/auth/google', {
      method: 'POST',
      body: JSON.stringify({ email, name }),
    });
  },

  /**
   * Fetches active user session.
   */
  async getMe() {
    return apiFetch('/auth/me');
  },
};

