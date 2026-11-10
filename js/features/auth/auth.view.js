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
<<<<<<< HEAD
      authModalTabs:      $('authModalTabs'),
      authModalTitle:     $('authModalTitle'),

      tabLogin:           $('tabLogin'),
      tabRegister:        $('tabRegister'),

=======
      tabLogin:           $('tabLogin'),
      tabRegister:        $('tabRegister'),
      authModalTitle:     $('authModalTitle'),
>>>>>>> 40a6cbec65d125425b1b30fe8365ff70bdec4ba2
      loginForm:          $('loginForm'),
      loginEmail:         $('loginEmail'),
      loginPassword:      $('loginPassword'),
      loginSubmitBtn:     $('loginSubmitBtn'),
      toggleLoginPassword:$('toggleLoginPassword'),
      linkToRegister:     $('linkToRegister'),

      registerForm:       $('registerForm'),
      regStep1:           $('regStep1'),
      regStep2:           $('regStep2'),
      regName:            $('regName'),
      regEmail:           $('regEmail'),
      regPassword:        $('regPassword'),
      regSubmitBtn:       $('regSubmitBtn'),
<<<<<<< HEAD
      toggleRegPassword:  $('toggleRegPassword'),
      linkToLogin:        $('linkToLogin'),

=======
      regOtpNotice:       $('regOtpNotice'),
      regOtpCode:         $('regOtpCode'),
      regVerifyBtn:       $('regVerifyBtn'),
      regBackBtn:         $('regBackBtn'),
>>>>>>> 40a6cbec65d125425b1b30fe8365ff70bdec4ba2
      googleAuthBtn:      $('googleAuthBtn'),
      googleBtnText:      $('googleBtnText'),
      googleRegBtn:       $('googleRegBtn'),
      googleRegBtnText:   $('googleRegBtnText'),

      otpSection:         $('otpSection'),
      otpBadge:           $('otpBadge'),
      otpHeading:         $('otpHeading'),
      otpNotice:          $('otpNotice'),
      otpTargetEmail:     $('otpTargetEmail'),
      otpVerifyForm:      $('otpVerifyForm'),
      otpCodeInput:       $('otpCodeInput'),
      otpVerifyBtn:       $('otpVerifyBtn'),
      otpResendBtn:       $('otpResendBtn'),
      otpCancelBtn:       $('otpCancelBtn'),
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

  openModal(mode = 'login') {
    const modal = $('authModal');
    if (!modal) return;
    modal.style.display = 'flex';
    modal.setAttribute('aria-hidden', 'false');

    this.hideOtpVerification();
    if (mode === 'register') {
      this.switchToRegisterTab();
    } else {
      this.switchToLoginTab();
    }
  },

  closeModal() {
    const modal = $('authModal');
    if (!modal) return;
    modal.style.display = 'none';
    modal.setAttribute('aria-hidden', 'true');
<<<<<<< HEAD
    this.hideOtpVerification();
=======
>>>>>>> 40a6cbec65d125425b1b30fe8365ff70bdec4ba2
  },

  switchToLoginTab() {
    const el = this.elements;
    this.hideOtpVerification();
    if (el.tabLogin) el.tabLogin.classList.add('modal__tab--active');
    if (el.tabRegister) el.tabRegister.classList.remove('modal__tab--active');
    if (el.loginForm) el.loginForm.style.display = 'flex';
    if (el.registerForm) el.registerForm.style.display = 'none';
    if (el.authModalTitle) el.authModalTitle.textContent = 'Masuk ke Akun';
  },

  switchToRegisterTab(prefillEmail = '', prefillName = '') {
    const el = this.elements;
    this.hideOtpVerification();
    if (el.tabRegister) el.tabRegister.classList.add('modal__tab--active');
    if (el.tabLogin) el.tabLogin.classList.remove('modal__tab--active');
    if (el.loginForm) el.loginForm.style.display = 'none';
    if (el.registerForm) el.registerForm.style.display = 'flex';
    if (el.authModalTitle) el.authModalTitle.textContent = 'Daftar Akun Baru';
<<<<<<< HEAD

    if (prefillEmail && el.regEmail) {
      el.regEmail.value = prefillEmail;
    }
    if (prefillName && el.regName) {
      el.regName.value = prefillName;
    }
    if (el.regPassword) {
      el.regPassword.focus();
    }
  },

  showOtpVerification({ email, purpose, message = '' }) {
    const el = this.elements;
    if (el.authModalTabs) el.authModalTabs.style.display = 'none';
    if (el.loginForm) el.loginForm.style.display = 'none';
    if (el.registerForm) el.registerForm.style.display = 'none';

    if (el.otpSection) el.otpSection.style.display = 'block';
    if (el.authModalTitle) el.authModalTitle.textContent = 'Verifikasi Kode OTP';

    if (purpose === 'login_new_device') {
      if (el.otpBadge) el.otpBadge.textContent = 'Perangkat Baru Terdeteksi';
      if (el.otpHeading) el.otpHeading.textContent = 'Verifikasi Keamanan Perangkat';
    } else {
      if (el.otpBadge) el.otpBadge.textContent = 'Aktivasi Pendaftaran';
      if (el.otpHeading) el.otpHeading.textContent = 'Verifikasi Pendaftaran Akun';
    }

    if (el.otpNotice) {
      el.otpNotice.textContent = message || 'Masukkan 6 digit kode OTP yang telah dikirimkan ke email Anda:';
    }
    if (el.otpTargetEmail) {
      el.otpTargetEmail.textContent = email;
    }
    if (el.otpCodeInput) {
      el.otpCodeInput.value = '';
      el.otpCodeInput.focus();
    }
  },

  hideOtpVerification() {
    const el = this.elements;
    if (el.otpSection) el.otpSection.style.display = 'none';
    if (el.authModalTabs) el.authModalTabs.style.display = 'flex';
  },

