'use strict';

const API = '/api';
const REFRESH_MS = 5000;

const $ = (sel) => document.querySelector(sel);
const el = {
  statusDot: $('#status-dot'),
  statusText: $('#status-text'),
  body: $('#flags-body'),
  table: $('#flags-table'),
  empty: $('#flags-empty'),
  form: $('#create-form'),
  createMsg: $('#create-msg'),
  toast: $('#toast'),
  autoRefresh: $('#auto-refresh'),
  refreshNote: $('#refresh-note'),
  refreshBtn: $('#refresh-btn'),
};

let flags = [];
let timer = null;

// ---------------------------------------------------------------- helpers

function toast(msg, kind) {
  el.toast.textContent = msg;
  el.toast.className = 'toast ' + (kind || '');
  clearTimeout(toast._t);
  toast._t = setTimeout(() => el.toast.classList.add('hidden'), 3200);
}

function setStatus(state, text) {
  el.statusDot.className = 'dot ' + state;
  el.statusText.textContent = text;
}

async function api(path, opts = {}) {
  const res = await fetch(API + path, {
    headers: { 'Content-Type': 'application/json' },
    ...opts,
  });
  if (res.status === 204) return null;
  const text = await res.text();
  let data = null;
  try { data = text ? JSON.parse(text) : null; } catch (_) { /* non-JSON error body */ }
  if (!res.ok) {
    throw new Error((data && data.error) || res.status + ' ' + res.statusText);
  }
  return data;
}

function esc(s) {
  return String(s == null ? '' : s).replace(/[&<>"']/g, (c) => (
    { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]
  ));
}

function fmtTime(ts) {
  if (!ts) return '—';
  const d = new Date(ts);
  if (isNaN(d.getTime())) return '—';
  return d.toLocaleString();
}

// ---------------------------------------------------------------- render

function render() {
  const has = flags.length > 0;
  el.table.classList.toggle('hidden', !has);
  el.empty.classList.toggle('hidden', has);

  // Do not clobber an input the user is currently editing.
  const active = document.activeElement;
  const activeName = active && active.dataset ? active.dataset.field : null;

  el.body.innerHTML = flags.map((f) => {
    const kill = f.kill_switch
      ? '<span class="pill kill">KILLED</span>'
      : '<span class="pill off">armed</span>';
    const enabled = f.enabled
      ? '<span class="pill on">ON</span>'
      : '<span class="pill off">OFF</span>';
    const rollout = f.enabled && !f.kill_switch && f.rollout_percent > 0 && f.rollout_percent < 100
      ? ` <span class="pill warn">${f.rollout_percent}%</span>` : '';
    return `<tr data-name="${esc(f.name)}">
      <td class="name">${esc(f.name)}</td>
      <td class="muted">${esc(f.description)}</td>
      <td><button class="small" data-act="toggle">${enabled}</button>${rollout}</td>
      <td class="rollout">
        <input type="number" min="0" max="100" value="${f.rollout_percent}"
               data-field="rollout" aria-label="rollout percent">
      </td>
      <td><button class="small" data-act="kill">${kill}</button></td>
      <td class="muted">${fmtTime(f.updated_at)}</td>
      <td style="text-align:right">
        <button class="small danger" data-act="delete">delete</button>
      </td>
    </tr>`;
  }).join('');

  if (activeName) {
    const restored = el.body.querySelector(`[data-field="${activeName}"]`);
    if (restored) {
      restored.focus();
      if (restored.select) restored.select();
    }
  }
}

// ---------------------------------------------------------------- data

async function refresh() {
  try {
    flags = await api('/flags') || [];
    flags.sort((a, b) => a.name.localeCompare(b.name));
    setStatus('ok', `connected · ${flags.length} flag${flags.length === 1 ? '' : 's'}`);
    render();
  } catch (err) {
    setStatus('bad', 'control plane unreachable');
    toast('Failed to load flags: ' + err.message, 'err');
  }
}

