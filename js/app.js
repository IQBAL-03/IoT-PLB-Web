const CONFIG = {
  API_BASE: '/api',
  POLL_INTERVAL: 4000,
  MOCK_MODE: false,
  KEEPALIVE_INTERVAL: 4 * 60 * 1000,
};

const state = {
  isManual: false,
  pumpOn: false,
  moisture: null,
  connected: false,
  history: [],
  historyFilter: 'all',
  lastFetch: 0,
  lastTimestamp: null,
  currentDeviceId: localStorage.getItem('plantcare_device') || 'DEV-PLB-001',
  devices: [],
  token: localStorage.getItem('plantcare_token') || null,
  user: null,
};

try {
  state.user = JSON.parse(localStorage.getItem('plantcare_user') || 'null');
} catch {
  state.user = null;
}

const $ = (id) => document.getElementById(id);

const dom = {
  connectionBadge:    $('connectionBadge'),
  connectionText:     $('connectionText'),
  currentTime:        $('currentTime'),
  refreshBtn:         $('refreshBtn'),
  refreshIcon:        $('refreshIcon'),
  pumpCard:           $('pumpCard'),
  pumpIndicator:      $('pumpIndicator'),
  pumpDot:            $('pumpDot'),
  pumpState:          $('pumpState'),
  pumpDetail:         $('pumpDetail'),
  gaugeArc:           $('gaugeArc'),
  moistureValue:      $('moistureValue'),
  moistureStatus:     $('moistureStatus'),
  moistureBar:        $('moistureBar'),
  modeSwitch:         $('modeSwitch'),
  modeCurrent:        $('modeCurrent'),
  modeDesc:           $('modeDesc'),
  modeManualLabel:    $('modeManualLabel'),
  modeAutoLabel:      $('modeAutoLabel'),
  manualBadgeStatus:  $('manualBadgeStatus'),
  controlPanel:       $('controlPanel'),
  controlInfo:        $('controlInfo'),
  pumpOnBtn:          $('pumpOnBtn'),
  pumpOffBtn:         $('pumpOffBtn'),
  historySkeleton:    $('historySkeleton'),
  historyTableWrap:   $('historyTableWrap'),
  historyBody:        $('historyBody'),
  historyEmpty:       $('historyEmpty'),
  historyFilter:      $('historyFilter'),
  clearHistoryBtn:    $('clearHistoryBtn'),
  testForm:           $('testForm'),
  testMoisture:       $('testMoisture'),
  testPump:           $('testPump'),
  testMode:           $('testMode'),
  testSendBtn:        $('testSendBtn'),
  toastContainer:     $('toastContainer'),

  // Multi-Device & Auth elements
  deviceSelector:     $('deviceSelector'),
  addDeviceBtn:       $('addDeviceBtn'),
  userAuthArea:       $('userAuthArea'),

  // Auth Modal
  authModal:          $('authModal'),
  authModalBackdrop:  $('authModalBackdrop'),
  closeAuthModal:     $('closeAuthModal'),
  tabLogin:           $('tabLogin'),
  tabRegister:        $('tabRegister'),
  authModalTitle:     $('authModalTitle'),
  loginForm:          $('loginForm'),
  loginEmail:         $('loginEmail'),
  loginPassword:      $('loginPassword'),
  loginSubmitBtn:     $('loginSubmitBtn'),
  toggleLoginPassword:$('toggleLoginPassword'),
  registerForm:       $('registerForm'),
  regName:            $('regName'),
  regEmail:           $('regEmail'),
  regPassword:        $('regPassword'),
  regSubmitBtn:       $('regSubmitBtn'),
  toggleRegPassword:  $('toggleRegPassword'),
  googleAuthBtn:      $('googleAuthBtn'),
  googleBtnText:      $('googleBtnText'),

  // Claim Modal
  claimModal:         $('claimModal'),
  claimModalBackdrop: $('claimModalBackdrop'),
  closeClaimModal:    $('closeClaimModal'),
  claimForm:          $('claimForm'),
  claimDeviceId:      $('claimDeviceId'),
  claimPairingCode:   $('claimPairingCode'),
  claimDeviceName:    $('claimDeviceName'),
  claimSubmitBtn:     $('claimSubmitBtn'),
};

