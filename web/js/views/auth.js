// Auth Studio View Module
views = window.views || {};

views.auth = {
  async render(container) {
    container.innerHTML = `
      <div style="display: grid; grid-template-columns: 1fr 1fr; gap: var(--spacing-lg);">
        <!-- JWT Token Generator Card -->
        <div class="card">
          <div class="card-header">
            <div>
              <div class="card-title">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
                Mock JWT Generator
              </div>
              <div class="card-subtitle">Generate realistic HMAC-SHA256 tokens for frontend auth testing</div>
            </div>
          </div>

          <form id="jwt-gen-form" onsubmit="views.auth.generateToken(event)">
            <div class="form-group">
              <label class="form-label">Secret Key (HMAC-SHA256)</label>
              <input type="text" class="form-input" id="jwt-secret-input" value="mockforge-super-token-2026" required>
            </div>

            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: var(--spacing-sm);">
              <div class="form-group">
                <label class="form-label">Subject (User ID / sub)</label>
                <input type="text" class="form-input" id="jwt-sub-input" value="usr_987654">
              </div>
              <div class="form-group">
                <label class="form-label">Validity Duration</label>
                <select class="form-select" id="jwt-duration-input">
                  <option value="60">1 Hour</option>
                  <option value="1440" selected>24 Hours</option>
                  <option value="10080">7 Days</option>
                  <option value="43200">30 Days</option>
                </select>
              </div>
            </div>

            <div class="form-group">
              <label class="form-label">Custom Payload Claims (JSON)</label>
              <textarea class="form-textarea" id="jwt-claims-input" style="height: 110px;">{
  "name": "Alice Developer",
  "email": "alice@linear.app",
  "role": "admin",
  "permissions": ["users:read", "users:write", "analytics:view"]
}</textarea>
            </div>

            <button type="submit" class="btn btn-primary" style="width: 100%;">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>
              Generate Signed Token
            </button>
          </form>

          <!-- Generated Output Box -->
          <div id="jwt-generated-result" style="margin-top: var(--spacing-md); display: none;">
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px;">
              <label class="form-label" style="margin-bottom: 0; color: var(--color-semantic-success); font-weight: 600;">Signed Bearer Token</label>
              <div style="display: flex; gap: 4px;">
                <button class="btn btn-secondary btn-sm" onclick="views.auth.copyGeneratedToken()">Copy</button>
                <button class="btn btn-secondary btn-sm" onclick="views.auth.useInConsole()">Use in Console</button>
              </div>
            </div>
            <div class="code-box" id="jwt-token-box" style="word-break: break-all; font-size: 11px; color: var(--color-primary-hover);"></div>
          </div>
        </div>

        <!-- Token Inspector / Decoder Card -->
        <div class="card">
          <div class="card-header">
            <div>
              <div class="card-title">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
                JWT Token Inspector
              </div>
              <div class="card-subtitle">Paste and inspect payload claims, expiration, and header metadata</div>
            </div>
          </div>

          <div class="form-group">
            <label class="form-label">Paste Token to Inspect</label>
            <textarea class="form-textarea" id="jwt-inspect-input" placeholder="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." style="height: 90px;" oninput="views.auth.inspectToken()"></textarea>
          </div>

          <div id="jwt-inspect-output">
            <div style="color: var(--color-ink-subtle); font-size: 12px; text-align: center; padding: 40px 0;">Paste a JWT token above to view decoded claims</div>
          </div>
        </div>
      </div>
    `;
  },

  async generateToken(event) {
    event.preventDefault();
    const secret = document.getElementById('jwt-secret-input').value.trim();
    const sub = document.getElementById('jwt-sub-input').value.trim();
    const duration = parseInt(document.getElementById('jwt-duration-input').value, 10) || 1440;
    const claimsRaw = document.getElementById('jwt-claims-input').value.trim();

    let claims = {};
    if (claimsRaw) {
      try {
        claims = JSON.parse(claimsRaw);
      } catch (err) {
        app.showToast('Invalid Claims JSON: ' + err.message, 'error');
        return;
      }
    }
    if (sub) {
      claims.sub = sub;
    }

    try {
      const res = await api.generateAuthToken({
        secret,
        durationMinutes: duration,
        claims,
      });

      this.latestGeneratedToken = res.token;
      const resultContainer = document.getElementById('jwt-generated-result');
      const tokenBox = document.getElementById('jwt-token-box');
      if (resultContainer && tokenBox) {
        tokenBox.textContent = res.token;
        resultContainer.style.display = 'block';
      }

      // Auto inspect
      const inspectInput = document.getElementById('jwt-inspect-input');
      if (inspectInput) {
        inspectInput.value = res.token;
        this.inspectToken();
      }

      app.showToast('JWT Token generated successfully', 'success');
    } catch (err) {
      app.showToast('Generation failed: ' + err.message, 'error');
    }
  },

  copyGeneratedToken() {
    if (this.latestGeneratedToken) {
      navigator.clipboard.writeText(this.latestGeneratedToken);
      app.showToast('Token copied to clipboard', 'success');
    }
  },

  useInConsole() {
    if (this.latestGeneratedToken) {
      views.tester.savedAuthHeader = `Bearer ${this.latestGeneratedToken}`;
      app.navigate('tester');
      app.showToast('Bearer token set in Test Console', 'success');
    }
  },

  inspectToken() {
    const input = document.getElementById('jwt-inspect-input');
    const output = document.getElementById('jwt-inspect-output');
    if (!input || !output) return;

    const token = input.value.trim();
    if (!token) {
      output.innerHTML = `<div style="color: var(--color-ink-subtle); font-size: 12px; text-align: center; padding: 40px 0;">Paste a JWT token above to view decoded claims</div>`;
      return;
    }

    const parts = token.split('.');
    if (parts.length !== 3) {
      output.innerHTML = `<div style="color: var(--color-semantic-danger); font-size: 12px;">Invalid JWT: must contain 3 dot-separated segments.</div>`;
      return;
    }

    try {
      const header = JSON.parse(atob(parts[0].replace(/-/g, '+').replace(/_/g, '/')));
      const payload = JSON.parse(atob(parts[1].replace(/-/g, '+').replace(/_/g, '/')));

      let expStatus = 'No expiration claim (exp)';
      if (payload.exp) {
        const expDate = new Date(payload.exp * 1000);
        const isExpired = Date.now() > expDate.getTime();
        expStatus = isExpired ? `<span style="color: var(--color-semantic-danger);">Expired on ${expDate.toLocaleString()}</span>` : `<span style="color: var(--color-semantic-success);">Valid until ${expDate.toLocaleString()}</span>`;
      }

      output.innerHTML = `
        <div style="display: flex; flex-direction: column; gap: var(--spacing-sm);">
          <div style="font-size: 12px; background: var(--color-surface-2); padding: 8px 12px; border-radius: var(--radius-md); border: 1px solid var(--color-hairline);">
            Status: ${expStatus}
          </div>
          <div>
            <div style="font-size: 11px; font-weight: 600; text-transform: uppercase; color: var(--color-ink-subtle); margin-bottom: 4px;">Header (Algorithm & Type)</div>
            <div class="code-box" style="font-size: 11px;">${escapeHTML(JSON.stringify(header, null, 2))}</div>
          </div>
          <div>
            <div style="font-size: 11px; font-weight: 600; text-transform: uppercase; color: var(--color-ink-subtle); margin-bottom: 4px;">Payload (Claims & User Data)</div>
            <div class="code-box" style="font-size: 11px;">${escapeHTML(JSON.stringify(payload, null, 2))}</div>
          </div>
        </div>
      `;
    } catch (e) {
      output.innerHTML = `<div style="color: var(--color-semantic-danger); font-size: 12px;">Failed to decode JWT segments: ${escapeHTML(e.message)}</div>`;
    }
  }
};
