'use strict';

const screencastSockets = {};

function getScreencastParams() {
  return '&quality=' + screencastSettings.quality + '&maxWidth=' + screencastSettings.maxWidth + '&fps=' + screencastSettings.fps;
}

function connectScreencast(tabId, baseUrl) {
  const wsProto = baseUrl ? 'ws:' : (location.protocol === 'https:' ? 'wss:' : 'ws:');
  const wsHost = baseUrl ? baseUrl.replace(/^https?:\/\//, '') : location.host;
  const wsUrl = wsProto + '//' + wsHost + '/screencast?tabId=' + tabId + getScreencastParams();

  const socket = new WebSocket(wsUrl);
  socket.binaryType = 'arraybuffer';
  screencastSockets[tabId] = socket;

  const canvas = document.getElementById('canvas-' + tabId);
  if (!canvas) return;
  const ctx2d = canvas.getContext('2d');

  let frameCount = 0;
  let lastFpsTime = Date.now();
  const statusEl = document.getElementById('status-' + tabId);
  const fpsEl = document.getElementById('fps-' + tabId);
  const sizeEl = document.getElementById('size-' + tabId);

  socket.onopen = () => { if (statusEl) statusEl.className = 'tile-status streaming'; };
  socket.onmessage = (evt) => {
    const blob = new Blob([evt.data], { type: 'image/jpeg' });
    const url = URL.createObjectURL(blob);
    const img = new Image();
    img.onload = () => {
      canvas.width = img.width;
      canvas.height = img.height;
      ctx2d.drawImage(img, 0, 0);
      URL.revokeObjectURL(url);
    };
    img.src = url;
    frameCount++;
    const now = Date.now();
    if (now - lastFpsTime >= 1000) {
      if (fpsEl) fpsEl.textContent = frameCount + ' fps';
      if (sizeEl) sizeEl.textContent = (evt.data.byteLength / 1024).toFixed(0) + ' KB/frame';
      frameCount = 0;
      lastFpsTime = now;
    }
  };
  socket.onerror = () => { if (statusEl) statusEl.className = 'tile-status error'; };
  socket.onclose = () => { if (statusEl) statusEl.className = 'tile-status error'; };
}

function startScreencast(tabId) { connectScreencast(tabId, null); }

async function refreshTabs() {
  Object.values(screencastSockets).forEach(s => s.close());
  Object.keys(screencastSockets).forEach(k => delete screencastSockets[k]);

  try {
    const res = await fetch('/screencast/tabs');
    const tabs = await res.json();
    const grid = document.getElementById('screencast-grid');
    document.getElementById('live-tab-count').textContent = tabs.length + ' tab(s)';

    if (tabs.length === 0) {
      grid.innerHTML = '<div class="empty-state"><div class="crab">🦀</div>No tabs open.</div>';
      return;
    }

    grid.innerHTML = tabs.map(t => `
      <div class="screen-tile" id="tile-${t.id}">
        <div class="tile-header">
          <span>
            <span class="tile-id">${t.id.substring(0, 8)}</span>
            <span class="tile-status connecting" id="status-${t.id}"></span>
          </span>
          <span class="tile-url" id="url-${t.id}">${esc(t.url || 'about:blank')}</span>
        </div>
        <canvas id="canvas-${t.id}" width="800" height="600"></canvas>
        <div class="tile-footer">
          <span id="fps-${t.id}">—</span>
          <span id="size-${t.id}">—</span>
        </div>
      </div>
    `).join('');

    tabs.forEach(t => startScreencast(t.id));
  } catch (e) {
    console.error('Failed to load tabs', e);
  }
}

async function viewInstanceLive(id, port) {
  switchView('live');
  try {
    const res = await fetch('/instances/tabs');
    const tabs = await res.json();
    const instTabs = tabs.filter(t => t.instancePort === port);
    const grid = document.getElementById('screencast-grid');
    document.getElementById('live-tab-count').textContent = instTabs.length + ' tab(s) on ' + id;

    if (instTabs.length === 0) {
      grid.innerHTML = '<div class="empty-state"><div class="crab">🦀</div>No tabs in this instance.</div>';
      return;
    }

    grid.innerHTML = instTabs.map(t => `
      <div class="screen-tile" id="tile-${t.tabId}">
        <div class="tile-header">
          <span>
            <span class="tile-id">${esc(t.instanceName)}:${t.tabId.substring(0, 6)}</span>
            <span class="tile-status connecting" id="status-${t.tabId}"></span>
          </span>
          <span class="tile-url" id="url-${t.tabId}">${esc(t.url || 'about:blank')}</span>
        </div>
        <canvas id="canvas-${t.tabId}" width="800" height="600"></canvas>
        <div class="tile-footer">
          <span id="fps-${t.tabId}">—</span>
          <span id="size-${t.tabId}">—</span>
        </div>
      </div>
    `).join('');

    instTabs.forEach(t => {
      connectScreencast(t.tabId, 'http://localhost:' + t.instancePort);
    });
  } catch (e) {
    console.error('Failed to load instance tabs', e);
  }
}

  window.addEventListener('beforeunload', () => {
  Object.values(screencastSockets).forEach(s => s.close());
});