const MOCK = {
  status: () => ({
    device_id: state.currentDeviceId,
    pump_on:   state.pumpOn,
    moisture:  state.moisture ?? 45,
    mode:      state.isManual ? 'manual' : 'auto',
    threshold: 50,
    timestamp: new Date().toISOString(),
  }),
  history: [],
};

const controller = new AbortController();

async function apiFetch(endpoint, options = {}) {
  if (CONFIG.MOCK_MODE) return mockFetch(endpoint, options);

  const headers = {
    'Content-Type': 'application/json',
    'Accept-Encoding': 'gzip',
    ...(options.headers || {}),
  };
  if (state.token) {
    headers['Authorization'] = `Bearer ${state.token}`;
  }

  const res = await fetch(`${CONFIG.API_BASE}${endpoint}`, {
    signal: AbortSignal.timeout(8000),
    ...options,
    headers,
  });
  if (!res.ok) {
    const errorBody = await res.json().catch(() => ({}));
    throw new Error(errorBody.error || errorBody.detail || `HTTP ${res.status}`);
  }
  return res.json();
}

async function mockFetch(endpoint, options) {
  await delay(150);
  if (endpoint.startsWith('/status'))  return MOCK.status();
  if (endpoint.startsWith('/history')) return MOCK.history;
  if (endpoint.startsWith('/devices')) return [{ device_id: 'DEV-PLB-001', name: 'Unit Utama', is_owner: true }];
  if (endpoint === '/ping')            return { status: 'ok' };
  if (endpoint === '/pump/on')         { state.pumpOn = true;  return { ok: true, message: 'Pompa dinyalakan' }; }
  if (endpoint === '/pump/off')        { state.pumpOn = false; return { ok: true, message: 'Pompa dimatikan' }; }
  if (endpoint === '/mode')            return { ok: true };
  throw new Error('Unknown endpoint');
}

const delay = (ms) => new Promise(r => setTimeout(r, ms));

function updateClock() {
  dom.currentTime.textContent = new Date().toLocaleTimeString('id-ID', { hour12: false });
}

function setConnected(ok) {
  state.connected = ok;
  dom.connectionBadge.className = `connection-badge ${ok ? 'connected' : 'error'}`;
  dom.connectionText.textContent = ok ? 'Terhubung' : 'Terputus';
}

function renderPump(on) {
  state.pumpOn = on;
  dom.pumpCard.classList.remove('pump-on', 'pump-off');
  dom.pumpCard.classList.add(on ? 'pump-on' : 'pump-off');
  dom.pumpIndicator.classList.remove('pump-on', 'pump-off');
  dom.pumpIndicator.classList.add(on ? 'pump-on' : 'pump-off');
  dom.pumpState.textContent  = on ? 'AKTIF' : 'MATI';
  dom.pumpDetail.textContent = on ? 'Pompa sedang menyiram tanaman' : 'Pompa dalam keadaan standby';
}

function renderMoisture(pct) {
  if (pct === null || pct === undefined) return;
  state.moisture = pct;
  const circumference = 314;
  dom.gaugeArc.style.strokeDashoffset = circumference - (pct / 100) * circumference;
  dom.moistureValue.textContent = pct;
  dom.moistureBar.style.width   = `${pct}%`;

  let statusText, barColor, strokeColor;
  if (pct < 30) {
    statusText  = '🔴 Sangat Kering — Perlu Disiram!';
    barColor    = 'linear-gradient(90deg, #e63946, #f4a261)';
    strokeColor = '#e63946';
  } else if (pct < 50) {
    statusText  = '🟡 Cukup Kering — Pantau Terus';
    barColor    = 'linear-gradient(90deg, #f4a261, #ffd166)';
    strokeColor = '#f4a261';
  } else if (pct < 75) {
    statusText  = '🟢 Lembap — Kondisi Baik';
    barColor    = 'linear-gradient(90deg, var(--green-600), var(--green-400))';
    strokeColor = '#52B788';
  } else {
    statusText  = '💧 Sangat Lembap — Tidak Perlu Disiram';
    barColor    = 'linear-gradient(90deg, #0096c7, #52B788)';
    strokeColor = '#0096c7';
  }

  dom.moistureStatus.textContent   = statusText;
  dom.moistureBar.style.background  = barColor;
  dom.gaugeArc.style.stroke         = strokeColor;
}

