// MockForge API Client
const api = {
  async getMetrics() {
    const res = await fetch('/api/admin/metrics');
    return res.json();
  },

  async getEndpoints() {
    const res = await fetch('/api/admin/endpoints');
    return res.json();
  },

  async createEndpoint(ep) {
    const res = await fetch('/api/admin/endpoints', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(ep),
    });
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.detail || 'Failed to create endpoint');
    }
    return res.json();
  },

  async updateEndpoint(id, ep) {
    const res = await fetch(`/api/admin/endpoints/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(ep),
    });
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.detail || 'Failed to update endpoint');
    }
    return res.json();
  },

  async deleteEndpoint(id) {
    const res = await fetch(`/api/admin/endpoints/${id}`, { method: 'DELETE' });
    if (!res.ok) throw new Error('Failed to delete endpoint');
    return res.json();
  },

  async getCollections() {
    const res = await fetch('/api/admin/collections');
    return res.json();
  },

  async createCollection(col) {
    const res = await fetch('/api/admin/collections', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(col),
    });
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.detail || 'Failed to create collection');
    }
    return res.json();
  },

  async updateCollection(id, col) {
    const res = await fetch(`/api/admin/collections/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(col),
    });
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.detail || 'Failed to update collection');
    }
    return res.json();
  },

  async deleteCollection(id) {
    const res = await fetch(`/api/admin/collections/${id}`, { method: 'DELETE' });
    if (!res.ok) throw new Error('Failed to delete collection');
    return res.json();
  },

  async getTraffic(limit = 100) {
    const res = await fetch(`/api/admin/traffic?limit=${limit}`);
    return res.json();
  },

  async clearTraffic() {
    const res = await fetch('/api/admin/traffic', { method: 'DELETE' });
    return res.json();
  },

  async generateAuthToken(data) {
    const res = await fetch('/api/admin/auth/generate-token', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return res.json();
  },

  async importOpenAPI(specJSON) {
    const res = await fetch('/api/admin/openapi/import', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: typeof specJSON === 'string' ? specJSON : JSON.stringify(specJSON),
    });
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.detail || 'Failed to import OpenAPI');
    }
    return res.json();
  },

  async exportOpenAPI() {
    const res = await fetch('/api/admin/openapi/export');
    return res.json();
  },

  async exportWorkspace() {
    const res = await fetch('/api/admin/workspace/export');
    return res.json();
  },

  async importWorkspace(wsJSON) {
    const res = await fetch('/api/admin/workspace/import', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: typeof wsJSON === 'string' ? wsJSON : JSON.stringify(wsJSON),
    });
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.detail || 'Failed to import workspace');
    }
    return res.json();
  },

  async resetDemo() {
    const res = await fetch('/api/admin/reset-demo', { method: 'POST' });
    return res.json();
  },

  async sendTestRequest({ method, url, headers = {}, body = null }) {
    const start = performance.now();
    try {
      const options = {
        method,
        headers: { ...headers },
      };
      if (body && method !== 'GET' && method !== 'HEAD') {
        options.body = body;
      }
      const res = await fetch(url, options);
      const duration = Math.round(performance.now() - start);

      const respHeaders = {};
      res.headers.forEach((v, k) => { respHeaders[k] = v; });

      const text = await res.text();
      let parsed = text;
      try {
        parsed = JSON.parse(text);
      } catch (e) {}

      return {
        status: res.status,
        statusText: res.statusText,
        headers: respHeaders,
        body: parsed,
        rawBody: text,
        duration,
      };
    } catch (err) {
      const duration = Math.round(performance.now() - start);
      return {
        status: 0,
        statusText: 'Network Error',
        headers: {},
        body: { error: err.message },
        rawBody: err.message,
        duration,
      };
    }
  }
};
