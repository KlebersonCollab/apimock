// MockForge Core SPA Controller & Router
function escapeHTML(str) {
  if (!str) return '';
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

const app = {
  currentView: 'dashboard',
  sseSource: null,

  init() {
    this.setupNavigation();
    this.connectSSE();
    this.navigate('dashboard');
  },

  setupNavigation() {
    window.addEventListener('popstate', (e) => {
      if (e.state && e.state.view) {
        this.renderView(e.state.view, false);
      }
    });
  },

  navigate(viewName, pushState = true) {
    this.currentView = viewName;
    if (pushState) {
      history.pushState({ view: viewName }, '', `#${viewName}`);
    }

    // Update active tab UI
    document.querySelectorAll('.nav-tab').forEach(tab => {
      if (tab.getAttribute('data-view') === viewName) {
        tab.classList.add('active');
      } else {
        tab.classList.remove('active');
      }
    });

    this.renderView(viewName);
  },

  async renderView(viewName) {
    const container = document.getElementById('main-content');
    if (!container) return;

    const view = views[viewName];
    if (view && typeof view.render === 'function') {
      await view.render(container);
    } else {
      container.innerHTML = `<div class="card"><div class="card-title">View '${viewName}' not found</div></div>`;
    }
  },

  connectSSE() {
    if (this.sseSource) {
      this.sseSource.close();
    }

    const sseStatus = document.getElementById('sse-status');

    try {
      this.sseSource = new EventSource('/api/admin/traffic/stream');

      this.sseSource.addEventListener('connected', () => {
        if (sseStatus) sseStatus.textContent = 'Live Connected';
      });

      this.sseSource.addEventListener('traffic', (e) => {
        try {
          const entry = JSON.parse(e.data);
          if (views.traffic && typeof views.traffic.handleLiveEntry === 'function') {
            views.traffic.handleLiveEntry(entry);
          }
        } catch (err) {
          console.error('SSE JSON error', err);
        }
      });

      this.sseSource.onerror = () => {
        if (sseStatus) sseStatus.textContent = 'Reconnecting...';
      };
    } catch (e) {
      console.warn('SSE connection failed', e);
    }
  },

  openModal(title, bodyHTML, footerActionsHTML = '') {
    const backdrop = document.getElementById('modal-container');
    const dialog = document.getElementById('modal-dialog-content');
    if (!backdrop || !dialog) return;

    dialog.innerHTML = `
      <div class="modal-header">
        <div class="card-title" style="font-size: 16px;">${escapeHTML(title)}</div>
        <button class="btn btn-secondary btn-sm" onclick="app.closeModal()" style="padding: 4px 8px;">✕</button>
      </div>
      <div class="modal-body">
        ${bodyHTML}
      </div>
      ${footerActionsHTML ? `
        <div class="modal-footer">
          ${footerActionsHTML}
        </div>
      ` : ''}
    `;

    backdrop.classList.add('show');
  },

  closeModal() {
    const backdrop = document.getElementById('modal-container');
    if (backdrop) backdrop.classList.remove('show');
  },

  showToast(message, type = 'info') {
    const container = document.getElementById('toast-container');
    if (!container) return;

    const toast = document.createElement('div');
    toast.className = 'toast';
    let icon = 'ℹ️';
    if (type === 'success') icon = '✅';
    if (type === 'error') icon = '❌';

    toast.innerHTML = `<span>${icon}</span> <span>${escapeHTML(message)}</span>`;
    container.appendChild(toast);

    setTimeout(() => {
      toast.style.opacity = '0';
      toast.style.transform = 'translateY(10px)';
      toast.style.transition = 'all 0.2s ease';
      setTimeout(() => toast.remove(), 200);
    }, 3200);
  },

  async resetDemoData() {
    if (!confirm('Reset all endpoints and collections to default demo seeds?')) return;
    try {
      await api.resetDemo();
      this.showToast('Demo data restored successfully', 'success');
      this.navigate(this.currentView);
    } catch (err) {
      this.showToast('Reset failed: ' + err.message, 'error');
    }
  }
};

document.addEventListener('DOMContentLoaded', () => {
  app.init();
});
