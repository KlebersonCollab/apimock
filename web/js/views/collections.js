// Collections View Module
views = window.views || {};

views.collections = {
  collections: [],
  selectedCollection: null,

  async render(container) {
    container.innerHTML = `
      <div style="display: grid; grid-template-columns: 280px 1fr; gap: var(--spacing-lg);">
        <!-- Collections Sidebar List -->
        <div class="card" style="padding: var(--spacing-md);">
          <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-md);">
            <div class="card-title" style="font-size: 15px;">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg>
              Resources
            </div>
            <button class="btn btn-primary btn-sm" onclick="views.collections.openCreateCollectionModal()">+ New</button>
          </div>
          <div id="collections-list" style="display: flex; flex-direction: column; gap: 4px;">Loading...</div>
        </div>

        <!-- Selected Collection Data & CRUD View -->
        <div class="card" id="collection-detail-panel">
          <div style="color: var(--color-ink-subtle); text-align: center; padding: 60px 0;">Select a collection to view items</div>
        </div>
      </div>
    `;

    await this.load();
  },

  async load() {
    try {
      this.collections = await api.getCollections();
      const badgeEl = document.getElementById('badge-collection-count');
      if (badgeEl) badgeEl.textContent = this.collections.length;

      this.renderList();
      if (this.collections.length > 0 && !this.selectedCollection) {
        this.selectCollection(this.collections[0].name);
      } else if (this.selectedCollection) {
        this.selectCollection(this.selectedCollection);
      }
    } catch (err) {
      app.showToast('Failed to load collections: ' + err.message, 'error');
    }
  },

  renderList() {
    const listEl = document.getElementById('collections-list');
    if (!listEl) return;

    if (this.collections.length === 0) {
      listEl.innerHTML = `<div style="font-size: 12px; color: var(--color-ink-subtle); padding: 8px;">No collections created yet.</div>`;
      return;
    }

    listEl.innerHTML = this.collections.map(col => `
      <div onclick="views.collections.selectCollection('${col.name}')" 
           style="padding: 8px 12px; border-radius: var(--radius-md); cursor: pointer; display: flex; align-items: center; justify-content: space-between; transition: all 0.15s ease; ${this.selectedCollection === col.name ? 'background: var(--color-surface-3); border: 1px solid var(--color-hairline-strong);' : 'background: transparent;'}">
        <div style="font-weight: 500; font-size: 13px; color: ${this.selectedCollection === col.name ? 'var(--color-primary-hover)' : 'var(--color-ink)'};">
          /${escapeHTML(col.name)}
        </div>
        <span class="tab-badge">${col.items ? col.items.length : 0}</span>
      </div>
    `).join('');
  },

  selectCollection(colName) {
    this.selectedCollection = colName;
    this.renderList();
    const col = this.collections.find(c => c.name === colName);
    const detailPanel = document.getElementById('collection-detail-panel');
    if (!detailPanel || !col) return;

    const items = col.items || [];
    const keys = items.length > 0 ? Object.keys(items[0]) : ['id', 'name', 'createdAt'];

    detailPanel.innerHTML = `
      <div class="card-header" style="margin-bottom: var(--spacing-sm);">
        <div>
          <div class="card-title">
            /${escapeHTML(col.name)}
            <span class="brand-badge">${items.length} records</span>
          </div>
          <div class="card-subtitle">${escapeHTML(col.description || 'Auto-CRUD REST Resource')}</div>
        </div>
        <div style="display: flex; gap: var(--spacing-xs);">
          <button class="btn btn-secondary btn-sm" onclick="views.collections.deleteCollection('${col.id}')">Delete Resource</button>
          <button class="btn btn-primary btn-sm" onclick="views.collections.openAddItemModal('${col.name}')">+ Add Record</button>
        </div>
      </div>

      <!-- Live REST API Endpoints Strip -->
      <div style="background: var(--color-surface-2); padding: 8px 12px; border-radius: var(--radius-md); margin-bottom: var(--spacing-md); font-family: var(--font-mono); font-size: 12px; display: flex; align-items: center; justify-content: space-between; border: 1px solid var(--color-hairline);">
        <div style="color: var(--color-ink-muted);">
          REST URL: <span style="color: var(--color-primary-hover);">/api/resources/${escapeHTML(col.name)}</span>
        </div>
        <div style="display: flex; gap: 6px;">
          <button class="btn btn-secondary btn-sm" style="padding: 2px 8px; font-size: 11px;" onclick="views.tester.testEndpoint('GET', '/api/resources/${col.name}')">Test GET</button>
          <button class="btn btn-secondary btn-sm" style="padding: 2px 8px; font-size: 11px;" onclick="navigator.clipboard.writeText(location.origin + '/api/resources/${col.name}'); app.showToast('Copied URL', 'success')">Copy URL</button>
        </div>
      </div>

      <!-- Records Table -->
      ${items.length === 0 ? `
        <div style="text-align: center; padding: 40px 0; color: var(--color-ink-subtle);">No items in this collection yet. Click "+ Add Record" to insert data.</div>
      ` : `
        <div class="table-container">
          <table class="data-table">
            <thead>
              <tr>
                ${keys.map(k => `<th>${escapeHTML(k)}</th>`).join('')}
                <th style="text-align: right;">Actions</th>
              </tr>
            </thead>
            <tbody>
              ${items.map(item => `
                <tr>
                  ${keys.map(k => `
                    <td style="max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 12px;">
                      ${typeof item[k] === 'object' ? JSON.stringify(item[k]) : escapeHTML(String(item[k] !== undefined ? item[k] : ''))}
                    </td>
                  `).join('')}
                  <td style="text-align: right;">
                    <button class="btn btn-danger btn-sm" onclick="views.collections.deleteItem('${col.name}', '${item.id}')">Delete</button>
                  </td>
                </tr>
              `).join('')}
            </tbody>
          </table>
        </div>
      `}
    `;
  },

  openCreateCollectionModal() {
    const html = `
      <form id="create-collection-form" onsubmit="views.collections.saveCollection(event)">
        <div class="form-group">
          <label class="form-label">Resource Name (e.g. users, products, orders)</label>
          <input type="text" class="form-input" id="col-form-name" placeholder="products" required>
        </div>
        <div class="form-group">
          <label class="form-label">Description</label>
          <input type="text" class="form-input" id="col-form-desc" placeholder="Product catalog collection">
        </div>
        <div class="form-group">
          <label class="form-label">Initial Seed Items (JSON Array)</label>
          <textarea class="form-textarea" id="col-form-items" style="height: 140px;">[
  {
    "title": "Ergonomic Mechanical Keyboard",
    "price": 149.99,
    "category": "hardware",
    "inStock": true
  }
]</textarea>
        </div>
      </form>
    `;

    const actions = `
      <button type="button" class="btn btn-secondary" onclick="app.closeModal()">Cancel</button>
      <button type="submit" form="create-collection-form" class="btn btn-primary">Create Collection</button>
    `;

    app.openModal('Create Stateful Resource Collection', html, actions);
  },

  async saveCollection(event) {
    event.preventDefault();
    const name = document.getElementById('col-form-name').value.trim();
    const desc = document.getElementById('col-form-desc').value.trim();
    const itemsRaw = document.getElementById('col-form-items').value.trim();

    let items = [];
    if (itemsRaw) {
      try {
        items = JSON.parse(itemsRaw);
        if (!Array.isArray(items)) throw new Error('Items must be a JSON array');
      } catch (err) {
        app.showToast('Invalid JSON: ' + err.message, 'error');
        return;
      }
    }

    try {
      await api.createCollection({
        name,
        description: desc,
        items,
      });
      app.showToast('Collection created successfully', 'success');
      app.closeModal();
      this.selectedCollection = name.toLowerCase();
      await this.load();
    } catch (err) {
      app.showToast('Create failed: ' + err.message, 'error');
    }
  },

  openAddItemModal(colName) {
    const html = `
      <form id="add-item-form" onsubmit="views.collections.saveItem(event, '${colName}')">
        <div class="form-group">
          <label class="form-label">Record JSON Payload</label>
          <textarea class="form-textarea" id="item-form-json" style="height: 180px;">{
  "title": "New Sample Item",
  "status": "active",
  "tags": ["featured"]
}</textarea>
        </div>
      </form>
    `;

    const actions = `
      <button type="button" class="btn btn-secondary" onclick="app.closeModal()">Cancel</button>
      <button type="submit" form="add-item-form" class="btn btn-primary">Insert Record</button>
    `;

    app.openModal(`Add Record to /${colName}`, html, actions);
  },

  async saveItem(event, colName) {
    event.preventDefault();
    const jsonStr = document.getElementById('item-form-json').value;
    try {
      const data = JSON.parse(jsonStr);
      await fetch(`/api/resources/${colName}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      });
      app.showToast('Record added', 'success');
      app.closeModal();
      await this.load();
    } catch (err) {
      app.showToast('Failed to add record: ' + err.message, 'error');
    }
  },

  async deleteItem(colName, id) {
    if (!confirm(`Delete record #${id}?`)) return;
    try {
      await fetch(`/api/resources/${colName}/${id}`, { method: 'DELETE' });
      app.showToast('Record deleted', 'success');
      await this.load();
    } catch (err) {
      app.showToast('Delete failed: ' + err.message, 'error');
    }
  },

  async deleteCollection(id) {
    if (!confirm('Are you sure you want to delete this entire collection?')) return;
    try {
      await api.deleteCollection(id);
      app.showToast('Collection deleted', 'success');
      this.selectedCollection = null;
      await this.load();
    } catch (err) {
      app.showToast('Delete failed: ' + err.message, 'error');
    }
  }
};
