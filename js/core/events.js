/**
 * @file Lightweight publish-subscribe event emitter for decoupled feature communication.
 */

class EventEmitter {
  constructor() {
    this.events = {};
  }

  /**
   * Subscribes a listener to an event topic.
   * @param {string} event
   * @param {Function} listener
   */
  on(event, listener) {
    if (!this.events[event]) {
      this.events[event] = [];
    }
    this.events[event].push(listener);
    return () => this.off(event, listener);
  }

  /**
   * Unsubscribes a listener from an event.
   * @param {string} event
   * @param {Function} listener
   */
  off(event, listener) {
    if (!this.events[event]) return;
    this.events[event] = this.events[event].filter(l => l !== listener);
  }

  /**
   * Emits an event with optional payload to all subscribed listeners.
   * @param {string} event
   * @param {any} [data]
   */
  emit(event, data) {
    if (!this.events[event]) return;
    this.events[event].forEach(listener => {
      try {
        listener(data);
      } catch (err) {
        console.error(`[EventBus] Error in listener for event "${event}":`, err);
      }
    });
  }
}

export const eventBus = new EventEmitter();