function renderMode(isManual) {
  state.isManual = isManual;
  dom.modeManualLabel.classList.toggle('mode-label--active', isManual);
  dom.modeAutoLabel.classList.toggle('mode-label--active', !isManual);
  dom.modeCurrent.textContent = isManual ? 'Mode: Manual' : 'Mode: Otomatis';
  dom.modeDesc.textContent    = isManual
    ? 'Pompa dikendalikan secara manual melalui tombol di bawah.'
    : 'Pompa bekerja otomatis berdasarkan pembacaan sensor kelembapan.';
  dom.pumpOnBtn.disabled  = !isManual;
  dom.pumpOffBtn.disabled = !isManual;
  dom.controlPanel.classList.toggle('is-active', isManual);
  dom.controlInfo.textContent = isManual
    ? 'Mode manual aktif. Gunakan tombol di bawah untuk mengontrol pompa.'
    : 'Aktifkan mode manual untuk mengontrol pompa secara langsung.';
  dom.manualBadgeStatus.textContent = isManual ? 'Mode Manual Aktif' : 'Mode Otomatis Aktif';
  dom.manualBadgeStatus.className   = `section__badge${isManual ? ' manual' : ''}`;
}

function formatTimestamp(iso) {
  try {
    return new Date(iso).toLocaleString('id-ID', {
      day: '2-digit', month: 'short', year: 'numeric',
      hour: '2-digit', minute: '2-digit', hour12: false,
    });
  } catch {
    return iso;
  }
}

function renderHistory(data) {
  state.history = data;
  dom.historySkeleton.style.display = 'none';
  const filtered = state.historyFilter === 'all'
    ? data
    : data.filter(r => r.mode === state.historyFilter);

  if (!filtered.length) {
    dom.historyTableWrap.style.display = 'none';
    dom.historyEmpty.style.display     = 'flex';
    return;
  }

  dom.historyEmpty.style.display     = 'none';
  dom.historyTableWrap.style.display = 'block';
  dom.historyBody.innerHTML = filtered.map(r => `
    <tr>
      <td>${formatTimestamp(r.timestamp)}</td>
      <td><span class="badge badge--${r.action.toLowerCase()}">${r.action}</span></td>
      <td><span class="badge badge--${r.mode}">${r.mode === 'auto' ? 'Otomatis' : 'Manual'}</span></td>
      <td>${r.moisture}%</td>
      <td>${r.duration || '—'}</td>
    </tr>
  `).join('');
}

// ─── Multi-Device Management ──────────────────────────────────────────────────

async function fetchDevices() {
  try {
    const list = await apiFetch('/devices');
    state.devices = list || [];
    renderDeviceSelector();
  } catch (err) {
    console.warn('[PlantCare] Could not fetch devices:', err);
  }
}

function renderDeviceSelector() {
  if (!dom.deviceSelector) return;
  dom.deviceSelector.innerHTML = state.devices.map(d => `
    <option value="${d.device_id}" ${d.device_id === state.currentDeviceId ? 'selected' : ''}>
      ${d.is_owner ? '🌱' : '🌿'} ${d.name} (${d.device_id})
    </option>
  `).join('');

  // Validate active device ID
  const exists = state.devices.some(d => d.device_id === state.currentDeviceId);
  if (!exists && state.devices.length > 0) {
    state.currentDeviceId = state.devices[0].device_id;
    localStorage.setItem('plantcare_device', state.currentDeviceId);
    dom.deviceSelector.value = state.currentDeviceId;
  }
}

