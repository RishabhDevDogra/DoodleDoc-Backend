// Frontend configuration constants
// In production with Docker/nginx proxying, use relative paths. In production
// with separate hosted frontend/backend services, set REACT_APP_API_URL.
const isProduction = process.env.NODE_ENV === 'production';
export const API_URL = process.env.REACT_APP_API_URL || (isProduction ? '' : 'http://localhost:8080');
export const HUB_URL = `${API_URL}/hubs/document`;

function buildWebSocketURL() {
	if (API_URL) {
		return `${API_URL.replace(/^http/, 'ws')}/hubs/document`;
	}

	const wsProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
	return `${wsProtocol}://${window.location.host}/hubs/document`;
}

export const WS_HUB_URL = buildWebSocketURL();

// Auto-save delays (milliseconds)
export const TITLE_SAVE_DELAY_MS = 1000;
export const CANVAS_SAVE_DELAY_MS = 500;
