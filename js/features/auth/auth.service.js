/**
 * @file Authentication API service.
 */

import { apiFetch } from '../../core/api.js';

export const authService = {
  /**
   * Logs in a user with email, password, and device identification.
   * @param {string} email
   * @param {string} password
   * @param {string} deviceToken
   * @param {string} deviceName
   */
  async login(email, password, deviceToken = '', deviceName = '') {
    return apiFetch('/auth/login', {
      method: 'POST',
      body: JSON.stringify({
        email,
        password,
        device_token: deviceToken,
        device_name: deviceName,
      }),
    });
  },

  /**
   * Initiates account registration.
   * @param {string} name
   * @param {string} email
   * @param {string} password
   * @param {string} deviceToken
   * @param {string} deviceName
   */
  async register(name, email, password, deviceToken = '', deviceName = '') {
    return apiFetch('/auth/register', {
      method: 'POST',
      body: JSON.stringify({
        name,
        email,
        password,
        device_token: deviceToken,
        device_name: deviceName,
      }),
    });
  },

  /**
   * Authenticates using Google OAuth profile data.
   * @param {string} email
   * @param {string} name
   * @param {string} deviceToken
   * @param {string} deviceName
   * @param {string} [action="login"]
   */
  async googleAuth(email, name, deviceToken = '', deviceName = '', action = 'login') {
    return apiFetch('/auth/google', {
      method: 'POST',
      body: JSON.stringify({
        email,
        name,
        device_token: deviceToken,
        device_name: deviceName,
        action,
      }),
    });
  },

  /**
   * Verifies an OTP code for registration or new device trust.
   * @param {string} email
   * @param {string} code
   * @param {string} purpose
   * @param {string} deviceToken
   * @param {string} deviceName
   */
  async verifyOTP(email, code, purpose, deviceToken = '', deviceName = '') {
    return apiFetch('/auth/otp/verify', {
      method: 'POST',
      body: JSON.stringify({
        email,
        code,
        purpose,
        device_token: deviceToken,
        device_name: deviceName,
      }),
    });
  },

  /**
   * Requests a new OTP code if expired or cooldown has elapsed.
   * @param {string} email
   * @param {string} purpose
   */
  async resendOTP(email, purpose) {
    return apiFetch('/auth/otp/resend', {
      method: 'POST',
      body: JSON.stringify({ email, purpose }),
    });
  },

  /**
   * Fetches the active authenticated user profile.
   */
  async getMe() {
    return apiFetch('/auth/me');
  },
};
