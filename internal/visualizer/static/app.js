const timeline = document.getElementById('timeline');
const emptyState = document.getElementById('empty-state');
const statusIndicator = document.getElementById('connection-status');
const searchInput = document.getElementById('search-input');
const autoscrollToggle = document.getElementById('autoscroll-toggle');
const clearBtn = document.getElementById('clear-btn');
const filterTabs = document.querySelectorAll('.filter-tabs .tab');

const metricTotal = document.getElementById('metric-total');
const metricRequests = document.getElementById('metric-requests');
const metricResponses = document.getElementById('metric-responses');
const metricErrors = document.getElementById('metric-errors');
const metricDuration = document.getElementById('metric-duration');

let eventsData = [];
let activeFilter = 'all';
let searchQuery = '';

const statusTexts = {
  200: '200 OK',
  201: '201 Created',
  204: '204 No Content',
  301: '301 Moved',
  302: '302 Found',
  304: '304 Not Modified',
  400: '400 Bad Request',
  401: '401 Unauthorized',
  403: '403 Forbidden',
  404: '404 Not Found',
  413: '413 Payload Too Large',
  500: '500 Server Error',
  502: '502 Bad Gateway',
  503: '503 Unavailable',
  504: '504 Timeout',
  0: '0 Conn Error'
};

function triggerPacket(packetId, animClass) {
  const el = document.getElementById(packetId);
  if (!el) return;
  el.classList.remove('anim-req-forward', 'anim-resp-backward');
  void el.offsetWidth;
  el.classList.add(animClass);
}

function updateMetrics() {
  let reqs = 0;
  let resps = 0;
  let errs = 0;
  let maxOffset = 0;

  eventsData.forEach(ev => {
    if (ev.type === 'request') {
      reqs++;
    } else {
      resps++;
      if (ev.status === 0 || ev.status >= 400) {
        errs++;
      }
    }
    if (ev.offsetMs > maxOffset) {
      maxOffset = ev.offsetMs;
    }
  });

  metricTotal.textContent = eventsData.length;
  metricRequests.textContent = reqs;
  metricResponses.textContent = resps;
  metricErrors.textContent = errs;
  metricDuration.textContent = maxOffset + 'ms';
}

function matchFilter(ev) {
  if (activeFilter === 'request' && ev.type !== 'request') return false;
  if (activeFilter === 'response' && ev.type !== 'response') return false;
  if (activeFilter === 'error' && (ev.type !== 'response' || (ev.status < 400 && ev.status !== 0))) return false;

  if (!searchQuery) return true;
  const q = searchQuery.toLowerCase();
  const path = (ev.path || '').toLowerCase();
  const method = (ev.method || '').toLowerCase();
  const status = String(ev.status || '');
  const index = '#' + ev.index;
  return path.includes(q) || method.includes(q) || status.includes(q) || index.includes(q);
}

function renderRow(msg) {
  if (emptyState && emptyState.parentNode) {
    emptyState.style.display = 'none';
  }

  const row = document.createElement('div');
  row.className = 'timeline-row';
  row.dataset.type = msg.type;
  row.dataset.index = msg.index;
  row.dataset.status = msg.status || 0;

  const isReq = msg.type === 'request';
  
  let methodStatusHtml = '';
  let pathInfoHtml = '';

  if (isReq) {
    const method = msg.method || 'GET';
    methodStatusHtml = `<span class="method-badge method-${method}">${method}</span>`;
    pathInfoHtml = `<span class="event-path">${escapeHtml(msg.path || '/')}</span>`;
  } else {
    const st = msg.status || 0;
    let stClass = 'status-0';
    if (st >= 200 && st < 300) stClass = 'status-2xx';
    else if (st >= 300 && st < 400) stClass = 'status-3xx';
    else if (st >= 400 && st < 500) stClass = 'status-4xx';
    else if (st >= 500) stClass = 'status-5xx';

    const label = statusTexts[st] || (st + ' Status');
    methodStatusHtml = `<span class="status-badge ${stClass}">${label}</span>`;
    pathInfoHtml = `<span class="event-path" style="color: var(--text-muted);">Response payload to #${Math.max(0, msg.index - 1)}</span>`;
  }

  const badgeTypeClass = isReq ? 'badge-req' : 'badge-resp';
  const badgeLabel = isReq ? 'REQ' : 'RESP';

  row.innerHTML = `
    <div class="col-index"><span class="event-index">#${msg.index}</span></div>
    <div class="col-offset"><span class="event-offset">+${msg.offsetMs}ms</span></div>
    <div class="col-type"><span class="event-badge ${badgeTypeClass}">${badgeLabel}</span></div>
    <div class="col-method-status">${methodStatusHtml}</div>
    <div class="col-path">${pathInfoHtml}</div>
  `;

  if (!matchFilter(msg)) {
    row.classList.add('hidden');
  }

  timeline.appendChild(row);

  if (autoscrollToggle.checked) {
    timeline.scrollTop = timeline.scrollHeight;
  }
}

function escapeHtml(str) {
  const div = document.createElement('div');
  div.textContent = str;
  return div.innerHTML;
}

function initSSE() {
  const source = new EventSource('/events');

  source.onopen = function() {
    statusIndicator.className = 'status-indicator connected';
    statusIndicator.querySelector('.status-label').textContent = 'Live Stream';
  };

  source.onerror = function() {
    statusIndicator.className = 'status-indicator disconnected';
    statusIndicator.querySelector('.status-label').textContent = 'Disconnected';
  };

  source.onmessage = function(e) {
    try {
      const msg = JSON.parse(e.data);
      eventsData.push(msg);

      if (msg.type === 'request') {
        triggerPacket('packet-1', 'anim-req-forward');
        setTimeout(() => triggerPacket('packet-2', 'anim-req-forward'), 200);
      } else {
        triggerPacket('packet-3', 'anim-resp-backward');
        setTimeout(() => triggerPacket('packet-4', 'anim-resp-backward'), 200);
      }

      renderRow(msg);
      updateMetrics();
    } catch (err) {
      console.error('Error handling SSE message:', err);
    }
  };
}

filterTabs.forEach(tab => {
  tab.addEventListener('click', () => {
    filterTabs.forEach(t => t.classList.remove('active'));
    tab.classList.add('active');
    activeFilter = tab.dataset.filter;
    applyCurrentFilters();
  });
});

searchInput.addEventListener('input', (e) => {
  searchQuery = e.target.value.trim();
  applyCurrentFilters();
});

clearBtn.addEventListener('click', () => {
  timeline.innerHTML = '';
  if (emptyState) {
    emptyState.style.display = 'flex';
    timeline.appendChild(emptyState);
  }
  eventsData = [];
  updateMetrics();
});

function applyCurrentFilters() {
  const rows = timeline.querySelectorAll('.timeline-row');
  rows.forEach((row, i) => {
    const ev = eventsData[i];
    if (ev && matchFilter(ev)) {
      row.classList.remove('hidden');
    } else {
      row.classList.add('hidden');
    }
  });
}

initSSE();
