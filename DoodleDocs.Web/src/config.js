// Frontend configuration constants
// In production (Docker), use relative paths; in dev, use localhost
const isProduction = process.env.NODE_ENV === 'production';
export const API_URL = process.env.REACT_APP_API_URL || (isProduction ? '' : 'http://localhost:8080');
export const HUB_URL = `${API_URL}/hubs/document`;

// WebSocket URL for real-time updates (ws:// in dev, wss:// in production HTTPS)
const wsProtocol = isProduction ? (window.location.protocol === 'https:' ? 'wss' : 'ws') : 'ws';
const wsHost = isProduction ? window.location.host : 'localhost:8080';
export const WS_HUB_URL = `${wsProtocol}://${wsHost}/hubs/document`;

// Auto-save delays (milliseconds)
export const TITLE_SAVE_DELAY_MS = 1000;
export const CANVAS_SAVE_DELAY_MS = 500;
