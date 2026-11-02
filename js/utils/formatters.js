/**
 * @file Formatting helper utilities for dates and times.
 */

/**
 * Formats an ISO date string into locale-friendly Indonesian datetime.
 * @param {string} iso - ISO date string
 * @returns {string}
 */
export function formatTimestamp(iso) {
  try {
    return new Date(iso).toLocaleString('id-ID', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
    });
  } catch {
    return iso;
  }
}

