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
   * Initiates account registration by sending OTP to email.
   * Backend handles rate limiting and OTP generation.
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
   * Backend determines whether to log in, register, or require OTP.
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
   * @param {string} purpose - "registration", "login_new_device", or "google_registration"
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
   * Requests a new OTP code if the cooldown period has elapsed.
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
   * Fetches the active authenticated user profile from token.
   */
  async getMe() {
    return apiFetch('/auth/me');
  },
};
