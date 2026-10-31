const CONFIG = {
  API_BASE: '/api',
  POLL_INTERVAL: 5000,
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
};

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
};

const MOCK = {
  status: () => ({
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

  const res = await fetch(`${CONFIG.API_BASE}${endpoint}`, {
    headers: {
      'Content-Type': 'application/json',
      'Accept-Encoding': 'gzip',
    },
    signal: AbortSignal.timeout(8000),
    ...options,
  });
  if (!res.ok) {
    const errorBody = await res.json().catch(() => ({}));
    throw new Error(errorBody.detail || errorBody.error || `HTTP ${res.status}`);
  }
  return res.json();
}

async function mockFetch(endpoint, options) {
  await delay(200 + Math.random() * 100);
  if (endpoint === '/status')   return MOCK.status();
  if (endpoint === '/history')  return MOCK.history;
  if (endpoint === '/history/clear') { MOCK.history = []; return { ok: true }; }
  if (endpoint === '/ping')     return { status: 'ok' };
  if (endpoint === '/pump/on')  { state.pumpOn = true;  return { ok: true, message: 'Pompa dinyalakan' }; }
  if (endpoint === '/pump/off') { state.pumpOn = false; return { ok: true, message: 'Pompa dimatikan' }; }
  if (endpoint === '/mode')     return { ok: true };
  if (endpoint === '/webhook') {
    let body = {};
    if (options && options.body) {
      try { body = JSON.parse(options.body); } catch(e) {}
    }
    const moist = body.moisture ?? 40;
    const action = body.action ?? (body.pump_on ? 'ON' : 'OFF');
    const mode = body.mode ?? 'auto';
    state.moisture = moist;
    state.pumpOn = body.pump_on ?? (action === 'ON');
    MOCK.history.unshift({
      id: Date.now(),
      timestamp: new Date().toISOString(),
      action: action,
      mode: mode,
      moisture: moist,
      duration: '—'
    });
    return { ok: true, moisture: moist, pump_on: state.pumpOn };
  }
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

async function fetchStatus() {
  const now = Date.now();
  if (document.hidden && now - state.lastFetch < 30000) return;
  state.lastFetch = now;

  try {
    const data = await apiFetch('/status');
    setConnected(true);
    renderPump(data.pump_on);
    renderMoisture(data.moisture);
    if (data.mode && !dom.modeSwitch._userChanged) {
      const isManual = data.mode === 'manual';
      if (dom.modeSwitch.checked !== !isManual) {
        dom.modeSwitch.checked = !isManual;
        renderMode(isManual);
      }
    }
  } catch (err) {
    setConnected(false);
    console.error('[PlantCare] Status fetch error:', err);
  }
}

async function fetchHistory() {
  try {
    const data = await apiFetch('/history');
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
    const res = await apiFetch(on ? '/pump/on' : '/pump/off', { method: 'POST' });
    renderPump(on);
    MOCK.history.unshift({
      id:        Date.now(),
      timestamp: new Date().toISOString(),
      action:    on ? 'ON' : 'OFF',
      mode:      'manual',
      moisture:  state.moisture ?? 0,
      duration:  '—',
    });
    renderHistory(MOCK.history);
    showToast(res.message || (on ? 'Pompa dinyalakan' : 'Pompa dimatikan'), 'success');
  } catch (err) {
    showToast('Gagal mengontrol pompa. Coba lagi.', 'error');
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
      body:   JSON.stringify({ mode: !isAuto ? 'manual' : 'auto' }),
    });
    showToast(`Mode diubah ke: ${!isAuto ? 'Manual' : 'Otomatis'}`, 'info');
  } catch (err) {
    console.error('[PlantCare] Mode change error:', err);
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
  await Promise.all([fetchStatus(), fetchHistory()]);
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
    const res = await apiFetch('/webhook', {
      method: 'POST',
      body: JSON.stringify({
        moisture: moisture,
        pump_on: pumpOn,
        action: action,
        mode: mode,
        force_log: true,
      }),
    });

    showToast(`Data terkirim ke DB! Kelembapan: ${moisture}%, Pompa: ${action}`, 'success');
    await Promise.all([fetchStatus(), fetchHistory()]);
  } catch (err) {
    showToast(`Gagal mengirim ke DB: ${err.message}`, 'error');
    console.error('[PlantCare] Test send error:', err);
  } finally {
    dom.testSendBtn.innerHTML = originalHTML;
    dom.testSendBtn.disabled = false;
  }
}

async function handleClearHistory() {
  if (!confirm('Apakah Anda yakin ingin menghapus seluruh riwayat penyiraman di database?')) return;
  try {
    await apiFetch('/history/clear', { method: 'POST' });
    showToast('Riwayat berhasil dibersihkan dari DB', 'info');
    await fetchHistory();
  } catch (err) {
    showToast(`Gagal membersihkan riwayat: ${err.message}`, 'error');
    console.error('[PlantCare] Clear history error:', err);
  }
}

function bindEvents() {
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

  document.addEventListener('visibilitychange', () => {
    if (!document.hidden) fetchStatus();
  });
}

async function init() {
  updateClock();
  setInterval(updateClock, 1000);
  renderMode(false);
  renderPump(false);
  bindEvents();
  await fetchStatus();
  await fetchHistory();
  setInterval(fetchStatus, CONFIG.POLL_INTERVAL);
  setInterval(keepAlive, CONFIG.KEEPALIVE_INTERVAL);
}

document.addEventListener('DOMContentLoaded', init);
