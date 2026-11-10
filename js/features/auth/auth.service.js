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
<<<<<<< HEAD
   * Initiates account registration.
=======
   * Requests an OTP code for new user registration.
>>>>>>> 40a6cbec65d125425b1b30fe8365ff70bdec4ba2
   * @param {string} name
   * @param {string} email
   * @param {string} password
   * @param {string} deviceToken
   * @param {string} deviceName
   */
<<<<<<< HEAD
  async register(name, email, password, deviceToken = '', deviceName = '') {
    return apiFetch('/auth/register', {
=======
  async requestRegisterOTP(name, email, password) {
    return apiFetch('/auth/register/request-otp', {
>>>>>>> 40a6cbec65d125425b1b30fe8365ff70bdec4ba2
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
<<<<<<< HEAD
   * Authenticates using Google OAuth profile data.
=======
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
>>>>>>> 40a6cbec65d125425b1b30fe8365ff70bdec4ba2
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
<<<<<<< HEAD
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
=======
   * Fetches active user session.
>>>>>>> 40a6cbec65d125425b1b30fe8365ff70bdec4ba2
   */
  async getMe() {
    return apiFetch('/auth/me');
  },
};
