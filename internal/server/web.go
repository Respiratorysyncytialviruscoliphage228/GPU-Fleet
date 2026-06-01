package server

const dashboardHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>GPUFleet</title>
  <style>
    :root {
      color-scheme: light dark;
      --bg: #f6f7f9;
      --panel: #ffffff;
      --text: #17202a;
      --muted: #657282;
      --line: #d9dee7;
      --good: #168a4a;
      --warn: #ad6a00;
      --bad: #c53232;
      --accent: #1769aa;
      --shadow: 0 14px 32px rgba(23,32,42,.08);
    }
    @media (prefers-color-scheme: dark) {
      :root {
        --bg: #111418;
        --panel: #191f26;
        --text: #eef2f6;
        --muted: #a5b0bd;
        --line: #2c3540;
        --accent: #62a8e5;
        --shadow: 0 18px 40px rgba(0,0,0,.28);
      }
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      background: var(--bg);
      color: var(--text);
      letter-spacing: 0;
    }
    .app {
      min-height: 100vh;
      display: grid;
      grid-template-columns: 240px 1fr;
    }
    .side {
      padding: 24px 18px;
      border-right: 1px solid var(--line);
      background: color-mix(in srgb, var(--panel) 86%, transparent);
      position: sticky;
      top: 0;
      height: 100vh;
    }
    .brand {
      display: flex;
      align-items: center;
      gap: 10px;
      font-weight: 800;
      font-size: 20px;
      margin-bottom: 28px;
    }
    .mark {
      width: 30px;
      height: 30px;
      border-radius: 8px;
      background: var(--accent);
      display: grid;
      place-items: center;
      color: #fff;
      font-weight: 900;
    }
    .nav {
      display: grid;
      gap: 8px;
    }
    .nav button, .icon-btn, .primary {
      border: 1px solid var(--line);
      background: var(--panel);
      color: var(--text);
      border-radius: 8px;
      min-height: 40px;
      cursor: pointer;
      transition: transform .16s ease, border-color .16s ease, background .16s ease;
    }
    .nav button {
      width: 100%;
      text-align: left;
      padding: 10px 12px;
      display: flex;
      align-items: center;
      gap: 10px;
    }
    .nav button.active, .primary {
      background: var(--accent);
      border-color: var(--accent);
      color: #fff;
    }
    button:hover { transform: translateY(-1px); }
    .main {
      padding: 24px;
      display: grid;
      gap: 18px;
      align-content: start;
    }
    .topbar {
      display: flex;
      justify-content: space-between;
      gap: 16px;
      align-items: center;
    }
    h1 {
      font-size: clamp(24px, 4vw, 34px);
      line-height: 1.1;
      margin: 0;
    }
    .sub {
      color: var(--muted);
      margin-top: 6px;
      font-size: 14px;
    }
    .grid {
      display: grid;
      gap: 14px;
    }
    .stats {
      grid-template-columns: repeat(5, minmax(140px, 1fr));
    }
    .columns {
      grid-template-columns: minmax(0, 1.35fr) minmax(320px, .65fr);
      align-items: start;
    }
    .panel, .card {
      background: var(--panel);
      border: 1px solid var(--line);
      border-radius: 8px;
      box-shadow: var(--shadow);
    }
    .panel {
      padding: 16px;
    }
    .stat {
      padding: 14px;
    }
    .label {
      color: var(--muted);
      font-size: 12px;
      text-transform: uppercase;
    }
    .value {
      font-size: 26px;
      font-weight: 800;
      margin-top: 8px;
      overflow-wrap: anywhere;
    }
    .row {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
      padding: 12px 0;
      border-bottom: 1px solid var(--line);
    }
    .row:last-child { border-bottom: 0; }
    .pill {
      border-radius: 999px;
      padding: 4px 8px;
      font-size: 12px;
      border: 1px solid var(--line);
      color: var(--muted);
      white-space: nowrap;
    }
    .online { color: var(--good); border-color: color-mix(in srgb, var(--good) 45%, var(--line)); }
    .offline { color: var(--bad); border-color: color-mix(in srgb, var(--bad) 45%, var(--line)); }
    .warning { color: var(--warn); border-color: color-mix(in srgb, var(--warn) 45%, var(--line)); }
    .gpu-list {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
      gap: 12px;
    }
    .gpu {
      padding: 14px;
    }
    .meter {
      height: 9px;
      background: color-mix(in srgb, var(--line) 55%, transparent);
      border-radius: 999px;
      overflow: hidden;
      margin-top: 12px;
    }
    .meter span {
      display: block;
      height: 100%;
      background: var(--accent);
      width: var(--v);
    }
    .chart {
      height: 210px;
      width: 100%;
      display: block;
      margin-top: 10px;
    }
    .login {
      min-height: 100vh;
      display: grid;
      place-items: center;
      padding: 24px;
    }
    .login form {
      width: min(420px, 100%);
      padding: 24px;
    }
    input {
      width: 100%;
      min-height: 42px;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: var(--panel);
      color: var(--text);
      padding: 10px 12px;
      margin-top: 8px;
      margin-bottom: 14px;
    }
    .hidden { display: none !important; }
    @media (max-width: 980px) {
      .app { grid-template-columns: 1fr; }
      .side {
        height: auto;
        position: static;
        border-right: 0;
        border-bottom: 1px solid var(--line);
        padding: 14px;
      }
      .nav {
        grid-template-columns: repeat(4, minmax(0, 1fr));
      }
      .nav button {
        justify-content: center;
        text-align: center;
        font-size: 13px;
      }
      .columns, .stats { grid-template-columns: 1fr 1fr; }
      .main { padding: 16px; }
    }
    @media (max-width: 620px) {
      .topbar { align-items: flex-start; flex-direction: column; }
      .columns, .stats { grid-template-columns: 1fr; }
      .nav { grid-template-columns: repeat(2, minmax(0, 1fr)); }
      .value { font-size: 23px; }
    }
  </style>
