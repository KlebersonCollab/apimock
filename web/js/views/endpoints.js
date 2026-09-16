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
              <th>Scenarios</th>
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
                  ${ep.scenarios && ep.scenarios.length > 0 ? `
                    <span class="badge-scenarios" title="${ep.scenarios.length} Conditional Scenarios Configured">
                      🎯 ${ep.scenarios.length} Rule${ep.scenarios.length > 1 ? 's' : ''}
                    </span>
                  ` : `
                    <span style="color: var(--color-ink-tertiary); font-size: 11px;">1 Default</span>
                  `}
                </td>
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

  currentEditingScenarios: [],
  currentModalTab: 'general',

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
      auth: { type: 'none', token: '' },
      scenarios: []
    };

    if (id) {
      const found = this.endpoints.find(e => e.id === id);
      if (found) ep = JSON.parse(JSON.stringify(found));
    }

    this.currentEditingScenarios = ep.scenarios && Array.isArray(ep.scenarios) ? JSON.parse(JSON.stringify(ep.scenarios)) : [];
    this.currentModalTab = 'general';

    const html = `
      <!-- Modal Subtabs Header -->
      <div class="modal-subtabs" style="margin: -16px -16px 16px -16px; border-bottom: 1px solid var(--color-hairline);">
        <button type="button" class="modal-subtab active" data-subtab="general" onclick="views.endpoints.switchModalTab('general')">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
          Route Info
        </button>
        <button type="button" class="modal-subtab" data-subtab="response" onclick="views.endpoints.switchModalTab('response')">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/></svg>
          Default Response
        </button>
        <button type="button" class="modal-subtab" data-subtab="scenarios" onclick="views.endpoints.switchModalTab('scenarios')">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
          Conditional Scenarios <span class="subtab-badge" id="ep-modal-scenario-count">${this.currentEditingScenarios.length}</span>
        </button>
        <button type="button" class="modal-subtab" data-subtab="auth" onclick="views.endpoints.switchModalTab('auth')">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
          Auth & Resilience
        </button>
      </div>

      <form id="endpoint-form" onsubmit="views.endpoints.saveEditor(event, '${id || ''}')">

        <!-- TAB 1: Route Info -->
        <div id="subtab-content-general" class="subtab-pane">
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

          <div class="form-group">
            <label class="form-label">Endpoint Name</label>
            <input type="text" class="form-input" id="ep-form-name" value="${escapeHTML(ep.name || '')}" placeholder="e.g. Authenticate User or Get Customer">
          </div>

          <div class="form-group">
            <label class="form-label">Description (Optional)</label>
            <input type="text" class="form-input" id="ep-form-description" value="${escapeHTML(ep.description || '')}" placeholder="Brief overview of the API route behavior">
          </div>
        </div>

        <!-- TAB 2: Default Response -->
        <div id="subtab-content-response" class="subtab-pane" style="display: none;">
          <div class="form-group" style="max-width: 200px;">
            <label class="form-label">Default Status Code</label>
            <input type="number" class="form-input" id="ep-form-status" value="${ep.response.statusCode || 200}">
          </div>

          <div class="form-group">
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px;">
              <label class="form-label" style="margin-bottom: 0;">Default Response Body (Faker & Template)</label>
              <div style="display: flex; gap: 4px; flex-wrap: wrap;">
                <button type="button" class="btn btn-secondary btn-sm" onclick="views.endpoints.insertTag('{{faker.name}}')">+ Name</button>
                <button type="button" class="btn btn-secondary btn-sm" onclick="views.endpoints.insertTag('{{faker.email}}')">+ Email</button>
                <button type="button" class="btn btn-secondary btn-sm" onclick="views.endpoints.insertTag('{{faker.uuid}}')">+ UUID</button>
                <button type="button" class="btn btn-secondary btn-sm" onclick="views.endpoints.insertTag('{{faker.price(10, 500)}}')">+ Price</button>
                <button type="button" class="btn btn-secondary btn-sm" onclick="views.endpoints.insertTag('{{req.params.id}}')">+ :id Param</button>
                <button type="button" class="btn btn-secondary btn-sm" onclick="views.endpoints.insertRepeatBlock()">+ Repeat List</button>
              </div>
            </div>
            <textarea class="form-textarea" id="ep-form-body" style="height: 220px;">${escapeHTML(ep.response.body || '')}</textarea>
          </div>
        </div>

        <!-- TAB 3: Conditional Scenarios -->
        <div id="subtab-content-scenarios" class="subtab-pane" style="display: none;">
          <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-md);">
            <div>
              <div style="font-weight: 600; font-size: 13px; color: var(--color-ink);">🎯 Multi-Scenario Branch Rules</div>
              <div style="font-size: 12px; color: var(--color-ink-subtle);">Rules evaluate sequentially (First-Match-Wins). If no rule matches, Default Response executes.</div>
            </div>
            <button type="button" class="btn btn-primary btn-sm" onclick="views.endpoints.addScenario()">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
              + Add Scenario
            </button>
          </div>

          <div id="scenarios-container" style="max-height: 400px; overflow-y: auto; padding-right: 4px;">
            <!-- Rendered dynamically -->
          </div>
        </div>

        <!-- TAB 4: Auth & Resilience -->
        <div id="subtab-content-auth" class="subtab-pane" style="display: none;">
          <!-- Auth Guard -->
          <div class="form-group" style="background: var(--color-surface-2); padding: var(--spacing-md); border-radius: var(--radius-md); border: 1px solid var(--color-hairline); margin-bottom: var(--spacing-md);">
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

          <!-- Latency & Chaos -->
          <div style="display: grid; grid-template-columns: 1fr 1fr; gap: var(--spacing-md); background: var(--color-surface-2); padding: var(--spacing-md); border-radius: var(--radius-md); border: 1px solid var(--color-hairline);">
            <div>
              <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 8px;">
                <input type="checkbox" id="ep-form-lat-enabled" ${ep.latency.enabled ? 'checked' : ''}>
                <label for="ep-form-lat-enabled" style="font-weight: 600; font-size: 13px; color: var(--color-ink);">⏱ Simulated Latency</label>
              </div>
              <input type="number" class="form-input" id="ep-form-lat-fixed" placeholder="Fixed Ms (e.g. 200)" value="${ep.latency.fixedMs || 150}">
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
        </div>

      </form>
    `;

    const actions = `
      <button type="button" class="btn btn-secondary" onclick="app.closeModal()">Cancel</button>
      <button type="submit" form="endpoint-form" class="btn btn-primary">Save Mock Endpoint</button>
    `;

    app.openModal(id ? 'Edit Mock Endpoint' : 'Create New Mock Endpoint', html, actions);
    this.renderScenariosList();
  },

  switchModalTab(tabKey) {
    this.currentModalTab = tabKey;
    document.querySelectorAll('.modal-subtab').forEach(t => {
      if (t.getAttribute('data-subtab') === tabKey) {
        t.classList.add('active');
      } else {
        t.classList.remove('active');
      }
    });

    ['general', 'response', 'scenarios', 'auth'].forEach(k => {
      const el = document.getElementById(`subtab-content-${k}`);
      if (el) {
        el.style.display = k === tabKey ? 'block' : 'none';
      }
    });
  },

  renderScenariosList() {
    const container = document.getElementById('scenarios-container');
    const badge = document.getElementById('ep-modal-scenario-count');
    if (badge) badge.textContent = this.currentEditingScenarios.length;
    if (!container) return;

    if (this.currentEditingScenarios.length === 0) {
      container.innerHTML = `
        <div style="text-align: center; padding: 36px 16px; background: var(--color-surface-2); border: 1px dashed var(--color-hairline-strong); border-radius: var(--radius-md);">
          <div style="font-size: 24px; margin-bottom: 8px;">🎯</div>
          <div style="font-weight: 500; font-size: 13px; color: var(--color-ink); margin-bottom: 4px;">No Conditional Scenarios Yet</div>
          <div style="font-size: 12px; color: var(--color-ink-subtle); max-width: 440px; margin: 0 auto 16px auto;">
            Create branch rules to simulate error codes (401, 404), validation failures (400, 422), or role-based responses based on query params, headers, or body fields.
          </div>
          <button type="button" class="btn btn-secondary btn-sm" onclick="views.endpoints.addScenario()">+ Add First Scenario</button>
        </div>
      `;
      return;
    }

    container.innerHTML = this.currentEditingScenarios.map((sc, scIdx) => `
      <div class="scenario-card ${!sc.enabled ? 'disabled' : ''}">
        <div class="scenario-header">
          <div style="display: flex; align-items: center; gap: 8px; flex: 1;">
            <span class="scenario-priority">#${scIdx + 1}</span>
            <input type="checkbox" ${sc.enabled ? 'checked' : ''} onchange="views.endpoints.updateScenarioField(${scIdx}, 'enabled', this.checked)" title="Enable / Disable Rule">
            <input type="text" class="form-input" style="font-weight: 600; padding: 4px 8px; font-size: 13px;" value="${escapeHTML(sc.name)}" placeholder="Scenario Name (e.g. Invalid Password 401)" oninput="views.endpoints.updateScenarioField(${scIdx}, 'name', this.value)">
          </div>

          <div style="display: flex; align-items: center; gap: 6px;">
            <select class="form-select" style="padding: 3px 8px; font-size: 11px; width: 150px;" onchange="views.endpoints.updateScenarioField(${scIdx}, 'matchMode', this.value)">
              <option value="all" ${sc.matchMode === 'all' ? 'selected' : ''}>Match ALL (AND)</option>
              <option value="any" ${sc.matchMode === 'any' ? 'selected' : ''}>Match ANY (OR)</option>
            </select>

            <button type="button" class="btn btn-secondary btn-sm" style="padding: 2px 6px;" onclick="views.endpoints.moveScenario(${scIdx}, -1)" ${scIdx === 0 ? 'disabled' : ''} title="Move Up (Higher Priority)">↑</button>
            <button type="button" class="btn btn-secondary btn-sm" style="padding: 2px 6px;" onclick="views.endpoints.moveScenario(${scIdx}, 1)" ${scIdx === this.currentEditingScenarios.length - 1 ? 'disabled' : ''} title="Move Down">↓</button>
            <button type="button" class="btn btn-danger btn-sm" style="padding: 2px 6px;" onclick="views.endpoints.removeScenario(${scIdx})" title="Delete Scenario">✕</button>
          </div>
        </div>

        <!-- Conditions Table -->
        <div style="margin-bottom: var(--spacing-sm);">
          <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px;">
            <label class="form-label" style="font-size: 11px; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 0;">Match Conditions</label>
            <button type="button" class="btn btn-secondary btn-sm" style="padding: 2px 6px; font-size: 11px;" onclick="views.endpoints.addCondition(${scIdx})">+ Condition</button>
          </div>

          ${sc.conditions && sc.conditions.length > 0 ? sc.conditions.map((cond, condIdx) => `
            <div class="condition-row">
              <select class="form-select" style="font-size: 12px; padding: 4px 6px;" onchange="views.endpoints.updateCondition(${scIdx}, ${condIdx}, 'source', this.value)">
                <option value="query" ${cond.source === 'query' ? 'selected' : ''}>Query Param</option>
                <option value="header" ${cond.source === 'header' ? 'selected' : ''}>Header</option>
                <option value="param" ${cond.source === 'param' ? 'selected' : ''}>Path Param</option>
                <option value="body" ${cond.source === 'body' ? 'selected' : ''}>JSON Body</option>
              </select>

              <input type="text" class="form-input" style="font-size: 12px; padding: 4px 6px;" placeholder="Property (e.g. role)" value="${escapeHTML(cond.property || '')}" oninput="views.endpoints.updateCondition(${scIdx}, ${condIdx}, 'property', this.value)">

              <select class="form-select" style="font-size: 12px; padding: 4px 6px;" onchange="views.endpoints.updateCondition(${scIdx}, ${condIdx}, 'operator', this.value)">
                <option value="equals" ${cond.operator === 'equals' ? 'selected' : ''}>== Equals</option>
                <option value="not_equals" ${cond.operator === 'not_equals' ? 'selected' : ''}>!= Not Equals</option>
                <option value="contains" ${cond.operator === 'contains' ? 'selected' : ''}>Contains</option>
                <option value="regex" ${cond.operator === 'regex' ? 'selected' : ''}>Regex Match</option>
                <option value="gt" ${cond.operator === 'gt' ? 'selected' : ''}>&gt; Greater Than</option>
                <option value="gte" ${cond.operator === 'gte' ? 'selected' : ''}>&gt;= Greater Equal</option>
                <option value="lt" ${cond.operator === 'lt' ? 'selected' : ''}>&lt; Less Than</option>
                <option value="lte" ${cond.operator === 'lte' ? 'selected' : ''}>&lt;= Less Equal</option>
                <option value="is_empty" ${cond.operator === 'is_empty' ? 'selected' : ''}>Is Empty / Null</option>
                <option value="is_not_empty" ${cond.operator === 'is_not_empty' ? 'selected' : ''}>Is Not Empty</option>
              </select>

              <input type="text" class="form-input" style="font-size: 12px; padding: 4px 6px;" placeholder="Target Value..." value="${escapeHTML(cond.value || '')}" oninput="views.endpoints.updateCondition(${scIdx}, ${condIdx}, 'value', this.value)">

              <button type="button" class="btn-remove" onclick="views.endpoints.removeCondition(${scIdx}, ${condIdx})" title="Remove condition">✕</button>
            </div>
          `).join('') : `
            <div style="font-size: 12px; color: var(--color-ink-subtle); padding: 4px 0;">No conditions added. Click '+ Condition' to set a rule.</div>
          `}
        </div>

        <!-- Scenario Response Override -->
        <div style="background: var(--color-surface-3); padding: var(--spacing-sm); border-radius: var(--radius-sm); border: 1px solid var(--color-hairline);">
          <div style="display: flex; gap: var(--spacing-sm); align-items: center; margin-bottom: 6px;">
            <div style="width: 130px;">
              <label class="form-label" style="font-size: 11px; margin-bottom: 2px;">Status Code</label>
              <input type="number" class="form-input" style="padding: 4px 6px; font-size: 12px;" value="${sc.response.statusCode || 200}" oninput="views.endpoints.updateScenarioResponse(${scIdx}, 'statusCode', parseInt(this.value, 10))">
            </div>
            <div style="flex: 1;">
              <label class="form-label" style="font-size: 11px; margin-bottom: 2px;">Scenario Response Body (JSON or Template)</label>
            </div>
          </div>
          <textarea class="form-textarea" style="height: 80px; font-size: 11px; font-family: var(--font-mono);" oninput="views.endpoints.updateScenarioResponse(${scIdx}, 'body', this.value)">${escapeHTML(sc.response.body || '')}</textarea>
        </div>
      </div>
    `).join('');
  },

  addScenario() {
    this.currentEditingScenarios.push({
      id: 'sc_' + Date.now(),
      name: `Scenario #${this.currentEditingScenarios.length + 1}`,
      enabled: true,
      priority: this.currentEditingScenarios.length + 1,
      matchMode: 'all',
      conditions: [
        { source: 'query', property: '', operator: 'equals', value: '' }
      ],
      response: {
        statusCode: 400,
        contentType: 'application/json',
        headers: { 'Content-Type': 'application/json' },
        body: '{\n  "error": "Conditional scenario error",\n  "code": "BAD_REQUEST"\n}'
      }
    });
    this.renderScenariosList();
  },

  removeScenario(idx) {
    this.currentEditingScenarios.splice(idx, 1);
    this.renderScenariosList();
  },

  moveScenario(idx, delta) {
    const targetIdx = idx + delta;
    if (targetIdx < 0 || targetIdx >= this.currentEditingScenarios.length) return;
    const temp = this.currentEditingScenarios[idx];
    this.currentEditingScenarios[idx] = this.currentEditingScenarios[targetIdx];
    this.currentEditingScenarios[targetIdx] = temp;
    this.currentEditingScenarios.forEach((sc, i) => sc.priority = i + 1);
    this.renderScenariosList();
  },

  updateScenarioField(scIdx, field, val) {
    if (this.currentEditingScenarios[scIdx]) {
      this.currentEditingScenarios[scIdx][field] = val;
    }
  },

  updateScenarioResponse(scIdx, field, val) {
    if (this.currentEditingScenarios[scIdx] && this.currentEditingScenarios[scIdx].response) {
      this.currentEditingScenarios[scIdx].response[field] = val;
    }
  },

  addCondition(scIdx) {
    if (this.currentEditingScenarios[scIdx]) {
      if (!this.currentEditingScenarios[scIdx].conditions) {
        this.currentEditingScenarios[scIdx].conditions = [];
      }
      this.currentEditingScenarios[scIdx].conditions.push({
        source: 'query',
        property: '',
        operator: 'equals',
        value: ''
      });
      this.renderScenariosList();
    }
  },

  removeCondition(scIdx, condIdx) {
    if (this.currentEditingScenarios[scIdx] && this.currentEditingScenarios[scIdx].conditions) {
      this.currentEditingScenarios[scIdx].conditions.splice(condIdx, 1);
      this.renderScenariosList();
    }
  },

  updateCondition(scIdx, condIdx, field, val) {
    if (this.currentEditingScenarios[scIdx] && this.currentEditingScenarios[scIdx].conditions && this.currentEditingScenarios[scIdx].conditions[condIdx]) {
      this.currentEditingScenarios[scIdx].conditions[condIdx][field] = val;
    }
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
    const description = document.getElementById('ep-form-description') ? document.getElementById('ep-form-description').value.trim() : '';
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
      description,
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
      scenarios: this.currentEditingScenarios,
    };

    try {
      if (id) {
        await api.updateEndpoint(id, ep);
        app.showToast('Mock endpoint updated with scenarios', 'success');
      } else {
        await api.createEndpoint(ep);
        app.showToast('Mock endpoint created with scenarios', 'success');
      }
      app.closeModal();
      await this.load();
    } catch (err) {
      app.showToast('Save failed: ' + err.message, 'error');
    }
  }
};
