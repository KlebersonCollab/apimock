// OpenAPI & Workspace Backup View Module
views = window.views || {};

views.openapi = {
  async render(container) {
    container.innerHTML = `
      <div style="display: grid; grid-template-columns: 1fr 1fr; gap: var(--spacing-lg);">
        <!-- OpenAPI 3.0 Import / Export Card -->
        <div class="card">
          <div class="card-header">
            <div>
              <div class="card-title">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
                OpenAPI 3.0 Specification
              </div>
              <div class="card-subtitle">Import API definitions or export your mock endpoints as OpenAPI 3.0 JSON</div>
            </div>
            <button class="btn btn-secondary btn-sm" onclick="views.openapi.exportOpenAPI()">Export OpenAPI</button>
          </div>

          <form id="openapi-import-form" onsubmit="views.openapi.importSpec(event)">
            <div class="form-group">
              <label class="form-label">Paste OpenAPI 3.0 JSON Specification</label>
              <textarea class="form-textarea" id="openapi-import-text" placeholder='{\n  "openapi": "3.0.0",\n  "paths": { ... }\n}' style="height: 180px;" required></textarea>
            </div>
            <div style="display: flex; justify-content: space-between; align-items: center;">
              <input type="file" id="openapi-file-input" accept=".json,.yaml,.yml" onchange="views.openapi.onFileUpload(event)" style="font-size: 12px; color: var(--color-ink-muted);">
              <button type="submit" class="btn btn-primary">Import Endpoints</button>
            </div>
          </form>
        </div>

        <!-- Full Workspace Backup & Restore Card -->
        <div class="card">
          <div class="card-header">
            <div>
              <div class="card-title">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                Workspace Backup & Migration
              </div>
              <div class="card-subtitle">Export or restore all mock routes, collections, auth configs, and latency settings</div>
            </div>
            <button class="btn btn-secondary btn-sm" onclick="views.openapi.exportWorkspace()">Download Backup</button>
          </div>

          <form id="workspace-import-form" onsubmit="views.openapi.importWorkspace(event)">
            <div class="form-group">
              <label class="form-label">Paste Workspace JSON Backup</label>
              <textarea class="form-textarea" id="workspace-import-text" placeholder='{\n  "name": "MockForge Workspace",\n  "endpoints": [ ... ],\n  "collections": [ ... ]\n}' style="height: 180px;" required></textarea>
            </div>
            <div style="display: flex; justify-content: space-between; align-items: center;">
              <input type="file" id="workspace-file-input" accept=".json" onchange="views.openapi.onWorkspaceUpload(event)" style="font-size: 12px; color: var(--color-ink-muted);">
              <button type="submit" class="btn btn-primary">Restore Workspace</button>
            </div>
          </form>
        </div>
      </div>
    `;
  },

  async importSpec(event) {
    event.preventDefault();
    const raw = document.getElementById('openapi-import-text').value.trim();
    try {
      const res = await api.importOpenAPI(raw);
      app.showToast(`Imported ${res.importedCount} endpoints from OpenAPI specification`, 'success');
      document.getElementById('openapi-import-text').value = '';
      app.navigate('endpoints');
    } catch (err) {
      app.showToast('Import failed: ' + err.message, 'error');
    }
  },

  onFileUpload(event) {
    const file = event.target.files[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (e) => {
      document.getElementById('openapi-import-text').value = e.target.result;
    };
    reader.readAsText(file);
  },

  async exportOpenAPI() {
    try {
      const spec = await api.exportOpenAPI();
      const blob = new Blob([JSON.stringify(spec, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'mockforge-openapi.json';
      a.click();
      URL.revokeObjectURL(url);
      app.showToast('OpenAPI 3.0 specification exported', 'success');
    } catch (err) {
      app.showToast('Export failed: ' + err.message, 'error');
    }
  },

  async exportWorkspace() {
    try {
      const ws = await api.exportWorkspace();
      const blob = new Blob([JSON.stringify(ws, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'mockforge-workspace.json';
      a.click();
      URL.revokeObjectURL(url);
      app.showToast('Workspace backup exported', 'success');
    } catch (err) {
      app.showToast('Export failed: ' + err.message, 'error');
    }
  },

  async importWorkspace(event) {
    event.preventDefault();
    const raw = document.getElementById('workspace-import-text').value.trim();
    try {
      const res = await api.importWorkspace(raw);
      app.showToast(`Workspace restored with ${res.endpoints} endpoints and ${res.collections} collections`, 'success');
      document.getElementById('workspace-import-text').value = '';
      app.navigate('endpoints');
    } catch (err) {
      app.showToast('Workspace restore failed: ' + err.message, 'error');
    }
  },

  onWorkspaceUpload(event) {
    const file = event.target.files[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (e) => {
      document.getElementById('workspace-import-text').value = e.target.result;
    };
    reader.readAsText(file);
  }
};
