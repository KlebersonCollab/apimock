// Test Console View Module
views = window.views || {};

views.tester = {
  savedAuthHeader: '',
  targetMethod: 'GET',
  targetPath: '/api/v1/users',

  async render(container) {
    container.innerHTML = `
      <div class="card">
        <div class="card-header">
          <div>
            <div class="card-title">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg>
              Interactive API Test Console
            </div>
            <div class="card-subtitle">Test and verify simulated routes, headers, and stateful collections in real-time</div>
          </div>
          <div style="display: flex; gap: 6px; flex-wrap: wrap;">
            <button class="btn btn-secondary btn-sm" onclick="views.tester.testEndpoint('GET', '/api/v1/users')">GET Users</button>
            <button class="btn btn-secondary btn-sm" onclick="views.tester.testEndpoint('GET', '/api/v1/users?status=archived')">GET Users (Empty State)</button>
            <button class="btn btn-secondary btn-sm" onclick="views.tester.testEndpoint('POST', '/api/v1/auth/login', '{\n  \"email\": \"user@app.com\",\n  \"password\": \"123456\"\n}')">POST Login (Valid 200)</button>
            <button class="btn btn-secondary btn-sm" onclick="views.tester.testEndpoint('POST', '/api/v1/auth/login', '{\n  \"email\": \"user@app.com\",\n  \"password\": \"wrong\"\n}')">POST Login (Wrong Pass 401)</button>
            <button class="btn btn-secondary btn-sm" onclick="views.tester.testEndpoint('POST', '/api/v1/auth/login', '{\n  \"password\": \"123456\"\n}')">POST Login (No Email 400)</button>
            <button class="btn btn-secondary btn-sm" onclick="views.tester.testEndpoint('POST', '/api/v1/auth/login', '{\n  \"email\": \"admin@mockforge.io\",\n  \"password\": \"secret\"\n}')">POST Login (Admin 200)</button>
          </div>
        </div>

        <form id="tester-form" onsubmit="views.tester.sendRequest(event)">
          <!-- URL and Method Input Bar -->
          <div style="display: grid; grid-template-columns: 120px 1fr 120px; gap: var(--spacing-sm); margin-bottom: var(--spacing-md);">
            <select class="form-select" id="tester-method">
              ${['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS', 'HEAD'].map(m => `
                <option value="${m}" ${this.targetMethod === m ? 'selected' : ''}>${m}</option>
              `).join('')}
            </select>
            <input type="text" class="form-input" id="tester-path" value="${escapeHTML(this.targetPath)}" placeholder="/api/v1/resource..." required>
            <button type="submit" class="btn btn-primary" id="tester-send-btn">Send Call</button>
          </div>

          <!-- Headers & Body Tabbed Split -->
          <div style="display: grid; grid-template-columns: 1fr 1fr; gap: var(--spacing-md); margin-bottom: var(--spacing-md);">
            <div class="form-group">
              <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px;">
                <label class="form-label" style="margin-bottom: 0;">Request Headers (JSON format)</label>
                <button type="button" class="btn btn-secondary btn-sm" style="padding: 2px 8px; font-size: 11px;" onclick="views.tester.addAuthHeaderPreset()">+ Set Bearer Auth</button>
              </div>
              <textarea class="form-textarea" id="tester-headers" style="height: 120px;">{\n  "Content-Type": "application/json"\n}</textarea>
            </div>

            <div class="form-group">
              <label class="form-label">Request Body (JSON format)</label>
              <textarea class="form-textarea" id="tester-body" style="height: 120px;" placeholder="Optional request payload..."></textarea>
            </div>
          </div>
        </form>

        <!-- Response Console Output -->
        <div id="tester-response-container" style="background: var(--color-surface-2); border: 1px solid var(--color-hairline); border-radius: var(--radius-md); padding: var(--spacing-md); display: none;">
          <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-sm);">
            <div style="display: flex; align-items: center; gap: 8px;">
              <span class="badge-status" id="tester-res-status">-</span>
              <span id="tester-res-statustext" style="font-size: 12px; font-weight: 500; color: var(--color-ink);">OK</span>
              <span id="tester-res-duration" style="font-size: 11px; color: var(--color-ink-subtle); margin-left: 8px;">⏱ 0ms</span>
            </div>
            <button class="btn btn-secondary btn-sm" onclick="navigator.clipboard.writeText(document.getElementById('tester-res-body').textContent); app.showToast('Copied Response', 'success')">Copy Body</button>
          </div>
          <div class="code-box" id="tester-res-body" style="max-height: 350px; font-size: 12px;"></div>
        </div>
      </div>
    `;

    if (this.savedAuthHeader) {
      this.addAuthHeaderPreset();
    }
  },

  testEndpoint(method, path, body = '') {
    this.targetMethod = method;
    this.targetPath = path;
    this.targetBody = body;
    if (app.currentView !== 'tester') {
      app.navigate('tester');
      setTimeout(() => {
        const bEl = document.getElementById('tester-body');
        if (bEl && body) bEl.value = body;
      }, 50);
    } else {
      const mEl = document.getElementById('tester-method');
      const pEl = document.getElementById('tester-path');
      const bEl = document.getElementById('tester-body');
      if (mEl) mEl.value = method;
      if (pEl) pEl.value = path;
      if (bEl) bEl.value = body;
      this.sendRequest(new Event('submit'));
    }
  },

  addAuthHeaderPreset() {
    const headersEl = document.getElementById('tester-headers');
    if (!headersEl) return;
    try {
      let current = JSON.parse(headersEl.value || '{}');
      current['Authorization'] = this.savedAuthHeader || 'Bearer mockforge-super-token-2026';
      headersEl.value = JSON.stringify(current, null, 2);
      app.showToast('Authorization header added', 'success');
    } catch (e) {
      headersEl.value = `{\n  "Content-Type": "application/json",\n  "Authorization": "${this.savedAuthHeader || 'Bearer mockforge-super-token-2026'}"\n}`;
    }
  },

  async sendRequest(event) {
    if (event && event.preventDefault) event.preventDefault();

    const method = document.getElementById('tester-method').value;
    const path = document.getElementById('tester-path').value.trim();
    const headersRaw = document.getElementById('tester-headers').value.trim();
    const bodyRaw = document.getElementById('tester-body').value.trim();

    let headers = {};
    if (headersRaw) {
      try {
        headers = JSON.parse(headersRaw);
      } catch (err) {
        app.showToast('Invalid Headers JSON: ' + err.message, 'error');
        return;
      }
    }

    const sendBtn = document.getElementById('tester-send-btn');
    if (sendBtn) sendBtn.textContent = 'Sending...';

    const url = path.startsWith('http') ? path : location.origin + (path.startsWith('/') ? path : '/' + path);

    const res = await api.sendTestRequest({
      method,
      url,
      headers,
      body: bodyRaw || undefined,
    });

    if (sendBtn) sendBtn.textContent = 'Send Call';

    const respContainer = document.getElementById('tester-response-container');
    const statusBadge = document.getElementById('tester-res-status');
    const statusText = document.getElementById('tester-res-statustext');
    const durationEl = document.getElementById('tester-res-duration');
    const bodyEl = document.getElementById('tester-res-body');

    if (respContainer && statusBadge && statusText && durationEl && bodyEl) {
      statusBadge.textContent = res.status;
      statusBadge.className = `badge-status s${String(res.status)[0]}xx`;
      statusText.textContent = res.statusText || (res.status === 200 ? 'OK' : 'Response');
      durationEl.textContent = `⏱ ${res.duration}ms`;

      let formatted = res.rawBody;
      try {
        if (typeof res.body === 'object') {
          formatted = JSON.stringify(res.body, null, 2);
        }
      } catch (e) {}

      bodyEl.textContent = formatted;
      respContainer.style.display = 'block';
    }
  }
};