async function handleDeviceChange(e) {
  state.currentDeviceId = e.target.value;
  localStorage.setItem('plantcare_device', state.currentDeviceId);
  showToast(`Beralih ke unit: ${state.currentDeviceId}`, 'info');
  await Promise.all([fetchStatus(), fetchHistory()]);
}

// ─── User Authentication UI & Actions ─────────────────────────────────────────

function renderAuthUI() {
  if (!dom.userAuthArea) return;
  if (state.user && state.token) {
    dom.userAuthArea.innerHTML = `
      <div class="user-badge">
        <span>👤 ${state.user.name}</span>
        <button class="user-badge__btn" id="logoutBtn" title="Keluar akun">Keluar</button>
      </div>
    `;
    const logoutBtn = $('logoutBtn');
    if (logoutBtn) logoutBtn.addEventListener('click', handleLogout);
  } else {
    dom.userAuthArea.innerHTML = `
      <button class="btn btn--ghost btn--xs" id="authBtn">👤 Masuk</button>
    `;
    const authBtn = $('authBtn');
    if (authBtn) authBtn.addEventListener('click', openAuthModal);
  }
}

function openAuthModal() {
  dom.authModal.style.display = 'flex';
  dom.authModal.setAttribute('aria-hidden', 'false');
}

function closeAuthModal() {
  dom.authModal.style.display = 'none';
  dom.authModal.setAttribute('aria-hidden', 'true');
}

function openClaimModal() {
  if (!state.token) {
    showToast('Silakan login terlebih dahulu untuk mendaftarkan unit.', 'info');
    openAuthModal();
    return;
  }
  dom.claimModal.style.display = 'flex';
  dom.claimModal.setAttribute('aria-hidden', 'false');
}

function closeClaimModal() {
  dom.claimModal.style.display = 'none';
  dom.claimModal.setAttribute('aria-hidden', 'true');
}

async function handleLogin(e) {
  e.preventDefault();
  dom.loginSubmitBtn.disabled = true;
  dom.loginSubmitBtn.textContent = 'Memverifikasi...';

  try {
    const res = await apiFetch('/auth/login', {
      method: 'POST',
      body: JSON.stringify({
        email: dom.loginEmail.value,
        password: dom.loginPassword.value,
      }),
    });

    state.token = res.token;
    state.user = res.user;
    localStorage.setItem('plantcare_token', res.token);
    localStorage.setItem('plantcare_user', JSON.stringify(res.user));

    renderAuthUI();
    closeAuthModal();
    showToast(`Selamat datang kembali, ${res.user.name}!`, 'success');
    await fetchDevices();
    await Promise.all([fetchStatus(), fetchHistory()]);
  } catch (err) {
    showToast(err.message, 'error');
  } finally {
    dom.loginSubmitBtn.disabled = false;
    dom.loginSubmitBtn.textContent = 'Masuk';
  }
}

async function handleRegister(e) {
  e.preventDefault();
  dom.regSubmitBtn.disabled = true;
  dom.regSubmitBtn.textContent = 'Mendaftar...';

  try {
    const res = await apiFetch('/auth/register', {
      method: 'POST',
      body: JSON.stringify({
        name: dom.regName.value,
        email: dom.regEmail.value,
        password: dom.regPassword.value,
      }),
    });

    state.token = res.token;
    state.user = res.user;
    localStorage.setItem('plantcare_token', res.token);
    localStorage.setItem('plantcare_user', JSON.stringify(res.user));

    renderAuthUI();
    closeAuthModal();
    showToast(`Akun berhasil dibuat! Halo, ${res.user.name}!`, 'success');
    await fetchDevices();
  } catch (err) {
    showToast(err.message, 'error');
  } finally {
    dom.regSubmitBtn.disabled = false;
    dom.regSubmitBtn.textContent = 'Daftar Akun';
  }
}

