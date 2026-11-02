/**
 * @file Simulator panel view helper.
 */

import { $ } from '../../utils/dom.js';

export const simulatorView = {
  get elements() {
    return {
      testForm:     $('testForm'),
      testMoisture: $('testMoisture'),
      testPump:     $('testPump'),
      testMode:     $('testMode'),
      testSendBtn:  $('testSendBtn'),
    };
  },

  setLoading(loading) {
    const btn = this.elements.testSendBtn;
    if (!btn) return;

    if (loading) {
      btn._orig = btn.innerHTML;
      btn.disabled = true;
      btn.innerHTML = 'Mengirim...';
    } else {
      if (btn._orig) btn.innerHTML = btn._orig;
      btn.disabled = false;
    }
  },
};

