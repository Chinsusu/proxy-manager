class PGWManager {
  constructor() {
    this.apiBase = '/api';
    this.agentBase = '/agent';
    this.proxies = [];
    this.clients = [];
    this.mappings = [];
    this.nodes = [];
    this.loading = false;
    // sorting state (persisted)
    this.pSort = 'address'; this.pAsc = true;
    this.mSort = 'client'; this.mAsc = true;
    try {
      const sp = JSON.parse(localStorage.getItem('pgw_sort_p2') || '{}');
      if (sp && sp.k) { this.pSort = sp.k; this.pAsc = !!sp.a; }
      const sm = JSON.parse(localStorage.getItem('pgw_sort_m2') || '{}');
      if (sm && sm.k) { this.mSort = sm.k; this.mAsc = !!sm.a; }
    } catch (_) { }

    this.init();
  }

  init() {
    this.bindEvents();
    this.loadData();

    // Auto refresh every 30 seconds
    setInterval(() => this.loadData(), 30000);
  }

  bindEvents() {
    // Refresh button
    document.getElementById('btn-refresh')?.addEventListener('click', () => {
      this.loadData();
    });

    // Health check all proxies
    document.getElementById('btn-health-all')?.addEventListener('click', () => {
      this.healthCheckAll();
    });

    // Reconcile rules
    document.getElementById('btn-reconcile')?.addEventListener('click', () => {
      this.reconcileRules();
    });

    // Create proxy form
    document.getElementById('form-proxy')?.addEventListener('submit', (e) => {
      e.preventDefault();
      this.createProxy();
    });


    // Import proxies (bulk)
    document.getElementById("btn-import-proxies")?.addEventListener("click", (e) => {
      e.preventDefault();
      this.importProxies();
    });
    // Create mapping form
    document.getElementById('form-mapping')?.addEventListener('submit', (e) => {
      e.preventDefault();
      this.createMapping();
    });
  }

  async apiCall(url, options = {}) {
    try {
      const controller = new AbortController();
      const timeout = setTimeout(() => controller.abort(), 10000);
      const response = await fetch(url, {
        headers: {
          'Content-Type': 'application/json',
          ...options.headers
        },
        signal: controller.signal,
        ...options
      });
      clearTimeout(timeout);

      // Auto-redirect to login on 401 Unauthorized
      if (response.status === 401) {
        window.location.href = '/login';
        return null;
      }

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      if (response.status === 204) {
        return null;
      }

      return await response.json();
    } catch (error) {
      if (error.name === 'AbortError') {
        console.error('API call timed out:', url);
        this.showAlert('API request timed out — is the server running?', 'warning');
      } else {
        console.error('API call failed:', error);
        this.showAlert('API call failed: ' + error.message, 'danger');
      }
      throw error;
    }
  }

  async loadData() {
    if (this.loading) return;

    this.loading = true;
    let _spinnerTO = setTimeout(() => this.showLoading(true), 700);


    try {
      const [proxies, clients, mappings, nodes] = await Promise.all([
        this.apiCall(`${this.apiBase}/v1/proxies`),
        this.apiCall(`${this.apiBase}/v1/clients`),
        this.apiCall(`${this.apiBase}/v1/mappings/active`),
        this.apiCall(`${this.apiBase}/v1/nodes`).catch(() => []),
      ]);

      this.proxies = proxies || [];
      this.clients = clients || [];
      this.mappings = mappings || [];
      this.nodes = nodes || [];

      this.renderStats();
      this.renderProxies();
      this.renderProxySummary();
      this.renderMappings();
      this.renderClients();
      this.updateCounts();
      this.updateLastRefresh();
      this.checkServices();

    } catch (error) {
      console.error('Failed to load data:', error);
    } finally {
      this.loading = false;
      clearTimeout(_spinnerTO);
      this.showLoading(false);
    }
  }

  async checkServices() {
    // API health
    const apiEl = document.getElementById('api-status');
    if (apiEl) {
      try {
        const r = await fetch(`${this.apiBase}/v1/health`, { method: 'GET' });
        if (r.ok) {
          apiEl.textContent = 'Running';
          apiEl.className = 'badge text-bg-success';
        } else {
          apiEl.textContent = 'Error';
          apiEl.className = 'badge text-bg-warning';
        }
      } catch {
        apiEl.textContent = 'Down';
        apiEl.className = 'badge text-bg-danger';
      }
    }

    // Agent health
    const agentEl = document.getElementById('agent-status');
    if (agentEl) {
      try {
        const r = await fetch(`${this.agentBase}/health`, { method: 'HEAD' });
        if (r.ok) {
          agentEl.textContent = 'Running';
          agentEl.className = 'badge text-bg-success';
        } else {
          agentEl.textContent = 'Error';
          agentEl.className = 'badge text-bg-warning';
        }
      } catch {
        agentEl.textContent = 'Down';
        agentEl.className = 'badge text-bg-danger';
      }
    }

    // Forwarder status: inferred from applied mappings
    const fwdEl = document.getElementById('fwd-status');
    if (fwdEl) {
      const applied = (this.mappings || []).filter(m => m.state === 'APPLIED').length;
      if (applied > 0) {
        fwdEl.textContent = `${applied} active`;
        fwdEl.className = 'badge text-bg-success';
      } else if ((this.mappings || []).length > 0) {
        fwdEl.textContent = 'Pending';
        fwdEl.className = 'badge text-bg-warning';
      } else {
        fwdEl.textContent = 'No mappings';
        fwdEl.className = 'badge text-bg-secondary';
      }
    }
  }

  renderStats() {
    const okProxies = this.proxies.filter(p => p.status === 'OK').length;
    const activeMappings = this.mappings.filter(m => m.client?.enabled && m.proxy?.enabled).length;

    this.updateElement('stat-proxies', this.proxies.length);
    this.updateElement('stat-proxies-ok', okProxies);
    this.updateElement('stat-clients', this.clients.length);
    this.updateElement('stat-mappings', activeMappings);
  }

  renderProxies() {
    const tbody = document.getElementById('tbody-proxies');
    if (!tbody) return;
    // sort
    const key = this.pSort, asc = this.pAsc;
    const val = (p) => {
      if (key === 'id') return (p.id || '');
      if (key === 'type') return (p.type || '');
      if (key === 'address') return ((p.host || '') + ':' + p.port).toLowerCase();
      if (key === 'status') return (p.status || '');
      if (key === 'latency') return (p.latency_ms == null ? Infinity : p.latency_ms);
      if (key === 'exit') return (p.exit_ip || '');
      if (key === 'last') return (p.last_checked_at || '');
      return ((p.host || '') + ':' + p.port).toLowerCase();
    };
    const sorted = (this.proxies || []).slice().sort((a, b) => { const va = val(a), vb = val(b); if (va < vb) return asc ? -1 : 1; if (va > vb) return asc ? 1 : -1; return 0; });
    // header icons + click
    const thead = tbody.parentElement?.querySelector('thead');
    if (thead) {
      const arrow = asc ? ' ▲' : ' ▼';
      thead.innerHTML = '<tr>'
        + '<th data-k="id" class="sortable">ID' + (key === 'id' ? arrow : '') + '</th>'
        + '<th data-k="type" class="sortable">Type' + (key === 'type' ? arrow : '') + '</th>'
        + '<th data-k="address" class="sortable">Address' + (key === 'address' ? arrow : '') + '</th>'
        + '<th data-k="status" class="sortable">Status' + (key === 'status' ? arrow : '') + '</th>'
        + '<th data-k="latency" class="sortable">Latency' + (key === 'latency' ? arrow : '') + '</th>'
        + '<th data-k="exit" class="sortable">Exit IP' + (key === 'exit' ? arrow : '') + '</th>'
        + '<th>Region</th>'
        + '<th>ISP</th>'
        + '<th data-k="last" class="sortable">Last Check' + (key === 'last' ? arrow : '') + '</th>'
        + '<th>Node VPS</th>'
        + '<th>Actions</th>'
        + '</tr>';
      thead.querySelectorAll('th.sortable').forEach((th) => {
        th.style.cursor = 'pointer'; th.onclick = () => {
          const k = th.getAttribute('data-k');
          if (this.pSort === k) this.pAsc = !this.pAsc; else { this.pSort = k; this.pAsc = true; }
          localStorage.setItem('pgw_sort_p2', JSON.stringify({ k: this.pSort, a: this.pAsc }));
          this.renderProxies();
        };
      });
    }

    tbody.innerHTML = '';

    if (sorted.length === 0) {
      tbody.innerHTML = '<tr><td colspan="11" class="text-center">No proxies configured</td></tr>';
      return;
    }

    sorted.forEach(proxy => {
      const row = this.createProxyRow(proxy);
      tbody.appendChild(row);
    });
  }

  renderProxySummary() {
    const tbody = document.getElementById('tbody-proxy-summary');
    if (!tbody) return;

    tbody.innerHTML = '';

    if (this.proxies.length === 0) {
      tbody.innerHTML = '<tr><td colspan="5" class="text-center">No proxies configured</td></tr>';
      return;
    }

    this.proxies.forEach(proxy => {
      const tr = document.createElement('tr');
      const statusBadge = this.createStatusBadge(proxy.status);
      const latencyBadge = this.createLatencyBadge(proxy.latency_ms);
      const lastChecked = proxy.last_checked_at
        ? new Date(proxy.last_checked_at).toLocaleTimeString()
        : '—';

      tr.innerHTML = `
        <td>${proxy.host}:${proxy.port}</td>
        <td>${statusBadge}</td>
        <td>${latencyBadge}</td>
        <td>${proxy.exit_ip || '—'}</td>
        <td>${lastChecked}</td>
      `;
      tbody.appendChild(tr);
    });
  }

  createProxyRow(proxy) {
    const tr = document.createElement('tr');

    const statusBadge = this.createStatusBadge(proxy.status);
    const latencyBadge = this.createLatencyBadge(proxy.latency_ms);
    const lastChecked = proxy.last_checked_at
      ? new Date(proxy.last_checked_at).toLocaleTimeString()
      : '—';

    // Build node inline-select for proxy assignment
    const currentNode = proxy.node_id || '';
    const nodeOpts = [`<option value="">— no node —</option>`,
      ...(this.nodes || []).map(n => `<option value="${n.id}" ${n.id === currentNode ? 'selected' : ''}>${n.status === 'online' ? '🟢' : '🔴'} ${n.name || n.ssh_host}</option>`)
    ].join('');
    const nodeSelect = `<select class="form-select form-select-sm" style="min-width:130px"
      onchange="pgw.assignProxyNode('${proxy.id}', this.value)">${nodeOpts}</select>`;

    const truncISP = isp => {
      if (!isp) return '—';
      return isp.replace(/\s*(Telecommunications?|Communications?|Technologies?|Technology|Corp(?:oration)?|Ltd|Limited|Inc(?:orporated)?|S\.A\.|LLC|Co\.)\s*$/gi, '').trim() || isp;
    };

    tr.innerHTML = `
      <td><code>${proxy.id.slice(0, 8)}</code></td>
      <td>${proxy.type}</td>
      <td>${proxy.host}:${proxy.port}</td>
      <td>${statusBadge}</td>
      <td>${latencyBadge}</td>
      <td>${proxy.exit_ip || '—'}</td>
      <td><small class="text-muted">${proxy.region || '—'}</small></td>
      <td><small class="text-muted">${truncISP(proxy.isp)}</small></td>
      <td>${lastChecked}</td>
      <td>${nodeSelect}</td>
      <td>
        <button class="btn btn-sm btn-secondary" onclick="pgw.checkProxyHealth('${proxy.id}')" data-tooltip="Health check">
          Check
        </button>
        <button class="btn btn-sm btn-danger" onclick="pgw.deleteProxy('${proxy.id}')" data-tooltip="Delete proxy">
          ×
        </button>
      </td>
    `;

    return tr;
  }

  createStatusBadge(status) {
    const statusClass = {
      'OK': 'text-bg-success',
      'DEGRADED': 'text-bg-warning',
      'DOWN': 'text-bg-danger'
    }[status] || 'text-bg-secondary';

    return `<span class="badge ${statusClass}">${status || 'Unknown'}</span>`;
  }

  createLatencyBadge(ms) {
    if (ms == null || isNaN(ms)) return '—';
    let cls = 'text-bg-danger';
    if (ms < 300) cls = 'text-bg-success';
    else if (ms < 900) cls = 'text-bg-warning';
    return `<span class="badge ${cls}">${ms}ms</span>`;
  }


  renderMappings() {
    const tbody = document.getElementById('tbody-mappings');
    if (!tbody) return;
    // sort
    const key = this.mSort, asc = this.mAsc;
    const val = (m) => {
      if (key === 'id') return (m.id || '');
      if (key === 'client') return ((m.client?.ip_cidr) || '');
      if (key === 'proxy') { const p = m.proxy || {}; return ((p.host || '') + ':' + (p.port ?? '')); }
      if (key === 'state') return (m.state || '');
      if (key === 'port') return (m.local_redirect_port ?? 0);
      return ((m.client?.ip_cidr) || '');
    };
    const sorted = (this.mappings || []).slice().sort((a, b) => { const va = val(a), vb = val(b); if (va < vb) return asc ? -1 : 1; if (va > vb) return asc ? 1 : -1; return 0; });
    // header icons + click
    const thead = tbody.parentElement?.querySelector('thead');
    if (thead) {
      const arrow = asc ? ' ▲' : ' ▼';
      thead.innerHTML = '<tr>'
        + '<th data-k="id" class="sortable">ID' + (key === 'id' ? arrow : '') + '</th>'
        + '<th data-k="client" class="sortable">Client IP/CIDR' + (key === 'client' ? arrow : '') + '</th>'
        + '<th data-k="proxy" class="sortable">Proxy Server' + (key === 'proxy' ? arrow : '') + '</th>'
        + '<th data-k="state" class="sortable">State' + (key === 'state' ? arrow : '') + '</th>'
        + '<th data-k="port" class="sortable">Local Port' + (key === 'port' ? arrow : '') + '</th>'
        + '<th>Actions</th>'
        + '</tr>';
      thead.querySelectorAll('th.sortable').forEach((th) => {
        th.style.cursor = 'pointer'; th.onclick = () => {
          const k = th.getAttribute('data-k');
          if (this.mSort === k) this.mAsc = !this.mAsc; else { this.mSort = k; this.mAsc = true; }
          localStorage.setItem('pgw_sort_m2', JSON.stringify({ k: this.mSort, a: this.mAsc }));
          this.renderMappings();
        };
      });
    }

    tbody.innerHTML = '';

    if (sorted.length === 0) {
      tbody.innerHTML = '<tr><td colspan="6" class="text-center">No mappings configured</td></tr>';
      return;
    }

    sorted.forEach(mapping => {
      const row = this.createMappingRow(mapping);
      tbody.appendChild(row);
    });
  }

  createMappingRow(mapping) {
    const tr = document.createElement('tr');

    const proxyAddress = mapping.proxy
      ? `${mapping.proxy.host}:${mapping.proxy.port}`
      : '—';

    const stateBadge = this.createStatusBadge(mapping.state || 'PENDING');

    tr.innerHTML = `
      <td><code>${mapping.id.slice(0, 8)}</code></td>
      <td>${mapping.client?.ip_cidr || '—'}</td>
      <td>${proxyAddress}</td>
      <td>${stateBadge}</td>
      <td>${mapping.local_redirect_port || '—'}</td>
      <td>
        <button class="btn btn-sm btn-danger" onclick="pgw.deleteMapping('${mapping.id}')">
          Delete
        </button>
      </td>
    `;

    return tr;
  }

  renderClients() {
    const select = document.getElementById('select-proxy');
    if (select) {
      select.innerHTML = '<option value="">Select proxy server...</option>';
      const used = new Set((this.mappings || []).map(m => m && m.proxy ? m.proxy.id : null).filter(Boolean));
      const available = (this.proxies || []).filter(p => !used.has(p.id));
      if (!available || available.length === 0) {
        const opt = document.createElement('option');
        opt.disabled = true;
        opt.textContent = 'No available proxies (all mapped)';
        select.appendChild(opt);
      } else {
        available.forEach(proxy => {
          const option = document.createElement('option');
          option.value = proxy.id;
          const statusIndicator = proxy.status === 'OK' ? '✓' : proxy.status === 'DEGRADED' ? '⚠' : '✗';
          option.textContent = `${statusIndicator} ${proxy.host}:${proxy.port} (${proxy.type})`;
          select.appendChild(option);
        });
      }
    }

    // Populate node dropdown in create-mapping form
    const nodeSelect = document.getElementById('select-node');
    if (nodeSelect) {
      nodeSelect.innerHTML = '<option value="">-- Không gán node --</option>';
      (this.nodes || []).forEach(n => {
        const opt = document.createElement('option');
        opt.value = n.id;
        const status = n.status === 'online' ? '🟢' : n.status === 'deploying' ? '🟡' : '🔴';
        opt.textContent = `${status} ${n.label || n.ssh_host}`;
        nodeSelect.appendChild(opt);
      });
    }
  }
  detectProxyType(originalLine, host, port) {
    // Check for explicit type prefix
    if (originalLine.includes("socks5://")) return "socks5";
    if (originalLine.includes("http://")) return "http";

    // Auto-detect based on common SOCKS5 ports
    const socksCommonPorts = [1080, 9050, 9150];
    if (socksCommonPorts.includes(port)) return "socks5";

    // Default to HTTP
    return "http";
  }

  parseProxyLine(line) {
    const cleanLine = line.replace(/^(https?|socks5):\/\//, ''); const m = cleanLine.trim().match(/^([^:\s]+):(\d{1,5}):([^:]*):([^:]*)$/);
    if (!m) return null;
    const host = m[1];
    const port = parseInt(m[2], 10);
    const username = m[3] || "";
    const password = m[4] || "";
    if (!host || !port || port <= 0 || port > 65535) return null;
    return { type: this.detectProxyType(line, host, port), host, port, username, password, enabled: true };
  }

  async importProxies() {
    const textarea = document.getElementById("import-proxies");
    if (!textarea) return;
    const raw = textarea.value || "";
    const lines = raw.split(/\r?\n/).map(l => l.trim()).filter(Boolean);
    if (lines.length === 0) {
      this.showAlert("No proxies to import", "warning");
      return;
    }

    let ok = 0, skipped = 0;
    for (const [idx, line] of lines.entries()) {
      if (line.startsWith("#")) { skipped++; continue; }
      const data = this.parseProxyLine(line);
      if (!data) { skipped++; continue; }
      try {
        const created = await this.apiCall(`${this.apiBase}/v1/proxies`, { method: "POST", body: JSON.stringify(data) });
        ok++;
        setTimeout(() => this.checkProxyHealth(created.id), 500);
      } catch (e) {
        console.error("Import failed for line", idx + 1, line, e);
        skipped++;
      }
    }

    this.showAlert(`Imported ${ok} proxies${skipped ? `, skipped ${skipped}` : ""}`, ok > 0 ? "success" : "warning");
    if (ok > 0) this.loadData();
  }


  updateCounts() {
    this.updateElement('proxy-count', `${this.proxies.length} proxies`);
    this.updateElement('mapping-count', `${this.mappings.length} mappings`);
  }

  async createProxy() {
    const form = document.getElementById('form-proxy');
    const formData = new FormData(form);

    const proxyData = {
      type: formData.get('type'),
      host: formData.get('host'),
      port: parseInt(formData.get('port')),
      username: formData.get('username') || '',
      password: formData.get('password') || '',
      enabled: true
    };

    try {
      const newProxy = await this.apiCall(`${this.apiBase}/v1/proxies`, {
        method: 'POST',
        body: JSON.stringify(proxyData)
      });

      this.showAlert('Proxy created successfully', 'success');
      form.reset();
      this.loadData();

      // Auto health check the new proxy
      setTimeout(() => this.checkProxyHealth(newProxy.id), 1000);

    } catch (error) {
      console.error('Failed to create proxy:', error);
    }
  }

  async createMapping() {
    const form = document.getElementById('form-mapping');
    const formData = new FormData(form);

    const clientIP = (formData.get('client_ip') || '').trim();
    const proxyId = formData.get('proxy_id');

    // Frontend validation: IPv4 only, forbid CIDR
    const ipv4Re = /^(25[0-5]|2[0-4]\d|[01]?\d\d?)\.(25[0-5]|2[0-4]\d|[01]?\d\d?)\.(25[0-5]|2[0-4]\d|[01]?\d\d?)\.(25[0-5]|2[0-4]\d|[01]?\d\d?)$/;
    if (clientIP.includes('/')) {
      this.showAlert('CIDR is not allowed. Please enter a single IPv4 address (e.g., 192.168.2.3).', 'warning');
      return;
    }
    if (clientIP && !ipv4Re.test(clientIP)) {
      this.showAlert('Invalid IPv4 address format.', 'warning');
      return;
    }

    if (!clientIP || !proxyId) {
      this.showAlert('Please fill all required fields', 'warning');
      return;
    }

    try {
      // First create client if not exists
      let clientId;
      const existingClient = this.clients.find(c => c.ip_cidr === `${clientIP}/32`);

      if (existingClient) {
        clientId = existingClient.id;
      } else {
        const client = await this.apiCall(`${this.apiBase}/v1/clients`, {
          method: 'POST',
          body: JSON.stringify({
            ip_cidr: clientIP, // API will auto-add /32
            enabled: true
          })
        });
        clientId = client.id;
      }

      // Create mapping
      await this.apiCall(`${this.apiBase}/v1/mappings`, {
        method: 'POST',
        body: JSON.stringify({
          client_id: clientId,
          proxy_id: proxyId
        })
      });

      this.showAlert('Mapping created successfully', 'success');
      form.reset();
      this.loadData();

      // Auto reconcile after creating mapping
      setTimeout(() => this.reconcileRules(), 1000);

    } catch (error) {
      console.error('Failed to create mapping:', error);
    }
  }

  async assignProxyNode(proxyId, nodeId) {
    try {
      await this.apiCall(`${this.apiBase}/v1/proxies/${proxyId}/node`, {
        method: 'PUT',
        body: JSON.stringify({ node_id: nodeId })
      });
      this.showAlert(nodeId ? 'Proxy assigned to node' : 'Proxy unassigned from node', 'success');
      this.loadData();
    } catch (error) {
      console.error('Failed to assign proxy node:', error);
    }
  }

  async assignNode(mappingId, nodeId) {
    try {
      await this.apiCall(`${this.apiBase}/v1/mappings/${mappingId}/node`, {
        method: 'PUT',
        body: JSON.stringify({ node_id: nodeId })
      });
      this.showAlert(nodeId ? 'Node assigned' : 'Node unassigned', 'success');
      this.loadData();
    } catch (error) {
      console.error('Failed to assign node:', error);
    }
  }

  async checkProxyHealth(proxyId) {
    try {
      await this.apiCall(`${this.apiBase}/v1/proxies/${proxyId}/check`, {
        method: 'POST'
      });

      this.showAlert('Health check completed', 'success');
      this.loadData();
    } catch (error) {
      console.error('Health check failed:', error);
    }
  }

  async healthCheckAll() {
    if (this.proxies.length === 0) {
      this.showAlert('No proxies to check', 'warning');
      return;
    }

    this.showAlert('Running health checks...', 'info');

    const checkPromises = this.proxies.map(proxy =>
      this.checkProxyHealth(proxy.id).catch(e => console.error(`Health check failed for ${proxy.id}:`, e))
    );

    try {
      await Promise.all(checkPromises);
      this.showAlert('All health checks completed', 'success');
    } catch (error) {
      console.error('Some health checks failed:', error);
    }
  }

  async deleteProxy(proxyId) {
    if (!confirm('Are you sure you want to delete this proxy? This will also remove any associated mappings.')) {
      return;
    }

    try {
      await this.apiCall(`${this.apiBase}/v1/proxies/${proxyId}`, {
        method: 'DELETE'
      });

      this.showAlert('Proxy deleted successfully', 'success');
      this.loadData();
    } catch (error) {
      console.error('Failed to delete proxy:', error);
    }
  }

  async deleteMapping(mappingId) {
    if (!confirm('Are you sure you want to delete this mapping?')) {
      return;
    }

    try {
      await this.apiCall(`${this.apiBase}/v1/mappings/${mappingId}`, {
        method: 'DELETE'
      });

      this.showAlert('Mapping deleted successfully', 'success');
      this.loadData();

      // Auto reconcile after deleting mapping
      setTimeout(() => this.reconcileRules(), 1000);

    } catch (error) {
      console.error('Failed to delete mapping:', error);
    }
  }

  async reconcileRules() {
    try {
      const response = await fetch(`${this.agentBase}/reconcile`);

      if (response.ok) {
        this.showAlert('Rules reconciled successfully', 'success');
        this.updateElement('last-reconcile', new Date().toLocaleTimeString());
        this.loadData();
      } else {
        throw new Error('Reconcile failed');
      }
    } catch (error) {
      console.error('Reconcile failed:', error);
      this.showAlert('Failed to reconcile rules', 'danger');
    }
  }

  exportProxies() {
    if (this.proxies.length === 0) {
      this.showAlert('No proxies to export', 'warning');
      return;
    }

    const csvContent = [
      'ID,Type,Host,Port,Status,Latency,Exit IP,Last Check',
      ...this.proxies.map(p => [
        p.id,
        p.type,
        p.host,
        p.port,
        p.status || 'Unknown',
        p.latency_ms || '',
        p.exit_ip || '',
        p.last_checked_at || ''
      ].join(','))
    ].join('\n');

    this.downloadFile(csvContent, 'pgw-proxies.csv', 'text/csv');
  }

  exportMappings() {
    if (this.mappings.length === 0) {
      this.showAlert('No mappings to export', 'warning');
      return;
    }

    const csvContent = [
      'ID,Client IP,Proxy Host,Proxy Port,State,Local Port',
      ...this.mappings.map(m => [
        m.id,
        m.client?.ip_cidr || '',
        m.proxy?.host || '',
        m.proxy?.port || '',
        m.state || 'PENDING',
        m.local_redirect_port || ''
      ].join(','))
    ].join('\n');

    this.downloadFile(csvContent, 'pgw-mappings.csv', 'text/csv');
  }

  downloadFile(content, filename, mimeType) {
    const blob = new Blob([content], { type: mimeType });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);

    this.showAlert(`${filename} downloaded`, 'success');
  }

  showAlert(message, type = 'info') {
    const container = document.getElementById('toast-container');
    if (!container) return;

    // Ensure container is overlay and toast-ready (CSS handles positioning)
    container.classList.add('toast-stack');

    const bsType = ['success', 'danger', 'warning', 'info', 'primary', 'secondary'].includes(type) ? type : 'info';
    const toast = document.createElement('div');
    toast.className = `toast align-items-center text-bg-${bsType} border-0`;
    toast.setAttribute('role', 'alert');
    toast.setAttribute('aria-live', 'assertive');
    toast.setAttribute('aria-atomic', 'true');

    const inner = document.createElement('div');
    inner.className = 'd-flex';
    const body = document.createElement('div');
    body.className = 'toast-body';
    body.textContent = message;
    const btn = document.createElement('button');
    btn.type = 'button';
    btn.className = 'btn-close btn-close-white me-2 m-auto';
    btn.setAttribute('data-bs-dismiss', 'toast');
    btn.setAttribute('aria-label', 'Close');

    inner.appendChild(body);
    inner.appendChild(btn);
    toast.appendChild(inner);

    container.appendChild(toast);

    const inst = bootstrap.Toast.getOrCreateInstance(toast, { delay: 3500 });
    inst.show();
    toast.addEventListener('hidden.bs.toast', () => toast.remove());
  }

  // alias used by sub-managers (EmailManager, PayPalManager, IncomeManager)
  showToast(message, type = 'info') { this.showAlert(message, type); }

  showLoading(show) {
    const loadingEl = document.getElementById('loading-indicator');
    if (loadingEl) {
      loadingEl.style.display = show ? 'flex' : 'none';
    }
  }

  updateElement(id, value) {
    const element = document.getElementById(id);
    if (element) {
      element.textContent = value;
    }
  }

  updateLastRefresh() {
    this.updateElement('last-refresh', new Date().toLocaleTimeString());
  }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  window.pgw = new PGWManager();
});