function setupPasswordToggle(buttonId, inputId) {
  const btn = $(buttonId);
  const input = $(inputId);
  if (!btn || !input) return;

  btn.addEventListener('click', (e) => {
    e.preventDefault();
    const isPass = input.type === 'password';
    input.type = isPass ? 'text' : 'password';
    const openEye = btn.querySelector('.eye-open');
    const closedEye = btn.querySelector('.eye-closed');
    if (openEye && closedEye) {
      openEye.style.display = isPass ? 'none' : 'block';
      closedEye.style.display = isPass ? 'block' : 'none';
    }
    btn.setAttribute('title', isPass ? 'Sembunyikan password' : 'Lihat password');
  });
}

const GOOGLE_CLIENT_ID = '97161960012-snb6l6t80ddbfkjfcdp7nmjsc11kcg0m.apps.googleusercontent.com';
let googleTokenClient = null;

function initGoogleSignIn() {
  if (typeof google === 'undefined' || !google.accounts || !google.accounts.oauth2) return;

  try {
    googleTokenClient = google.accounts.oauth2.initTokenClient({
      client_id: GOOGLE_CLIENT_ID,
      scope: 'email profile openid',
      callback: async (tokenResponse) => {
        if (tokenResponse && tokenResponse.error) {
          showToast(`Google login: ${tokenResponse.error_description || tokenResponse.error}`, 'error');
          return;
        }
        if (tokenResponse && tokenResponse.access_token) {
          await processGoogleToken(tokenResponse.access_token);
        }
      },
      error_callback: (nonOAuthError) => {
        console.warn('[PlantCare] Google OAuth non-OAuth error:', nonOAuthError);
        showToast(`Google login error: ${nonOAuthError.type || 'Inisialisasi popup gagal'}`, 'error');
      },
    });
  } catch (err) {
    console.warn('[PlantCare] Google OAuth init error:', err);
  }
}

async function processGoogleToken(accessToken) {
  dom.googleAuthBtn.disabled = true;
  const originalText = dom.googleBtnText.textContent;
  dom.googleBtnText.textContent = 'Memverifikasi Google...';

  try {
    const userInfoRes = await fetch('https://www.googleapis.com/oauth2/v3/userinfo', {
      headers: { Authorization: `Bearer ${accessToken}` },
    });
    if (!userInfoRes.ok) throw new Error('Gagal mengambil data akun Google');
    const googleUser = await userInfoRes.json();

    const res = await apiFetch('/auth/google', {
      method: 'POST',
      body: JSON.stringify({
        email: googleUser.email,
        name: googleUser.name || googleUser.given_name || googleUser.email.split('@')[0],
      }),
    });

    state.token = res.token;
    state.user = res.user;
    localStorage.setItem('plantcare_token', res.token);
    localStorage.setItem('plantcare_user', JSON.stringify(res.user));

    renderAuthUI();
    closeAuthModal();
    showToast(`Berhasil masuk sebagai ${res.user.name}!`, 'success');
    await fetchDevices();
    await Promise.all([fetchStatus(), fetchHistory()]);
  } catch (err) {
    showToast(`Google login gagal: ${err.message}`, 'error');
    console.error(err);
  } finally {
    dom.googleAuthBtn.disabled = false;
    dom.googleBtnText.textContent = originalText;
  }
}

async function handleGoogleAuth() {
  if (typeof google !== 'undefined' && google.accounts && google.accounts.oauth2) {
    if (!googleTokenClient) {
      initGoogleSignIn();
    }
    if (googleTokenClient) {
      googleTokenClient.requestAccessToken({ prompt: 'select_account' });
      return;
    }
  }

  // Fallback prompt jika script Google diblokir adblocker
  const inputEmail = prompt(
    'Masuk dengan Akun Google:\nMasukkan email Google Anda (contoh: user@gmail.com):',
    state.user ? state.user.email : 'user@gmail.com'
  );
  if (!inputEmail || !inputEmail.trim()) return;

  const email = inputEmail.trim();
  const name = email.split('@')[0];

  dom.googleAuthBtn.disabled = true;
  const originalText = dom.googleBtnText.textContent;
  dom.googleBtnText.textContent = 'Menghubungkan ke Google...';

  try {
    const res = await apiFetch('/auth/google', {
      method: 'POST',
      body: JSON.stringify({ email, name }),
    });

    state.token = res.token;
    state.user = res.user;
    localStorage.setItem('plantcare_token', res.token);
    localStorage.setItem('plantcare_user', JSON.stringify(res.user));

    renderAuthUI();
    closeAuthModal();
    showToast(`Berhasil masuk dengan Google (${res.user.email})!`, 'success');
    await fetchDevices();
    await Promise.all([fetchStatus(), fetchHistory()]);
  } catch (err) {
    showToast(`Google login gagal: ${err.message}`, 'error');
  } finally {
    dom.googleAuthBtn.disabled = false;
    dom.googleBtnText.textContent = originalText;
  }
}

