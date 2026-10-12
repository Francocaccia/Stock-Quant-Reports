// Backend base URL. Empty VITE_API_URL falls back to a local dev server.
export const API = import.meta.env.VITE_API_URL || 'http://localhost:8080';

// Comma-separated list of emails that see the admin UI (empty = nobody).
export const ADMIN_EMAILS = (import.meta.env.VITE_ADMIN_EMAILS || '')
  .split(',')
  .map((e) => e.trim())
  .filter(Boolean);
