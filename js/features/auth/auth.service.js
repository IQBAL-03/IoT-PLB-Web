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
   * Requests an OTP code for new user registration.
   * @param {string} name
   * @param {string} email
   * @param {string} password
   */
  async requestRegisterOTP(name, email, password) {
    return apiFetch('/auth/register/request-otp', {
      method: 'POST',
      body: JSON.stringify({ name, email, password }),
    });
  },

  /**
   * Verifies registration OTP code and completes account creation.
   * @param {string} name
   * @param {string} email
   * @param {string} password
   * @param {string} code
   */
  async verifyRegisterOTP(name, email, password, code) {
    return apiFetch('/auth/register/verify-otp', {
      method: 'POST',
      body: JSON.stringify({ name, email, password, code }),
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