function handleLogout() {
  state.token = null;
  state.user = null;
  localStorage.removeItem('plantcare_token');
  localStorage.removeItem('plantcare_user');
  renderAuthUI();
  showToast('Anda telah keluar.', 'info');
  fetchDevices();
}

async function handleClaimDevice(e) {
  e.preventDefault();
  dom.claimSubmitBtn.disabled = true;
  dom.claimSubmitBtn.textContent = 'Menghubungkan...';

  try {
    const res = await apiFetch('/devices/claim', {
      method: 'POST',
      body: JSON.stringify({
        device_id: dom.claimDeviceId.value,
        pairing_code: dom.claimPairingCode.value,
        name: dom.claimDeviceName.value,
      }),
    });

    showToast(res.message || 'Unit berhasil dihubungkan!', 'success');
    closeClaimModal();
    dom.claimForm.reset();

    state.currentDeviceId = res.device_id;
    localStorage.setItem('plantcare_device', res.device_id);

    await fetchDevices();
    await Promise.all([fetchStatus(), fetchHistory()]);
  } catch (err) {
    showToast(err.message, 'error');
  } finally {
    dom.claimSubmitBtn.disabled = false;
    dom.claimSubmitBtn.textContent = 'Hubungkan ke Akun';
  }
}

// ─── Status, Control & Polling ────────────────────────────────────────────────

async function fetchStatus() {
  const now = Date.now();
  if (document.hidden && now - state.lastFetch < 10000) return;
  state.lastFetch = now;

  try {
    const data = await apiFetch(`/status?device_id=${encodeURIComponent(state.currentDeviceId)}`);
    setConnected(true);
    renderPump(data.pump_on);
    renderMoisture(data.moisture);

    // Sync mode if user is not actively toggling
    if (data.mode && !dom.modeSwitch._userChanged) {
      const isManual = data.mode === 'manual';
      if (dom.modeSwitch.checked !== !isManual) {
        dom.modeSwitch.checked = !isManual;
        renderMode(isManual);
      }
    }

    // Cross-device sync: trigger history fetch if status timestamp changed
    if (data.timestamp && data.timestamp !== state.lastTimestamp) {
      state.lastTimestamp = data.timestamp;
      fetchHistory();
    }
  } catch (err) {
    setConnected(false);
    console.error('[PlantCare] Status fetch error:', err);
  }
}

async function fetchHistory() {
  try {
    const data = await apiFetch(`/history?device_id=${encodeURIComponent(state.currentDeviceId)}`);
    renderHistory(data);
  } catch (err) {
    dom.historySkeleton.style.display = 'none';
    dom.historyEmpty.style.display    = 'flex';
    console.error('[PlantCare] History fetch error:', err);
  }
}

async function keepAlive() {
  if (CONFIG.MOCK_MODE) return;
  try { await apiFetch('/ping'); } catch { }
}

