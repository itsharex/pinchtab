'use strict';

// ---------------------------------------------------------------------------
// Shared utilities
// ---------------------------------------------------------------------------

function esc(s) {
  if (!s) return '';
  const d = document.createElement('div');
  d.textContent = s;
  return d.innerHTML;
}

function timeAgo(d) {
  const s = Math.floor((Date.now() - d.getTime()) / 1000);
  if (s < 5) return 'just now';
  if (s < 60) return s + 's ago';
  if (s < 3600) return Math.floor(s / 60) + 'm ago';
  return Math.floor(s / 3600) + 'h ago';
}

function appConfirm(message, title, isDanger) {
  return new Promise((resolve) => {
    document.getElementById('confirm-title').textContent = title || 'Confirm';
    document.getElementById('confirm-message').textContent = message;
    const okBtn = document.getElementById('confirm-ok');
    okBtn.textContent = 'Confirm';
    okBtn.style.display = '';
    okBtn.className = isDanger !== false ? 'danger' : '';
    document.getElementById('confirm-cancel').textContent = 'Cancel';
    document.getElementById('confirm-modal').classList.add('open');
    const cleanup = () => { document.getElementById('confirm-modal').classList.remove('open'); };
    okBtn.onclick = () => { cleanup(); resolve(true); };
    document.getElementById('confirm-cancel').onclick = () => { cleanup(); resolve(false); };
  });
}

function appAlert(message, title) {
  return new Promise((resolve) => {
    document.getElementById('confirm-title').textContent = title || 'Notice';
    document.getElementById('confirm-message').textContent = message;
    document.getElementById('confirm-ok').style.display = 'none';
    document.getElementById('confirm-cancel').textContent = 'OK';
    document.getElementById('confirm-modal').classList.add('open');
    document.getElementById('confirm-cancel').onclick = () => {
      document.getElementById('confirm-modal').classList.remove('open');
      resolve();
    };
  });
}

function closeModal() {
  document.getElementById('modal').classList.remove('open');
}

// Reusable modal — pass title + HTML content + optional footer buttons
function showModal(title, bodyHtml, buttons) {
  const modal = document.getElementById('modal');
  document.getElementById('modal-title').textContent = title;
  let html = bodyHtml;
  if (buttons) {
    html += '<div class="btn-row" style="margin-top:16px">' + buttons + '</div>';
  } else {
    html += '<div class="btn-row" style="margin-top:16px"><button class="secondary" onclick="closeModal()">Close</button></div>';
  }
  document.getElementById('modal-body').innerHTML = html;
  modal.classList.add('open');
}

// Fetch with error handling — returns {ok, data} or shows alert on failure
async function apiFetch(url, options) {
  try {
    const res = await fetch(url, options);
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: res.statusText }));
      await appAlert(err.error || 'Request failed', 'Error');
      return { ok: false };
    }
    const data = await res.json().catch(() => null);
    return { ok: true, data };
  } catch (e) {
    await appAlert('Network error: ' + e.message, 'Error');
    return { ok: false };
  }
}
