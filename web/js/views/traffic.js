// Traffic Live Stream & Request Inspector View Module
views = window.views || {};

views.traffic = {
  trafficLogs: [],

  async render(container) {
    container.innerHTML = `
      <div class="card">
        <div class="card-header">
          <div>
            <div class="card-title">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>
              Live Traffic Telemetry & Request Inspector
            </div>
            <div class="card-subtitle">Real-time stream of all HTTP requests intercepted by MockForge</div>
          </div>
          <div style="display: flex; gap: var(--spacing-xs);">
            <button class="btn btn-secondary btn-sm" onclick="views.traffic.clearTraffic()">Clear Stream</button>
            <button class="btn btn-secondary btn-sm" onclick="views.traffic.load()">Refresh</button>
          </div>
        </div>

        <div id="traffic-table-container">Loading...</div>
      </div>
    `;

    await this.load();
  },

  async load() {
    try {
      this.trafficLogs = await api.getTraffic(100);
      this.renderTable();
    } catch (err) {
      app.showToast('Failed to load traffic logs: ' + err.message, 'error');
    }
  },

  handleLiveEntry(entry) {
    this.trafficLogs.unshift(entry);
    if (this.trafficLogs.length > 200) {
      this.trafficLogs.pop();
    }
    // Only re-render if current active view is traffic or dashboard
    if (app.currentView === 'traffic') {
      this.renderTable();
    } else if (app.currentView === 'dashboard') {
      views.dashboard.update();
    }
  },

  renderTable() {
    const container = document.getElementById('traffic-table-container');
    if (!container) return;

    if (this.trafficLogs.length === 0) {
      container.innerHTML = `
        <div style="text-align: center; padding: 60px 0; color: var(--color-ink-subtle);">
          <div style="font-size: 14px; margin-bottom: 6px;">No requests recorded yet</div>
          <div style="font-size: 12px; color: var(--color-ink-tertiary);">Send HTTP requests to your mock endpoints or test console to see live logs.</div>
        </div>
      `;
      return;
    }

    container.innerHTML = `
      <div class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th>Time</th>
              <th>Method</th>
              <th>Path & Query</th>
              <th>Status</th>
              <th>Total Duration</th>
              <th>Simulated Delay</th>
              <th>Matched Mock</th>
              <th style="text-align: right;">Action</th>
            </tr>
          </thead>
          <tbody>
            ${this.trafficLogs.map(t => {
              const timeStr = new Date(t.timestamp).toLocaleTimeString();
              return `
                <tr style="cursor: pointer;" onclick="views.traffic.openInspector('${t.id}')">
                  <td style="font-size: 11px; color: var(--color-ink-tertiary); font-family: var(--font-mono);">${timeStr}</td>
                  <td><span class="badge-method ${t.method.toLowerCase()}">${t.method}</span></td>
                  <td class="route-cell" style="max-width: 280px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">
                    ${escapeHTML(t.path)}
                  </td>
                  <td><span class="badge-status s${String(t.responseStatus)[0]}xx">${t.responseStatus}</span></td>
                  <td style="font-size: 12px; color: var(--color-ink-muted);">${t.durationMs}ms</td>
                  <td style="font-size: 12px; color: ${t.simulatedDelayMs > 0 ? 'var(--color-primary-hover)' : 'var(--color-ink-tertiary)'};">
                    ${t.simulatedDelayMs > 0 ? `+${t.simulatedDelayMs}ms` : '0ms'}
                  </td>
                  <td style="font-size: 12px; color: var(--color-ink-subtle);">
                    ${escapeHTML(t.matchedEndpoint || t.matchedCollection || '404')}
                  </td>
                  <td style="text-align: right;">
                    <button class="btn btn-secondary btn-sm" onclick="event.stopPropagation(); views.traffic.openInspector('${t.id}')">Inspect</button>
                  </td>
                </tr>
              `;
            }).join('')}
          </tbody>
        </table>
      </div>
    `;
  },

  async clearTraffic() {
    try {
      await api.clearTraffic();
      this.trafficLogs = [];
      this.renderTable();
      app.showToast('Traffic stream cleared', 'success');
    } catch (err) {
      app.showToast('Failed to clear traffic: ' + err.message, 'error');
    }
  },

  openInspector(id) {
    const entry = this.trafficLogs.find(t => t.id === id);
    if (!entry) return;

    let formattedReqBody = entry.requestBody || '';
    try {
      if (formattedReqBody) {
        formattedReqBody = JSON.stringify(JSON.parse(formattedReqBody), null, 2);
      }
    } catch (e) {}

    let formattedRespBody = entry.responseBody || '';
    try {
      if (formattedRespBody) {
        formattedRespBody = JSON.stringify(JSON.parse(formattedRespBody), null, 2);
      }
    } catch (e) {}

    const curlCmd = `curl -X ${entry.method} "${location.origin}${entry.path}"`;

    const html = `
      <div style="display: flex; flex-direction: column; gap: var(--spacing-md);">
        <!-- Summary Header Strip -->
        <div style="display: flex; align-items: center; justify-content: space-between; background: var(--color-surface-2); padding: var(--spacing-md); border-radius: var(--radius-md); border: 1px solid var(--color-hairline);">
          <div style="display: flex; align-items: center; gap: 12px;">
            <span class="badge-method ${entry.method.toLowerCase()}">${entry.method}</span>
            <span style="font-family: var(--font-mono); font-size: 14px; font-weight: 600; color: var(--color-ink);">${escapeHTML(entry.path)}</span>
          </div>
          <div style="display: flex; align-items: center; gap: 8px;">
            <span class="badge-status s${String(entry.responseStatus)[0]}xx" style="font-size: 13px; padding: 4px 8px;">${entry.responseStatus}</span>
            <span style="color: var(--color-ink-muted); font-size: 12px;">Duration: ${entry.durationMs}ms</span>
          </div>
        </div>

        <!-- Request Details -->
        <div>
          <div style="font-weight: 600; font-size: 13px; color: var(--color-ink); margin-bottom: 6px;">Request Headers</div>
          <div class="code-box" style="max-height: 120px;">${escapeHTML(JSON.stringify(entry.requestHeaders || {}, null, 2))}</div>
        </div>

        ${entry.requestBody ? `
          <div>
            <div style="font-weight: 600; font-size: 13px; color: var(--color-ink); margin-bottom: 6px;">Request Body</div>
            <div class="code-box" style="max-height: 140px;">${escapeHTML(formattedReqBody)}</div>
          </div>
        ` : ''}

        <!-- Response Details -->
        <div>
          <div style="font-weight: 600; font-size: 13px; color: var(--color-ink); margin-bottom: 6px;">Response Payload</div>
          <div class="code-box" style="max-height: 200px;">${escapeHTML(formattedRespBody)}</div>
        </div>

        <!-- Generated cURL -->
        <div>
          <div style="font-weight: 600; font-size: 13px; color: var(--color-ink); margin-bottom: 6px;">cURL Command</div>
          <div class="code-box" style="display: flex; justify-content: space-between; align-items: center;">
            <code>${escapeHTML(curlCmd)}</code>
            <button class="btn btn-secondary btn-sm" onclick="navigator.clipboard.writeText('${curlCmd}'); app.showToast('Copied cURL', 'success')">Copy</button>
          </div>
        </div>
      </div>
    `;

    const actions = `
      <button class="btn btn-secondary" onclick="app.closeModal()">Close</button>
      <button class="btn btn-primary" onclick="app.closeModal(); views.tester.testEndpoint('${entry.method}', '${entry.path}')">Replay in Test Console</button>
    `;

    app.openModal(`Request Inspector [${entry.id}]`, html, actions);
  }
};