// Set active navigation
document.addEventListener('DOMContentLoaded', () => {
  const currentPath = window.location.pathname;
  const navLinks = document.querySelectorAll('.nav-link');

  navLinks.forEach(link => {
    link.classList.remove('active');
    if (link.getAttribute('href') === currentPath) {
      link.classList.add('active');
    }
  });
});

// ============================================================
// Email Management
// ============================================================
class EmailManager {
  constructor(pgw) { this.pgw = pgw; this.data = []; this.paypals = []; this.sortKey = 'address'; this.sortAsc = true; }

  async init() {
    document.getElementById('form-email')?.addEventListener('submit', e => { e.preventDefault(); this.addEmail(); });
    await this.loadPayPals();
    await this.load();
  }

  async loadPayPals() {
    try {
      const res = await fetch('/api/v1/paypals');
      this.paypals = res.ok ? (await res.json() || []) : [];
    } catch (_) { this.paypals = []; }
  }

  async load() {
    try {
      const res = await fetch('/api/v1/emails');
      if (!res.ok) throw new Error(res.statusText);
      this.data = await res.json() || [];
      this.render();
    } catch (e) { console.error('emails load:', e); }
  }

  async addEmail() {
    const address = document.getElementById('email-address').value.trim();
    const provider = document.getElementById('email-provider').value;
    const password = document.getElementById('email-password').value.trim();
    const recovery = document.getElementById('email-recovery').value.trim();
    const note = document.getElementById('email-note').value.trim();
    if (!address) return;
    try {
      const res = await fetch('/api/v1/emails', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ address, provider, password, recovery_email: recovery, note, status: 'active' })
      });
      if (!res.ok) throw new Error(await res.text());
      document.getElementById('form-email').reset();
      this.pgw.showToast('Email added', 'success');
      await this.load();
    } catch (e) { this.pgw.showToast('Error: ' + e.message, 'danger'); }
  }

  async deleteEmail(id) {
    if (!confirm('Xóa email này?')) return;
    try {
      await fetch('/api/v1/emails/' + id, { method: 'DELETE' });
      this.pgw.showToast('Deleted', 'success');
      await this.load();
    } catch (e) { this.pgw.showToast('Error', 'danger'); }
  }

  // Format a date — returns '' if zero/invalid
  fmtDate(v) {
    if (!v) return '—';
    const d = new Date(v);
    if (isNaN(d) || d.getFullYear() < 1970) return '—';
    return d.toLocaleDateString('vi-VN');
  }

  render() {
    const tbody = document.getElementById('tbody-emails');
    const badge = document.getElementById('email-count');
    if (!tbody) return;
    if (badge) badge.textContent = this.data.length + ' emails';
    if (!this.data.length) {
      tbody.innerHTML = '<tr><td colspan="8" class="text-center text-muted py-4">No emails yet</td></tr>'; return;
    }
    const statusBadge = s => {
      const m = { active: 'success', disabled: 'secondary', banned: 'danger' };
      return `<span class="badge bg-label-${m[s] || 'secondary'}">${s}</span>`;
    };
    tbody.innerHTML = this.data.map(e => {
      const linkedPayPal = e.paypal_id
        ? (this.paypals.find(p => p.id === e.paypal_id)?.email || e.paypal_id)
        : '—';
      return `
      <tr>
        <td><span class="fw-medium">${e.address}</span></td>
        <td><span class="badge bg-label-primary">${e.provider || 'other'}</span></td>
        <td>${statusBadge(e.status || 'active')}</td>
        <td class="text-muted small">${linkedPayPal !== '—' ? `<span class="fw-medium text-body">${linkedPayPal}</span>` : '—'}</td>
        <td class="text-muted small">${e.note || '—'}</td>
        <td class="text-muted small">${this.fmtDate(e.created_at)}</td>
        <td>
          <div class="d-flex gap-1">
            <button class="btn btn-sm btn-icon btn-label-warning" onclick="window.emailMgr.editEmail('${e.id}')" title="Edit">
              <i class="bi bi-pencil"></i>
            </button>
            <button class="btn btn-sm btn-icon btn-label-danger" onclick="window.emailMgr.deleteEmail('${e.id}')" title="Delete">
              <i class="bi bi-trash"></i>
            </button>
          </div>
        </td>
      </tr>`;
    }).join('');
  }

  editEmail(id) {
    const e = this.data.find(x => x.id === id);
    if (!e) return;
    document.getElementById('edit-email-id').value = e.id;
    document.getElementById('edit-email-address').value = e.address || '';
    document.getElementById('edit-email-provider').value = e.provider || 'gmail';
    document.getElementById('edit-email-status').value = e.status || 'active';
    document.getElementById('edit-email-password').value = e.password || '';
    document.getElementById('edit-email-recovery').value = e.recovery_email || '';
    document.getElementById('edit-email-note').value = e.note || '';
    // Populate PayPal dropdown
    const sel = document.getElementById('edit-email-paypal');
    sel.innerHTML = '<option value="">— None —</option>' +
      this.paypals.map(p => `<option value="${p.id}" ${e.paypal_id === p.id ? 'selected' : ''}>${p.email}</option>`).join('');
    const modal = new bootstrap.Modal(document.getElementById('editEmailModal'));
    modal.show();
  }

  async saveEmail() {
    const id = document.getElementById('edit-email-id').value;
    const paypal_id = document.getElementById('edit-email-paypal').value || null;
    const body = {
      address: document.getElementById('edit-email-address').value.trim(),
      provider: document.getElementById('edit-email-provider').value,
      status: document.getElementById('edit-email-status').value,
      password: document.getElementById('edit-email-password').value.trim(),
      recovery_email: document.getElementById('edit-email-recovery').value.trim(),
      note: document.getElementById('edit-email-note').value.trim(),
      paypal_id,
    };
    try {
      const res = await fetch('/api/v1/emails/' + id, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      });
      if (!res.ok) throw new Error(await res.text());
      bootstrap.Modal.getInstance(document.getElementById('editEmailModal'))?.hide();
      this.pgw.showToast('Email updated', 'success');
      await this.load();
    } catch (err) { this.pgw.showToast('Error: ' + err.message, 'danger'); }
  }
}