=======
    if (el.googleBtnText) el.googleBtnText.textContent = 'Daftar dengan Google';
    this.showRegisterStep(1);
  },

  showRegisterStep(step) {
    const el = this.elements;
    if (step === 2) {
      if (el.regStep1) el.regStep1.style.display = 'none';
      if (el.regStep2) el.regStep2.style.display = 'block';
      if (el.regOtpCode) {
        el.regOtpCode.value = '';
        el.regOtpCode.focus();
      }
    } else {
      if (el.regStep1) el.regStep1.style.display = 'block';
      if (el.regStep2) el.regStep2.style.display = 'none';
    }
  },

>>>>>>> 40a6cbec65d125425b1b30fe8365ff70bdec4ba2
  setLoginLoading(loading) {
    const btn = this.elements.loginSubmitBtn;
    if (!btn) return;
    btn.disabled = loading;
    btn.textContent = loading ? 'Memverifikasi...' : 'Masuk';
  },

  setRegisterRequestLoading(loading) {
    const btn = this.elements.regSubmitBtn;
    if (!btn) return;
    btn.disabled = loading;
<<<<<<< HEAD
    btn.textContent = loading ? 'Memproses Pendaftaran...' : 'Daftar Akun';
=======
    btn.textContent = loading ? 'Mengirim OTP...' : 'Daftar Akun (Kirim OTP)';
  },

  setRegisterVerifyLoading(loading) {
    const btn = this.elements.regVerifyBtn;
    if (!btn) return;
    btn.disabled = loading;
    btn.textContent = loading ? 'Membuat Akun...' : 'Verifikasi & Buat Akun';
>>>>>>> 40a6cbec65d125425b1b30fe8365ff70bdec4ba2
  },

  setGoogleLoading(loading, btnId = 'googleAuthBtn', originalText = 'Masuk dengan Google') {
    const btn = $(btnId);
    if (!btn) return;
    btn.disabled = loading;
    const textSpan = btn.querySelector('span');
    if (textSpan) {
      textSpan.textContent = loading ? 'Menghubungkan ke Google...' : originalText;
    }
  },

  setVerifyOtpLoading(loading) {
    const btn = this.elements.otpVerifyBtn;
    if (!btn) return;
    btn.disabled = loading;
    btn.textContent = loading ? 'Memeriksa OTP...' : 'Verifikasi & Lanjutkan';
  },

  setResendCooldown(seconds) {
    const btn = this.elements.otpResendBtn;
    if (!btn) return;

    if (seconds <= 0) {
      btn.disabled = false;
      btn.textContent = 'Kirim Ulang Kode';
      return;
    }

    btn.disabled = true;
    btn.textContent = `Kirim Ulang Kode (${seconds}s)`;
  },

  initPasswordToggles() {
    setupPasswordToggle('toggleLoginPassword', 'loginPassword');
    setupPasswordToggle('toggleRegPassword', 'regPassword');
  },
};