async function setPump(on) {
  const btn = on ? dom.pumpOnBtn : dom.pumpOffBtn;
  const originalHTML = btn.innerHTML;
  btn.disabled = true;
  btn.innerHTML = `
    <svg class="btn__icon spinning" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
      <polyline points="23 4 23 10 17 10"></polyline>
      <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10"></path>
    </svg>
    ${on ? 'Menyalakan...' : 'Mematikan...'}
  `;

  try {
    const res = await apiFetch(on ? '/pump/on' : '/pump/off', {
      method: 'POST',
      body: JSON.stringify({ device_id: state.currentDeviceId }),
    });
    renderPump(on);
    showToast(res.message || (on ? 'Pompa dinyalakan' : 'Pompa dimatikan'), 'success');
    await Promise.all([fetchStatus(), fetchHistory()]);
  } catch (err) {
    showToast(`Gagal: ${err.message}`, 'error');
    console.error('[PlantCare] Pump control error:', err);
  } finally {
    btn.innerHTML = originalHTML;
    btn.disabled  = !state.isManual;
  }
}

async function setMode(isAuto) {
  renderMode(!isAuto);
  try {
    await apiFetch('/mode', {
      method: 'POST',
      body:   JSON.stringify({ device_id: state.currentDeviceId, mode: !isAuto ? 'manual' : 'auto' }),
    });
    showToast(`Mode diubah ke: ${!isAuto ? 'Manual' : 'Otomatis'}`, 'info');
    await fetchStatus();
  } catch (err) {
    showToast(`Gagal: ${err.message}`, 'error');
    console.error('[PlantCare] Mode change error:', err);
  } finally {
    setTimeout(() => { dom.modeSwitch._userChanged = false; }, 1500);
  }
}

function showToast(message, type = 'info') {
  const icons = { success: '✅', error: '❌', info: 'ℹ️' };
  const toast = document.createElement('div');
  toast.className = `toast toast--${type}`;
  toast.setAttribute('role', 'alert');
  toast.innerHTML = `<span>${icons[type]}</span> <span>${message}</span>`;
  dom.toastContainer.appendChild(toast);
  setTimeout(() => {
    toast.classList.add('toast--out');
    toast.addEventListener('animationend', () => toast.remove(), { once: true });
  }, 3500);
}

async function handleRefresh() {
  dom.refreshIcon.classList.add('spinning');
  dom.refreshBtn.disabled = true;
  await Promise.all([fetchStatus(), fetchHistory(), fetchDevices()]);
  dom.refreshIcon.classList.remove('spinning');
  dom.refreshBtn.disabled = false;
  showToast('Data diperbarui', 'success');
}

async function handleTestSend(e) {
  e.preventDefault();
  const moisture = parseFloat(dom.testMoisture.value) || 0;
  const action = dom.testPump.value;
  const pumpOn = action === 'ON';
  const mode = dom.testMode.value;

  dom.testSendBtn.disabled = true;
  const originalHTML = dom.testSendBtn.innerHTML;
  dom.testSendBtn.innerHTML = `Mengirim...`;

  try {
    await apiFetch('/webhook', {
      method: 'POST',
      body: JSON.stringify({
        device_id: state.currentDeviceId,
        moisture: moisture,
        pump_on: pumpOn,
        action: action,
        mode: mode,
        force_log: true,
      }),
    });

    showToast(`Data unit ${state.currentDeviceId} tersimpan! Kelembapan: ${moisture}%, Pompa: ${action}`, 'success');
    await Promise.all([fetchStatus(), fetchHistory()]);
  } catch (err) {
    showToast(`Gagal: ${err.message}`, 'error');
    console.error('[PlantCare] Test send error:', err);
  } finally {
    dom.testSendBtn.innerHTML = originalHTML;
    dom.testSendBtn.disabled = false;
  }
}

async function handleClearHistory() {
  if (!confirm(`Hapus seluruh riwayat untuk unit ${state.currentDeviceId}?`)) return;
  try {
    await apiFetch(`/history/clear?device_id=${encodeURIComponent(state.currentDeviceId)}`, { method: 'POST' });
    showToast('Riwayat unit berhasil dibersihkan', 'info');
    await fetchHistory();
  } catch (err) {
    showToast(`Gagal: ${err.message}`, 'error');
    console.error('[PlantCare] Clear history error:', err);
  }
}