</head>
<body>
  <div id="login" class="login">
    <form class="panel" id="loginForm">
      <div class="brand"><span class="mark">G</span><span>GPUFleet</span></div>
      <h1>登录面板</h1>
      <p class="sub">使用服务端启动时配置或生成的管理员密码。</p>
      <label>用户名<input name="username" value="admin" autocomplete="username"></label>
      <label>密码<input name="password" type="password" autocomplete="current-password"></label>
      <button class="primary" type="submit" style="width:100%">登录</button>
      <p class="sub" id="loginError"></p>
    </form>
  </div>
  <div id="app" class="app hidden">
    <aside class="side">
      <div class="brand"><span class="mark">G</span><span>GPUFleet</span></div>
      <nav class="nav">
        <button class="active">总览</button>
        <button>设备</button>
        <button>GPU</button>
        <button>设置</button>
      </nav>
    </aside>
    <main class="main">
      <section class="topbar">
        <div>
          <h1>GPU 资源总览</h1>
          <div class="sub" id="serverTime">等待数据</div>
        </div>
        <button id="refresh" class="icon-btn" title="刷新" style="width:42px">↻</button>
      </section>
      <section class="grid stats">
        <div class="panel stat"><div class="label">在线设备</div><div class="value" id="onlineDevices">0</div></div>
        <div class="panel stat"><div class="label">GPU 数量</div><div class="value" id="gpuCount">0</div></div>
        <div class="panel stat"><div class="label">平均利用率</div><div class="value" id="avgUtil">0%</div></div>
        <div class="panel stat"><div class="label">显存占用</div><div class="value" id="memUse">0%</div></div>
        <div class="panel stat"><div class="label">磁盘保护</div><div class="value" id="diskStatus">OK</div></div>
      </section>
      <section class="grid columns">
        <div class="panel">
          <div class="row"><strong>GPU 状态</strong><span class="pill" id="gpuUpdated">-</span></div>
          <div class="gpu-list" id="gpuList"></div>
          <canvas id="chart" class="chart" width="900" height="230"></canvas>
        </div>
        <div class="panel">
          <div class="row"><strong>设备</strong><span class="pill" id="deviceCount">0</span></div>
          <div id="deviceList"></div>
        </div>
      </section>
    </main>
  </div>
  <script>
    const login = document.getElementById('login');
    const app = document.getElementById('app');
    const fmtBytes = (n) => {
      if (!n) return '0 B';
      const units = ['B','KiB','MiB','GiB','TiB'];
      let i = 0, v = n;
      while (v >= 1024 && i < units.length - 1) { v /= 1024; i++; }
      return v.toFixed(i ? 1 : 0) + ' ' + units[i];
    };
    const pct = (n) => Number.isFinite(n) ? Math.round(n) + '%' : '-';
    const api = async (url, options = {}) => {
      const res = await fetch(url, {
        headers: {'Content-Type':'application/json', ...(options.headers || {})},
        credentials: 'same-origin',
        ...options
      });
      if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || res.statusText);
      return res.json();
    };
    document.getElementById('loginForm').addEventListener('submit', async (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      try {
        await api('/api/v1/auth/login', {
          method: 'POST',
          body: JSON.stringify({username: form.get('username'), password: form.get('password')})
        });
        login.classList.add('hidden');
        app.classList.remove('hidden');
        refresh();
      } catch (err) {
        document.getElementById('loginError').textContent = err.message;
      }
    });
    document.getElementById('refresh').addEventListener('click', refresh);
    async function refresh() {
      try {
        const data = await api('/api/v1/overview');
        login.classList.add('hidden');
        app.classList.remove('hidden');
        render(data);
      } catch (err) {
        if (err.message.includes('login')) {
          app.classList.add('hidden');
          login.classList.remove('hidden');
        }
      }
    }
    function render(data) {
      document.getElementById('serverTime').textContent = '服务端时间 ' + new Date(data.server_time).toLocaleString();
      document.getElementById('onlineDevices').textContent = data.online_device_count + ' / ' + data.device_count;
      document.getElementById('gpuCount').textContent = data.gpu_count;
      document.getElementById('avgUtil').textContent = pct(data.average_utilization);
      document.getElementById('memUse').textContent = data.memory_total_bytes ? pct(data.memory_used_bytes / data.memory_total_bytes * 100) : '0%';
      const disk = document.getElementById('diskStatus');
      disk.textContent = data.disk.status.toUpperCase();
      disk.style.color = data.disk.status === 'critical' ? 'var(--bad)' : data.disk.status === 'warning' ? 'var(--warn)' : 'var(--good)';
      document.getElementById('deviceCount').textContent = data.devices.length;
      document.getElementById('gpuUpdated').textContent = data.latest_gpus.length ? new Date(data.latest_gpus[0].timestamp).toLocaleTimeString() : '-';
      renderDevices(data.devices);
      renderGPUs(data.latest_gpus);
      drawChart(data.latest_gpus);
    }
    function renderDevices(devices) {
      const list = document.getElementById('deviceList');
      list.innerHTML = devices.map(d => {
        const status = d.status || 'offline';
        return '<div class="row"><div><strong>' + esc(d.alias || d.id) + '</strong><div class="sub">' +
          esc([d.hostname, d.os, d.agent_version].filter(Boolean).join(' · ') || d.id) +
          '</div></div><span class="pill ' + status + '">' + status + '</span></div>';
      }).join('') || '<p class="sub">暂无设备</p>';
    }
    function renderGPUs(items) {
      const list = document.getElementById('gpuList');
      list.innerHTML = items.map(item => {
        const gpu = item.gpu;
        const util = gpu.utilization_gpu_percent ?? 0;
        const mem = gpu.memory_total_bytes ? gpu.memory_used_bytes / gpu.memory_total_bytes * 100 : 0;
        const temp = gpu.temperature_celsius ?? null;
        return '<article class="card gpu"><div class="row"><div><strong>' + esc(gpu.name || gpu.gpu_id) +
          '</strong><div class="sub">' + esc(item.device_id + ' · ' + gpu.gpu_id) + '</div></div><span class="pill">' +
          pct(util) + '</span></div><div class="label">GPU 利用率</div><div class="meter" style="--v:' + util +
          '%"><span></span></div><div class="row"><span>显存 ' + pct(mem) + '</span><span>' +
          fmtBytes(gpu.memory_used_bytes) + ' / ' + fmtBytes(gpu.memory_total_bytes) + '</span></div><div class="row"><span>温度</span><span>' +
          (temp === null ? '-' : Math.round(temp) + '°C') + '</span></div><div class="row"><span>功耗</span><span>' +
          (gpu.power_draw_watts == null ? '-' : gpu.power_draw_watts.toFixed(1) + ' W') + '</span></div></article>';
      }).join('') || '<p class="sub">等待 Agent 上报 GPU 数据</p>';
    }
    function drawChart(items) {
      const canvas = document.getElementById('chart');
      const ctx = canvas.getContext('2d');
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      ctx.strokeStyle = getComputedStyle(document.documentElement).getPropertyValue('--line');
      ctx.lineWidth = 1;
      for (let i = 0; i <= 4; i++) {
        const y = 20 + i * 45;
        ctx.beginPath(); ctx.moveTo(36, y); ctx.lineTo(canvas.width - 20, y); ctx.stroke();
      }
      const bars = items.slice(0, 24);
      const w = Math.max(18, (canvas.width - 70) / Math.max(1, bars.length) - 8);
      bars.forEach((item, i) => {
        const util = item.gpu.utilization_gpu_percent ?? 0;
        const h = util / 100 * 170;
        const x = 44 + i * (w + 8);
        const y = 200 - h;
        ctx.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--accent');
        ctx.fillRect(x, y, w, h);
      });
    }
    function esc(value) {
      return String(value ?? '').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
    }
    refresh();
    setInterval(refresh, 10000);
  </script>
</body>
</html>`
