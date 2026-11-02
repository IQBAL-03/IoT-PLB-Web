/**
 * @file Google Identity Services (GIS) OAuth integration and token processor.
 */

import { CONFIG } from '../../config/constants.js';
import { showToast } from '../../utils/toast.js';

let googleTokenClient = null;

/**
 * Initializes Google OAuth2 Token Client.
 * @param {Function} onTokenReceived - Callback when access token is successfully received
 */
export function initGoogleOAuth(onTokenReceived) {
  if (typeof google === 'undefined' || !google.accounts || !google.accounts.oauth2) return;

  try {
    googleTokenClient = google.accounts.oauth2.initTokenClient({
      client_id: CONFIG.GOOGLE_CLIENT_ID,
      scope: 'email profile openid',
      callback: async (tokenResponse) => {
        if (tokenResponse && tokenResponse.error) {
          if (tokenResponse.error === 'access_denied') return;
          showToast(`Google login: ${tokenResponse.error_description || tokenResponse.error}`, 'error');
          return;
        }
        if (tokenResponse && tokenResponse.access_token) {
          await onTokenReceived(tokenResponse.access_token);
        }
      },
      error_callback: (nonOAuthError) => {
        console.warn('[PlantCare] Google OAuth:', nonOAuthError);
        if (nonOAuthError && (nonOAuthError.type === 'popup_closed' || nonOAuthError.type === 'popup_closed_by_user')) {
          return;
        }
        if (nonOAuthError && nonOAuthError.type === 'popup_blocked_by_browser') {
          showToast('Popup Google diblokir browser. Mohon izinkan popup untuk web ini.', 'warning');
          return;
        }
        showToast(`Google login error: ${nonOAuthError.type || 'Inisialisasi popup gagal'}`, 'error');
      },
    });
  } catch (err) {
    console.warn('[PlantCare] Google OAuth init error:', err);
  }
}

/**
 * Prompts user to authenticate with Google.
 * @param {Function} onTokenReceived
 */
export function requestGoogleSignIn(onTokenReceived) {
  if (typeof google !== 'undefined' && google.accounts && google.accounts.oauth2) {
    if (!googleTokenClient) {
      initGoogleOAuth(onTokenReceived);
    }
    if (googleTokenClient) {
      googleTokenClient.requestAccessToken({ prompt: 'select_account' });
      return;
    }
  }

  // Fallback prompt for adblockers or offline SDK
  return null;
}

