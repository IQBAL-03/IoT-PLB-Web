/**
 * @file Authentication modal and badge UI view manipulation.
 */

import { $, setupPasswordToggle } from '../../utils/dom.js';

export const authView = {
  get elements() {
    return {
      userAuthArea:       $('userAuthArea'),
      authModal:          $('authModal'),
      authModalBackdrop:  $('authModalBackdrop'),
      closeAuthModal:     $('closeAuthModal'),
      tabOtp:             $('tabOtp'),
      tabLogin:           $('tabLogin'),
      tabRegister:        $('tabRegister'),
      authModalTitle:     $('authModalTitle'),
      otpForm:            $('otpForm'),
      otpStep1:           $('otpStep1'),
      otpStep2:           $('otpStep2'),
      otpEmail:           $('otpEmail'),
      otpCode:            $('otpCode'),
      otpName:            $('otpName'),
      requestOtpBtn:      $('requestOtpBtn'),
      verifyOtpBtn:       $('verifyOtpBtn'),
      otpBackBtn:         $('otpBackBtn'),
      otpNotice:          $('otpNotice'),
      loginForm:          $('loginForm'),
      loginEmail:         $('loginEmail'),
      loginPassword:      $('loginPassword'),
      loginSubmitBtn:     $('loginSubmitBtn'),
      registerForm:       $('registerForm'),
      regName:            $('regName'),
      regEmail:           $('regEmail'),
      regPassword:        $('regPassword'),
      regSubmitBtn:       $('regSubmitBtn'),
      googleAuthBtn:      $('googleAuthBtn'),
      googleBtnText:      $('googleBtnText'),
    };
  },

  /**
   * Renders the user badge in the header or the "Masuk" button.
   * @param {object|null} user
   * @param {Function} onLogout
   * @param {Function} onOpenModal
   */
  renderUserBadge(user, onLogout, onOpenModal) {
    const area = $('userAuthArea');
    if (!area) return;

    if (user) {
      area.innerHTML = `
        <div class="user-badge">
          <span>${user.name}</span>
          <button class="user-badge__btn" id="logoutBtn" title="Keluar akun">Keluar</button>
        </div>
      `;
      const logoutBtn = $('logoutBtn');
      if (logoutBtn) logoutBtn.addEventListener('click', onLogout);
    } else {
      area.innerHTML = `
        <button class="btn btn--ghost btn--xs" id="authBtn">Masuk</button>
      `;
      const authBtn = $('authBtn');
      if (authBtn) authBtn.addEventListener('click', onOpenModal);
    }
  },

  openModal() {
    const modal = $('authModal');
    if (!modal) return;
    modal.style.display = 'flex';
    modal.setAttribute('aria-hidden', 'false');
  },

  closeModal() {
    const modal = $('authModal');
    if (!modal) return;
    modal.style.display = 'none';
    modal.setAttribute('aria-hidden', 'true');
  },

  switchToOtpTab() {
    const el = this.elements;
    if (el.tabOtp) el.tabOtp.classList.add('modal__tab--active');
    if (el.tabLogin) el.tabLogin.classList.remove('modal__tab--active');
    if (el.tabRegister) el.tabRegister.classList.remove('modal__tab--active');
    if (el.otpForm) el.otpForm.style.display = 'flex';
    if (el.loginForm) el.loginForm.style.display = 'none';
    if (el.registerForm) el.registerForm.style.display = 'none';
    if (el.authModalTitle) el.authModalTitle.textContent = 'Masuk dengan OTP';
    if (el.googleBtnText) el.googleBtnText.textContent = 'Masuk dengan Google';
  },

  switchToLoginTab() {
    const el = this.elements;
    if (el.tabLogin) el.tabLogin.classList.add('modal__tab--active');
    if (el.tabOtp) el.tabOtp.classList.remove('modal__tab--active');
    if (el.tabRegister) el.tabRegister.classList.remove('modal__tab--active');
    if (el.otpForm) el.otpForm.style.display = 'none';
    if (el.loginForm) el.loginForm.style.display = 'flex';
    if (el.registerForm) el.registerForm.style.display = 'none';
    if (el.authModalTitle) el.authModalTitle.textContent = 'Masuk ke Akun';
    if (el.googleBtnText) el.googleBtnText.textContent = 'Masuk dengan Google';
  },

  switchToRegisterTab() {
    const el = this.elements;
    if (el.tabRegister) el.tabRegister.classList.add('modal__tab--active');
    if (el.tabOtp) el.tabOtp.classList.remove('modal__tab--active');
    if (el.tabLogin) el.tabLogin.classList.remove('modal__tab--active');
    if (el.otpForm) el.otpForm.style.display = 'none';
    if (el.loginForm) el.loginForm.style.display = 'none';
    if (el.registerForm) el.registerForm.style.display = 'flex';
    if (el.authModalTitle) el.authModalTitle.textContent = 'Daftar Akun Baru';
    if (el.googleBtnText) el.googleBtnText.textContent = 'Daftar dengan Google';
  },

  showOtpStep(step) {
    const el = this.elements;
    if (step === 2) {
      if (el.otpStep1) el.otpStep1.style.display = 'none';
      if (el.otpStep2) el.otpStep2.style.display = 'block';
      if (el.otpCode) el.otpCode.focus();
    } else {
      if (el.otpStep1) el.otpStep1.style.display = 'block';
      if (el.otpStep2) el.otpStep2.style.display = 'none';
    }
  },

  setRequestOtpLoading(loading) {
    const btn = this.elements.requestOtpBtn;
    if (!btn) return;
    btn.disabled = loading;
    btn.textContent = loading ? 'Mengirim Kode...' : 'Kirim Kode OTP';
  },

  setVerifyOtpLoading(loading) {
    const btn = this.elements.verifyOtpBtn;
    if (!btn) return;
    btn.disabled = loading;
    btn.textContent = loading ? 'Memverifikasi...' : 'Verifikasi & Masuk';
  },

  setLoginLoading(loading) {
    const btn = this.elements.loginSubmitBtn;
    if (!btn) return;
    btn.disabled = loading;
    btn.textContent = loading ? 'Memverifikasi...' : 'Masuk';
  },

  setRegisterLoading(loading) {
    const btn = this.elements.regSubmitBtn;
    if (!btn) return;
    btn.disabled = loading;
    btn.textContent = loading ? 'Mendaftar...' : 'Daftar Akun';
  },

  setGoogleLoading(loading, originalText = 'Masuk dengan Google') {
    const el = this.elements;
    if (!el.googleAuthBtn) return;
    el.googleAuthBtn.disabled = loading;
    if (el.googleBtnText) {
      el.googleBtnText.textContent = loading ? 'Memverifikasi Google...' : originalText;
    }
  },

  initPasswordToggles() {
    setupPasswordToggle('toggleLoginPassword', 'loginPassword');
    setupPasswordToggle('toggleRegPassword', 'regPassword');
  },
};