async function createFlag(ev) {
  ev.preventDefault();
  el.createMsg.textContent = '';
  el.createMsg.className = 'msg';
  const payload = {
    name: $('#f-name').value.trim(),
    description: $('#f-desc').value.trim(),
    enabled: $('#f-enabled').checked,
    rollout_percent: parseInt($('#f-rollout').value || '0', 10),
    kill_switch: $('#f-kill').checked,
  };
  if (!payload.name) return;
  try {
    await api('/flags', { method: 'POST', body: JSON.stringify(payload) });
    el.form.reset();
    el.createMsg.textContent = 'created “' + payload.name + '”';
    el.createMsg.className = 'msg ok';
    toast('Flag created', 'ok');
    refresh();
  } catch (err) {
    el.createMsg.textContent = err.message;
    el.createMsg.className = 'msg err';
  }
}

async function toggleEnabled(name) {
  const f = flags.find((x) => x.name === name);
  if (!f) return;
  await mutate(name, {
    name: f.name,
    description: f.description,
    enabled: !f.enabled,
    rollout_percent: f.rollout_percent,
    kill_switch: f.kill_switch,
  }, f.enabled ? 'disabled' : 'enabled');
}

async function setRollout(name, value) {
  const pct = parseInt(value, 10);
  if (isNaN(pct) || pct < 0 || pct > 100) {
    toast('Rollout must be between 0 and 100', 'err');
    refresh();
    return;
  }
  try {
    await api(`/flags/${encodeURIComponent(name)}/rollout`, {
      method: 'PUT',
      body: JSON.stringify({ rollout_percent: pct }),
    });
    toast(`Rollout for ${name} set to ${pct}%`, 'ok');
    refresh();
  } catch (err) {
    toast('Failed to set rollout: ' + err.message, 'err');
  }
}

async function toggleKill(name) {
  const f = flags.find((x) => x.name === name);
  if (!f) return;
  const next = !f.kill_switch;
  try {
    await api(`/flags/${encodeURIComponent(name)}/kill-switch`, {
      method: 'POST',
      body: JSON.stringify({ kill_switch: next }),
    });
    toast(next ? `Kill switch ON for ${name}` : `Kill switch cleared for ${name}`,
      next ? 'err' : 'ok');
    refresh();
  } catch (err) {
    toast('Failed to toggle kill switch: ' + err.message, 'err');
  }
}

async function removeFlag(name) {
  if (!confirm(`Delete flag “${name}”?`)) return;
  try {
    await api(`/flags/${encodeURIComponent(name)}`, { method: 'DELETE' });
    toast('Flag deleted', 'ok');
    refresh();
  } catch (err) {
    toast('Failed to delete: ' + err.message, 'err');
  }
}

async function mutate(name, payload, verb) {
  try {
    await api(`/flags/${encodeURIComponent(name)}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    });
    toast(`${name} ${verb}`, 'ok');
    refresh();
  } catch (err) {
    toast('Update failed: ' + err.message, 'err');
  }
}

// ---------------------------------------------------------------- events

el.form.addEventListener('submit', createFlag);
el.refreshBtn.addEventListener('click', refresh);

el.autoRefresh.addEventListener('change', () => {
  if (el.autoRefresh.checked) {
    timer = setInterval(refresh, REFRESH_MS);
    el.refreshNote.textContent = 'every 5s';
  } else {
    clearInterval(timer);
    timer = null;
    el.refreshNote.textContent = 'paused';
  }
});

el.body.addEventListener('click', (ev) => {
  const btn = ev.target.closest('button[data-act]');
  if (!btn) return;
  const name = btn.closest('tr').dataset.name;
  if (btn.dataset.act === 'toggle') toggleEnabled(name);
  if (btn.dataset.act === 'kill') toggleKill(name);
  if (btn.dataset.act === 'delete') removeFlag(name);
});

el.body.addEventListener('change', (ev) => {
  const input = ev.target.closest('input[data-field="rollout"]');
  if (!input) return;
  setRollout(input.closest('tr').dataset.name, input.value);
});

refresh();
timer = setInterval(refresh, REFRESH_MS);
