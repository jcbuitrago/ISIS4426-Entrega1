// Environment-based API configuration
const getApiBaseUrl = () => {
  // In production, use ALB endpoint
  if (import.meta.env.PROD) {
    return import.meta.env.VITE_API_BASE_URL || 'http://anb-alb-1580832969.us-east-1.elb.amazonaws.com';
  }
  // In development, use relative URLs (Vite proxy)
  return '';
};

const API_BASE_URL = getApiBaseUrl();

export function getAccessToken() {
  const tok = localStorage.getItem("access_token");
  const expAt = Number(localStorage.getItem("access_token_expires_at") || 0);
  if (!tok) return null;
  if (expAt && Date.now() > expAt) return null;
  return tok;
}

export function decodeAccessToken() {
  const token = getAccessToken();
  if (!token) return null;
  const parts = token.split(".");
  if (parts.length !== 3) return null;
  try {
    const payload = JSON.parse(atob(parts[1].replace(/-/g, "+").replace(/_/g, "/")));
    return payload || null;
  } catch {
    return null;
  }
}

function authHeaders() {
  const t = getAccessToken();
  return t ? { Authorization: `Bearer ${t}` } : {};
}

async function handleJson(res) {
  const text = await res.text();
  let data = null;
  try { data = text ? JSON.parse(text) : null; } catch {}
  if (!res.ok) {
    throw new Error((typeof data === "string" && data) || data?.message || text || "Error de servidor");
  }
  return data;
}

export const api = {
  // Auth
  async login({ email, password, remember }) {
    const res = await fetch(`${API_BASE_URL}/api/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password, remember: !!remember }),
    });
    return handleJson(res);
  },
  async signup({ first_name, last_name, email, password1, password2, city, country }) {
    const res = await fetch(`${API_BASE_URL}/api/auth/signup`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ first_name, last_name, email, password1, password2, city, country }),
    });
    return handleJson(res);
  },

  // Profile (protected)
  async getMe() {
    const res = await fetch(`${API_BASE_URL}/api/me`, { headers: { ...authHeaders() } });
    return handleJson(res);
  },
  async updateMe({ first_name, last_name, city, country }) {
    const res = await fetch(`${API_BASE_URL}/api/me`, {
      method: "PUT",
      headers: { "Content-Type": "application/json", ...authHeaders() },
      body: JSON.stringify({ first_name, last_name, city, country }),
    });
    return handleJson(res);
  },
  async uploadAvatar(file) {
    const form = new FormData();
    form.set("avatar", file);
    const res = await fetch(`${API_BASE_URL}/api/me/avatar`, {
      method: "POST",
      headers: { ...authHeaders() },
      body: form,
    });
    return handleJson(res);
  },

  // Videos (protected)
  async listMyVideos({ limit, offset, user_id } = {}) {
    const qs = new URLSearchParams();
    if (limit) qs.set("limit", String(limit));
    if (offset) qs.set("offset", String(offset));
    if (user_id) qs.set("user_id", String(user_id));
    const res = await fetch(`${API_BASE_URL}/api/videos?${qs.toString()}`,
      { headers: { ...authHeaders() } });
    return handleJson(res);
  },
  async uploadVideo({ title, file }) {
    const form = new FormData();
    if (title) form.set("title", title);
    form.set("video_file", file);
    const res = await fetch(`${API_BASE_URL}/api/videos`, {
      method: "POST",
      headers: { ...authHeaders() },
      body: form,
    });
    return handleJson(res);
  },
  async deleteVideo(id) {
    const res = await fetch(`${API_BASE_URL}/api/videos/${id}`, {
      method: "DELETE",
      headers: { ...authHeaders() },
    });
    return handleJson(res);
  },
  async getVideo(id) {
    const res = await fetch(`${API_BASE_URL}/api/videos/${id}`, { headers: { ...authHeaders() } });
    return handleJson(res);
  },

  // Jobs (public)
  async getJob(id) {
    const res = await fetch(`${API_BASE_URL}/api/jobs/${id}`);
    return handleJson(res);
  },

  // Public
  async listPublicVideos({ limit, offset } = {}) {
    const qs = new URLSearchParams();
    if (limit) qs.set("limit", String(limit));
    if (offset) qs.set("offset", String(offset));
    const res = await fetch(`${API_BASE_URL}/api/public/videos?${qs.toString()}`);
    return handleJson(res);
  },
  async voteVideo(id) {
    const res = await fetch(`${API_BASE_URL}/api/public/videos/${id}/vote`, {
      method: "POST",
      headers: { ...authHeaders() },
    });
    return handleJson(res);
  },
  async rankings({ city } = {}) {
    const qs = new URLSearchParams();
    if (city) qs.set("city", city);
    const res = await fetch(`${API_BASE_URL}/api/public/rankings?${qs.toString()}`);
    return handleJson(res);
  },
};
