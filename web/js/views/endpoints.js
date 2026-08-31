// Endpoints Studio View Module
views = window.views || {};

views.endpoints = {
  endpoints: [],
  filterText: '',
  filterMethod: 'ALL',

  async render(container) {
    container.innerHTML = `
      <div class="card">
        <div class="card-header">
          <div>
            <div class="card-title">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/></svg>
              Mock Endpoint Studio
            </div>
            <div class="card-subtitle">Create and customize simulated REST routes with realistic fake templates</div>
          </div>
          <button class="btn btn-primary" onclick="views.endpoints.openEditor()">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            Create Endpoint
          </button>
        </div>

        <div style="display: flex; gap: var(--spacing-sm); margin-bottom: var(--spacing-md); align-items: center;">
          <input type="text" class="form-input" placeholder="Filter by path or name..." style="max-width: 320px;" id="ep-filter-input" oninput="views.endpoints.onFilterChange()">
          <select class="form-select" style="max-width: 140px;" id="ep-filter-method" onchange="views.endpoints.onFilterChange()">
            <option value="ALL">All Methods</option>
            <option value="GET">GET</option>
            <option value="POST">POST</option>
            <option value="PUT">PUT</option>
            <option value="PATCH">PATCH</option>
            <option value="DELETE">DELETE</option>
          </select>
        </div>

        <div id="endpoints-table-container">Loading...</div>
      </div>
    `;

    await this.load();
  },

  async load() {
    try {
      this.endpoints = await api.getEndpoints();
      const badgeEl = document.getElementById('badge-endpoint-count');
      if (badgeEl) badgeEl.textContent = this.endpoints.length;
      this.renderTable();
    } catch (err) {
      app.showToast('Failed to load endpoints: ' + err.message, 'error');
    }
  },

  onFilterChange() {
    const input = document.getElementById('ep-filter-input');
    const select = document.getElementById('ep-filter-method');
    if (input) this.filterText = input.value.toLowerCase().trim();
    if (select) this.filterMethod = select.value;
    this.renderTable();
  },

  renderTable() {
    const container = document.getElementById('endpoints-table-container');
    if (!container) return;

    let filtered = this.endpoints;
    if (this.filterMethod !== 'ALL') {
      filtered = filtered.filter(ep => ep.method === this.filterMethod);
    }
    if (this.filterText) {
      filtered = filtered.filter(ep => ep.path.toLowerCase().includes(this.filterText) || ep.name.toLowerCase().includes(this.filterText));
    }

    if (filtered.length === 0) {
      container.innerHTML = `<div style="text-align: center; padding: 40px 0; color: var(--color-ink-subtle);">No matching mock endpoints found. Click "Create Endpoint" to add one.</div>`;
      return;
    }

    container.innerHTML = `
      <div class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th style="width: 70px;">State</th>
              <th style="width: 80px;">Method</th>
              <th>Route Path & Name</th>
              <th style="width: 80px;">Status</th>
              <th>Auth Guard</th>
              <th>Latency / Chaos</th>
              <th style="text-align: right;">Actions</th>
            </tr>
          </thead>
          <tbody>
            ${filtered.map(ep => `
              <tr>
                <td>
                  <input type="checkbox" ${ep.enabled ? 'checked' : ''} onchange="views.endpoints.toggleEnabled('${ep.id}', this.checked)" title="Enable / Disable">
                </td>
                <td><span class="badge-method ${ep.method.toLowerCase()}">${ep.method}</span></td>
                <td>
                  <div class="route-cell">${escapeHTML(ep.path)}</div>
                  <div style="font-size: 12px; color: var(--color-ink-subtle); margin-top: 2px;">${escapeHTML(ep.name || '')}</div>
                </td>
                <td><span class="badge-status s${String(ep.response.statusCode)[0]}xx">${ep.response.statusCode}</span></td>
                <td>
                  <span style="font-size: 12px; color: ${ep.auth.type !== 'none' ? 'var(--color-primary-hover)' : 'var(--color-ink-tertiary)'}; font-weight: 500;">
                    ${ep.auth.type === 'none' ? 'Public' : `${ep.auth.type.toUpperCase()} Auth`}
                  </span>
                </td>
                <td>
                  <div style="display: flex; gap: 4px; flex-wrap: wrap;">
                    ${ep.latency.enabled ? `<span style="font-size: 11px; background: var(--color-surface-3); padding: 1px 6px; border-radius: var(--radius-pill); color: var(--color-ink-muted);">⏱ ${ep.latency.mode === 'random' ? `${ep.latency.minMs}-${ep.latency.maxMs}ms` : `${ep.latency.fixedMs}ms`}</span>` : ''}
                    ${ep.chaos.enabled ? `<span style="font-size: 11px; background: rgba(229, 72, 77, 0.15); padding: 1px 6px; border-radius: var(--radius-pill); color: #fca5a5;">💥 ${Math.round(ep.chaos.rate * 100)}% Chaos</span>` : ''}
                    ${!ep.latency.enabled && !ep.chaos.enabled ? '<span style="color: var(--color-ink-tertiary); font-size: 12px;">Standard</span>' : ''}
                  </div>
                </td>
                <td style="text-align: right;">
                  <button class="btn btn-secondary btn-sm" onclick="views.tester.testEndpoint('${ep.method}', '${ep.path}')" title="Test Request">Test</button>
                  <button class="btn btn-secondary btn-sm" onclick="views.endpoints.openEditor('${ep.id}')" title="Edit Configuration">Edit</button>
                  <button class="btn btn-danger btn-sm" onclick="views.endpoints.deleteEndpoint('${ep.id}')" title="Delete Endpoint">Delete</button>
                </td>
              </tr>
            `).join('')}
          </tbody>
        </table>
      </div>
    `;
  },

  async toggleEnabled(id, enabled) {
    const ep = this.endpoints.find(e => e.id === id);
    if (!ep) return;
    ep.enabled = enabled;
    try {
      await api.updateEndpoint(id, ep);
      app.showToast(`Endpoint ${enabled ? 'enabled' : 'disabled'}`, 'success');
    } catch (err) {
      app.showToast('Update failed: ' + err.message, 'error');
      await this.load();
    }
  },

  async deleteEndpoint(id) {
    if (!confirm('Are you sure you want to delete this mock endpoint?')) return;
    try {
      await api.deleteEndpoint(id);
      app.showToast('Endpoint deleted successfully', 'success');
      await this.load();
    } catch (err) {
      app.showToast('Delete failed: ' + err.message, 'error');
    }
  },

  openEditor(id = null) {
    let ep = {
      id: '',
      name: '',
      description: '',
      method: 'GET',
      path: '/api/v1/resource',
      enabled: true,
      tags: [],
      response: {
        statusCode: 200,
        contentType: 'application/json',
        headers: { 'Content-Type': 'application/json' },
        body: '{\n  "status": "success",\n  "data": {\n    "id": "{{faker.uuid}}",\n    "name": "{{faker.name}}",\n    "email": "{{faker.email}}"\n  }\n}',
      },
      latency: { enabled: false, mode: 'fixed', fixedMs: 150, minMs: 50, maxMs: 300 },
      chaos: { enabled: false, rate: 0.1, statusCode: 500, responseBody: '{"error": "Simulated Chaos Error"}' },
      auth: { type: 'none', token: '' }
    };

    if (id) {
      const found = this.endpoints.find(e => e.id === id);
      if (found) ep = JSON.parse(JSON.stringify(found));
    }

    const html = `
      <form id="endpoint-form" onsubmit="views.endpoints.saveEditor(event, '${id || ''}')">
        <div style="display: grid; grid-template-columns: 120px 1fr; gap: var(--spacing-sm);">
          <div class="form-group">
            <label class="form-label">HTTP Method</label>
            <select class="form-select" id="ep-form-method">
              ${['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS', 'HEAD'].map(m => `
                <option value="${m}" ${ep.method === m ? 'selected' : ''}>${m}</option>
              `).join('')}
            </select>
          </div>
          <div class="form-group">
            <label class="form-label">Route Path (supports :id, *wildcard)</label>
            <input type="text" class="form-input" id="ep-form-path" value="${escapeHTML(ep.path)}" required>
          </div>
        </div>

        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: var(--spacing-sm);">
          <div class="form-group">
            <label class="form-label">Endpoint Name</label>
            <input type="text" class="form-input" id="ep-form-name" value="${escapeHTML(ep.name || '')}" placeholder="e.g. Get User Profile">
          </div>
          <div class="form-group">
            <label class="form-label">Response Status Code</label>
            <input type="number" class="form-input" id="ep-form-status" value="${ep.response.statusCode || 200}">
          </div>
        </div>

        <!-- Response Body Template Section with Faker Snippet Helper -->
        <div class="form-group">
          <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px;">
            <label class="form-label" style="margin-bottom: 0;">Response Body (Template & Dynamic Faker)</label>
            <div style="display: flex; gap: 4px; flex-wrap: wrap;">
              <button type="button" class="btn btn-secondary btn-sm" onclick="views.endpoints.insertTag('{{faker.name}}')">+ Name</button>
              <button type="button" class="btn btn-secondary btn-sm" onclick="views.endpoints.insertTag('{{faker.email}}')">+ Email</button>
              <button type="button" class="btn btn-secondary btn-sm" onclick="views.endpoints.insertTag('{{faker.uuid}}')">+ UUID</button>
              <button type="button" class="btn btn-secondary btn-sm" onclick="views.endpoints.insertTag('{{faker.price(10, 500)}}')">+ Price</button>
              <button type="button" class="btn btn-secondary btn-sm" onclick="views.endpoints.insertTag('{{req.params.id}}')">+ :id Param</button>
              <button type="button" class="btn btn-secondary btn-sm" onclick="views.endpoints.insertRepeatBlock()">+ Repeat List</button>
            </div>
          </div>
          <textarea class="form-textarea" id="ep-form-body" style="height: 180px;">${escapeHTML(ep.response.body || '')}</textarea>
        </div>

        <!-- Latency & Chaos Settings Accordion -->
        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: var(--spacing-md); background: var(--color-surface-2); padding: var(--spacing-md); border-radius: var(--radius-md); margin-bottom: var(--spacing-md); border: 1px solid var(--color-hairline);">
          <div>
            <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 8px;">
              <input type="checkbox" id="ep-form-lat-enabled" ${ep.latency.enabled ? 'checked' : ''}>
              <label for="ep-form-lat-enabled" style="font-weight: 600; font-size: 13px; color: var(--color-ink);">⏱ Simulated Latency</label>
            </div>
            <div style="display: flex; gap: 8px;">
              <input type="number" class="form-input" id="ep-form-lat-fixed" placeholder="Fixed Ms (e.g. 200)" value="${ep.latency.fixedMs || 150}">
            </div>
          </div>

          <div>
            <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 8px;">
              <input type="checkbox" id="ep-form-chaos-enabled" ${ep.chaos.enabled ? 'checked' : ''}>
              <label for="ep-form-chaos-enabled" style="font-weight: 600; font-size: 13px; color: var(--color-ink);">💥 Chaos Error Injection</label>
            </div>
            <div style="display: flex; gap: 8px;">
              <input type="number" step="0.05" min="0" max="1" class="form-input" id="ep-form-chaos-rate" placeholder="Rate (0.1 = 10%)" value="${ep.chaos.rate || 0.1}">
              <input type="number" class="form-input" id="ep-form-chaos-status" placeholder="Status (500)" value="${ep.chaos.statusCode || 500}">
            </div>
          </div>
        </div>

        <!-- Auth Guard Settings -->
        <div class="form-group" style="background: var(--color-surface-2); padding: var(--spacing-md); border-radius: var(--radius-md); border: 1px solid var(--color-hairline);">
          <label class="form-label" style="font-weight: 600; font-size: 13px; color: var(--color-ink);">🔒 Authentication Guard</label>
          <div style="display: grid; grid-template-columns: 140px 1fr; gap: var(--spacing-sm); margin-top: 8px;">
            <select class="form-select" id="ep-form-auth-type" onchange="views.endpoints.onAuthTypeChange(this.value)">
              <option value="none" ${ep.auth.type === 'none' ? 'selected' : ''}>None (Public)</option>
              <option value="bearer" ${ep.auth.type === 'bearer' ? 'selected' : ''}>Bearer Token</option>
              <option value="apikey" ${ep.auth.type === 'apikey' ? 'selected' : ''}>API Key</option>
              <option value="basic" ${ep.auth.type === 'basic' ? 'selected' : ''}>Basic Auth</option>
              <option value="jwt" ${ep.auth.type === 'jwt' ? 'selected' : ''}>JWT Verification</option>
            </select>
            <input type="text" class="form-input" id="ep-form-auth-token" placeholder="Bearer token / API key / JWT secret..." value="${escapeHTML(ep.auth.token || ep.auth.jwtSecret || '')}">
          </div>
        </div>
      </form>
    `;

    const actions = `
      <button type="button" class="btn btn-secondary" onclick="app.closeModal()">Cancel</button>
      <button type="submit" form="endpoint-form" class="btn btn-primary">Save Mock Endpoint</button>
    `;

    app.openModal(id ? 'Edit Mock Endpoint' : 'Create New Mock Endpoint', html, actions);
  },

  insertTag(tag) {
    const textarea = document.getElementById('ep-form-body');
    if (!textarea) return;
    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const text = textarea.value;
    textarea.value = text.substring(0, start) + tag + text.substring(end);
    textarea.focus();
    textarea.selectionEnd = start + tag.length;
  },

  insertRepeatBlock() {
    const repeatSnippet = `[\n  {{#repeat 5}}\n  {\n    "id": {{@iteration}},\n    "name": "{{faker.name}}",\n    "email": "{{faker.email}}",\n    "avatar": "{{faker.avatar}}",\n    "price": {{faker.price(10, 100)}}\n  }\n  {{/repeat}}\n]`;
    const textarea = document.getElementById('ep-form-body');
    if (textarea) textarea.value = repeatSnippet;
  },

  onAuthTypeChange(val) {
    const tokenInput = document.getElementById('ep-form-auth-token');
    if (!tokenInput) return;
    if (val === 'none') {
      tokenInput.style.display = 'none';
    } else {
      tokenInput.style.display = 'block';
      if (val === 'bearer') tokenInput.placeholder = 'Required Bearer token (leave blank to allow any)';
      if (val === 'apikey') tokenInput.placeholder = 'Required API Key value';
      if (val === 'jwt') tokenInput.placeholder = 'JWT Verification Secret (HMAC-SHA256)';
      if (val === 'basic') tokenInput.placeholder = 'username:password';
    }
  },

  async saveEditor(event, id) {
    event.preventDefault();
    const method = document.getElementById('ep-form-method').value;
    const path = document.getElementById('ep-form-path').value.trim();
    const name = document.getElementById('ep-form-name').value.trim();
    const statusCode = parseInt(document.getElementById('ep-form-status').value, 10) || 200;
    const body = document.getElementById('ep-form-body').value;

    const latEnabled = document.getElementById('ep-form-lat-enabled').checked;
    const latFixed = parseInt(document.getElementById('ep-form-lat-fixed').value, 10) || 0;

    const chaosEnabled = document.getElementById('ep-form-chaos-enabled').checked;
    const chaosRate = parseFloat(document.getElementById('ep-form-chaos-rate').value) || 0.1;
    const chaosStatus = parseInt(document.getElementById('ep-form-chaos-status').value, 10) || 500;

    const authType = document.getElementById('ep-form-auth-type').value;
    const authToken = document.getElementById('ep-form-auth-token').value.trim();

    const ep = {
      id: id || undefined,
      name: name || `${method} ${path}`,
      method,
      path,
      enabled: true,
      response: {
        statusCode,
        contentType: 'application/json',
        headers: { 'Content-Type': 'application/json' },
        body,
      },
      latency: {
        enabled: latEnabled,
        mode: 'fixed',
        fixedMs: latFixed,
      },
      chaos: {
        enabled: chaosEnabled,
        rate: chaosRate,
        statusCode: chaosStatus,
        responseBody: '{"error":"Simulated Chaos Error"}',
      },
      auth: {
        type: authType,
        token: authToken,
        jwtSecret: authType === 'jwt' ? authToken : undefined,
      },
    };

    try {
      if (id) {
        await api.updateEndpoint(id, ep);
        app.showToast('Mock endpoint updated', 'success');
      } else {
        await api.createEndpoint(ep);
        app.showToast('Mock endpoint created', 'success');
      }
      app.closeModal();
      await this.load();
    } catch (err) {
      app.showToast('Save failed: ' + err.message, 'error');
    }
  }
};
