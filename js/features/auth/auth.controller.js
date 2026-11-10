/**
 * @file Authentication controller coordinating user interaction, services, and view state.
 */

import { state } from '../../core/state.js';
import { eventBus } from '../../core/events.js';
import { showToast } from '../../utils/toast.js';
import { getDeviceToken, getDeviceName } from '../../utils/device.js';
import { authService } from './auth.service.js';
import { authView } from './auth.view.js';
import { initGoogleOAuth, requestGoogleSignIn, fetchGoogleUserProfile } from './google-auth.js';

export const authController = {
  activeSession: { email: '', purpose: '', name: '' },
  cooldownInterval: null,

  init() {
    this.render();
    authView.initPasswordToggles();
    this.bindEvents();

    const onToken = this.handleGoogleTokenReceived.bind(this);
    if (typeof google !== 'undefined' && google.accounts) {
      initGoogleOAuth(onToken);
    } else {
      window.addEventListener('load', () => initGoogleOAuth(onToken));
    }
  },

  render() {
    authView.renderUserBadge(state.user, this.handleLogout.bind(this), authView.openModal.bind(authView));
  },

  bindEvents() {
    const el = authView.elements;
    if (el.closeAuthModal) el.closeAuthModal.addEventListener('click', () => authView.closeModal());
    if (el.authModalBackdrop) el.authModalBackdrop.addEventListener('click', () => authView.closeModal());

    if (el.tabLogin) el.tabLogin.addEventListener('click', () => authView.switchToLoginTab());
    if (el.tabRegister) el.tabRegister.addEventListener('click', () => authView.switchToRegisterTab());

    if (el.linkToRegister) el.linkToRegister.addEventListener('click', (e) => { e.preventDefault(); authView.switchToRegisterTab(); });
    if (el.linkToLogin) el.linkToLogin.addEventListener('click', (e) => { e.preventDefault(); authView.switchToLoginTab(); });

    if (el.loginForm) el.loginForm.addEventListener('submit', this.handleLogin.bind(this));
    if (el.registerForm) el.registerForm.addEventListener('submit', this.handleRegister.bind(this));
    if (el.otpVerifyForm) el.otpVerifyForm.addEventListener('submit', this.handleVerifyOtp.bind(this));

    if (el.otpResendBtn) el.otpResendBtn.addEventListener('click', this.handleResendOtp.bind(this));
    if (el.otpCancelBtn) el.otpCancelBtn.addEventListener('click', this.handleCancelOtp.bind(this));

    if (el.googleAuthBtn) el.googleAuthBtn.addEventListener('click', () => this.handleGoogleClick('login'));
    if (el.googleRegBtn) el.googleRegBtn.addEventListener('click', () => this.handleGoogleClick('register'));
  },

  async handleLogin(e) {
    e.preventDefault();
    const el = authView.elements;
    const email = el.loginEmail ? el.loginEmail.value.trim() : '';
    const password = el.loginPassword ? el.loginPassword.value : '';

    if (!email || !password) {
      showToast('Email dan password wajib diisi.', 'warning');
      return;
    }

    authView.setLoginLoading(true);
    try {
      const res = await authService.login(email, password, getDeviceToken(), getDeviceName());
      if (res.require_otp) {
        this.openOtpStep(res.email || email, res.purpose || 'login_new_device', res.message);
        return;
      }
      this.completeAuth(res.token, res.user, `Selamat datang kembali, ${res.user.name}!`);
    } catch (err) {
      if (err.data && err.data.error === 'user_not_found') {
        showToast('Email belum terdaftar. Mengalihkan Anda ke pendaftaran akun...', 'warning');
        authView.switchToRegisterTab(email);
        return;
      }
      if (err.data && err.data.error === 'unverified_account') {
        showToast('Akun belum menyelesaikan verifikasi OTP.', 'warning');
        this.openOtpStep(email, 'registration', 'Akun Anda belum aktif. Masukkan kode OTP pendaftaran:');
        return;
      }
      showToast(err.message || 'Login gagal.', 'error');
    } finally {
      authView.setLoginLoading(false);
    }
  },

  async handleRegister(e) {
    e.preventDefault();
    const el = authView.elements;
    const name = el.regName ? el.regName.value.trim() : '';
    const email = el.regEmail ? el.regEmail.value.trim() : '';
    const password = el.regPassword ? el.regPassword.value : '';

    if (!name || !email || !password) {
      showToast('Seluruh formulir pendaftaran wajib diisi.', 'warning');
      return;
    }
    if (password.length < 6) {
      showToast('Password minimal 6 karakter.', 'warning');
      return;
    }

    authView.setRegisterLoading(true);
    try {
      const res = await authService.register(name, email, password, getDeviceToken(), getDeviceName());
      this.openOtpStep(email, res.purpose || 'registration', res.message, name);
      showToast('Kode verifikasi telah dikirimkan ke email Anda.', 'success');
    } catch (err) {
      showToast(err.message || 'Pendaftaran gagal.', 'error');
    } finally {
      authView.setRegisterLoading(false);
    }
  },

  async handleVerifyOtp(e) {
    e.preventDefault();
    const el = authView.elements;
    const code = el.otpCodeInput ? el.otpCodeInput.value.trim() : '';

    if (!code || code.length !== 6) {
      showToast('Masukkan 6 digit kode OTP lengkap.', 'warning');
      return;
    }

    authView.setVerifyOtpLoading(true);
    try {
      const res = await authService.verifyOTP(
        this.activeSession.email,
        code,
        this.activeSession.purpose,
        getDeviceToken(),
        getDeviceName()
      );
      this.completeAuth(res.token, res.user, res.message || `Selamat datang, ${res.user.name}!`);
    } catch (err) {
      showToast(err.message || 'Verifikasi OTP gagal.', 'error');
    } finally {
      authView.setVerifyOtpLoading(false);
    }
  },

  async handleResendOtp() {
    if (!this.activeSession.email) {
      showToast('Sesi verifikasi tidak ditemukan.', 'warning');
      return;
    }
    try {
      const res = await authService.resendOTP(this.activeSession.email, this.activeSession.purpose);
      showToast(res.message || 'Kode OTP baru telah dikirimkan.', 'success');
      this.startCooldown(60);
    } catch (err) {
      showToast(err.message || 'Gagal mengirim ulang OTP.', 'error');
    }
  },

  handleCancelOtp() {
    clearInterval(this.cooldownInterval);
    authView.hideOtpVerification();
  },

  openOtpStep(email, purpose, message = '', name = '') {
    this.activeSession = { email, purpose, name };
    authView.showOtpVerification({ email, purpose, message });
    this.startCooldown(60);
  },

  startCooldown(seconds) {
    clearInterval(this.cooldownInterval);
    let remaining = seconds;
    authView.setResendCooldown(remaining);
    this.cooldownInterval = setInterval(() => {
      remaining -= 1;
      authView.setResendCooldown(remaining);
      if (remaining <= 0) clearInterval(this.cooldownInterval);
    }, 1000);
  },

  completeAuth(token, user, message = '') {
    clearInterval(this.cooldownInterval);
    state.setAuth(token, user);
    this.render();
    authView.closeModal();
    if (message) showToast(message, 'success');
    eventBus.emit('auth:changed', state.user);
  },

  async handleGoogleClick(action = 'login') {
    this.pendingGoogleAction = action;
    const initiated = requestGoogleSignIn((token) => this.handleGoogleTokenReceived(token, action));
    if (initiated) return;

    const promptMsg = action === 'register' ? 'Daftar dengan Google:\nMasukkan email Google Anda:' : 'Masuk dengan Google:\nMasukkan email Google Anda:';
    const inputEmail = prompt(promptMsg, state.user ? state.user.email : 'user@gmail.com');
    if (!inputEmail || !inputEmail.trim()) return;

    await this.processGoogleProfile(inputEmail.trim(), inputEmail.trim().split('@')[0], action);
  },

  async handleGoogleTokenReceived(accessToken, action = 'login') {
    authView.setGoogleLoading(true, action === 'register' ? 'googleRegBtn' : 'googleAuthBtn');
    try {
      const profile = await fetchGoogleUserProfile(accessToken);
      await this.processGoogleProfile(profile.email, profile.name, action || this.pendingGoogleAction || 'login');
    } catch (err) {
      showToast(`Google login error: ${err.message}`, 'error');
    } finally {
      authView.setGoogleLoading(false, action === 'register' ? 'googleRegBtn' : 'googleAuthBtn');
    }
  },

  async processGoogleProfile(email, name, action) {
    try {
      const res = await authService.googleAuth(email, name, getDeviceToken(), getDeviceName(), action);
      if (res.require_otp) {
        this.openOtpStep(res.email || email, res.purpose || (action === 'register' ? 'registration' : 'login_new_device'), res.message, name);
        showToast(res.message, 'info');
        return;
      }
      this.completeAuth(res.token, res.user, `Berhasil masuk sebagai ${res.user.name}!`);
    } catch (err) {
      if (err.data && err.data.error === 'user_not_found') {
        showToast('Akun Google ini belum terdaftar. Mengalihkan Anda ke formulir pendaftaran...', 'info');
        authView.switchToRegisterTab(email, name);
        return;
      }
      showToast(err.message || 'Google Auth gagal.', 'error');
    }
  },

  handleLogout() {
    state.setAuth(null, null);
    this.render();
    showToast('Anda telah keluar dari akun.', 'info');
    eventBus.emit('auth:changed', null);
  },
};
