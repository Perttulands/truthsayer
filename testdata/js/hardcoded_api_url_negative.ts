// Negative: URLs from env vars or config
const API = process.env.API_URL;
const BASE = config.get('apiUrl');
fetch(`${API}/users`);