// ============================================================
// PayPal Management
// ============================================================
class PayPalManager {
  constructor(pgw) { this.pgw = pgw; this.data = []; }

  async init() {
    document.getElementById('form-paypal')?.addEventListener('submit', e => { e.preventDefault(); this.addPayPal(); });
    await this.load();
  }

  async load() {
    try {
      const res = await fetch('/api/v1/paypals');
      if (!res.ok) throw new Error(res.statusText);
      this.data = await res.json() || [];
      this.render();
      return this.data;
    } catch (e) { console.error('paypals load:', e); return []; }
  }

  async addPayPal() {
    const email = document.getElementById('paypal-email').value.trim();
    const owner_name = document.getElementById('paypal-owner').value.trim();
    const balance = parseFloat(document.getElementById('paypal-balance').value) || 0;
    const currency = document.getElementById('paypal-currency').value;
    const status = document.getElementById('paypal-status').value;
    const verified = document.getElementById('paypal-verified').checked;
    const note = document.getElementById('paypal-note').value.trim();
    if (!email) return;
    try {
      const res = await fetch('/api/v1/paypals', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, owner_name, balance, currency, status, verified, note })
      });
      if (!res.ok) throw new Error(await res.text());
      document.getElementById('form-paypal').reset();
      this.pgw.showToast('PayPal added', 'success');
      await this.load();
    } catch (e) { this.pgw.showToast('Error: ' + e.message, 'danger'); }
  }

  async deletePayPal(id) {
    if (!confirm('Xóa PayPal account này? Income liên quan sẽ bị xóa.')) return;
    try {
      await fetch('/api/v1/paypals/' + id, { method: 'DELETE' });
      this.pgw.showToast('Deleted', 'success');
      await this.load();
    } catch (e) { this.pgw.showToast('Error', 'danger'); }
  }

  fmtDate(v) {
    if (!v) return '—';
    const d = new Date(v);
    if (isNaN(d) || d.getFullYear() < 1970) return '—';
    return d.toLocaleDateString('vi-VN');
  }

  render() {
    const tbody = document.getElementById('tbody-paypals');
    const badge = document.getElementById('paypal-count');
    if (!tbody) return;
    if (badge) badge.textContent = this.data.length + ' accounts';
    if (!this.data.length) {
      tbody.innerHTML = '<tr><td colspan="9" class="text-center text-muted py-4">No PayPal accounts yet</td></tr>'; return;
    }
    const statusBadge = s => {
      const m = { active: 'success', limited: 'warning', suspended: 'danger' };
      return `<span class="badge bg-label-${m[s] || 'secondary'}">${s}</span>`;
    };
    tbody.innerHTML = this.data.map(p => `
      <tr>
        <td><span class="fw-medium">${p.email}</span></td>
        <td>${p.owner_name || '<span class="text-muted">—</span>'}</td>
        <td class="fw-semibold text-success">$${(p.balance || 0).toFixed(2)} <span class="text-muted small">${p.currency || 'USD'}</span></td>
        <td>${statusBadge(p.status || 'active')}</td>
        <td>${p.verified ? '<i class="bi bi-check-circle-fill text-success"></i>' : '<i class="bi bi-x-circle text-muted"></i>'}</td>
        <td class="text-muted small">${p.note || '—'}</td>
        <td class="text-muted small">${this.fmtDate(p.created_at)}</td>
        <td>
          <div class="d-flex gap-1">
            <button class="btn btn-sm btn-icon btn-label-warning" onclick="window.paypalMgr.editPayPal('${p.id}')" title="Edit">
              <i class="bi bi-pencil"></i>
            </button>
            <button class="btn btn-sm btn-icon btn-label-danger" onclick="window.paypalMgr.deletePayPal('${p.id}')" title="Delete">
              <i class="bi bi-trash"></i>
            </button>
          </div>
        </td>
      </tr>`).join('');
  }

  editPayPal(id) {
    const p = this.data.find(x => x.id === id);
    if (!p) return;
    document.getElementById('edit-paypal-id').value = p.id;
    document.getElementById('edit-paypal-email').value = p.email || '';
    document.getElementById('edit-paypal-owner').value = p.owner_name || '';
    document.getElementById('edit-paypal-balance').value = p.balance || 0;
    document.getElementById('edit-paypal-currency').value = p.currency || 'USD';
    document.getElementById('edit-paypal-status').value = p.status || 'active';
    document.getElementById('edit-paypal-verified').checked = !!p.verified;
    document.getElementById('edit-paypal-note').value = p.note || '';
    const modal = new bootstrap.Modal(document.getElementById('editPayPalModal'));
    modal.show();
  }

  async savePayPal() {
    const id = document.getElementById('edit-paypal-id').value;
    const body = {
      email: document.getElementById('edit-paypal-email').value.trim(),
      owner_name: document.getElementById('edit-paypal-owner').value.trim(),
      balance: parseFloat(document.getElementById('edit-paypal-balance').value) || 0,
      currency: document.getElementById('edit-paypal-currency').value,
      status: document.getElementById('edit-paypal-status').value,
      verified: document.getElementById('edit-paypal-verified').checked,
      note: document.getElementById('edit-paypal-note').value.trim(),
    };
    try {
      const res = await fetch('/api/v1/paypals/' + id, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      });
      if (!res.ok) throw new Error(await res.text());
      bootstrap.Modal.getInstance(document.getElementById('editPayPalModal'))?.hide();
      this.pgw.showToast('PayPal updated', 'success');
      await this.load();
    } catch (err) { this.pgw.showToast('Error: ' + err.message, 'danger'); }
  }
}

