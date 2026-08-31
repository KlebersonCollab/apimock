// Dashboard View Module
views = window.views || {};

views.dashboard = {
  async render(container) {
    container.innerHTML = `
      <div class="metrics-grid">
        <div class="metric-card">
          <div class="metric-label">Active Mock Endpoints</div>
          <div class="metric-value" id="dash-endpoints-count">-</div>
          <div class="metric-sub">REST routes configured</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">Stateful Collections</div>
          <div class="metric-value" id="dash-collections-count">-</div>
          <div class="metric-sub">Auto-CRUD resources</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">Requests Intercepted</div>
          <div class="metric-value" id="dash-requests-count">-</div>
          <div class="metric-sub" id="dash-success-rate">100% success rate</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">Avg Simulated Latency</div>
          <div class="metric-value" id="dash-latency-avg">-</div>
          <div class="metric-sub">Real-time telemetry</div>
        </div>
      </div>

      <div style="display: grid; grid-template-columns: 2fr 1fr; gap: var(--spacing-lg);">
        <div class="card">
          <div class="card-header">
            <div>
              <div class="card-title">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/></svg>
                Active Mock Endpoints
              </div>
              <div class="card-subtitle">Frequently accessed simulated endpoints</div>
            </div>
            <button class="btn btn-secondary btn-sm" onclick="app.navigate('endpoints')">View All</button>
          </div>
          <div id="dash-endpoints-table">Loading...</div>
        </div>

        <div class="card">
          <div class="card-header">
            <div>
              <div class="card-title">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>
                Recent Live Traffic
              </div>
              <div class="card-subtitle">Inbound requests stream</div>
            </div>
            <button class="btn btn-secondary btn-sm" onclick="app.navigate('traffic')">Open Stream</button>
          </div>
          <div id="dash-traffic-stream" style="display: flex; flex-direction: column; gap: 8px;">
            <div style="color: var(--color-ink-subtle); font-size: 12px;">Waiting for incoming traffic...</div>
          </div>
        </div>
      </div>
    `;

    this.update();
  },

  async update() {
    try {
      const [metrics, endpoints, traffic] = await Promise.all([
        api.getMetrics(),
        api.getEndpoints(),
        api.getTraffic(10)
      ]);

      const epCountEl = document.getElementById('dash-endpoints-count');
      const colCountEl = document.getElementById('dash-collections-count');
      const reqCountEl = document.getElementById('dash-requests-count');
      const succRateEl = document.getElementById('dash-success-rate');
      const latAvgEl = document.getElementById('dash-latency-avg');

      if (epCountEl) epCountEl.textContent = metrics.activeEndpoints;
      if (colCountEl) colCountEl.textContent = metrics.activeCollections;
      if (reqCountEl) reqCountEl.textContent = metrics.trafficStats.totalRequests;
      
      if (succRateEl && metrics.trafficStats.totalRequests > 0) {
        const rate = ((metrics.trafficStats.successRequests / metrics.trafficStats.totalRequests) * 100).toFixed(1);
        succRateEl.textContent = `${rate}% success (${metrics.trafficStats.errorRequests} errors)`;
      }

      if (latAvgEl) {
        latAvgEl.textContent = `${Math.round(metrics.trafficStats.avgSimulatedDelay)} ms`;
      }

      // Render Endpoints summary
      const epTableEl = document.getElementById('dash-endpoints-table');
      if (epTableEl) {
        if (!endpoints || endpoints.length === 0) {
          epTableEl.innerHTML = `<div style="color: var(--color-ink-subtle); padding: 12px 0;">No mock endpoints configured yet.</div>`;
        } else {
          epTableEl.innerHTML = `
            <div class="table-container">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>Method</th>
                    <th>Route Path</th>
                    <th>Status</th>
                    <th>Latency</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  ${endpoints.slice(0, 5).map(ep => `
                    <tr>
                      <td><span class="badge-method ${ep.method.toLowerCase()}">${ep.method}</span></td>
                      <td class="route-cell">${escapeHTML(ep.path)}</td>
                      <td><span class="badge-status s${String(ep.response.statusCode)[0]}xx">${ep.response.statusCode}</span></td>
                      <td style="color: var(--color-ink-muted); font-size: 12px;">
                        ${ep.latency.enabled ? `${ep.latency.mode === 'random' ? `${ep.latency.minMs}-${ep.latency.maxMs}ms` : `${ep.latency.fixedMs}ms`}` : 'None'}
                      </td>
                      <td>
                        <button class="btn btn-secondary btn-sm" onclick="views.tester.testEndpoint('${ep.method}', '${ep.path}')">Test</button>
                      </td>
                    </tr>
                  `).join('')}
                </tbody>
              </table>
            </div>
          `;
        }
      }

      // Render Traffic stream summary
      const trafficStreamEl = document.getElementById('dash-traffic-stream');
      if (trafficStreamEl && traffic) {
        if (traffic.length === 0) {
          trafficStreamEl.innerHTML = `<div style="color: var(--color-ink-subtle); font-size: 12px;">No requests recorded yet. Make a request to see live telemetry.</div>`;
        } else {
          trafficStreamEl.innerHTML = traffic.slice(0, 6).map(t => `
            <div style="display: flex; align-items: center; justify-content: space-between; padding: 8px 12px; background: var(--color-surface-2); border-radius: var(--radius-sm); font-size: 12px;">
              <div style="display: flex; align-items: center; gap: 8px;">
                <span class="badge-method ${t.method.toLowerCase()}" style="min-width: 44px;">${t.method}</span>
                <span style="font-family: var(--font-mono); color: var(--color-ink); max-width: 140px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">${escapeHTML(t.path)}</span>
              </div>
              <div style="display: flex; align-items: center; gap: 6px;">
                <span class="badge-status s${String(t.responseStatus)[0]}xx">${t.responseStatus}</span>
                <span style="color: var(--color-ink-tertiary); font-size: 11px;">${t.durationMs}ms</span>
              </div>
            </div>
          `).join('');
        }
      }

    } catch (e) {
      console.error('Error updating dashboard', e);
    }
  }
};