function bindEvents() {
  // Device selector
  if (dom.deviceSelector) {
    dom.deviceSelector.addEventListener('change', handleDeviceChange);
  }
  if (dom.addDeviceBtn) {
    dom.addDeviceBtn.addEventListener('click', openClaimModal);
  }

  // Modal close handlers
  if (dom.closeAuthModal) dom.closeAuthModal.addEventListener('click', closeAuthModal);
  if (dom.authModalBackdrop) dom.authModalBackdrop.addEventListener('click', closeAuthModal);
  if (dom.closeClaimModal) dom.closeClaimModal.addEventListener('click', closeClaimModal);
  if (dom.claimModalBackdrop) dom.claimModalBackdrop.addEventListener('click', closeClaimModal);

  // Tabs Login / Register
  if (dom.tabLogin && dom.tabRegister) {
    dom.tabLogin.addEventListener('click', () => {
      dom.tabLogin.classList.add('modal__tab--active');
      dom.tabRegister.classList.remove('modal__tab--active');
      dom.loginForm.style.display = 'flex';
      dom.registerForm.style.display = 'none';
      dom.authModalTitle.textContent = 'Masuk ke Akun';
      if (dom.googleBtnText) dom.googleBtnText.textContent = 'Masuk dengan Google';
    });
    dom.tabRegister.addEventListener('click', () => {
      dom.tabRegister.classList.add('modal__tab--active');
      dom.tabLogin.classList.remove('modal__tab--active');
      dom.loginForm.style.display = 'none';
      dom.registerForm.style.display = 'flex';
      dom.authModalTitle.textContent = 'Daftar Akun Baru';
      if (dom.googleBtnText) dom.googleBtnText.textContent = 'Daftar dengan Google';
    });
  }

  // Password visibility toggles
  setupPasswordToggle('toggleLoginPassword', 'loginPassword');
  setupPasswordToggle('toggleRegPassword', 'regPassword');

  // Google Sign-In
  if (dom.googleAuthBtn) {
    dom.googleAuthBtn.addEventListener('click', handleGoogleAuth);
  }

  // Auth form submits
  if (dom.loginForm) dom.loginForm.addEventListener('submit', handleLogin);
  if (dom.registerForm) dom.registerForm.addEventListener('submit', handleRegister);
  if (dom.claimForm) dom.claimForm.addEventListener('submit', handleClaimDevice);

  // Mode and pump controls
  dom.modeSwitch.addEventListener('change', (e) => {
    dom.modeSwitch._userChanged = true;
    setMode(e.target.checked);
  });
  dom.pumpOnBtn.addEventListener('click',  () => setPump(true));
  dom.pumpOffBtn.addEventListener('click', () => setPump(false));
  dom.refreshBtn.addEventListener('click', handleRefresh);
  dom.historyFilter.addEventListener('change', (e) => {
    state.historyFilter = e.target.value;
    renderHistory(state.history);
  });

  if (dom.testForm) {
    dom.testForm.addEventListener('submit', handleTestSend);
  }
  if (dom.clearHistoryBtn) {
    dom.clearHistoryBtn.addEventListener('click', handleClearHistory);
  }

  const syncAll = () => {
    fetchStatus();
    fetchHistory();
    fetchDevices();
  };
  window.addEventListener('focus', syncAll);
  document.addEventListener('visibilitychange', () => {
    if (!document.hidden) syncAll();
  });
}

async function init() {
  updateClock();
  setInterval(updateClock, 1000);
  renderAuthUI();
  renderMode(false);
  renderPump(false);
  bindEvents();

  await fetchDevices();
  await fetchStatus();
  await fetchHistory();

  setInterval(fetchStatus, CONFIG.POLL_INTERVAL);
  setInterval(fetchHistory, 12000);
  setInterval(keepAlive, CONFIG.KEEPALIVE_INTERVAL);

  // Initialize Google Sign-In when ready
  if (typeof google !== 'undefined' && google.accounts) {
    initGoogleSignIn();
  } else {
    window.addEventListener('load', initGoogleSignIn);
  }
}

document.addEventListener('DOMContentLoaded', init);