// ============================================================
// Income Management
// ============================================================
class IncomeManager {
  constructor(pgw) { this.pgw = pgw; this.data = []; this.paypals = []; }

  async init() {
    document.getElementById('form-income')?.addEventListener('submit', e => { e.preventDefault(); this.addIncome(); });
    // Set default received_at to now
    const dtEl = document.getElementById('income-received-at');
    if (dtEl) dtEl.value = new Date().toISOString().slice(0, 16);

    await Promise.all([this.loadPayPals(), this.load()]);
    this.updateSummary();
  }

  async loadPayPals() {
    try {
      const res = await fetch('/api/v1/paypals');
      this.paypals = await res.json() || [];
      const sel = document.getElementById('income-paypal-id');
      if (sel) {
        const opts = this.paypals.map(p => `<option value="${p.id}">${p.email}</option>`).join('');
        sel.innerHTML = '<option value="">(none)</option>' + opts;
      }
    } catch (e) { }
  }

  async load() {
    try {
      const res = await fetch('/api/v1/income');
      if (!res.ok) throw new Error(res.statusText);
      this.data = await res.json() || [];
      this.render();
    } catch (e) { console.error('income load:', e); }
  }

  async addIncome() {
    const amount = parseFloat(document.getElementById('income-amount').value);
    const currency = document.getElementById('income-currency').value;
    const source = document.getElementById('income-source').value;
    const paypalIdEl = document.getElementById('income-paypal-id');
    const paypal_id = paypalIdEl?.value || null;
    const description = document.getElementById('income-description').value.trim();
    const receivedAtEl = document.getElementById('income-received-at');
    const received_at = receivedAtEl?.value ? new Date(receivedAtEl.value).toISOString() : new Date().toISOString();
    if (!amount || amount <= 0) return;
    try {
      const body = { amount, currency, source, description, received_at };
      if (paypal_id) body.paypal_id = paypal_id;
      const res = await fetch('/api/v1/income', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      });
      if (!res.ok) throw new Error(await res.text());
      document.getElementById('form-income').reset();
      if (receivedAtEl) receivedAtEl.value = new Date().toISOString().slice(0, 16);
      this.pgw.showToast('Income recorded', 'success');
      await this.load();
      this.updateSummary();
    } catch (e) { this.pgw.showToast('Error: ' + e.message, 'danger'); }
  }

  async deleteIncome(id) {
    if (!confirm('Xóa income entry này?')) return;
    try {
      await fetch('/api/v1/income/' + id, { method: 'DELETE' });
      this.pgw.showToast('Deleted', 'success');
      await this.load();
      this.updateSummary();
    } catch (e) { this.pgw.showToast('Error', 'danger'); }
  }

  updateSummary() {
    const total = this.data.reduce((s, i) => s + (i.amount || 0), 0);
    const now = new Date();
    const thisMonth = now.getFullYear() + '-' + String(now.getMonth() + 1).padStart(2, '0');
    const monthTotal = this.data
      .filter(i => i.received_at && i.received_at.startsWith(thisMonth))
      .reduce((s, i) => s + (i.amount || 0), 0);
    const paypalTotal = this.data
      .filter(i => i.source === 'paypal')
      .reduce((s, i) => s + (i.amount || 0), 0);

    const set = (id, v) => { const el = document.getElementById(id); if (el) el.textContent = v; };
    set('income-total', '$' + total.toFixed(2));
    set('income-month', '$' + monthTotal.toFixed(2));
    set('income-count', this.data.length);
    set('income-paypal-total', '$' + paypalTotal.toFixed(2));
  }

  render() {
    const tbody = document.getElementById('tbody-income');
    if (!tbody) return;
    if (!this.data.length) {
      tbody.innerHTML = '<tr><td colspan="6" class="text-center text-muted py-4">No income recorded yet</td></tr>'; return;
    }
    const srcBadge = s => {
      const m = { paypal: 'warning', bank: 'primary', crypto: 'info', other: 'secondary' };
      return `<span class="badge bg-label-${m[s] || 'secondary'}">${s || 'other'}</span>`;
    };
    const ppMap = {};
    this.paypals.forEach(p => ppMap[p.id] = p.email);
    tbody.innerHTML = this.data.map(i => `
      <tr>
        <td class="text-muted small">${new Date(i.received_at).toLocaleString('vi-VN')}</td>
        <td class="fw-semibold text-success">$${(i.amount || 0).toFixed(2)} <span class="text-muted small">${i.currency || 'USD'}</span></td>
        <td>${srcBadge(i.source)}</td>
        <td class="text-muted small">${i.paypal_id ? (ppMap[i.paypal_id] || i.paypal_id) : '—'}</td>
        <td class="text-muted small">${i.description || '—'}</td>
        <td>
          <button class="btn btn-sm btn-icon btn-label-danger" onclick="window.incomeMgr.deleteIncome('${i.id}')" title="Delete">
            <i class="bi bi-trash"></i>
          </button>
        </td>
      </tr>`).join('');
  }
}

// ============================================================
// Auto-init based on current page
// ============================================================
document.addEventListener('DOMContentLoaded', () => {
  const path = window.location.pathname;
  const pgw = window.pgw || { showToast: (m, t) => console.log(t, m), showAlert: (m, t) => console.log(t, m) };

  if (path === '/emails') {
    window.emailMgr = new EmailManager(pgw);
    window.emailMgr.init();
  } else if (path === '/paypal') {
    window.paypalMgr = new PayPalManager(pgw);
    window.paypalMgr.init();
  } else if (path === '/income') {
    window.incomeMgr = new IncomeManager(pgw);
    window.incomeMgr.init();
  }
});
