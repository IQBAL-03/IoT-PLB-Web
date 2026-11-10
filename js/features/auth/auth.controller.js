/**
 * @file Authentication controller coordinating user interaction, services, and view state.
 */

import { state } from '../../core/state.js';
import { eventBus } from '../../core/events.js';
import { showToast } from '../../utils/toast.js';
import { authService } from './auth.service.js';
import { authView } from './auth.view.js';
import { initGoogleOAuth, requestGoogleSignIn } from './google-auth.js';

export const authController = {
  /**
   * Initializes auth event listeners and UI state.
   */
  init() {
    this.render();
    authView.initPasswordToggles();
    this.bindEvents();

    if (typeof google !== 'undefined' && google.accounts) {
      initGoogleOAuth(this.handleGoogleTokenReceived.bind(this));
    } else {
      window.addEventListener('load', () => {
        initGoogleOAuth(this.handleGoogleTokenReceived.bind(this));
      });
    }
  },

  render() {
    authView.renderUserBadge(
      state.user,
      this.handleLogout.bind(this),
      authView.openModal.bind(authView)
    );
  },

  bindEvents() {
    const el = authView.elements;

    if (el.closeAuthModal) el.closeAuthModal.addEventListener('click', authView.closeModal);
    if (el.authModalBackdrop) el.authModalBackdrop.addEventListener('click', authView.closeModal);

    if (el.tabOtp) el.tabOtp.addEventListener('click', () => authView.switchToOtpTab());
    if (el.tabLogin) el.tabLogin.addEventListener('click', () => authView.switchToLoginTab());
    if (el.tabRegister) el.tabRegister.addEventListener('click', () => authView.switchToRegisterTab());

    if (el.requestOtpBtn) el.requestOtpBtn.addEventListener('click', this.handleRequestOtp.bind(this));
    if (el.otpForm) el.otpForm.addEventListener('submit', this.handleVerifyOtp.bind(this));
    if (el.otpBackBtn) el.otpBackBtn.addEventListener('click', () => authView.showOtpStep(1));

    if (el.loginForm) el.loginForm.addEventListener('submit', this.handleLogin.bind(this));
    if (el.registerForm) el.registerForm.addEventListener('submit', this.handleRegister.bind(this));
    if (el.googleAuthBtn) el.googleAuthBtn.addEventListener('click', this.handleGoogleClick.bind(this));
  },

  async handleRequestOtp(e) {
    if (e) e.preventDefault();
    const el = authView.elements;
    const email = el.otpEmail ? el.otpEmail.value.trim() : '';
    if (!email) {
      showToast('Masukkan alamat email Anda terlebih dahulu.', 'warning');
      return;
    }

    authView.setRequestOtpLoading(true);
    try {
      const res = await authService.sendOTP(email);
      showToast(res.message || 'Kode OTP telah dikirimkan ke email Anda!', 'success');
      authView.showOtpStep(2);
    } catch (err) {
      showToast(err.message, 'error');
    } finally {
      authView.setRequestOtpLoading(false);
    }
  },

  async handleVerifyOtp(e) {
    e.preventDefault();
    const el = authView.elements;
    const email = el.otpEmail ? el.otpEmail.value.trim() : '';
    const code = el.otpCode ? el.otpCode.value.trim() : '';
    const name = el.otpName ? el.otpName.value.trim() : '';

    if (!email || !code) {
      showToast('Email dan 6-digit kode OTP wajib diisi.', 'warning');
      return;
    }

    authView.setVerifyOtpLoading(true);
    try {
      const res = await authService.verifyOTP(email, code, name);
      state.setAuth(res.token, res.user);
      this.render();
      authView.closeModal();
      showToast(`Berhasil masuk! Selamat datang, ${res.user.name}!`, 'success');
      eventBus.emit('auth:changed', state.user);
    } catch (err) {
      showToast(err.message, 'error');
    } finally {
      authView.setVerifyOtpLoading(false);
    }
  },

  async handleLogin(e) {
    e.preventDefault();
    const el = authView.elements;
    authView.setLoginLoading(true);

    try {
      const res = await authService.login(el.loginEmail.value, el.loginPassword.value);
      state.setAuth(res.token, res.user);
      this.render();
      authView.closeModal();
      showToast(`Selamat datang kembali, ${res.user.name}!`, 'success');
      eventBus.emit('auth:changed', state.user);
    } catch (err) {
      showToast(err.message, 'error');
    } finally {
      authView.setLoginLoading(false);
    }
  },

  async handleRegister(e) {
    e.preventDefault();
    const el = authView.elements;
    authView.setRegisterLoading(true);

    try {
      const res = await authService.register(el.regName.value, el.regEmail.value, el.regPassword.value);
      state.setAuth(res.token, res.user);
      this.render();
      authView.closeModal();
      showToast(`Akun berhasil dibuat! Halo, ${res.user.name}!`, 'success');
      eventBus.emit('auth:changed', state.user);
    } catch (err) {
      showToast(err.message, 'error');
    } finally {
      authView.setRegisterLoading(false);
    }
  },

  handleLogout() {
    state.setAuth(null, null);
    this.render();
    showToast('Anda telah keluar.', 'info');
    eventBus.emit('auth:changed', null);
  },

  async handleGoogleTokenReceived(accessToken) {
    const el = authView.elements;
    const originalText = el.googleBtnText ? el.googleBtnText.textContent : 'Masuk dengan Google';
    authView.setGoogleLoading(true);

    try {
      const userInfoRes = await fetch('https://www.googleapis.com/oauth2/v3/userinfo', {
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      if (!userInfoRes.ok) throw new Error('Gagal mengambil data akun Google');
      const googleUser = await userInfoRes.json();

      const res = await authService.googleAuth(
        googleUser.email,
        googleUser.name || googleUser.given_name || googleUser.email.split('@')[0]
      );

      state.setAuth(res.token, res.user);
      this.render();
      authView.closeModal();
      showToast(`Berhasil masuk sebagai ${res.user.name}!`, 'success');
      eventBus.emit('auth:changed', state.user);
    } catch (err) {
      showToast(`Google login gagal: ${err.message}`, 'error');
      console.error(err);
    } finally {
      authView.setGoogleLoading(false, originalText);
    }
  },

  async handleGoogleClick() {
    const initiated = requestGoogleSignIn(this.handleGoogleTokenReceived.bind(this));
    if (initiated !== null) return;

    // Fallback prompt jika script Google diblokir
    const inputEmail = prompt(
      'Masuk dengan Akun Google:\nMasukkan email Google Anda (contoh: user@gmail.com):',
      state.user ? state.user.email : 'user@gmail.com'
    );
    if (!inputEmail || !inputEmail.trim()) return;

    const email = inputEmail.trim();
    const name = email.split('@')[0];
    const el = authView.elements;
    const originalText = el.googleBtnText ? el.googleBtnText.textContent : 'Masuk dengan Google';

    authView.setGoogleLoading(true);
    try {
      const res = await authService.googleAuth(email, name);
      state.setAuth(res.token, res.user);
      this.render();
      authView.closeModal();
      showToast(`Berhasil masuk dengan Google (${res.user.email})!`, 'success');
      eventBus.emit('auth:changed', state.user);
    } catch (err) {
      showToast(`Google login gagal: ${err.message}`, 'error');
    } finally {
      authView.setGoogleLoading(false, originalText);
    }
  },
};

