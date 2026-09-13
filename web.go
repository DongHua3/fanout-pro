package main

import "net/http"

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	_, _ = w.Write([]byte(indexHTML))
}

const indexHTML = `<!DOCTYPE html>
<html lang="zh-CN" class="dark">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>fanout • 工业级控制台</title>
<script>
(function(){
  try{
    var t = localStorage.getItem('fanout_theme');
    if(!t){
      t = window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark';
    }
    document.documentElement.className = t === 'light' ? 'light' : 'dark';
  }catch(e){}
})();
</script>
<style>
/* ==========================================================================
   高对比度设计系统 (Dark High-Contrast / Light Tech)
   ========================================================================== */
:root {
  --bg-canvas: #0b0f19;
  --bg-header: rgba(11, 15, 25, 0.92);
  --bg-card: #111827;
  --bg-card-hover: #162032;
  --bg-surface: #172033;
  --bg-surface-sub: #1e293b;
  --bg-input: #0e1524;
  
  --border-card: #1f293d;
  --border-card-hover: #334155;
  --border-subtle: #1e293b;
  --border-strong: #334155;

  --text-main: #ffffff;
  --text-body: #e2e8f0;
  --text-muted: #94a3b8;
  --text-dim: #64748b;

  --accent-blue: #38bdf8;
  --accent-blue-bg: rgba(56, 189, 248, 0.12);
  --accent-blue-border: rgba(56, 189, 248, 0.3);

  --accent-purple: #a78bfa;
  --accent-purple-bg: rgba(167, 139, 250, 0.12);
  --accent-purple-border: rgba(167, 139, 250, 0.3);

  --status-up: #10b981;
  --status-up-bg: rgba(16, 185, 129, 0.12);
  --status-up-border: rgba(16, 185, 129, 0.3);

  --status-warn: #f59e0b;
  --status-warn-bg: rgba(245, 158, 11, 0.12);
  --status-warn-border: rgba(245, 158, 11, 0.3);

  --status-danger: #f43f5e;
  --status-danger-bg: rgba(244, 63, 94, 0.12);
  --status-danger-border: rgba(244, 63, 94, 0.3);

  --icon-stroke: #ffffff;
  --icon-stroke-dim: rgba(255, 255, 255, 0.75);

  --font-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Inter", "PingFang SC", sans-serif;
  --font-mono: "JetBrains Mono", ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;

  --radius-xs: 4px;
  --radius-sm: 8px;
  --radius-md: 12px;
  --radius-lg: 16px;
  --radius-full: 9999px;

  --shadow-card: 0 4px 20px -2px rgba(0, 0, 0, 0.45);
  --shadow-modal: 0 25px 50px -12px rgba(0, 0, 0, 0.7);

  /* 兼容既有组件变量 */
  --bg: var(--bg-canvas);
  --panel: var(--bg-card);
  --line: var(--border-card);
  --text: var(--text-body);
  --dim: var(--text-muted);
  --accent: var(--accent-blue);
  --ok: var(--status-up);
  --warn: var(--status-warn);
  --bad: var(--status-danger);
}

html.light {
  --bg-canvas: #f8fafc;
  --bg-header: rgba(255, 255, 255, 0.92);
  --bg-card: #ffffff;
  --bg-card-hover: #f1f5f9;
  --bg-surface: #f1f5f9;
  --bg-surface-sub: #e2e8f0;
  --bg-input: #ffffff;
  
  --border-card: #e2e8f0;
  --border-card-hover: #cbd5e1;
  --border-subtle: #e2e8f0;
  --border-strong: #cbd5e1;

  --text-main: #0f172a;
  --text-body: #334155;
  --text-muted: #64748b;
  --text-dim: #94a3b8;

  --icon-stroke: #0f172a;
  --icon-stroke-dim: #475569;

  --accent-blue: #0284c7;
  --accent-blue-bg: #e0f2fe;
  --accent-blue-border: #bae6fd;

  --accent-purple: #7c3aed;
  --accent-purple-bg: #ede9fe;
  --accent-purple-border: #ddd6fe;

  --status-up: #059669;
  --status-up-bg: #d1fae5;
  --status-up-border: #a7f3d0;

  --status-warn: #d97706;
  --status-warn-bg: #fef3c7;
  --status-warn-border: #fde68a;

  --status-danger: #e11d48;
  --status-danger-bg: #ffe4e6;
  --status-danger-border: #fecdd3;

  --shadow-card: 0 1px 3px 0 rgba(0, 0, 0, 0.08), 0 1px 2px -1px rgba(0, 0, 0, 0.08);
  --shadow-modal: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 8px 10px -6px rgba(0, 0, 0, 0.1);

  --bg: var(--bg-canvas);
  --panel: var(--bg-card);
  --line: var(--border-card);
  --text: var(--text-body);
  --dim: var(--text-muted);
  --accent: var(--accent-blue);
  --ok: var(--status-up);
  --warn: var(--status-warn);
  --bad: var(--status-danger);
}

* { box-sizing: border-box; margin: 0; padding: 0; }
body {
  margin: 0;
  background: var(--bg-canvas);
  color: var(--text-body);
  font-family: var(--font-sans);
  font-size: 13px;
  line-height: 1.5;
  letter-spacing: -0.01em;
  min-height: 100vh;
  -webkit-font-smoothing: antialiased;
  transition: background-color 0.2s, color 0.2s;
}
.mono {
  font-family: var(--font-mono);
  font-variant-numeric: tabular-nums;
}

/* 纯白矢量 SVG 图标与发光微动效 */
.icon, .icon-sm, .icon-xs {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  stroke-width: 1.8;
  stroke: var(--icon-stroke);
  fill: none;
  stroke-linecap: round;
  stroke-linejoin: round;
  flex-shrink: 0;
  transition: stroke 0.15s, filter 0.15s;
}
.icon { width: 16px; height: 16px; }
.icon-sm { width: 14px; height: 14px; stroke-width: 1.8; }
.icon-xs { width: 12px; height: 12px; stroke-width: 2; }

.btn:hover .icon, .btn:hover .icon-sm, .btn:hover .icon-xs,
button:hover .icon, button:hover .icon-sm, button:hover .icon-xs,
.copy-btn:hover .icon-xs, .socks-badge:hover .icon-xs,
.theme-toggle:hover .icon, .theme-toggle:hover .icon-sm {
  filter: drop-shadow(0 0 5px rgba(255, 255, 255, 0.7));
}
html.light .btn:hover .icon, html.light .btn:hover .icon-sm, html.light .btn:hover .icon-xs,
html.light button:hover .icon, html.light button:hover .icon-sm, html.light button:hover .icon-xs,
html.light .copy-btn:hover .icon-xs, html.light .socks-badge:hover .icon-xs,
html.light .theme-toggle:hover .icon, html.light .theme-toggle:hover .icon-sm {
  filter: drop-shadow(0 0 3px rgba(0, 0, 0, 0.3));
}

svg {
  width: 14px; height: 14px;
  stroke: var(--icon-stroke);
  fill: none;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
  flex: none;
  transition: stroke 0.15s, filter 0.15s;
}

/* 按钮与基础控制 */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  padding: 7px 14px;
  font-size: 13px;
  font-weight: 500;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-card);
  background: var(--bg-card);
  color: var(--text-main);
  cursor: pointer;
  transition: all 0.15s;
  text-decoration: none;
  white-space: nowrap;
}
.btn:hover {
  background: var(--bg-card-hover);
  border-color: var(--border-card-hover);
}
.btn-primary {
  background: #0284c7;
  border-color: #0284c7;
  color: #fff;
  font-weight: 600;
}
.btn-primary:hover {
  background: #0369a1;
  border-color: #0369a1;
}
.btn-crystal {
  background: var(--accent-blue-bg);
  border: 1px solid var(--accent-blue-border);
  color: var(--accent-blue);
  font-weight: 600;
}
.btn-crystal:hover {
  background: var(--accent-blue-border);
}
.btn-ghost {
  background: transparent;
  border-color: transparent;
  color: var(--text-muted);
}
.btn-ghost:hover {
  background: var(--bg-surface);
  color: var(--text-main);
}
.theme-toggle {
  padding: 6px 12px;
  font-size: 12px;
  font-weight: 600;
  border-radius: var(--radius-full);
  border: 1px solid var(--border-card);
  background: var(--bg-surface);
  color: var(--text-main);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: all 0.15s;
}
.theme-toggle:hover {
  border-color: var(--accent-blue);
}

/* 徽章与标签 */
.metric-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: var(--radius-full);
  font-weight: 600;
  letter-spacing: 0.02em;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
.badge-blue { background: var(--accent-blue-bg); color: var(--accent-blue); border: 1px solid var(--accent-blue-border); }
.badge-purple { background: var(--accent-purple-bg); color: var(--accent-purple); border: 1px solid var(--accent-purple-border); }
.badge-green { background: var(--status-up-bg); color: var(--status-up); border: 1px solid var(--status-up-border); }
.badge-amber { background: var(--status-warn-bg); color: var(--status-warn); border: 1px solid var(--status-warn-border); }
.country-tag {
  font-size: 11px;
  font-weight: 800;
  padding: 3px 8px;
  border-radius: var(--radius-xs);
  background: var(--bg-surface);
  border: 1px solid var(--border-strong);
  color: var(--text-main);
  font-family: var(--font-mono);
  letter-spacing: 0.05em;
}
.quality-tag {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: var(--radius-full);
  display: inline-flex;
  align-items: center;
  gap: 5px;
}
.tag-residential { background: var(--status-up-bg); color: var(--status-up); border: 1px solid var(--status-up-border); }
.tag-datacenter { background: var(--bg-surface); color: var(--text-muted); border: 1px solid var(--border-card); }
.dot-indicator { width: 6px; height: 6px; border-radius: 50%; background: currentColor; }
.metric-tag {
  font-size: 11px;
  font-weight: 600;
  font-family: var(--font-mono);
  padding: 2px 8px;
  border-radius: var(--radius-full);
  display: inline-flex;
  align-items: center;
  gap: 4px;
  white-space: nowrap;
}
.tag-speed {
  background: var(--accent-blue-bg);
  color: var(--accent-blue);
  border: 1px solid var(--accent-blue-border);
}
.tag-ping {
  background: var(--bg-surface);
  color: var(--text-muted);
  border: 1px solid var(--border-card);
}
.tag-ping.ping-good {
  background: var(--status-up-bg);
  color: var(--status-up);
  border: 1px solid var(--status-up-border);
}
.tag-ping.ping-med {
  background: var(--status-warn-bg);
  color: var(--status-warn);
  border: 1px solid var(--status-warn-border);
}
.tag-ping.ping-slow {
  background: var(--status-danger-bg);
  color: var(--status-danger);
  border: 1px solid var(--status-danger-border);
}
.icon-nano { width: 12px; height: 12px; stroke: currentColor; fill: none; stroke-width: 2.2; stroke-linecap: round; stroke-linejoin: round; }

header{display:flex;align-items:center;padding:12px 24px;
  border-bottom:1px solid var(--border-card);background:var(--bg-header);
  backdrop-filter:blur(16px);position:sticky;top:0;z-index:40}
.header-inner{max-width:1280px;margin:0 auto;display:flex;align-items:center;justify-content:space-between;gap:16px;width:100%}
.brand{display:flex;align-items:center;gap:10px;text-decoration:none}
.logo-mark{width:28px;height:28px;border-radius:var(--radius-xs);background:linear-gradient(135deg,#0284c7,#4f46e5);
  display:flex;align-items:center;justify-content:center;color:#fff;flex-shrink:0}
.logo-title{font-size:15px;font-weight:700;letter-spacing:-0.02em;color:var(--text-main);display:flex;align-items:center;gap:8px}
.pro-badge{font-size:10px;font-weight:700;padding:1px 6px;border-radius:var(--radius-full);
  background:var(--accent-blue-bg);border:1px solid var(--accent-blue-border);color:var(--accent-blue);
  text-transform:uppercase;letter-spacing:0.05em}
.host-pill{display:inline-flex;align-items:center;gap:8px;padding:4px 12px;background:var(--bg-surface);
  border:1px solid var(--border-card);border-radius:var(--radius-full);font-size:12px;color:var(--text-muted)}
.pulse-dot{width:8px;height:8px;border-radius:50%;background:var(--status-up);box-shadow:0 0 8px var(--status-up)}
.header-actions{display:flex;align-items:center;gap:8px}

/* 指标概览看板 (Hero Metrics Grid) */
.metrics-grid{display:grid;grid-template-columns:repeat(4,1fr);gap:16px;margin-bottom:24px}
.metric-card{background:var(--bg-card);border:1px solid var(--border-card);border-radius:var(--radius-md);
  padding:18px 20px;box-shadow:var(--shadow-card);display:flex;flex-direction:column;justify-content:space-between;transition:all .15s}
.metric-card:hover{border-color:var(--border-card-hover);transform:translateY(-1px)}
.metric-header{display:flex;align-items:center;justify-content:space-between;margin-bottom:10px}
.metric-title{font-size:13px;font-weight:600;color:var(--text-muted);display:flex;align-items:center;gap:7px}
.metric-val{font-size:20px;font-weight:700;letter-spacing:-0.02em;color:var(--text-main);margin-bottom:4px}
.metric-desc{font-size:12px;color:var(--text-muted);display:flex;align-items:center;gap:6px}


/* ==========================================================================
   5. 核心入站与流量拓扑管线 (Section Card & Pipeline Component)
   ========================================================================== */
.section-card {
  background: var(--bg-card);
  border: 1px solid var(--border-card);
  border-radius: var(--radius-lg);
  padding: 20px 22px;
  box-shadow: var(--shadow-card);
  margin-bottom: 24px;
}
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-card);
}
.section-title {
  font-size: 15px;
  font-weight: 700;
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-main);
}
.section-subtitle {
  font-size: 12px;
  color: var(--text-muted);
  font-weight: 400;
  margin-left: 6px;
}
.pipeline-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  background: var(--bg-canvas);
  border: 1px solid var(--border-card);
  border-radius: var(--radius-md);
  padding: 18px 22px;
  margin-bottom: 16px;
}
.pipeline-stage {
  flex: 1;
  background: var(--bg-card);
  border: 1px solid var(--border-card);
  border-radius: var(--radius-md);
  padding: 16px;
  transition: all 0.2s;
}
.pipeline-stage.core {
  border-color: var(--accent-purple-border);
  background: var(--bg-card);
}
.pipeline-stage.exit-active {
  border-color: var(--status-up-border);
  background: var(--bg-card);
}
.pipeline-stage.exit-direct {
  border-color: var(--status-warn-border);
  background: var(--bg-card);
}
.pipeline-stage-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  gap: 6px;
}
.pipeline-stage-main {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-main);
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 4px;
}
.pipeline-stage-sub {
  font-size: 12px;
  color: var(--text-muted);
}
.pipeline-flow {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 0 6px;
  color: var(--text-dim);
  min-width: 70px;
}
.flow-line {
  width: 100%;
  height: 2px;
  background: var(--border-strong);
  position: relative;
  overflow: hidden;
  border-radius: 2px;
}
.flow-line::after {
  content: "";
  position: absolute;
  top: 0; left: -100%;
  width: 100%; height: 100%;
  background: linear-gradient(90deg, transparent, var(--accent-blue), transparent);
  animation: flowLight 2s infinite linear;
}
@keyframes flowLight {
  0% { left: -100%; }
  100% { left: 100%; }
}
.flow-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: var(--radius-full);
  background: var(--accent-blue-bg);
  border: 1px solid var(--accent-blue-border);
  color: var(--accent-blue);
  white-space: nowrap;
}
.pipeline-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  background: var(--bg-surface);
  padding: 12px 16px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-card);
}
.select-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}
.select-label {
  font-size: 12px;
  color: var(--text-main);
  font-weight: 600;
}
.pipeline-toolbar .select {
  background: var(--bg-input);
  border: 1px solid var(--border-strong);
  color: var(--text-main);
  padding: 6px 12px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  outline: none;
  cursor: pointer;
  max-width: 480px;
}
.pipeline-toolbar .select:focus {
  border-color: var(--accent-blue);
}
.subinbounds-area {
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px dashed var(--border-card);
}
.subinbounds-title {
  font-size: 12px;
  color: var(--text-muted);
  font-weight: 600;
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  gap: 6px;
}
.subinbound-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: var(--bg-surface);
  border: 1px solid var(--border-card);
  border-radius: var(--radius-sm);
  margin-bottom: 6px;
  font-size: 13px;
}

h1{font-size:14px;font-weight:700;margin:0;letter-spacing:0}
.spacer{flex:1}
button{font:inherit;color:var(--text-main);background:var(--bg-surface);border:1px solid var(--border-card);
  border-radius:var(--radius-xs);padding:5px 12px;cursor:pointer;display:inline-flex;
  align-items:center;gap:6px;white-space:nowrap;transition:all .15s}
button:hover:not(:disabled){border-color:var(--accent-blue);background:var(--bg-surface-sub)}
button:disabled{opacity:.45;cursor:default}
button.primary{background:var(--accent-blue);border-color:var(--accent-blue);color:#fff;font-weight:600}
button.primary:hover:not(:disabled){background:#0284c7;border-color:#0284c7}
button.icon{padding:4px 6px;background:transparent;border-color:transparent;color:var(--text-muted)}
button.icon:hover:not(:disabled){color:var(--text-main);border-color:var(--border-card);background:var(--bg-surface)}
button.icon.danger:hover:not(:disabled){color:var(--status-danger);border-color:var(--status-danger-border);background:var(--status-danger-bg)}
main{padding:20px 24px 60px;max-width:1280px;margin:0 auto}
/* ==========================================================================
   6. 出口隧道工具栏与卡片列表 (Toolbar & Exit Cards)
   ========================================================================== */
.toolbar{display:flex;align-items:center;justify-content:space-between;margin-bottom:14px;flex-wrap:wrap;gap:10px}
.toolbar-left{display:flex;align-items:center;gap:10px}
.pill-count{font-size:12px;padding:2px 8px;border-radius:var(--radius-full);background:var(--bg-surface);color:var(--text-muted);font-weight:600;border:1px solid var(--border-card)}
.toolbar-right{display:flex;align-items:center;gap:8px}

.exit-list{display:flex;flex-direction:column;gap:10px}
.exit-card{background:var(--bg-card);border:1px solid var(--border-card);border-radius:var(--radius-md);
  padding:14px 20px;box-shadow:var(--shadow-card);display:flex;align-items:center;justify-content:space-between;gap:16px;transition:all .15s}
.exit-main{display:flex;align-items:center;gap:14px;flex:1.3;min-width:0}
.exit-info{display:flex;flex-direction:column;gap:5px;min-width:0}
.exit-ip-row{font-size:14px;font-weight:700;color:var(--text-main);display:flex;align-items:center;gap:8px}
.exit-metrics-row{display:flex;align-items:center;gap:8px;flex-wrap:nowrap}
.copy-btn{font-size:11px;padding:2px 6px;border-radius:4px;background:var(--bg-surface);border:1px solid var(--border-card);color:var(--text-muted);cursor:pointer;display:inline-flex;align-items:center;gap:4px;transition:all .15s}
.copy-btn:hover{background:var(--bg-surface-sub);color:var(--text-main)}
.exit-meta{font-size:12px;color:var(--text-muted);display:flex;align-items:center;gap:8px}
.exit-time{display:inline-flex;align-items:center;gap:5px;color:var(--status-up);font-size:11px;font-weight:500}
.exit-bindings{flex:1.4;display:flex;align-items:center;gap:8px;flex-wrap:wrap}
.bound-pill{display:inline-flex;align-items:center;padding:3px 8px;border-radius:var(--radius-sm);background:var(--accent-blue-bg);border:1px solid var(--accent-blue-border);color:var(--accent-blue);font-size:12px;font-weight:600;gap:6px}
.bound-pill.core{background:var(--accent-purple-bg);border-color:var(--accent-purple-border);color:var(--accent-purple)}
.bound-unbind{font-size:12px;cursor:pointer;padding:1px 4px;border-radius:3px;color:var(--text-muted);transition:all .15s;line-height:1}
.bound-unbind:hover{background:var(--status-danger);color:#fff}
.btn-add-bind{border:1px dashed var(--border-strong);background:transparent;padding:3px 8px;font-size:12px;color:var(--text-muted);border-radius:var(--radius-sm);cursor:pointer;display:inline-flex;align-items:center;gap:4px;transition:all .15s}
.btn-add-bind:hover{border-color:var(--accent-blue);color:var(--accent-blue)}
.exit-actions{display:flex;align-items:center;gap:8px}
.socks-badge{display:inline-flex;align-items:center;gap:5px;padding:4px 10px;border-radius:var(--radius-sm);background:var(--bg-surface);border:1px solid var(--border-card);color:var(--text-muted);font-size:12px;cursor:pointer;transition:all .15s}
.socks-badge:hover{background:var(--bg-surface-sub);color:var(--text-main);border-color:var(--border-strong)}
.errline{padding:0 12px 9px;color:var(--status-danger);font-size:11px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.empty{border:1px dashed var(--border-card);border-radius:var(--radius-md);padding:40px 20px;text-align:center;color:var(--text-muted)}
.empty button{margin-top:14px}
.jobs{margin-bottom:12px}
.job{border:1px solid var(--border-card);border-radius:var(--radius-sm);background:var(--bg-card);
  padding:10px 12px;margin-bottom:8px}
.job .top{display:flex;align-items:center;gap:10px;margin-bottom:8px}
.job .top strong{font-weight:600;font-size:12px;color:var(--text-main)}
.steps{display:flex;flex-wrap:wrap;gap:6px}
.step{display:flex;align-items:center;gap:5px;font-size:11px;color:var(--text-muted);
  border:1px solid var(--border-card);border-radius:var(--radius-xs);padding:2px 7px;background:var(--bg-surface)}
.step.ok{color:var(--ok);border-color:rgba(63,166,107,.35)}
.step.failed{color:var(--bad);border-color:rgba(194,84,80,.35)}
.step.running{color:var(--warn);border-color:rgba(201,144,58,.35)}
.spin{animation:rot 1s linear infinite;transform-origin:center}
@keyframes rot{to{transform:rotate(360deg)}}
.links{display:flex;gap:14px;margin-right:4px}
.links a{color:var(--dim);text-decoration:none;font-size:12px}
.links a:hover{color:var(--accent)}
.modal{position:fixed;inset:0;background:rgba(0,0,0,.75);backdrop-filter:blur(8px);display:none;
  align-items:center;justify-content:center;z-index:50;padding:20px;animation:modalFadeIn .18s ease-out}
.modal.open{display:flex}
.sheet{background:var(--bg-card);border:1px solid var(--border-card);border-radius:var(--radius-lg);
  box-shadow:var(--shadow-modal);width:min(680px,100%);max-height:88vh;display:flex;flex-direction:column;overflow:hidden;
  animation:modalScaleIn .18s ease-out}
@keyframes modalFadeIn{from{opacity:0}to{opacity:1}}
@keyframes modalScaleIn{from{transform:scale(.97);opacity:0}to{transform:scale(1);opacity:1}}
.sheet .head{display:flex;align-items:center;gap:10px;padding:14px 20px;
  border-bottom:1px solid var(--border-card);background:var(--bg-card)}
.sheet .head h2{font-size:14px;margin:0;font-weight:700;color:var(--text-main);display:flex;align-items:center;gap:8px}
.sheet .body{overflow:auto;padding:20px;background:var(--bg-canvas)}
.sheet .foot{display:flex;align-items:center;gap:10px;padding:14px 20px;
  border-top:1px solid var(--border-card);background:var(--bg-card)}
/* 自定义确认操作弹窗 (Custom Confirm Modal) */
.confirm-sheet{max-width:440px;width:92%;padding:22px 24px;border-radius:var(--radius-md)}
.confirm-head{display:flex;align-items:flex-start;gap:16px;margin-bottom:20px}
.confirm-icon-box{width:42px;height:42px;border-radius:var(--radius-sm);
  background:var(--accent-blue-bg);border:1px solid var(--accent-blue-border);
  color:var(--accent-blue);display:flex;align-items:center;justify-content:center;flex-shrink:0}
.confirm-icon-box.danger{background:var(--status-danger-bg);border-color:var(--status-danger-border);color:var(--status-danger)}
.confirm-content h3{margin:0 0 6px;font-size:15px;font-weight:700;color:var(--text-main)}
.confirm-msg{margin:0;font-size:13px;color:var(--text-muted);line-height:1.5;word-break:break-word}
.confirm-actions{display:flex;align-items:center;justify-content:flex-end;gap:10px}
.btn-danger{background:var(--status-danger);border-color:var(--status-danger);color:#fff;font-weight:600}
.btn-danger:hover:not(:disabled){background:#dc2626;border-color:#dc2626}
.count{color:var(--text-muted);font-size:11px}
label.f{display:block;margin-bottom:16px}
label.f[hidden]{display:none}
label.f>span{display:block;color:var(--text-muted);font-size:11px;margin-bottom:6px}
.regions{display:grid;grid-template-columns:repeat(auto-fill,minmax(148px,1fr));
  gap:6px;max-height:224px;overflow:auto}
.rg{border:1px solid var(--border-card);background:var(--bg-surface);border-radius:var(--radius-xs);padding:7px 9px;
  cursor:pointer;text-align:left;display:block;width:100%;color:var(--text-main);transition:all .15s}
.rg:hover{border-color:var(--accent-blue);background:var(--bg-surface-sub)}
.rg.sel{border-color:var(--accent-blue);background:var(--accent-blue-bg)}
.rg b{font-weight:600;font-size:12px;display:block;overflow:hidden;
  text-overflow:ellipsis;white-space:nowrap;color:var(--text-main)}
.rg em{display:block;font-style:normal;color:var(--text-muted);font-size:11px;margin-top:2px}
.stepper{display:flex;align-items:center;gap:0;width:fit-content;
  border:1px solid var(--border-card);border-radius:var(--radius-xs);overflow:hidden;background:var(--bg-surface)}
.stepper button{border:0;border-radius:0;background:transparent;padding:5px 11px;color:var(--text-main)}
select,input[type=search],input[type=text]{font:inherit;background:var(--bg-input);
  border:1px solid var(--border-card);color:var(--text-main);border-radius:var(--radius-xs);
  padding:6px 10px;width:100%;transition:border-color .15s}
select:focus,input[type=search]:focus,input[type=text]:focus{outline:none;border-color:var(--accent-blue)}
.stepper input[type=text]{width:56px;text-align:center;font:inherit;background:transparent;
  border:0;border-left:1px solid var(--border-card);border-right:1px solid var(--border-card);
  color:var(--text-main);padding:5px 0;font-variant-numeric:tabular-nums}
.stepper input:focus{outline:none}
.hint{color:var(--text-muted);font-size:11px;margin-top:6px}
.setrow{display:grid;grid-template-columns:1fr 1fr;gap:12px;margin-top:16px}
.setrow input,.setrow select{width:100%}
.updsec{margin-top:18px;padding-top:14px;border-top:1px solid var(--line)}
.updrow{display:flex;align-items:center;gap:10px}
.updver{font-size:12px;color:var(--text)}
.updver b{font-weight:600}
.updver span{color:var(--dim);margin-left:8px}
.updnotes{margin-top:10px;padding:10px;background:var(--bg-surface);border:1px solid var(--border-card);
  border-radius:4px;font-size:12px;line-height:1.6;color:var(--text-muted);white-space:pre-wrap;
  max-height:180px;overflow:auto}
label.chk{display:flex;align-items:center;gap:7px;color:var(--text-main);font-size:12px;
  cursor:pointer;margin:0}
label.chk input{margin:0}
.hint.bad{color:var(--bad)}
.kv{display:grid;grid-template-columns:76px 1fr;gap:5px 12px;margin:0 0 14px}
.kv dt{color:var(--text-muted)}
.kv dd{margin:0;word-break:break-all}
.share{padding:10px;background:var(--bg-surface);border:1px solid var(--border-card);
  border-radius:4px;word-break:break-all;font-size:12px;line-height:1.7;margin-bottom:8px;color:var(--text-main)}
.editbar{display:flex;align-items:flex-end;gap:12px;flex-wrap:wrap;
  padding:12px 0;border-top:1px solid var(--border-card);margin-top:4px}
.ef{display:block}
.ef>span{display:block;color:var(--text-muted);font-size:11px;margin-bottom:4px}
.ef input{width:150px}
.credrow{display:flex;align-items:flex-end;gap:12px;flex-wrap:wrap;margin-bottom:10px}
.credrow .ef input{width:190px}
.chead{display:flex;align-items:center;gap:10px;margin:14px 0 8px;
  padding-top:12px;border-top:1px solid var(--border-card)}
.chead h3{font-size:12px;margin:0;font-weight:600;color:var(--text-muted)}
.client{border:1px solid var(--border-card);border-radius:4px;padding:8px 10px;margin-bottom:8px;background:var(--bg-card)}
.orow{display:flex;align-items:center;gap:10px;padding:6px 0}
.orow select{width:200px}
.crow{display:flex;align-items:center;gap:10px}
.cemail{font-weight:600;font-size:12px;color:var(--text-main)}
.cid{color:var(--text-muted);font-size:11px;overflow:hidden;text-overflow:ellipsis;
  white-space:nowrap;max-width:280px}
.client .share{margin:8px 0 0}
.share button{margin-top:8px}
textarea{width:100%;min-height:300px;background:var(--bg-input);border:1px solid var(--border-card);
  color:var(--text-main);border-radius:4px;
  font:12px/1.8 ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;
  padding:10px 12px;resize:vertical}
textarea:focus{outline:none;border-color:var(--accent-blue)}
.toast{position:fixed;left:50%;bottom:24px;transform:translateX(-50%);
  background:var(--panel);border:1px solid var(--line);border-radius:4px;
  padding:8px 14px;font-size:12px;z-index:80;opacity:0;pointer-events:none;
  transition:opacity .18s}
.toast.show{opacity:1}
.toast.bad{border-color:rgba(194,84,80,.5);color:var(--bad)}

/* 端口推荐与占用指示 */
.port-recom-wrap{margin-top:6px;padding:6px 8px;background:var(--bg-surface);border:1px solid var(--border-card);border-radius:var(--radius-sm)}
.port-recom-title{font-size:11px;color:var(--text-muted);margin-bottom:5px}
.port-chips{display:flex;flex-wrap:wrap;gap:5px}
.pchip{font-size:11px;padding:2px 7px;border-radius:var(--radius-xs);background:var(--bg-surface-sub);border:1px solid var(--border-card);cursor:pointer;color:var(--text-main);display:inline-flex;align-items:center;gap:3px}
.pchip:hover:not(:disabled){border-color:var(--accent-blue);color:var(--accent-blue)}
.pchip.occupied{opacity:.5;border-color:var(--status-danger-border);color:var(--status-danger);cursor:not-allowed}
.pchip.free{border-color:rgba(16,185,129,.4);color:var(--status-up)}
.port-status{font-size:11px;margin-top:5px;min-height:16px}
.port-status.bad{color:var(--status-danger)}
.port-status.ok{color:var(--status-up)}

/* 快速解绑与直连交互 */
.chip-group{display:inline-flex;align-items:center;border:1px solid var(--border-card);border-radius:var(--radius-xs);background:var(--bg-surface);overflow:hidden}
.chip-group .chip{border:none;border-radius:0;background:transparent}
.chip-unbind{padding:1px 5px;background:transparent;border:none;border-left:1px solid var(--border-card);color:var(--text-muted);cursor:pointer;font-size:10px}
.chip-unbind:hover{color:var(--status-danger);background:var(--status-danger-bg)}

.chip.none-btn{border-style:dashed;cursor:pointer;background:var(--accent-blue-bg);color:var(--accent-blue);transition:all .15s}
.chip.none-btn:hover{border-color:var(--accent-blue);background:var(--accent-blue-bg)}
.bmode-item{border:1px solid var(--border-card);border-radius:var(--radius-sm);padding:12px;background:var(--bg-surface);cursor:pointer;transition:border-color .15s}
.bmode-item:hover{border-color:var(--accent-blue)}
.bmode-item.active{border-color:var(--accent-blue);background:var(--accent-blue-bg)}
.bmode-head{display:flex;align-items:center;gap:8px;font-size:13px;color:var(--text-main)}

.unbind-bar{display:flex;align-items:center;gap:10px;padding:8px 12px;background:var(--status-warn-bg);border:1px solid var(--status-warn-border);border-radius:var(--radius-xs);margin-bottom:12px}
.ub-hint{font-size:12px;color:var(--text-main);flex:1}
.btn-unbind{background:var(--status-danger);border-color:var(--status-danger);color:#fff;font-weight:600;font-size:12px;padding:4px 10px;border-radius:var(--radius-xs);cursor:pointer}
.btn-unbind:hover{opacity:.9}
.direct-bar{padding:8px 12px;background:var(--status-up-bg);border:1px solid var(--status-up-border);border-radius:var(--radius-xs);margin-bottom:12px;font-size:12px;color:var(--status-up)}

/* 直连看板 */
.direct-section{margin-top:20px;border:1px solid var(--border-card);border-radius:var(--radius-md);background:var(--bg-card);padding:12px 14px}
.direct-header{display:flex;align-items:center;gap:10px;margin-bottom:10px;flex-wrap:wrap}
.dh-title{display:flex;align-items:center;gap:6px}
.dh-title h3{font-size:13px;margin:0;font-weight:600;color:var(--text-main)}
.dh-ip{font-size:11px;color:var(--text-muted)}
.btn-subtle{font-size:11px;padding:3px 8px;background:var(--bg-surface-sub);border:1px solid var(--border-card);color:var(--text-main);border-radius:var(--radius-xs);cursor:pointer}
.btn-subtle:hover{border-color:var(--accent-blue);color:var(--accent-blue)}

/* 443 核心卡片 */
.core-inbound-card{border:1px solid var(--accent-blue-border);border-radius:var(--radius-sm);background:var(--bg-surface);padding:12px;margin-bottom:10px}
.cic-top{display:flex;align-items:center;gap:10px;flex-wrap:wrap}
.cic-badge{font-size:11px;font-weight:600;padding:2px 6px;border-radius:var(--radius-xs);background:var(--accent-blue-bg);color:var(--accent-blue)}
.cic-title{font-size:13px;font-weight:600;color:var(--text-main)}
.cic-status{font-size:11px;padding:2px 6px;border-radius:var(--radius-xs)}
.cic-status.direct{background:var(--status-up-bg);color:var(--status-up)}
.cic-status.bound{background:var(--status-warn-bg);color:var(--status-warn)}
.cic-acts{display:flex;gap:6px;align-items:center}
.btn-start-core-exit{font-size:11px;padding:3px 8px;border-radius:var(--radius-xs);background:var(--accent-blue);border:1px solid var(--accent-blue);color:#fff;font-weight:600;cursor:pointer}

/* 系统设置卡片 */
.set-card{background:var(--bg-surface);border:1px solid var(--border-card);border-radius:var(--radius-sm);padding:12px 14px}
.set-card-title{font-size:12px;font-weight:700;color:var(--text-main);display:flex;align-items:center;gap:6px;margin-bottom:10px}

/* 全节点大厅 */
.sheet-wide{max-width:960px!important;width:95vw}
.nex-controls{padding:10px 16px;border-bottom:1px solid var(--border-card);background:var(--bg-card);display:flex;flex-direction:column;gap:8px}
.nex-search-row{display:flex;align-items:center;gap:12px;flex-wrap:wrap}
.nex-search-box{display:flex;align-items:center;gap:6px;flex:1;min-width:240px;background:var(--bg-input);border:1px solid var(--border-card);border-radius:var(--radius-xs);padding:4px 8px;position:relative}
.nex-search-box input{border:none;background:transparent;padding:2px 0;width:100%;color:var(--text-main);font:inherit}
.nex-search-box input:focus{outline:none}
.btn-search-clear{background:transparent;border:0;color:var(--text-muted);cursor:pointer;padding:0 4px;font-size:12px;line-height:1}
.btn-search-clear:hover{color:var(--text-main)}
.nex-sort-group{display:flex;align-items:center;gap:6px}
.sort-label{font-size:11px;color:var(--text-muted)}
.nex-sort-btn{font-size:11px;padding:3px 8px;border-radius:var(--radius-xs);background:var(--bg-surface);border:1px solid var(--border-card);color:var(--text-muted);cursor:pointer;transition:all .15s}
.nex-sort-btn.active{background:var(--accent-blue);border-color:var(--accent-blue);color:#fff;font-weight:600}
.nex-filter-row{display:flex;align-items:center;gap:8px;flex-wrap:wrap}
.nex-filter-chip{font-size:11px;padding:2px 8px;border-radius:12px;background:var(--bg-surface);border:1px solid var(--border-card);color:var(--text-muted);cursor:pointer;transition:all .15s}
.nex-filter-chip.active{background:var(--accent-blue-bg);border-color:var(--accent-blue);color:var(--accent-blue);font-weight:600}
.nex-list-wrap{max-height:60vh;overflow-y:auto;padding:12px 16px}
.nex-list{display:flex;flex-direction:column;gap:8px}
.ncard{display:grid;grid-template-columns:auto 1fr auto;gap:10px 14px;align-items:center;padding:10px 14px;background:var(--bg-surface);border:1px solid var(--border-card);border-radius:var(--radius-sm);transition:all .15s}
.ncard:hover{border-color:var(--border-card-hover)}
.ncard-flag{font-size:22px;line-height:1}
.ncard-info{display:flex;flex-direction:column;gap:3px;overflow:hidden}
.ncard-title-row{display:flex;align-items:center;gap:8px;flex-wrap:wrap}
.ncard-host{font-weight:600;font-size:13px;color:var(--text-main)}
.ncard-ip{font-family:inherit;font-size:12px;color:var(--text-muted)}
.qbadge{font-size:11px;padding:1px 6px;border-radius:var(--radius-xs);font-weight:600}
.qbadge.res{background:var(--status-up-bg);color:var(--status-up);border:1px solid var(--status-up-border)}
.qbadge.dc{background:var(--bg-surface-sub);color:var(--text-muted);border:1px solid var(--border-card)}
.qbadge.mob{background:var(--status-warn-bg);color:var(--status-warn);border:1px solid var(--status-warn-border)}
.ncard-meta-row{display:flex;align-items:center;gap:12px;font-size:11px;color:var(--text-muted);flex-wrap:wrap}
.ncard-acts{display:flex;align-items:center;gap:6px}
.btn-start-exit{font-size:11px;padding:4px 8px;border-radius:var(--radius-xs);background:var(--bg-surface-sub);border:1px solid var(--border-card);color:var(--text-main);cursor:pointer;transition:all .15s}
.btn-start-exit:hover{border-color:var(--accent-blue);color:var(--accent-blue)}
.btn-mount-443{font-size:11px;padding:4px 9px;border-radius:var(--radius-xs);background:var(--accent-blue);border:1px solid var(--accent-blue);color:#fff;font-weight:600;cursor:pointer;transition:all .15s}
.btn-mount-443:hover{opacity:.9}
.tag-running{font-size:11px;padding:3px 8px;border-radius:var(--radius-xs);background:var(--status-up-bg);color:var(--status-up);border:1px solid var(--status-up-border)}
.badge{background:var(--accent-blue);color:#fff;font-size:10px;font-weight:700;padding:1px 5px;border-radius:8px;margin-left:3px}

/* 顶栏品牌与面板药丸 */
.brand{display:flex;align-items:center;gap:6px}
.brand h1{font-size:14px;font-weight:700;margin:0;letter-spacing:.3px;color:var(--text-main)}
.badge-pro{font-size:9px;font-weight:800;padding:1px 5px;background:var(--accent-blue-bg);color:var(--accent-blue);border:1px solid var(--accent-blue-border);border-radius:4px;letter-spacing:.5px}
.panel-badge{display:inline-flex;align-items:center;gap:6px;font-size:11px;color:var(--text-muted);padding:2px 8px;border-radius:12px;background:var(--bg-surface);border:1px solid var(--border-card)}
.panel-badge::before{content:"";width:6px;height:6px;border-radius:50%;background:var(--status-up);display:inline-block}

/* 导出与订阅弹窗样式 */
.sheet-export{width:min(780px,95vw)!important;max-height:88vh}
.sub-card{background:var(--bg-surface);border:1px solid var(--border-card);border-radius:var(--radius-sm);padding:12px 14px;margin-bottom:14px}
.sub-card-head{display:flex;align-items:center;gap:8px;margin-bottom:5px;flex-wrap:wrap}
.sub-card-head strong{font-size:12px;color:var(--text-main)}
.sub-badge{font-size:10px;padding:1px 6px;border-radius:var(--radius-xs);background:var(--accent-blue-bg);color:var(--accent-blue);font-weight:600}
.sub-card-desc{font-size:11px;color:var(--text-muted);margin-bottom:8px;line-height:1.5}
.sub-input-row{display:flex;align-items:center;gap:8px}
.sub-input-row input{font-size:11px;flex:1;background:var(--bg-input);border:1px solid var(--border-card);border-radius:var(--radius-xs);padding:6px 9px;color:var(--text-main)}
.ex-nav-bar{display:flex;align-items:center;gap:10px;margin-bottom:10px;flex-wrap:wrap}
.ex-view-tabs{display:inline-flex;border:1px solid var(--border-card);border-radius:var(--radius-xs);overflow:hidden;background:var(--bg-surface)}
.ex-tab-btn{border:0;border-radius:0;background:transparent;padding:4px 10px;font-size:11px;color:var(--text-muted);cursor:pointer}
.ex-tab-btn.active{background:var(--accent-blue);color:#fff;font-weight:700}
.ex-filters{display:inline-flex;gap:4px}
.ex-filter-btn{font-size:11px;padding:2px 8px;border-radius:12px;background:var(--bg-surface);border:1px solid var(--border-card);color:var(--text-muted);cursor:pointer}
.ex-filter-btn.active{background:var(--accent-blue-bg);border-color:var(--accent-blue);color:var(--accent-blue);font-weight:600}
.ex-cards-list{display:flex;flex-direction:column;gap:8px;max-height:48vh;overflow-y:auto}
.ex-card{display:flex;flex-direction:column;gap:6px;padding:10px 12px;background:var(--bg-surface);border:1px solid var(--border-card);border-radius:var(--radius-sm);transition:border-color .15s}
.ex-card:hover{border-color:var(--border-card-hover)}
.ex-card-top{display:flex;align-items:center;gap:8px;flex-wrap:wrap}
.ex-pill{font-size:10px;font-weight:700;padding:2px 7px;border-radius:var(--radius-xs);white-space:nowrap}
.ex-pill.direct{background:var(--status-up-bg);color:var(--status-up);border:1px solid var(--status-up-border)}
.ex-pill.exit{background:var(--accent-blue-bg);color:var(--accent-blue);border:1px solid var(--accent-blue-border)}
.ex-card-info{display:flex;flex-direction:column;gap:4px;overflow:hidden}
.ex-card-title{font-size:12px;font-weight:600;color:var(--text-main);overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.ex-card-link-preview{font-size:11px;color:var(--text-muted);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-family:inherit;background:var(--bg-input);border:1px solid var(--border-card);border-radius:var(--radius-xs);padding:4px 8px;flex:1}
.ex-card-acts{display:flex;align-items:center;gap:8px;margin-top:2px}
.ex-card-sublinks{display:flex;flex-direction:column;gap:5px;margin-top:4px}
.ex-sublink-row{display:flex;align-items:center;gap:8px;background:var(--bg-input);padding:4px 8px;border-radius:var(--radius-xs);border:1px solid var(--border-card)}
.ex-sublink-badge{font-size:10px;color:var(--accent-blue);background:var(--accent-blue-bg);padding:1px 5px;border-radius:var(--radius-xs);white-space:nowrap}
.btn-copy-one{font-size:11px;padding:4px 10px;border-radius:var(--radius-xs);background:var(--bg-surface-sub);border:1px solid var(--border-card);color:var(--text-main);cursor:pointer;white-space:nowrap;display:inline-flex;align-items:center;gap:4px}
.btn-copy-one:hover{border-color:var(--accent-blue);color:var(--accent-blue)}

/* SOCKS5 凭据弹窗增强 */
.cred-preview-card{background:var(--bg-surface);border:1px solid var(--border-card);border-radius:var(--radius-sm);padding:12px 14px}
.cred-preview-label{font-size:11px;color:var(--text-muted);margin-bottom:6px}
.cred-preview-row{display:flex;align-items:center;gap:8px}
.cred-fields-grid{display:grid;grid-template-columns:1fr 1fr;gap:10px}
.cfg-item{background:var(--bg-input);border:1px solid var(--border-card);border-radius:var(--radius-xs);padding:8px 10px}
.cfg-item span{display:block;font-size:10px;color:var(--text-muted);margin-bottom:4px}
.cfg-val-row{display:flex;align-items:center;justify-content:space-between;gap:6px}
.cfg-val-row code{font-size:12px;font-weight:600;color:var(--text-main);word-break:break-all}
.btn-mini{font-size:10px;padding:1px 6px;border-radius:var(--radius-xs);background:var(--bg-surface-sub);border:1px solid var(--border-card);color:var(--text-main);cursor:pointer;white-space:nowrap}
.btn-mini:hover{color:var(--accent-blue);border-color:var(--accent-blue)}
.route-note{font-size:11px;color:var(--text-muted);margin-top:12px;line-height:1.6;padding:8px 12px;background:var(--bg-surface);border-radius:var(--radius-xs);border-left:3px solid var(--accent-blue)}
.route-note code{color:var(--accent-blue)}

/* ==========================================================================
   7. 响应式断点系统 (Tablet 1024px, Mobile 768px & Small 480px)
   ========================================================================== */
@media (max-width: 1024px) {
  main { padding: 16px 20px 60px; }
  .metrics-grid { grid-template-columns: repeat(2, 1fr); gap: 14px; }
  .host-pill { display: none; }
  .exit-card { flex-direction: column; align-items: stretch; gap: 14px; }
  .exit-bindings { border-left: none; border-top: 1px dashed var(--border-card); padding-left: 0; padding-top: 10px; }
  .exit-actions { justify-content: flex-start; }
}

@media (max-width: 768px) {
  body { overflow-x: hidden; }
  main { padding: 12px 14px 60px; }
  
  header { padding: 10px 14px; }
  .header-inner { flex-wrap: wrap; gap: 10px; }
  .brand { flex-wrap: wrap; }
  .header-actions { width: 100%; justify-content: space-between; gap: 6px; flex-wrap: wrap; }
  .header-actions button { flex: 1; justify-content: center; min-height: 38px; padding: 6px 8px; font-size: 12px; }
  
  .metrics-grid { grid-template-columns: 1fr; gap: 10px; }
  .metric-card { padding: 14px 16px; }
  
  .pipeline-container { flex-direction: column; align-items: stretch; gap: 12px; padding: 14px; }
  .pipeline-flow { transform: rotate(90deg); margin: 10px auto; width: 32px; height: 32px; }
  .pipeline-stage { width: 100%; min-width: 0; }
  .pipeline-toolbar { flex-direction: column; align-items: stretch; gap: 10px; }
  .pipeline-toolbar .obind { flex-direction: column; align-items: stretch; width: 100%; gap: 6px; }
  .pipeline-toolbar .select { width: 100%; min-height: 38px; }
  .pipeline-actions { width: 100%; justify-content: stretch; gap: 8px; }
  .pipeline-actions button { flex: 1; text-align: center; justify-content: center; min-height: 38px; }
  
  .toolbar { flex-wrap: wrap; gap: 10px; }
  .toolbar-left { width: 100%; justify-content: space-between; }
  .toolbar-right { width: 100%; display: flex; }
  .toolbar-right button { flex: 1; justify-content: center; min-height: 38px; }
  
  .exit-card { flex-direction: column; align-items: stretch; gap: 12px; padding: 14px; }
  .exit-main { min-width: 0; }
  .exit-bindings { border-left: none; border-top: 1px dashed var(--border-card); padding-left: 0; padding-top: 10px; }
  .exit-actions { justify-content: flex-start; flex-wrap: wrap; border-top: 1px dashed var(--border-card); padding-top: 10px; width: 100%; gap: 8px; }
  .exit-actions button, .exit-actions .socks-badge { min-height: 38px; }
  
  .modal { padding: 10px; }
  .sheet { width: 96% !important; max-width: 96% !important; margin: 12px auto; padding: 16px 14px !important; }
  
  .ncard { grid-template-columns: auto 1fr; grid-template-areas: "flag info" "acts acts"; gap: 8px 10px; }
  .ncard-host { word-break: break-all; }
  .ncard-flag { grid-area: flag; }
  .ncard-info { grid-area: info; }
  .ncard-acts { grid-area: acts; width: 100%; justify-content: stretch; margin-top: 4px; }
  .ncard-acts button, .ncard-acts .tag-running { flex: 1; text-align: center; justify-content: center; min-height: 36px; }
  
  .nex-controls { padding: 8px 10px; }
  .nex-search-row { flex-direction: column; align-items: stretch; gap: 8px; }
  .nex-search-box { min-width: 100%; }
  .nex-sort-group { justify-content: space-between; width: 100%; }
  .nex-sort-btn { flex: 1; text-align: center; min-height: 34px; }
  
  .cred-fields-grid { grid-template-columns: 1fr; }
  .cred-preview-row { flex-direction: column; align-items: stretch; }
  .cred-preview-row button { width: 100%; min-height: 38px; }
  .credrow { flex-direction: column; align-items: stretch; }
  .credrow .ef { width: 100%; }
  .credrow .ef input { width: 100% !important; min-height: 38px; }
  .credrow button { width: 100%; min-height: 38px; }
  
  .sub-input-row { flex-direction: column; align-items: stretch; }
  .sub-input-row button { width: 100%; min-height: 38px; }
  
  .ex-card-acts { flex-direction: column; align-items: stretch; }
  .btn-copy-one { width: 100%; text-align: center; justify-content: center; min-height: 36px; }
  .ex-sublink-row { flex-wrap: wrap; }
}

@media (max-width: 480px) {
  .logo-title { font-size: 14px; }
  .header-actions button { padding: 6px 8px; }
}
</style>
</head>
<body>
<header>
  <div class="header-inner">
    <div class="brand">
      <div class="logo-mark">
        <svg class="icon" viewBox="0 0 24 24">
          <circle cx="6" cy="12" r="3"/>
          <circle cx="18" cy="6" r="3"/>
          <circle cx="18" cy="18" r="3"/>
          <path d="M9 12h3a3 3 0 0 0 3-3V6m-3 6a3 3 0 0 1 3 3v3"/>
        </svg>
      </div>
      <div class="logo-title">
        fanout
        <span class="pro-badge">PRO</span>
      </div>
      <div class="host-pill" id="headerHostPill">
        <span class="pulse-dot"></span>
        <span id="panel">正在连接后端...</span>
      </div>
    </div>

    <div class="header-actions">
      <button type="button" class="theme-toggle" id="themeToggleBtn" onclick="toggleTheme()" title="切换浅色/深色模式">
        <svg class="icon-sm" viewBox="0 0 24 24"><circle cx="12" cy="12" r="4"/><path d="M12 2v2"/><path d="M12 20v2"/><path d="m4.93 4.93 1.41 1.41"/><path d="m17.66 17.66 1.41 1.41"/><path d="M2 12h2"/><path d="M20 12h2"/><path d="m6.34 17.66-1.41 1.41"/><path d="m19.07 4.93-1.41 1.41"/></svg>
        <span id="themeToggleText">浅色模式</span>
      </button>

      <button type="button" class="btn btn-crystal" id="openExplorerBtn" title="全节点大厅（质量/速度降序、自由挑选住宅节点）">
        <svg class="icon-sm" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><polygon points="16.24 7.76 14.12 14.12 7.76 16.24 9.88 9.88 16.24 7.76"/></svg>
        <span>节点大厅</span>
        <span class="mono" id="navNodeCount" style="font-size:11px;font-weight:700"></span>
      </button>

      <button type="button" class="btn btn-ghost" id="exportAllNavBtn" onclick="openModal('export')" title="导出聚合订阅与节点链接">
        <svg class="icon-sm" viewBox="0 0 24 24"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" x2="12" y1="15" y2="3"/></svg>
        <span>导出订阅</span>
      </button>

      <button type="button" class="btn btn-ghost" id="settingsBtn" title="控制台系统设置">
        <svg class="icon-sm" viewBox="0 0 24 24"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
        <span>设置</span>
      </button>

      <button type="button" class="icon" id="logoutBtn" title="退出登录">
        <svg viewBox="0 0 24 24"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
      </button>
    </div>
  </div>
</header>

<main>
  <!-- 核心基础设施概览看板 (Hero Metrics Grid) -->
  <div class="metrics-grid" id="heroMetrics">
    <div class="metric-card">
      <div class="metric-header">
        <span class="metric-title">
          <svg class="icon-sm" viewBox="0 0 24 24"><rect width="20" height="8" x="2" y="2" rx="2" ry="2"/><rect width="20" height="8" x="2" y="14" rx="2" ry="2"/><line x1="6" x2="6.01" y1="6" y2="6"/><line x1="6" x2="6.01" y1="18" y2="18"/></svg>
          核心入站 (:443)
        </span>
        <span class="metric-badge badge-amber" id="heroCoreBadge">母机原生直连</span>
      </div>
      <div class="metric-val mono" id="heroCoreProto">—</div>
      <div class="metric-desc" id="heroCoreDesc">
        <span>—</span>
      </div>
    </div>

    <div class="metric-card">
      <div class="metric-header">
        <span class="metric-title">
          <svg class="icon-sm" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><line x1="2" x2="22" y1="12" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
          活跃住宅出口
        </span>
        <span class="metric-badge badge-green" id="heroExitsBadge">连通中</span>
      </div>
      <div class="metric-val mono" id="heroExitsCount">0 <span style="font-size:14px;color:var(--text-muted);font-weight:400">个在线</span></div>
      <div class="metric-desc" id="heroExitsDesc">
        <span>暂无运行中出口</span>
      </div>
    </div>

    <div class="metric-card">
      <div class="metric-header">
        <span class="metric-title">
          <svg class="icon-sm" viewBox="0 0 24 24"><circle cx="6" cy="19" r="3"/><path d="M9 19h8.5a4.5 4.5 0 0 0 0-9H11"/><circle cx="18" cy="5" r="3"/><path d="M18 8v3"/></svg>
          流量分流拓扑
        </span>
        <span class="metric-badge badge-blue" id="heroSplitBadge">100% 直连</span>
      </div>
      <div class="metric-val mono" id="heroSplitVal">0 挂载 · 0 直连</div>
      <div class="metric-desc" id="heroSplitDesc">
        <span>宿主机公网出网</span>
      </div>
    </div>

    <div class="metric-card">
      <div class="metric-header">
        <span class="metric-title">
          <svg class="icon-sm" viewBox="0 0 24 24"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10"/><path d="m9 12 2 2 4-4"/></svg>
          域名与 HTTPS
        </span>
        <span class="metric-badge badge-blue" id="heroSSLBadge">HTTP 明文</span>
      </div>
      <div class="metric-val mono" id="heroSSLHost" style="font-size:16px;font-weight:600">未绑定域名</div>
      <div class="metric-desc" id="heroSSLDesc">
        <span>IP 直连访问</span>
      </div>
    </div>
  </div>

  <!-- 核心入站与流量拓扑管线容器 -->
  <div id="pipelineContainer"></div>

  <div class="jobs" id="jobs"></div>

  <div class="toolbar bar">
    <div class="toolbar-left">
      <h2 style="font-size:15px;font-weight:700;color:var(--text-main);margin:0">出口隧道</h2>
      <span class="pill-count" id="ecount">0 个活跃出口</span>
    </div>
    <div class="toolbar-right">
      <button type="button" class="btn btn-primary" id="newexit" title="新建出口隧道">
        <svg class="icon-xs" viewBox="0 0 24 24"><line x1="12" x2="12" y1="5" y2="19"/><line x1="5" x2="19" y1="12" y2="12"/></svg>
        <span>新建出口</span>
      </button>
      <button type="button" class="btn btn-ghost" id="exportAll" title="导出全部节点链接">
        <svg class="icon-xs" viewBox="0 0 24 24"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" x2="12" y1="15" y2="3"/></svg>
        <span>导出与订阅</span>
      </button>
      <button type="button" class="btn btn-ghost" id="newnode" title="新建一个节点（协议与端口）">
        <svg class="icon-xs" viewBox="0 0 24 24"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/></svg>
        <span>新建节点</span>
      </button>
      <button type="button" class="btn btn-ghost" id="stopall" style="color:var(--status-danger)" title="停止所有出口">
        <svg class="icon-xs" viewBox="0 0 24 24"><rect width="14" height="14" x="5" y="5" rx="2"/></svg>
        <span>全部停止</span>
      </button>
    </div>
  </div>

  <div id="list"></div>
</main>

<div class="modal" id="wizard">
  <div class="sheet">
    <div class="head">
      <h2>新建出口</h2>
      <span class="spacer"></span>
      <button class="icon" data-close="wizard" title="关闭">
        <svg viewBox="0 0 24 24"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
      </button>
    </div>
    <div class="body">
      <label class="f">
        <span>地区</span>
        <input type="search" id="rgfilter" placeholder="筛选地区">
        <div class="regions" id="regions" style="margin-top:6px"></div>
      </label>
      <label class="f">
        <span>数量</span>
        <div class="stepper">
          <button id="minus" title="减少">
            <svg viewBox="0 0 24 24"><path d="M5 12h14"/></svg>
          </button>
          <input id="count" type="text" inputmode="numeric" value="3">
          <button id="plus" title="增加">
            <svg viewBox="0 0 24 24"><path d="M12 5v14"/><path d="M5 12h14"/></svg>
          </button>
        </div>
        <div class="hint" id="availhint"></div>
      </label>
      <label class="f" id="tplwrap">
        <span>节点链接</span>
        <select id="tpl"></select>
        <div class="hint" id="tplhint"></div>
      </label>
    </div>
    <div class="foot">
      <span class="count" id="wzhint"></span>
      <span class="spacer"></span>
      <button data-close="wizard">取消</button>
      <button class="primary" id="go">开始</button>
    </div>
  </div>
</div>

<div class="modal" id="newnodebox">
  <div class="sheet">
    <div class="head">
      <h2>新建节点</h2>
      <span class="spacer"></span>
      <button class="icon" data-close="newnodebox" title="关闭">
        <svg viewBox="0 0 24 24"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
      </button>
    </div>
    <div class="body">
      <label class="f">
        <span>协议</span>
        <select id="nproto">
          <option value="vless">VLESS</option>
          <option value="vmess">VMess</option>
          <option value="trojan">Trojan</option>
        </select>
      </label>
      <label class="f">
        <span>传输</span>
        <select id="nnet">
          <option value="tcp">TCP</option>
          <option value="ws">WebSocket</option>
          <option value="grpc">gRPC</option>
          <option value="httpupgrade">HTTPUpgrade</option>
          <option value="xhttp">XHTTP</option>
        </select>
      </label>
      <label class="f">
        <span>安全</span>
        <select id="nsec">
          <option value="none">无</option>
          <option value="tls">TLS</option>
          <option value="reality">REALITY</option>
        </select>
        <div class="hint" id="nsechint"></div>
      </label>
      <label class="f" id="nvisionwrap" hidden>
        <span>流控</span>
        <label class="chk"><input type="checkbox" id="nvision"> xtls-rprx-vision</label>
      </label>
      <label class="f" id="nsniwrap" hidden>
        <span>域名 SNI</span>
        <input id="nsni" type="text" placeholder="留空用 localhost，将生成自签证书">
      </label>
      <label class="f" id="ncertwrap" hidden>
        <span>证书路径</span>
        <input id="ncert" type="text" placeholder="留空生成自签证书，如 /etc/ssl/x.crt">
      </label>
      <label class="f" id="nkeywrap" hidden>
        <span>私钥路径</span>
        <input id="nkey" type="text" placeholder="与证书成对填写，如 /etc/ssl/x.key">
      </label>
      <label class="f" id="ndestwrap" hidden>
        <span>借用站点</span>
        <input id="ndest" type="text" placeholder="留空用 www.tesla.com:443">
      </label>
      <label class="f" id="npathwrap" hidden>
        <span id="npathlabel">路径</span>
        <input id="npath" type="text" placeholder="留空自动生成">
      </label>
      <label class="f">
        <span>端口</span>
        <input id="nport" type="text" inputmode="numeric" placeholder="留空随机分配">
      </label>
      <div class="port-recom-wrap">
        <div class="port-recom-title">常用推荐端口（点击填入）：</div>
        <div class="port-chips" id="nport-chips"></div>
        <div class="port-status" id="nport-status"></div>
      </div>
      <label class="f">
        <span>备注</span>
        <input id="nremark" type="text" placeholder="留空自动命名">
      </label>
    </div>
    <div class="foot">
      <span class="count" id="nnhint"></span>
      <span class="spacer"></span>
      <button data-close="newnodebox">取消</button>
      <button class="primary" id="ncreate">创建</button>
    </div>
  </div>
</div>

<div class="modal" id="detail">
  <div class="sheet">
    <div class="head">
      <h2 id="dtitle">节点</h2>
      <span class="spacer"></span>
      <button class="icon danger" id="ddel" title="删除这个入站">
        <svg viewBox="0 0 24 24"><path d="M3 6h18"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6"/><path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
      </button>
      <button class="icon" data-close="detail" title="关闭">
        <svg viewBox="0 0 24 24"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
      </button>
    </div>
    <div class="body" id="dbody"></div>
  </div>
</div>

<div class="modal" id="credbox">
  <div class="sheet">
    <div class="head">
      <h2>SOCKS5 访问凭据</h2>
      <span class="count" id="crtitle"></span>
      <span class="spacer"></span>
      <button class="icon" data-close="credbox" title="关闭">
        <svg viewBox="0 0 24 24"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
      </button>
    </div>
    <div class="body">
      <div class="cred-preview-card">
        <div class="cred-preview-label">SOCKS5 完整连接 URL：</div>
        <div class="cred-preview-row">
          <div class="share" id="crurl" style="margin:0;flex:1"></div>
          <button type="button" class="primary" id="crcopy" title="一键复制完整 SOCKS5 URL">复制 URL</button>
        </div>
        <div class="cred-fields-grid" style="margin-top:10px">
          <div class="cfg-item">
            <span>母机连接地址与端口</span>
            <div class="cfg-val-row">
              <code id="crhostport">—</code>
              <button type="button" class="btn-mini" id="crcopyhost" title="复制 host:port">复制</button>
            </div>
          </div>
          <div class="cfg-item">
            <span>出口真实 IP (VPN Gate)</span>
            <div class="cfg-val-row">
              <code id="crexitip">—</code>
              <button type="button" class="btn-mini" id="crcopyip" title="复制真实出口 IP">复制</button>
            </div>
          </div>
        </div>
      </div>

      <div class="credrow" style="margin-top:14px">
        <label class="ef"><span>用户名</span>
          <input id="cruser" type="text" spellcheck="false"></label>
        <label class="ef"><span>口令</span>
          <input id="crpass" type="text" spellcheck="false"></label>
        <button type="button" id="crrand" title="随机生成一套无歧义凭据">
          <svg viewBox="0 0 24 24"><path d="M21 12a9 9 0 1 1-3-6.7L21 8"/><path d="M21 3v5h-5"/></svg>
          随机生成
        </button>
      </div>
      <div class="route-note">
        <b>网络转发路径</b>：客户端连接 <code>母机公网IP:端口</code> ➔ 经由 OpenVPN 隧道隔离网络 ➔ 由 <code>VPN Gate 出口节点</code> 发往目标网络。<br>
        修改凭据后立即生效，已连上的活动会话不中断；新发起连接需使用新凭据。
      </div>
    </div>
    <div class="foot">
      <span class="spacer"></span>
      <button data-close="credbox">取消</button>
      <button class="primary" id="crsave">保存凭据</button>
    </div>
  </div>
</div>

<div class="modal" id="export">
  <div class="sheet sheet-export">
    <div class="head">
      <div style="display:flex;align-items:center;gap:8px">
        <h2>节点链接与客户端订阅</h2>
        <span class="count" id="excount"></span>
      </div>
      <span class="spacer"></span>
      <button id="copyall" title="复制所有分享链接">
        <svg viewBox="0 0 24 24"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
        全部复制
      </button>
      <button class="icon" data-close="export" title="关闭">
        <svg viewBox="0 0 24 24"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
      </button>
    </div>
    <div class="body">
      <div class="sub-card">
        <div class="sub-card-head">
          <svg class="icon-sm" viewBox="0 0 24 24"><path d="M4.9 19.1C1 15.2 1 8.8 4.9 4.9"/><path d="M7.8 16.2c-2.3-2.3-2.3-6.1 0-8.5"/><circle cx="12" cy="12" r="2"/><path d="M16.2 7.8c2.3 2.3 2.3 6.1 0 8.5"/><path d="M19.1 4.9C23 8.8 23 15.2 19.1 19.1"/></svg>
          <strong>客户端一键聚合订阅源 (/sub)</strong>
          <span class="sub-badge">免手动导入 · 自动同步</span>
        </div>
        <div class="sub-card-desc">
          支持 V2RayN、Shadowrocket、Clash、Sing-box 等客户端自动拉取母机直连与全部出口节点：
        </div>
        <div class="sub-input-row">
          <input type="text" id="subUrlInput" readonly spellcheck="false">
          <button type="button" class="primary" id="copySubUrlBtn">复制订阅链接</button>
        </div>
      </div>

      <div class="ex-nav-bar">
        <div class="ex-view-tabs">
          <button type="button" class="ex-tab-btn active" id="exTabCards">节点卡片明细</button>
          <button type="button" class="ex-tab-btn" id="exTabText">纯文本批量导入</button>
        </div>
        <span class="spacer"></span>
        <div class="ex-filters" id="exFilters">
          <button type="button" class="ex-filter-btn active" data-exfilter="all">全部</button>
          <button type="button" class="ex-filter-btn" data-exfilter="direct">仅直连</button>
          <button type="button" class="ex-filter-btn" data-exfilter="exit">仅出口</button>
        </div>
      </div>

      <div id="exCardsWrap" class="ex-cards-list"></div>

      <div id="exTextWrap" style="display:none">
        <textarea id="exbox" spellcheck="false" readonly style="min-height:220px"></textarea>
      </div>
    </div>
  </div>
</div>

<div class="modal" id="settings">
  <div class="sheet">
    <div class="head">
      <h2><svg class="icon-sm" viewBox="0 0 24 24"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg> 系统与面板设置</h2>
      <span class="spacer"></span>
      <button class="icon" data-close="settings" title="关闭">
        <svg viewBox="0 0 24 24"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
      </button>
    </div>
    <div class="body" style="display:flex;flex-direction:column;gap:12px">
      <!-- 卡片 1: 访问安全与路径 -->
      <div class="set-card">
        <div class="set-card-title">访问安全与路径</div>
        <label class="f"><span>访问口令</span>
          <input id="setPw" type="password" spellcheck="false" autocomplete="new-password" placeholder="留空则保持原口令不变"></label>
        <div class="hint">修改后仅对新登录会话生效，当前浏览器不会被强制退出。</div>

        <label class="f" style="margin-top:12px"><span>访问路径前缀</span>
          <input id="setPath" type="text" spellcheck="false" placeholder="留空则去掉路径前缀"></label>
        <div class="hint" id="setPathHint">挂在自定义路径下可隐藏面板，防全网探针扫描。只能用字母数字和 - _。</div>
      </div>

      <!-- 卡片 2: 域名与 SSL (HTTPS) -->
      <div class="set-card">
        <div class="set-card-title">域名与 HTTPS (SSL) 加密</div>
        <label class="f"><span>面板绑定域名</span>
          <input id="setDomain" type="text" spellcheck="false" placeholder="例如 panel.example.com（可选，留空使用 IP）"></label>
        <div class="hint">绑定域名后，节点分享链接和 /sub 聚合订阅源将优先使用此域名。</div>

        <label class="f" style="margin-top:12px"><span>SSL 模式</span>
          <select id="setSSLMode">
            <option value="none">外部反代 / HTTP（纯域名直连或外部 CDN 反代）</option>
            <option value="custom">原生自定义 SSL 证书（面板原生 HTTPS 监听）</option>
            <option value="caddy">一键 Caddy 自动化反代（母机 443 免端口纯净访问）</option>
          </select></label>

        <div id="setSSLCertWrap" style="margin-top:12px;display:none;background:var(--bg-input);border:1px solid var(--border-card);border-radius:var(--radius-sm);padding:12px">
          <label class="f"><span>证书文件绝对路径 (Cert / PEM)</span>
            <input id="setCertFile" type="text" spellcheck="false" placeholder="/etc/ssl/cert.pem 或 acme.sh 证书路径"></label>
          <label class="f" style="margin-top:8px"><span>私钥文件绝对路径 (Key)</span>
            <input id="setKeyFile" type="text" spellcheck="false" placeholder="/etc/ssl/key.pem 或 acme.sh 私钥路径"></label>
          <div style="display:flex;align-items:center;gap:8px;margin-top:10px">
            <button type="button" class="btn btn-ghost" id="setCertCheck" style="font-size:12px;padding:4px 10px"><svg class="icon-xs" viewBox="0 0 24 24"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10"/><path d="m9 12 2 2 4-4"/></svg> <span>检测证书有效性</span></button>
            <span id="setCertStatus" style="font-size:11px;color:var(--text-muted)"></span>
          </div>
        </div>
        <div class="hint" style="margin-top:8px">也可在终端使用 <code>f ssl</code> 一键申请免费证书，或 <code>f caddy</code> 部署 443 自动化反代。</div>
      </div>

      <!-- 卡片 3: 监听与后端模式 -->
      <div class="set-card">
        <div class="set-card-title">网络监听与后端模式</div>
        <label class="f"><span>节点后端</span>
          <select id="setBackend"></select></label>
        <div class="hint" id="setBackendHint">节点从哪来。装了 3x-ui 或 xray-cf-lite 就能直接接管，都没有就用自建。</div>

        <div class="setrow" style="margin-top:12px">
          <label class="f" style="margin:0"><span>监听端口</span>
            <input id="setPort" type="text" inputmode="numeric" spellcheck="false"></label>
          <label class="f" style="margin:0"><span>本地监听地址</span>
            <select id="setListen">
              <option value="0.0.0.0">所有网卡（0.0.0.0）</option>
              <option value="127.0.0.1">仅本机（127.0.0.1）</option>
            </select></label>
        </div>
        <div class="hint bad" id="setPortHint" style="margin-top:8px">改端口、监听地址或协议会切换监听，保存后要用新地址重新打开界面。</div>
      </div>

      <!-- 卡片 4: 版本与更新 -->
      <div class="set-card" style="margin-bottom:0">
        <div class="set-card-title">系统版本与维护</div>
        <div class="updrow">
          <div class="updver">当前版本 <b id="updCur">-</b><span id="updLatest"></span></div>
          <span class="spacer"></span>
          <button id="updCheck">检查更新</button>
          <button class="primary" id="updApply" hidden>更新到 <span id="updApplyVer"></span></button>
        </div>
        <div class="updnotes" id="updNotes" hidden style="margin-top:8px"></div>
        <div style="display:flex;justify-content:space-between;align-items:center;margin-top:12px;padding-top:10px;border-top:1px solid var(--border-card)">
          <span style="font-size:11px;color:var(--text-muted)">当前登录会话</span>
          <button type="button" class="btn-subtle" id="settingsLogoutBtn" style="color:var(--status-danger)" title="退出当前登录会话">
            退出登录
          </button>
        </div>
      </div>
    </div>
    <div class="foot">
      <span class="spacer"></span>
      <button data-close="settings">取消</button>
      <button class="primary" id="setSave">保存生效</button>
    </div>
  </div>
</div>

<div class="modal" id="nodeExplorer">
  <div class="sheet sheet-wide">
    <div class="head">
      <div style="display:flex;align-items:center;gap:10px">
        <h2><svg class="icon-sm" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><polygon points="16.24 7.76 14.12 14.12 7.76 16.24 9.88 9.88 16.24 7.76"/></svg> 全节点大厅 (Node Explorer)</h2>
        <span class="count" id="nex-count">正在加载节点...</span>
      </div>
      <span class="spacer"></span>
      <button class="icon" data-close="nodeExplorer" title="关闭">
        <svg viewBox="0 0 24 24"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
      </button>
    </div>
    <div class="nex-controls">
      <div class="nex-search-row">
        <div class="nex-search-box">
          <svg viewBox="0 0 24 24"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
          <input type="text" id="nex-kw" placeholder="快速搜索国家、城市、IP、ISP 运营商、ASN...">
          <button type="button" id="nex-kw-clear" class="btn-search-clear" style="display:none" title="清空搜索">✕</button>
        </div>
        <div class="nex-sort-group">
          <span class="sort-label">排序方式:</span>
          <button type="button" class="nex-sort-btn active" data-sort="quality" title="家宽优先，纯净度与信誉得分由高到低">质量优先</button>
          <button type="button" class="nex-sort-btn" data-sort="speed" title="带宽吞吐量由高到低">速度降序</button>
          <button type="button" class="nex-sort-btn" data-sort="ping" title="响应延迟由低到高">延迟升序</button>
        </div>
      </div>
      <div class="nex-filter-row">
        <div class="nex-types" id="nex-types">
          <button type="button" class="nex-filter-chip active" data-type="">全部类型</button>
          <button type="button" class="nex-filter-chip" data-type="residential">优质家庭宽带 (95+分)</button>
          <button type="button" class="nex-filter-chip" data-type="datacenter">机房/IDC</button>
        </div>
        <span class="spacer"></span>
        <div class="nex-regions" id="nex-regions" style="display:flex;gap:4px;flex-wrap:wrap"></div>
      </div>
    </div>
    <div class="nex-list-wrap">
      <div class="nex-list" id="nex-list"></div>
    </div>
  </div>
</div>

<div class="modal" id="bindExitModal">
  <div class="sheet">
    <div class="head">
      <h2 id="bem-title">为出口绑定节点</h2>
      <span class="spacer"></span>
      <button class="icon" data-close="bindExitModal" title="关闭">
        <svg viewBox="0 0 24 24"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
      </button>
    </div>
    <div class="body">
      <div id="bem-target" style="padding:8px 12px;background:var(--bg-surface);border:1px solid var(--border-card);border-radius:var(--radius-xs);font-size:12px;margin-bottom:14px;color:var(--text-main)"></div>

      <div class="bmode-item active" id="bmode-clone-card">
        <div class="bmode-head">
          <input type="radio" name="bmode" id="bmode-radio-clone" checked>
          <strong>复制原节点并绑定 (推荐 · 原节点继续直连)</strong>
        </div>
        <p class="hint" style="margin:6px 0 10px">基于现有节点复制一套全新入站（相同密码/UUID/传输协议），<b>原节点继续留作母机直连</b>，两者互不影响！</p>
        <div id="bmode-clone-fields">
          <label class="f"><span>选择模板节点</span>
            <select id="bem-tpl-node"></select>
          </label>
          <label class="f"><span>分配新端口（专走此出口）</span>
            <input type="text" id="bem-port" placeholder="输入或点击下方推荐端口" inputmode="numeric">
            <div id="bem-port-status" class="port-status"></div>
            <div class="port-recom-wrap" style="margin-top:6px">
              <div class="port-recom-title">推荐空闲端口（点击一键填入）：</div>
              <div class="port-chips" id="bem-port-chips"></div>
            </div>
          </label>
          <button class="primary" id="bem-btn-clone" style="width:100%;margin-top:8px">立即复制并绑定至该出口</button>
        </div>
      </div>

      <div class="bmode-item" id="bmode-move-card" style="margin-top:12px">
        <div class="bmode-head">
          <input type="radio" name="bmode" id="bmode-radio-move">
          <strong>直接转移现有直连节点</strong>
        </div>
        <p class="hint" style="margin:6px 0 10px">将现有的某个直连节点直接改绑至此出口（注意：该节点将不再直连，流量转向该出口）。</p>
        <div id="bmode-move-fields" style="display:none">
          <label class="f"><span>选择要转移改绑的节点</span>
            <select id="bem-direct-node"></select>
          </label>
          <button class="primary" id="bem-btn-move" style="width:100%;margin-top:8px">转移改绑至该出口</button>
        </div>
      </div>
    </div>
  </div>
</div>

<!-- 自定义操作确认模态框 -->
<div class="modal" id="confirmModal">
  <div class="sheet confirm-sheet">
    <div class="confirm-head">
      <div class="confirm-icon-box" id="confirmIconBox"></div>
      <div class="confirm-content">
        <h3 id="confirmTitle">操作确认</h3>
        <p class="confirm-msg" id="confirmMsg"></p>
      </div>
    </div>
    <div class="confirm-actions">
      <button type="button" class="btn" id="confirmCancelBtn" onclick="resolveConfirm(false)">取消</button>
      <button type="button" class="btn primary" id="confirmOkBtn" onclick="resolveConfirm(true)">确定</button>
    </div>
  </div>
</div>

<div class="toast" id="toast"></div>

<script>
const $ = s => document.querySelector(s);
const ICON = {
  sun: '<svg class="icon-sm" viewBox="0 0 24 24"><circle cx="12" cy="12" r="4"/><path d="M12 2v2"/><path d="M12 20v2"/><path d="m4.93 4.93 1.41 1.41"/><path d="m17.66 17.66 1.41 1.41"/><path d="M2 12h2"/><path d="M20 12h2"/><path d="m6.34 17.66-1.41 1.41"/><path d="m19.07 4.93-1.41 1.41"/></svg>',
  moon: '<svg class="icon-sm" viewBox="0 0 24 24"><path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"/></svg>',
  node: '<svg class="icon-sm" viewBox="0 0 24 24"><rect width="20" height="8" x="2" y="2" rx="2"/><rect width="20" height="8" x="2" y="14" rx="2"/><line x1="6" x2="6.01" y1="6" y2="6"/><line x1="6" x2="6.01" y1="18" y2="18"/></svg>',
  settings: '<svg class="icon-sm" viewBox="0 0 24 24"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>',
  export: '<svg class="icon-sm" viewBox="0 0 24 24"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" x2="12" y1="3" y2="15"/></svg>',
  cred: '<svg class="icon-sm" viewBox="0 0 24 24"><rect width="18" height="11" x="3" y="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>',
  lock: '<svg class="icon-xs lock" viewBox="0 0 24 24"><rect width="18" height="11" x="3" y="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>',
  swap: '<svg class="icon-xs" viewBox="0 0 24 24"><path d="m16 3 4 4-4 4"/><path d="M20 7H4"/><path d="m8 21-4-4 4-4"/><path d="M4 17h16"/></svg>',
  stop: '<svg class="icon-xs" viewBox="0 0 24 24"><rect x="6" y="6" width="12" height="12" rx="1"/></svg>',
  redo: '<svg class="icon-xs" viewBox="0 0 24 24"><path d="M21 12a9 9 0 1 1-3-6.7L21 8"/><path d="M21 3v5h-5"/></svg>',
  copy: '<svg class="icon-xs" viewBox="0 0 24 24"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>',
  plus: '<svg class="icon-xs" viewBox="0 0 24 24"><path d="M12 5v14"/><path d="M5 12h14"/></svg>',
  check: '<svg class="icon-xs" viewBox="0 0 24 24"><polyline points="20 6 9 17 4 12"/></svg>',
  x: '<svg class="icon-xs" viewBox="0 0 24 24"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>',
  direct: '<svg class="icon-sm" viewBox="0 0 24 24"><polyline points="13 17 18 12 13 7"/><polyline points="6 17 11 12 6 7"/></svg>',
  reality: '<svg class="icon-sm" viewBox="0 0 24 24"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>',
  flow: '<svg class="icon-xs" viewBox="0 0 24 24"><path d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/><path d="M3 12a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16"/><path d="M16 21h5v-5"/></svg>',
  trash: '<svg class="icon-xs" viewBox="0 0 24 24"><path d="M3 6h18"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6"/><path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>',
  ok: '<svg class="icon-xs" viewBox="0 0 24 24"><path d="M20 6 9 17l-5-5"/></svg>',
  bad: '<svg class="icon-xs" viewBox="0 0 24 24"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>',
  run: '<svg class="icon-xs spin" viewBox="0 0 24 24"><path d="M21 12a9 9 0 1 1-6.2-8.5"/></svg>',
  wait: '<svg class="icon-xs" viewBox="0 0 24 24"><circle cx="12" cy="12" r="9"/></svg>'
};

function toggleTheme(){
  const html = document.documentElement;
  const isDark = html.classList.contains('dark') || !html.classList.contains('light');
  const target = isDark ? 'light' : 'dark';
  html.classList.remove('dark', 'light');
  html.classList.add(target);
  try{ localStorage.setItem('fanout_theme', target); }catch(e){}
  updateThemeButton(target);
}

function updateThemeButton(theme){
  const btn = $('#themeToggleBtn');
  if(!btn) return;
  const isLight = theme === 'light';
  btn.innerHTML = (isLight ? ICON.moon : ICON.sun) + '<span id="themeToggleText">' + (isLight ? '深色模式' : '浅色模式') + '</span>';
  btn.title = isLight ? '切换为深色高对比模式' : '切换为纯净浅色模式';
}

function initTheme(){
  let t = 'dark';
  try{
    t = localStorage.getItem('fanout_theme');
    if(!t && window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches){
      t = 'light';
    }
  }catch(e){}
  if(!t) t = 'dark';
  document.documentElement.classList.remove('dark', 'light');
  document.documentElement.classList.add(t);
  updateThemeButton(t);
}
initTheme();

// 界面挂在随机前缀下，请求一律走相对路径
async function api(path, opts){
  const r = await fetch(path.replace(/^\//, ''), opts);
  const d = await r.json().catch(()=>({}));
  if(!r.ok) throw new Error(d.error || ('HTTP '+r.status));
  return d;
}
function esc(s){ return String(s == null ? '' : s).replace(/[&<>"']/g, c =>
  ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c])); }

let toastTimer;
function toast(msg, bad){
  const el = $('#toast');
  el.textContent = msg;
  el.className = 'toast show' + (bad ? ' bad' : '');
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => { el.className = 'toast'; }, 2400);
}
async function copy(text){
  // navigator.clipboard 只在 HTTPS/localhost 下存在，而面板通常是 http://IP 访问，
  // 所以必须留一条 execCommand 兜底路径，否则复制在正常使用场景里必然失败。
  if(navigator.clipboard && window.isSecureContext){
    try{ await navigator.clipboard.writeText(text); toast('已复制'); return; }
    catch(e){}
  }
  const ta = document.createElement('textarea');
  ta.value = text;
  ta.setAttribute('readonly', '');
  // 放在视口内但不可见：置于视口外会让 iOS 在聚焦时滚动页面
  ta.style.cssText = 'position:fixed;top:0;left:0;width:1px;height:1px;opacity:0;padding:0;border:0';
  document.body.appendChild(ta);
  const prev = document.activeElement;
  ta.focus();
  ta.setSelectionRange(0, ta.value.length);
  let ok = false;
  try{ ok = document.execCommand('copy'); }catch(e){}
  ta.remove();
  if(prev && prev.focus) prev.focus();
  toast(ok ? '已复制' : '复制失败，请手动选中', !ok);
}

let view = {exits:[], direct:[], panel:'', backend:'', public_ip:''};
let inbounds = [];

const RECOMMENDED_PORTS = [443, 8443, 2053, 2083, 2096, 80, 8080, 8880, 'random'];

function getOccupiedPorts(excludePort){
  const set = new Map();
  (view.exits || []).forEach(e => {
    if(e.port && e.port !== excludePort) set.set(e.port, '出口 SOCKS5 :' + e.port);
    (e.inbounds || []).forEach(ib => {
      if(ib.port && ib.port !== excludePort) set.set(ib.port, '入站 ' + (ib.remark || ib.protocol) + ' :' + ib.port);
    });
  });
  (view.direct || []).forEach(ib => {
    if(ib.port && ib.port !== excludePort) set.set(ib.port, '直连入站 ' + (ib.remark || ib.protocol) + ' :' + ib.port);
  });
  return set;
}

function getRandomHighPort(occupied){
  for(let i = 0; i < 60; i++){
    const p = Math.floor(Math.random() * 40000) + 20000;
    if(!occupied.has(p)) return p;
  }
  return 35443;
}

function renderPortRecommendations(containerId, statusId, inputId, saveBtnId, excludePort){
  const box = $('#' + containerId);
  const statusEl = $('#' + statusId);
  const inputEl = $('#' + inputId);
  if(!box || !inputEl) return;

  const occupied = getOccupiedPorts(excludePort);

  box.innerHTML = RECOMMENDED_PORTS.map(p => {
    if(p === 'random'){
      return '<button type="button" class="pchip" data-paction="random">🎲 随机高位</button>';
    }
    const isOcc = occupied.has(p);
    const label = (p === 443 ? '🔥 443' : '' + p) + (isOcc ? ' [已占用]' : '');
    return '<button type="button" class="pchip ' + (isOcc ? 'occupied' : 'free') + '" data-port="' + p + '"'
      + (isOcc ? ' disabled' : '') + '>' + label + '</button>';
  }).join('');

  box.onclick = e => {
    const chip = e.target.closest('.pchip');
    if(!chip || chip.disabled) return;
    if(chip.dataset.paction === 'random'){
      inputEl.value = getRandomHighPort(occupied);
    }else if(chip.dataset.port){
      inputEl.value = chip.dataset.port;
    }
    validatePortInput(inputEl, statusEl, saveBtnId, occupied);
  };

  inputEl.oninput = () => {
    validatePortInput(inputEl, statusEl, saveBtnId, occupied);
  };

  validatePortInput(inputEl, statusEl, saveBtnId, occupied);
}

let portCheckSeq = 0;
function validatePortInput(inputEl, statusEl, saveBtnId, occupied){
  const saveBtn = saveBtnId ? $('#' + saveBtnId) : null;
  const val = parseInt((inputEl.value || '').trim(), 10);
  if(!val || isNaN(val)){
    statusEl.className = 'port-status';
    statusEl.textContent = inputEl.placeholder ? '' : '请输入 1~65535 端口';
    if(saveBtn) saveBtn.disabled = false;
    return;
  }
  if(val < 1 || val > 65535){
    statusEl.className = 'port-status bad';
    statusEl.textContent = '❌ 端口超出合法范围 (1 ~ 65535)';
    if(saveBtn) saveBtn.disabled = true;
    return;
  }
  if(occupied.has(val)){
    statusEl.className = 'port-status bad';
    statusEl.textContent = '❌ 端口 ' + val + ' 已被占用：' + occupied.get(val);
    if(saveBtn) saveBtn.disabled = true;
    return;
  }

  statusEl.className = 'port-status';
  statusEl.textContent = '🔍 正在探测端口可用性...';
  const seq = ++portCheckSeq;
  api('/api/port/check?port=' + val).then(res => {
    if(seq !== portCheckSeq) return;
    if(res && res.available){
      statusEl.className = 'port-status ok';
      statusEl.textContent = '✅ 端口 ' + val + ' 可用';
      if(saveBtn) saveBtn.disabled = false;
    } else {
      statusEl.className = 'port-status bad';
      statusEl.textContent = '❌ ' + (res.reason || ('端口 ' + val + ' 不可用'));
      if(saveBtn) saveBtn.disabled = true;
    }
  }).catch(() => {
    if(seq !== portCheckSeq) return;
    statusEl.className = 'port-status ok';
    statusEl.textContent = '✅ 端口 ' + val + ' 格式有效';
    if(saveBtn) saveBtn.disabled = false;
  });
}

function getPrimaryInboundPort(){
  const allIb = [];
  (view.direct || []).forEach(i => allIb.push(i));
  (view.exits || []).forEach(e => (e.inbounds || []).forEach(i => allIb.push(i)));
  const found443 = allIb.find(i => i.port === 443);
  if(found443) return 443;
  if(allIb.length > 0) return allIb[0].port;
  return 443;
}

// 自建模式下入站由 fanout 自己管，界面要提供新建入口；
// 接管 3x-ui 时入站归面板管，这里只读不写。
function isNative(){ return view.backend === 'native'; }
// xray-cf-lite 模式下节点归它管，fanout 只改路由，界面不给新建入口
function isXCL(){ return view.backend === 'xray-cf-lite'; }
const BACKEND_NAME = {'native':'自建 Xray', '3x-ui':'3x-ui', 'xray-cf-lite':'xray-cf-lite'};
function backendName(){ return BACKEND_NAME[view.backend] || '3x-ui'; }

const STATUS = {up:'已连通', starting:'连接中', failed:'连接失败', stopped:'已停止'};

function formatSince(sinceStr){
  if(!sinceStr) return '';
  try{
    const t = new Date(sinceStr).getTime();
    if(isNaN(t) || t <= 0) return '';
    const ms = Date.now() - t;
    if(ms < 0) return '';
    const mins = Math.floor(ms / 60000);
    if(mins < 1) return '刚刚连线';
    if(mins < 60) return '连线 ' + mins + ' 分钟';
    const hrs = (mins / 60).toFixed(1);
    return '运行 ' + hrs + ' 小时';
  }catch(e){ return ''; }
}

function renderExits(){
  const list = $('#list');
  const n = (view.exits || []).length;
  const totalInbounds = (view.direct || []).length + (view.exits || []).reduce((acc, e) => acc + (e.inbounds || []).length, 0);
  if($('#exportAll')) $('#exportAll').disabled = !totalInbounds;
  if($('#stopall')) $('#stopall').disabled = !n;
  if($('#ecount')) $('#ecount').textContent = n + ' 个活跃出口';

  if(!n){
    list.innerHTML = '<div class="empty" style="border:1px dashed var(--border-card);border-radius:var(--radius-lg);padding:48px 24px;text-align:center;background:var(--bg-card)">'
      + '<div style="font-size:14px;color:var(--text-muted);margin-bottom:12px">暂无活跃的出口隧道</div>'
      + '<button type="button" class="btn btn-primary" id="newexit2">'
      + '<svg class="icon-sm" viewBox="0 0 24 24"><line x1="12" x2="12" y1="5" y2="19"/><line x1="5" x2="19" y1="12" y2="12"/></svg>'
      + '<span>新建出口</span></button></div>';
    return;
  }

  list.innerHTML = '<div class="exit-list">' + view.exits.map(e => {
    const label = e.exit_ip || (e.status === 'starting' ? '连接中…' : (e.status === 'failed' ? '连接失败' : '—'));
    const country = esc(e.country || e.region || 'KR');
    const isUp = e.status === 'up';

    // 质量画像徽章
    let qualityBadge = '';
    if(e.quality && e.quality.score){
      const isRes = e.quality.type === 'residential';
      qualityBadge = isRes
        ? '<span class="quality-tag tag-residential"><span class="dot-indicator"></span> 住宅 ' + e.quality.score + '分</span>'
        : '<span class="quality-tag tag-datacenter"><span class="dot-indicator"></span> 机房 ' + e.quality.score + '分</span>';
    }

    // 延迟与速度徽章 (性能与测速数据)
    let perfBadges = '';
    let pingHTML = '';
    if(e.ping && e.ping > 0){
      const pingVal = e.ping;
      const pingCls = pingVal < 80 ? 'ping-good' : (pingVal < 180 ? 'ping-med' : 'ping-slow');
      pingHTML = '<span class="metric-tag tag-ping ' + pingCls + '" title="节点真实网络延迟 (RTT): ' + pingVal + ' ms">'
        + '<svg class="icon-nano" viewBox="0 0 24 24"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>'
        + pingVal + ' ms</span>';
    }

    let speedHTML = '';
    if(e.speed_mbps && e.speed_mbps > 0){
      const spdVal = e.speed_mbps;
      const spdStr = spdVal >= 100 ? spdVal.toFixed(0) : spdVal.toFixed(1);
      speedHTML = '<span class="metric-tag tag-speed" title="节点测速带宽: ' + spdStr + ' Mbps">'
        + '<svg class="icon-nano" viewBox="0 0 24 24"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>'
        + spdStr + ' Mbps</span>';
    }
    perfBadges = pingHTML + speedHTML;

    // 运营商与连线时间
    const ispName = (e.quality && e.quality.isp) ? e.quality.isp : (e.isp || '');
    const metaHost = esc(e.host || '');
    const metaParts = [];
    if(ispName) metaParts.push(esc(ispName));
    if(metaHost) metaParts.push(metaHost);
    const metaStr = metaParts.join(' • ') || esc(e.region || '—');

    const sinceText = isUp ? formatSince(e.since) : '';
    const timeHTML = sinceText
      ? '<span class="exit-time"><span class="dot-indicator"></span> ' + esc(sinceText) + '</span>'
      : (e.status !== 'up' ? '<span class="count" style="color:' + (e.status === 'failed' ? 'var(--status-danger)' : 'var(--status-warn)') + '">' + (STATUS[e.status] || e.status) + '</span>' : '');

    // 挂载入站胶囊
    let bindingsHTML = '';
    const inbounds = e.inbounds || [];
    if(inbounds.length > 0){
      bindingsHTML = inbounds.map(i => {
        const isCore = i.port === 443;
        const protoText = esc(i.protocol || 'port') + ' :' + i.port;
        const pillClass = 'bound-pill' + (isCore ? ' core' : '');
        return '<span class="' + pillClass + '" title="' + esc((i.remark || i.protocol) + ' · ' + i.protocol + ' :' + i.port) + '">'
          + '<span class="mono">' + protoText + '</span>'
          + '<span class="bound-unbind" data-unbind-tag="' + esc(i.tag) + '" data-unbind-port="' + i.port + '" title="解除挂载，恢复母机原生直连">✕</span>'
          + '</span>';
      }).join('');
    } else {
      bindingsHTML = '<span style="font-size:12px;color:var(--text-muted);font-style:italic">未挂载入站</span>';
    }

    const addBindBtn = '<button type="button" class="btn-add-bind" data-bind-slot="' + e.slot
      + '" data-bind-host="' + esc(e.host)
      + '" data-bind-ip="' + esc(label)
      + '" data-bind-region="' + esc(e.region || '')
      + '" title="为该出口挂载入站节点"><svg class="icon-xs" viewBox="0 0 24 24"><line x1="12" x2="12" y1="5" y2="19"/><line x1="5" x2="19" y1="12" y2="12"/></svg><span>绑定入站</span></button>';

    const errHTML = (e.status === 'failed' && e.err)
      ? '<div class="errline" style="margin-top:6px;color:var(--status-danger);font-size:12px" title="' + esc(e.err) + '">' + esc(e.err) + '</div>'
      : '';

    // 第二排三个核心指标字段 (质量画像、实时延迟、测速带宽) 统一排列
    const row2Badges = qualityBadge + perfBadges;
    const metricsHTML = row2Badges ? ('<div class="exit-metrics-row">' + row2Badges + '</div>') : '';

    return '<div class="exit-card">'
      + '<div class="exit-main">'
      +   '<span class="country-tag">' + country + '</span>'
      +   '<div class="exit-info">'
      +     '<div class="exit-ip-row">'
      +       '<span class="mono"' + (e.status === 'failed' ? ' style="color:var(--status-danger);font-weight:700"' : (e.status === 'starting' ? ' style="color:var(--status-warn)"' : '')) + '>' + esc(label) + '</span>'
      +       (e.exit_ip ? '<button type="button" class="copy-btn" data-copy="' + esc(e.exit_ip) + '" title="复制出口 IP"><svg class="icon-xs" viewBox="0 0 24 24"><rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg><span>复制</span></button>' : '')
      +     '</div>'
      +     metricsHTML
      +     '<div class="exit-meta">'
      +       '<span>' + metaStr + '</span>'
      +       timeHTML
      +     '</div>'
      +   '</div>'
      + '</div>'
      + '<div class="exit-bindings">'
      +   '<span style="font-size:12px;color:var(--text-muted);font-weight:600">挂载入站:</span>'
      +   bindingsHTML
      +   addBindBtn
      + '</div>'
      + '<div class="exit-actions">'
      +   '<div class="socks-badge mono" data-cred="' + e.slot + '" title="查看 SOCKS5 访问凭据">'
      +     '<svg class="icon-xs" viewBox="0 0 24 24"><rect width="18" height="11" x="3" y="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>'
      +     '<span>:' + e.port + '</span>'
      +   '</div>'
      +   (e.status === 'failed' ? '<button type="button" class="btn btn-ghost" style="padding:5px 10px;font-size:12px;color:var(--primary)" data-retry="' + e.slot + '" title="重试连接该节点"><svg class="icon-xs" viewBox="0 0 24 24"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg><span>重试</span></button>' : '')
      +   '<button type="button" class="btn btn-ghost" style="padding:5px 10px;font-size:12px" data-swap="' + e.slot + '" title="同地区轮换节点">'
      +     '<svg class="icon-xs" viewBox="0 0 24 24"><path d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/><path d="M3 12a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16"/><path d="M16 21h5v-5"/></svg>'
      +     '<span>换节点</span>'
      +   '</button>'
      +   '<button type="button" class="btn btn-ghost" style="padding:5px 10px;font-size:12px;color:var(--status-danger)" data-stop="' + e.slot + '" title="停止此出口隧道">'
      +     '<svg class="icon-xs" viewBox="0 0 24 24"><rect width="14" height="14" x="5" y="5" rx="2"/></svg>'
      +     '<span>停止</span>'
      +   '</button>'
      + '</div>'
      + errHTML
      + '</div>';
  }).join('') + '</div>';
}

function renderPipeline(){
  const container = $('#pipelineContainer');
  if(!container) return;
  const directList = view.direct || [];
  const exitsList = view.exits || [];
  const allInbounds = [];
  directList.forEach(i => allInbounds.push({inbound: i, exit: null}));
  exitsList.forEach(e => (e.inbounds || []).forEach(i => allInbounds.push({inbound: i, exit: e})));

  if(allInbounds.length === 0){
    container.innerHTML = '<div class="section-card"><div class="section-header"><div class="section-title">'
      + '<svg class="icon" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><path d="M16 12H8"/><path d="m12 8 4 4-4 4"/></svg>'
      + '核心入站与流量拓扑管线</div></div>'
      + '<div class="empty">暂无入站节点。可在下方点击“新建节点”或在面板中配置 443 入站，即可开启流量拓扑调度。</div></div>';
    return;
  }

  // 1. 查找核心 443 入站（优先查找 443，若无则取首个入站）
  const coreItem = allInbounds.find(item => item.inbound.port === 443) || allInbounds[0];
  const coreNode = coreItem.inbound;
  const coreOwner = coreItem.exit;
  const otherItems = allInbounds.filter(item => item !== coreItem);

  const corePort = coreNode.port;
  const coreRemark = coreNode.remark ? esc(coreNode.remark) : ('入站 :' + corePort);
  const coreProto = (coreNode.protocol || 'PORT').toUpperCase();
  const coreListen = coreNode.listen ? (esc(coreNode.listen) + ':' + corePort) : ('0.0.0.0:' + corePort);

  // 阶段 3: 真实出网节点
  let stage3HTML = '';
  if(coreOwner){
    const country = esc(coreOwner.country || coreOwner.region || 'KR');
    const exitIP = esc(coreOwner.exit_ip || coreOwner.host || '—');
    const isp = esc((coreOwner.quality && coreOwner.quality.isp) ? coreOwner.quality.isp : (coreOwner.isp || '优质住宅网络'));
    const isRes = coreOwner.quality ? (coreOwner.quality.type === 'residential') : true;
    const score = coreOwner.quality && coreOwner.quality.score ? coreOwner.quality.score : 95;
    const qualityTag = isRes
      ? '<span class="quality-tag tag-residential"><span class="dot-indicator"></span> 住宅宽带 ' + score + '分</span>'
      : '<span class="quality-tag tag-datacenter"><span class="dot-indicator"></span> 机房节点 ' + score + '分</span>';

    let corePingHTML = '';
    if(coreOwner.ping && coreOwner.ping > 0){
      const corePing = coreOwner.ping;
      const pingCls = corePing < 80 ? 'ping-good' : (corePing < 180 ? 'ping-med' : 'ping-slow');
      corePingHTML = '<span class="metric-tag tag-ping ' + pingCls + '" title="节点实时网络延迟: ' + corePing + ' ms"><svg class="icon-nano" viewBox="0 0 24 24"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>' + corePing + ' ms</span>';
    }

    let coreSpdHTML = '';
    if(coreOwner.speed_mbps && coreOwner.speed_mbps > 0){
      const coreSpd = coreOwner.speed_mbps;
      const spdStr = coreSpd >= 100 ? coreSpd.toFixed(0) : coreSpd.toFixed(1);
      coreSpdHTML = '<span class="metric-tag tag-speed" title="节点测速带宽: ' + spdStr + ' Mbps"><svg class="icon-nano" viewBox="0 0 24 24"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>' + spdStr + ' Mbps</span>';
    }
    const corePerf = corePingHTML + coreSpdHTML;

    stage3HTML = '<div class="pipeline-stage exit-active" id="pipelineExitStage">'
      + '<div class="pipeline-stage-label">'
      +   '<span>真实出网节点</span>'
      +   '<span class="metric-badge badge-green" style="font-size:10px">动态住宅代理</span>'
      + '</div>'
      + '<div class="pipeline-stage-main">'
      +   '<span class="country-tag">' + country + '</span>'
      +   '<span class="mono">' + exitIP + '</span>'
      +   qualityTag
      +   corePerf
      + '</div>'
      + '<div class="pipeline-stage-sub">'
      +   '运营商: <span style="color:var(--text-main);font-weight:600">' + isp + '</span> • 保护母机 IP 绝不外泄'
      + '</div></div>';
  } else {
    const hostIP = esc(view.public_ip || '—');
    stage3HTML = '<div class="pipeline-stage exit-direct" id="pipelineExitStage">'
      + '<div class="pipeline-stage-label">'
      +   '<span>真实出网节点</span>'
      +   '<span class="metric-badge badge-amber" style="font-size:10px">母机公网直连</span>'
      + '</div>'
      + '<div class="pipeline-stage-main">'
      +   '<span class="country-tag">HOST</span>'
      +   '<span class="mono">' + hostIP + '</span>'
      +   '<span class="quality-tag tag-datacenter"><span class="dot-indicator"></span> 母机直连</span>'
      + '</div>'
      + '<div class="pipeline-stage-sub">'
      +   '母机公网直连 • 未绑定任何出口代理'
      + '</div></div>';
  }

  // 拓扑调度控制条选项
  let optionsHTML = '<option value=""' + (!coreOwner ? ' selected' : '') + '>原生直连 (母机公网出网: ' + esc(view.public_ip || '—') + ')</option>';
  exitsList.forEach(e => {
    const eCountry = esc(e.country || e.region || '');
    const eIP = esc(e.exit_ip || e.host);
    const eISP = esc((e.quality && e.quality.isp) ? e.quality.isp : (e.isp || ''));
    const eScore = e.quality && e.quality.score ? (' · ' + e.quality.score + '分') : '';
    const sel = (coreOwner && coreOwner.host === e.host) ? ' selected' : '';
    optionsHTML += '<option value="' + esc(e.host) + '"' + sel + '>[' + eCountry + '] ' + eIP + (eISP ? ' · ' + eISP + eScore : '') + '</option>';
  });

  const detailBtn = '<button type="button" class="btn btn-ghost" data-detail="' + coreNode.id + '" title="查看分享链接与节点参数">'
    + '<svg class="icon-xs" viewBox="0 0 24 24"><rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg>'
    + '<span>复制与详情</span></button>';

  const explorerBtn = '<button type="button" class="btn btn-crystal" id="openExplorerFromPipeline" title="从全节点大厅挑选并挂载">'
    + '<svg class="icon-xs" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><polygon points="16.24 7.76 14.12 14.12 7.76 16.24 9.88 9.88 16.24 7.76"/></svg>'
    + '<span>从大厅挑选挂载</span></button>';

  const unbindBtn = coreOwner
    ? '<button type="button" class="btn btn-ghost" style="color:var(--status-danger)" data-unbind-tag="' + esc(coreNode.tag) + '" data-unbind-port="' + coreNode.port + '" title="解除绑定，恢复母机原生直连">'
      + '<svg class="icon-xs" viewBox="0 0 24 24"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>'
      + '<span>解绑恢复直连</span></button>'
    : '';

  // 其他直连入站折叠区
  let subinboundsHTML = '';
  if(otherItems.length > 0){
    const subRows = otherItems.map(item => {
      const ib = item.inbound;
      const owner = item.exit;
      const ibPort = ib.port;
      const ibProto = (ib.protocol || 'PORT').toLowerCase();
      const ibRemark = ib.remark ? esc(ib.remark) : '自建入站';
      const statusBadge = owner
        ? '<span class="metric-badge badge-purple" style="font-size:10px">挂载至: ' + esc(owner.exit_ip || owner.host) + '</span>'
        : '<span class="metric-badge badge-amber" style="font-size:10px">直连出网: ' + esc(view.public_ip || '—') + '</span>';

      let rowOpts = '<option value=""' + (!owner ? ' selected' : '') + '>保持原生直连</option>';
      exitsList.forEach(e => {
        const sel = (owner && owner.host === e.host) ? ' selected' : '';
        rowOpts += '<option value="' + esc(e.host) + '"' + sel + '>挂载至 ' + esc(e.exit_ip || e.host) + ' (' + esc(e.region || '') + ')</option>';
      });

      const delBtn = isXCL() ? ''
        : '<button type="button" class="btn btn-ghost" style="padding:3px 6px;color:var(--status-danger)" data-delone="' + ib.id + '" data-name="'
          + esc((ib.remark || ib.protocol) + ' :' + ib.port) + '" title="删除这个入站">'
          + ICON.trash + '</button>';

      return '<div class="subinbound-row">'
        + '<div style="display:flex;align-items:center;gap:10px">'
        +   '<span class="mono" style="color:var(--accent-blue);font-weight:700">' + ibProto + ' :' + ibPort + '</span>'
        +   '<span>' + ibRemark + '</span>'
        +   statusBadge
        + '</div>'
        + '<div style="display:flex;align-items:center;gap:8px">'
        +   '<select class="obind select" style="padding:3px 8px;font-size:12px" data-tag="' + esc(ib.tag) + '" data-port="' + ibPort + '">' + rowOpts + '</select>'
        +   '<button type="button" class="btn btn-ghost" style="padding:3px 8px;font-size:12px" data-detail="' + ib.id + '">详情</button>'
        +   delBtn
        + '</div>'
        + '</div>';
    }).join('');

    subinboundsHTML = '<div class="subinbounds-area">'
      + '<div class="subinbounds-title">'
      +   '<span>其他入站节点 (共 ' + otherItems.length + ' 个)</span>'
      + '</div>'
      + subRows
      + '</div>';
  }

  const html = '<div class="section-card">'
    + '<div class="section-header">'
    +   '<div class="section-title">'
    +     '<svg class="icon" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><path d="M16 12H8"/><path d="m12 8 4 4-4 4"/></svg>'
    +     '核心入站与流量拓扑管线'
    +     '<span class="section-subtitle">客户端流量经由母机 :' + corePort + ' 入口流向住宅代理出口的全链路状态</span>'
    +   '</div>'
    + '</div>'
    + '<div class="pipeline-container">'
    +   '<!-- 阶段 1: 客户端入站入口 -->'
    +   '<div class="pipeline-stage core">'
    +     '<div class="pipeline-stage-label">'
    +       '<span>客户端入站入口</span>'
    +       '<span class="metric-badge badge-purple" style="font-size:10px">母机核心端口</span>'
    +     '</div>'
    +     '<div class="pipeline-stage-main">'
    +       '<span class="mono" style="color:var(--accent-purple)">:' + corePort + '</span>'
    +       '<span>' + coreRemark + '</span>'
    +     '</div>'
    +     '<div class="pipeline-stage-sub">'
    +       '协议: <span class="mono">' + coreProto + '</span> • 监听: <span class="mono">' + coreListen + '</span>'
    +     '</div>'
    +   '</div>'
    +   '<!-- 阶段 2: 动态调度引擎 -->'
    +   '<div class="pipeline-flow">'
    +     '<div class="flow-badge">分流调度</div>'
    +     '<div class="flow-line"></div>'
    +     '<span style="font-size:10px;color:var(--text-muted)">0ms 穿透</span>'
    +   '</div>'
    +   '<!-- 阶段 3: 真实出网节点 -->'
    +   stage3HTML
    + '</div>'
    + '<!-- 拓扑调度控制条 -->'
    + '<div class="pipeline-toolbar">'
    +   '<div class="select-wrap">'
    +     '<span class="select-label">调度出网出口:</span>'
    +     '<select class="obind select" data-tag="' + esc(coreNode.tag) + '" data-port="' + corePort + '">' + optionsHTML + '</select>'
    +   '</div>'
    +   '<div style="display:flex;gap:8px;flex-wrap:wrap">'
    +     detailBtn
    +     explorerBtn
    +     unbindBtn
    +   '</div>'
    + '</div>'
    + subinboundsHTML
    + '</div>';

  container.innerHTML = html;
}

function renderJobs(jobs){
  const box = $('#jobs');
  box.innerHTML = jobs.map(j => {
    const steps = j.steps.map(s => {
      const ic = {ok:ICON.ok, failed:ICON.bad, running:ICON.run}[s.status] || ICON.wait;
      const t = s.detail ? s.label + ' — ' + s.detail : s.label;
      return '<span class="step ' + s.status + '" title="' + esc(t) + '">' + ic
        + esc(s.status === 'ok' && s.detail ? s.detail : s.label) + '</span>';
    }).join('');
    const close = j.status === 'running' ? ''
      : '<button class="icon" data-job="' + esc(j.id) + '" title="关闭">' + ICON.x + '</button>';
    return '<div class="job"><div class="top"><strong>' + esc(j.summary) + '</strong>'
      + '<span class="count">' + j.done + '/' + j.total + '</span>'
      + '<span class="spacer"></span>' + close + '</div>'
      + '<div class="steps">' + steps + '</div></div>';
  }).join('');
}

function renderHeroMetrics(){
  const direct = view.direct || [];
  const exits = view.exits || [];
  const allInbounds = [];
  direct.forEach(i => allInbounds.push({inbound: i, exit: null}));
  exits.forEach(e => (e.inbounds || []).forEach(i => allInbounds.push({inbound: i, exit: e})));

  // 1. 核心入站 (:443)
  const coreItem = allInbounds.find(item => item.inbound.port === 443) || (allInbounds.length > 0 ? allInbounds[0] : null);
  const coreBadge = $('#heroCoreBadge');
  const coreProto = $('#heroCoreProto');
  const coreDesc = $('#heroCoreDesc');

  if(coreItem){
    const ib = coreItem.inbound;
    const isBound = Boolean(coreItem.exit);
    if(coreBadge){
      coreBadge.className = 'metric-badge ' + (isBound ? 'badge-purple' : 'badge-amber');
      coreBadge.textContent = isBound ? '已挂载住宅出口' : '母机原生直连';
    }
    if(coreProto){
      const pName = (ib.protocol || 'PORT').toUpperCase();
      coreProto.textContent = pName + (pName === 'VLESS' ? ' + Reality' : (' :' + ib.port));
    }
    if(coreDesc){
      const remark = ib.remark ? esc(ib.remark) : '核心端口 :443';
      const dest = isBound ? (esc(coreItem.exit.region || '') + ' ' + esc(coreItem.exit.exit_ip || coreItem.exit.host)) : '原生公网出网';
      coreDesc.innerHTML = '<span>' + remark + '</span><span style="color:var(--border-strong)">•</span><span class="mono" style="color:var(--accent-blue)">' + dest + '</span>';
    }
  } else {
    if(coreBadge){
      coreBadge.className = 'metric-badge badge-amber';
      coreBadge.textContent = '暂无 :443 入站';
    }
    if(coreProto) coreProto.textContent = ':443 未创建';
    if(coreDesc) coreDesc.innerHTML = '<span>可在面板中创建 443 节点</span>';
  }

  // 2. 活跃住宅出口
  const upExits = exits.filter(e => e.status === 'up');
  const totalExits = exits.length;
  const exitsBadge = $('#heroExitsBadge');
  const exitsCount = $('#heroExitsCount');
  const exitsDesc = $('#heroExitsDesc');

  if(exitsBadge){
    const allUp = totalExits > 0 && upExits.length === totalExits;
    exitsBadge.className = 'metric-badge ' + (totalExits === 0 ? 'badge-amber' : (allUp ? 'badge-green' : 'badge-amber'));
    exitsBadge.textContent = totalExits === 0 ? '暂无出口' : (allUp ? '100% 连通' : (upExits.length + '/' + totalExits + ' 连通'));
  }
  if(exitsCount){
    exitsCount.innerHTML = exits.length + ' <span style="font-size:14px;color:var(--text-muted);font-weight:400">个在线</span>';
  }
  if(exitsDesc){
    if(exits.length > 0){
      const regions = Array.from(new Set(exits.map(e => e.country || e.region).filter(Boolean))).slice(0, 2).join(' / ');
      const isps = exits.map(e => e.quality && e.quality.isp).filter(Boolean);
      const ispSample = isps.length > 0 ? isps[0] : '';
      exitsDesc.innerHTML = '<span>' + esc(regions || '全球出口') + (ispSample ? ' (' + esc(ispSample) + ')' : '') + '</span>';
    } else {
      exitsDesc.innerHTML = '<span>点击“新建出口”开启隧道</span>';
    }
  }

  // 3. 流量分流拓扑
  const totalIb = allInbounds.length;
  const boundCount = allInbounds.filter(item => Boolean(item.exit)).length;
  const directCount = allInbounds.filter(item => !item.exit).length;
  const splitPct = totalIb > 0 ? Math.round((boundCount / totalIb) * 100) : 0;
  const splitBadge = $('#heroSplitBadge');
  const splitVal = $('#heroSplitVal');
  const splitDesc = $('#heroSplitDesc');

  if(splitBadge){
    splitBadge.className = 'metric-badge ' + (boundCount > 0 ? 'badge-blue' : 'badge-amber');
    splitBadge.textContent = boundCount > 0 ? (splitPct + '% 住宅出网') : '100% 母机直连';
  }
  if(splitVal){
    splitVal.innerHTML = boundCount + ' <span style="font-size:14px;color:var(--text-muted);font-weight:400">出口挂载 · ' + directCount + ' 直连</span>';
  }
  if(splitDesc){
    splitDesc.innerHTML = '<span>宿主公网: <span class="mono" style="color:var(--text-main)">' + esc(view.public_ip || '未知') + '</span></span>';
  }

  // 4. 域名与 HTTPS
  const sslBadge = $('#heroSSLBadge');
  const sslHost = $('#heroSSLHost');
  const sslDesc = $('#heroSSLDesc');
  const hasDomain = Boolean(view.domain);
  const isTLS = Boolean(view.is_tls);
  const sslMode = view.ssl_mode || 'none';

  if(sslBadge){
    if(sslMode === 'caddy'){
      sslBadge.className = 'metric-badge badge-green';
      sslBadge.textContent = 'Caddy HTTPS';
    } else if(isTLS){
      sslBadge.className = 'metric-badge badge-green';
      sslBadge.textContent = '原生 HTTPS';
    } else if(hasDomain){
      sslBadge.className = 'metric-badge badge-blue';
      sslBadge.textContent = '域名外部反代';
    } else {
      sslBadge.className = 'metric-badge badge-amber';
      sslBadge.textContent = 'HTTP 明文';
    }
  }
  if(sslHost){
    sslHost.textContent = hasDomain ? view.domain : (view.public_ip || '未绑定域名');
  }
  if(sslDesc){
    if(hasDomain){
      sslDesc.innerHTML = '<span>' + (isTLS ? 'TLS 0ms证书热载' : '已映射分享与订阅源') + '</span>';
    } else {
      sslDesc.innerHTML = '<span>可在设置中绑定域名</span>';
    }
  }

  // 顶栏宿主机信息药丸
  const hostPill = $('#headerHostPill');
  if(hostPill){
    const pInfo = view.panel_info || backendName();
    const pip = view.public_ip || '—';
    hostPill.innerHTML = '<span class="pulse-dot"></span>'
      + '<span>' + esc(pInfo) + '</span>'
      + '<span style="color:var(--border-strong)">•</span>'
      + '<span class="mono" style="color:var(--text-dim)">宿主机: ' + esc(pip) + '</span>';
  }
}

async function poll(){
  try{
    view = await api('/api/exits');
    const pEl = $('#panel');
    if(pEl){
      pEl.textContent = view.panel
        ? (backendName() + ': ' + view.panel)
        : (view.panel_info || '');
    }
    // xray-cf-lite 的节点由它自己生成，fanout 这边只管把它们导到哪条出口
    if($('#newnode')) $('#newnode').hidden = isXCL();
    // 链接由 xray-cf-lite 的订阅体系发，fanout 这边导不出来
    if($('#exportAll')) $('#exportAll').hidden = isXCL();
    renderHeroMetrics();
    renderExits();
    renderPipeline();
  }catch(e){}
  try{ renderJobs(await api('/api/jobs') || []); }catch(e){}
}

// ---- 新建向导 ----
let regions = [], region = '', regionsLoaded = false;

function openModal(id){ $('#' + id).classList.add('open'); }
function closeModal(id){ $('#' + id).classList.remove('open'); }

let confirmResolver = null;
function showConfirm(msg, opts = {}){
  return new Promise(resolve => {
    const titleEl = $('#confirmTitle');
    const msgEl = $('#confirmMsg');
    const iconBox = $('#confirmIconBox');
    const okBtn = $('#confirmOkBtn');
    const cancelBtn = $('#confirmCancelBtn');

    titleEl.textContent = opts.title || '操作确认';
    msgEl.textContent = msg;

    if(opts.danger){
      iconBox.className = 'confirm-icon-box danger';
      iconBox.innerHTML = '<svg class="icon-md" viewBox="0 0 24 24"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>';
      okBtn.className = 'btn btn-danger';
      okBtn.textContent = opts.okText || '确认操作';
    } else {
      iconBox.className = 'confirm-icon-box';
      iconBox.innerHTML = '<svg class="icon-md" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>';
      okBtn.className = 'btn primary';
      okBtn.textContent = opts.okText || '确定';
    }
    cancelBtn.textContent = opts.cancelText || '取消';

    confirmResolver = resolve;
    openModal('confirmModal');
    setTimeout(() => { (opts.danger ? cancelBtn : okBtn).focus(); }, 40);
  });
}

function resolveConfirm(res){
  if(confirmResolver){
    const fn = confirmResolver;
    confirmResolver = null;
    closeModal('confirmModal');
    fn(res);
  }
}

document.addEventListener('click', e => {
  const c = e.target.closest('[data-close]');
  if(c) closeModal(c.dataset.close);
});
document.addEventListener('keydown', e => {
  if(e.key === 'Escape'){
    if($('#confirmModal') && $('#confirmModal').classList.contains('open')){
      resolveConfirm(false);
      return;
    }
    document.querySelectorAll('.modal.open').forEach(m => m.classList.remove('open'));
  }
});
document.querySelectorAll('.modal').forEach(m => {
  m.onclick = e => {
    if(e.target === m){
      if(m.id === 'confirmModal') resolveConfirm(false);
      else m.classList.remove('open');
    }
  };
});

function renderRegions(){
  const kw = $('#rgfilter').value.trim().toLowerCase();
  const list = regions.filter(r => !kw
    || r.code.toLowerCase().includes(kw) || r.name.toLowerCase().includes(kw));
  $('#regions').innerHTML = ['<button class="rg' + (region === '' ? ' sel' : '')
      + '" data-rg=""><b>不限地区</b><em>速度优先</em></button>']
    .concat(list.map(r => '<button class="rg' + (region === r.code ? ' sel' : '')
      + '" data-rg="' + esc(r.code) + '"><b>' + esc(r.code) + ' ' + esc(r.name) + '</b>'
      + '<em>' + r.available + ' 个空闲 · ' + r.best_speed_mbps.toFixed(0) + ' Mbps</em></button>'))
    .join('');
  updateAvail();
}

function availOf(code){
  if(code === '') return regions.reduce((a, r) => a + r.available, 0);
  const r = regions.find(x => x.code === code);
  return r ? r.available : 0;
}

function updateAvail(){
  const avail = availOf(region);
  const want = Number($('#count').value) || 0;
  const hint = $('#availhint');
  hint.textContent = avail ? '可用 ' + avail + ' 个节点' : '这个地区没有空闲节点';
  hint.className = 'hint' + (want > avail ? ' bad' : '');
  if(want > avail && avail) hint.textContent = '只剩 ' + avail + ' 个，将全部使用';
  $('#go').disabled = !avail;
}

async function loadWizard(){
  try{
    regions = await api('/api/regions') || [];
    regionsLoaded = true;
    renderRegions();
  }catch(e){ toast('读取地区失败: ' + e.message, true); }

  const sel = $('#tpl');
  // xray-cf-lite 模式不能复制节点，向导退化成"只开出口"，之后在节点详情里挑出口
  if(isXCL()){
    $('#tplwrap').hidden = true;
    sel.innerHTML = '<option value="0">只开出口，不建节点</option>';
    return;
  }
  $('#tplwrap').hidden = false;
  try{
    // 已经挂在出口上的多半是上一批复制出来的，拿它当模板会套娃，
    // 所以把没绑出口的排在前面并默认选中
    const v = await api('/api/exits');
    const free = v.direct || [];
    const bound = (v.exits || []).flatMap(e => e.inbounds || []);
    inbounds = free.concat(bound);
    if(!inbounds.length){
      sel.innerHTML = '<option value="0">还没有节点</option>';
      $('#tplhint').textContent = '先用上面的「新建节点」建一个，之后这里可以按它批量生成';
      return;
    }
    const opt = i => '<option value="' + i.id + '">'
      + esc(i.remark || ('端口 ' + i.port)) + ' · ' + esc(i.protocol)
      + ' :' + i.port + '</option>';
    sel.innerHTML =
      (free.length ? '<optgroup label="未绑定出口">' + free.map(opt).join('') + '</optgroup>' : '')
      + (bound.length ? '<optgroup label="已挂在出口上">' + bound.map(opt).join('') + '</optgroup>' : '')
      + '<option value="0">只开出口，不建节点</option>';
    $('#tplhint').textContent = '每个出口复制一份，客户端 UUID 保持一致，只有端口不同';
  }catch(e){
    sel.innerHTML = '<option value="0">' + backendName() + '不可用</option>';
    $('#tplhint').textContent = e.message;
  }
}

document.addEventListener('click', e => {
  if(e.target.closest('#newexit') || e.target.closest('#newexit2')){
    openModal('wizard');
    if(!regionsLoaded) loadWizard(); else { renderRegions(); loadWizard(); }
  }
  const rg = e.target.closest('[data-rg]');
  if(rg){ region = rg.dataset.rg; renderRegions(); }
});

// ---- 新建节点 ----
document.addEventListener('click', e => {
  if(e.target.closest('#newnode') || e.target.closest('#newnode2')){
    $('#nnhint').textContent = '';
    syncNodeForm();
    openModal('newnodebox');
    renderPortRecommendations('nport-chips', 'nport-status', 'nport', 'ncreate', 0);
  }
});

// 表单随协议/传输/安全层联动：只露出当前组合真正用得到的字段
function syncNodeForm(){
  const proto = $('#nproto').value;
  const net   = $('#nnet').value;
  const sec   = $('#nsec').value;

  // REALITY 靠模仿 TLS 握手工作，套在自带头部的传输上没有意义
  const realityOK = net === 'tcp' || net === 'xhttp' || net === 'grpc';
  const secSel = $('#nsec');
  for(const o of secSel.options){
    if(o.value === 'reality') o.disabled = !realityOK;
  }
  if(secSel.value === 'reality' && !realityOK) secSel.value = 'none';

  const cur = secSel.value;
  $('#nsniwrap').hidden  = cur !== 'tls';
  $('#ncertwrap').hidden = cur !== 'tls';
  $('#nkeywrap').hidden  = cur !== 'tls';
  $('#ndestwrap').hidden = cur !== 'reality';

  // Vision 只在 VLESS + 裸 TCP + TLS/REALITY 下有效
  const visionOK = proto === 'vless' && net === 'tcp' && cur !== 'none';
  $('#nvisionwrap').hidden = !visionOK;
  if(!visionOK) $('#nvision').checked = false;

  const needPath = net === 'ws' || net === 'httpupgrade' || net === 'xhttp' || net === 'grpc';
  $('#npathwrap').hidden = !needPath;
  $('#npathlabel').textContent = net === 'grpc' ? '服务名' : '路径';

  $('#nsechint').textContent =
    cur === 'reality' ? '密钥与 shortId 自动生成' :
    cur === 'tls'     ? '不填证书就用自签，链接会带证书指纹' : '';
}
$('#nproto').onchange = syncNodeForm;
$('#nnet').onchange = syncNodeForm;
$('#nsec').onchange = syncNodeForm;

$('#ncreate').onclick = async e => {
  const q = new URLSearchParams({
    protocol: $('#nproto').value,
    network:  $('#nnet').value,
    security: $('#nsec').value,
    port:     ($('#nport').value || '').trim(),
    remark:   ($('#nremark').value || '').trim(),
    path:     ($('#npath').value || '').trim(),
    sni:      ($('#nsni').value || '').trim(),
    cert:     ($('#ncert').value || '').trim(),
    key:      ($('#nkey').value || '').trim(),
    dest:     ($('#ndest').value || '').trim(),
  });
  if($('#nvision').checked) q.set('vision', '1');
  e.target.disabled = true;
  try{
    const r = await api('/api/panel/inbound/new?' + q.toString(), {method:'POST'});
    toast('已创建 ' + r.protocol + ' 节点，端口 ' + r.port);
    closeModal('newnodebox');
    $('#nport').value = '';
    $('#nremark').value = '';
    poll();
  }catch(err){ toast(err.message, true); }
  e.target.disabled = false;
};

$('#rgfilter').oninput = renderRegions;
$('#minus').onclick = () => { step(-1); };
$('#plus').onclick = () => { step(1); };
function step(d){
  const el = $('#count');
  el.value = Math.min(20, Math.max(1, (Number(el.value) || 1) + d));
  updateAvail();
}
$('#count').oninput = updateAvail;

$('#go').onclick = async e => {
  const want = Math.min(Number($('#count').value) || 1, availOf(region) || 1);
  const tpl = $('#tpl').value || '0';
  e.target.disabled = true;
  try{
    await api('/api/provision?count=' + want + '&region=' + encodeURIComponent(region)
      + '&template=' + tpl, {method:'POST'});
    closeModal('wizard');
    poll();
  }catch(err){ toast(err.message, true); }
  e.target.disabled = false;
};

// ---- 出口操作 ----
document.addEventListener('click', async e => {
  const stop = e.target.closest('[data-stop]');
  if(stop){
    stop.disabled = true;
    try{ await api('/api/stop?slot=' + stop.dataset.stop, {method:'POST'}); }
    catch(err){ toast(err.message, true); }
    poll();
    return;
  }
  const swap = e.target.closest('[data-swap]');
  if(swap){
    swap.disabled = true;
    try{
      await api('/api/swap?slot=' + swap.dataset.swap, {method:'POST'});
      toast('正在换节点');
    }catch(err){ toast(err.message, true); }
    poll();
    return;
  }
  const retry = e.target.closest('[data-retry]');
  if(retry){
    retry.disabled = true;
    try{
      await api('/api/retry?slot=' + retry.dataset.retry, {method:'POST'});
      toast('正在重新连接节点');
    }catch(err){ toast(err.message, true); }
    poll();
    return;
  }
  const cred = e.target.closest('[data-cred]');
  if(cred){ openCred(Number(cred.dataset.cred)); return; }
  const job = e.target.closest('[data-job]');
  if(job){
    try{ await api('/api/jobs/dismiss?id=' + job.dataset.job, {method:'POST'}); }catch(err){}
    poll();
    return;
  }
  const del = e.target.closest('[data-delorphans]');
  if(del){
    const list = view.direct || [];
    if(!(await showConfirm('确定要删除这 ' + list.length + ' 个未绑定节点？此操作不可撤销。', {danger:true, title:'清理未绑定节点', okText:'确认清理'}))) return;
    del.disabled = true;
    try{
      await api('/api/xui/delete?ids=' + list.map(i => i.id).join(','), {method:'POST'});
      toast('已清理 ' + list.length + ' 个入站');
    }catch(err){ toast(err.message, true); }
    poll();
    return;
  }

  const one = e.target.closest('[data-delone]');
  if(one){
    if(!(await showConfirm('确定要删除入站 ' + one.dataset.name + '？此操作不可撤销。', {danger:true, title:'删除入站', okText:'确认删除'}))) return;
    one.disabled = true;
    try{
      await api('/api/xui/delete?ids=' + one.dataset.delone, {method:'POST'});
      toast('已删除 ' + one.dataset.name);
    }catch(err){ toast(err.message, true); }
    poll();
  }
});

$('#stopall').onclick = async e => {
  if(!view.exits || !view.exits.length) return;
  if(!(await showConfirm('确定要停止全部 ' + view.exits.length + ' 个出口？已绑定的流量将被中断。', {danger:true, title:'停止全部出口', okText:'停止全部'}))) return;
  e.target.disabled = true;
  for(const x of view.exits){
    try{ await api('/api/stop?slot=' + x.slot, {method:'POST'}); }catch(err){}
  }
  poll();
};

// ---- 为出口绑定节点（支持复制原节点直连保留，或转移直连） ----
let curBindExit = null;

document.addEventListener('click', e => {
  const btn = e.target.closest('[data-bind-slot]');
  if(btn){
    openBindExitModal({
      slot: btn.dataset.bindSlot,
      host: btn.dataset.bindHost,
      ip: btn.dataset.bindIp,
      region: btn.dataset.bindRegion
    });
  }
});

function openBindExitModal(target){
  curBindExit = target;
  $('#bem-target').innerHTML = '目标出口：<b>' + esc(target.ip) + '</b> (' + esc(target.region || '—') + ' · ' + esc(target.host) + ')';

  // 收集可用的所有模板节点（包含 direct 和已绑定 exits 的节点）
  const allNodes = [];
  (view.direct || []).forEach(i => allNodes.push(i));
  (view.exits || []).forEach(ex => (ex.inbounds || []).forEach(i => allNodes.push(i)));

  // 填充模板节点下拉
  const tplSel = $('#bem-tpl-node');
  if(allNodes.length === 0){
    tplSel.innerHTML = '<option value="">(暂无可用模板节点，请先新建节点)</option>';
    $('#bem-btn-clone').disabled = true;
  } else {
    tplSel.innerHTML = allNodes.map(i =>
      '<option value="' + i.id + '">[' + esc(i.protocol) + '] :' + i.port + ' · ' + esc(i.remark || i.protocol) + '</option>'
    ).join('');
    $('#bem-btn-clone').disabled = false;
  }

  // 填充转移直连下拉
  const directSel = $('#bem-direct-node');
  const directNodes = view.direct || [];
  if(directNodes.length === 0){
    directSel.innerHTML = '<option value="">(暂无直连节点可供转移)</option>';
    $('#bem-btn-move').disabled = true;
  } else {
    directSel.innerHTML = directNodes.map(i =>
      '<option value="' + i.id + '" data-tag="' + esc(i.tag) + '" data-port="' + i.port + '">[' + esc(i.protocol) + '] :' + i.port + ' · ' + esc(i.remark || i.protocol) + '</option>'
    ).join('');
    $('#bem-btn-move').disabled = false;
  }

  // 初始化模式：默认克隆
  switchBindMode('clone');

  // 计算推荐端口
  const tplPort = allNodes.length > 0 ? allNodes[0].port : 2087;
  let nextPort = tplPort + 1;
  const occupied = new Set();
  allNodes.forEach(i => occupied.add(i.port));
  while(occupied.has(nextPort) && nextPort < 65535){
    nextPort++;
  }
  $('#bem-port').value = nextPort < 65535 ? nextPort : '';
  renderPortRecommendations('bem-port-chips', 'bem-port-status', 'bem-port', 'bem-btn-clone', 0);

  openModal('bindExitModal');
}

function switchBindMode(mode){
  if(mode === 'clone'){
    $('#bmode-radio-clone').checked = true;
    $('#bmode-radio-move').checked = false;
    $('#bmode-clone-card').classList.add('active');
    $('#bmode-move-card').classList.remove('active');
    $('#bmode-clone-fields').style.display = 'block';
    $('#bmode-move-fields').style.display = 'none';
  } else {
    $('#bmode-radio-clone').checked = false;
    $('#bmode-radio-move').checked = true;
    $('#bmode-clone-card').classList.remove('active');
    $('#bmode-move-card').classList.add('active');
    $('#bmode-clone-fields').style.display = 'none';
    $('#bmode-move-fields').style.display = 'block';
  }
}

document.addEventListener('click', e => {
  const card = e.target.closest('.bmode-item');
  if(!card) return;
  if(card.id === 'bmode-clone-card' && !e.target.closest('#bmode-clone-fields')){
    switchBindMode('clone');
  } else if(card.id === 'bmode-move-card' && !e.target.closest('#bmode-move-fields')){
    switchBindMode('move');
  }
});

document.addEventListener('click', async e => {
  if(e.target.closest('#bem-btn-clone')){
    const btn = $('#bem-btn-clone');
    if(!curBindExit) return;
    const tplId = $('#bem-tpl-node').value;
    if(!tplId){
      toast('请选择模板节点', true);
      return;
    }
    const port = parseInt($('#bem-port').value, 10) || 0;
    btn.disabled = true;
    try{
      const url = '/api/xui/clone?id=' + encodeURIComponent(tplId)
        + '&hosts=' + encodeURIComponent(curBindExit.host)
        + (port > 0 ? '&port=' + port : '');
      const res = await api(url, {method:'POST'});
      const newPort = (res.created && res.created.length) ? res.created[0] : port;
      toast('🎉 已复制原节点并绑定到出口！(端口 :' + newPort + ')');
      closeModal('bindExitModal');
      poll();
    }catch(err){
      toast(err.message, true);
    }
    btn.disabled = false;
    return;
  }

  if(e.target.closest('#bem-btn-move')){
    const btn = $('#bem-btn-move');
    if(!curBindExit) return;
    const sel = $('#bem-direct-node');
    const opt = sel.options[sel.selectedIndex];
    if(!opt || !opt.dataset.tag){
      toast('请选择要转移的节点', true);
      return;
    }
    btn.disabled = true;
    try{
      await api('/api/xui/bind?tag=' + encodeURIComponent(opt.dataset.tag)
        + '&port=' + encodeURIComponent(opt.dataset.port || '')
        + '&host=' + encodeURIComponent(curBindExit.host), {method:'POST'});
      toast('已转移改绑至该出口');
      closeModal('bindExitModal');
      poll();
    }catch(err){
      toast(err.message, true);
    }
    btn.disabled = false;
    return;
  }
});

// ---- 节点详情 ----
let curDetail = null;

// 详情弹窗的重绘要跟轮询解耦：正在编辑时被 poll 刷掉输入会很烦
async function openDetail(id){
  $('#dbody').innerHTML = '<div class="empty">读取中…</div>';
  curDetail = null;
  $('#ddel').disabled = true;
  openModal('detail');
  try{
    const d = await api('/api/xui/detail?id=' + id);
    curDetail = d;
    renderDetail(d);
    $('#ddel').hidden = isXCL();
    $('#ddel').disabled = isXCL();
  }catch(err){
    $('#dbody').innerHTML = '<div class="empty">读取失败: ' + esc(err.message) + '</div>';
  }
}

// 出口下拉：列出所有已连通的隧道，外加"直连"。绑定按 Xray 的 inboundTag 走。
function exitOptions(currentHost){
  const exits = (view.exits || []).filter(e => e.status === 'up' || (currentHost && e.host === currentHost));
  return '<option value=""' + (currentHost ? '' : ' selected') + '>🌐 直连（母机公网出网）</option>'
    + exits.map(e => '<option value="' + esc(e.host) + '"'
        + (e.host === currentHost ? ' selected' : '') + '>'
        + esc((e.exit_ip || e.host) + ' · ' + e.region) + (e.status !== 'up' ? ' (' + (STATUS[e.status]||e.status) + ')' : '') + '</option>').join('');
}

function renderDetail(d){
  const owner = view.exits.find(x => (x.inbounds || []).some(i => i.id === d.id));

  const clients = (d.clients || []).map((c, i) => {
    const link = (d.links || [])[i] || '';
    return '<div class="client">'
      + '<div class="crow">'
      +   '<span class="cemail">' + esc(c.email) + '</span>'
      +   '<span class="cid">' + esc(c.id) + '</span>'
      +   '<span class="spacer"></span>'
      +   (link ? '<button class="icon" data-copy="' + esc(link) + '" title="复制链接">' + ICON.copy + '</button>' : '')
      +   '<button class="icon" data-creset="' + esc(c.email) + '" title="换一套凭据，旧链接立即失效">' + ICON.redo + '</button>'
      +   '<button class="icon" data-cdel="' + esc(c.email) + '" title="删除这个客户端">' + ICON.trash + '</button>'
      + '</div>'
      + (link ? '<div class="share">' + esc(link) + '</div>' : '')
      + '</div>';
  }).join('');

  $('#dtitle').textContent = (d.remark || '节点') + '　:' + d.port;
  const editable = !isXCL();

  const unbindBar = owner
    ? '<div class="unbind-bar"><span class="ub-hint">已挂载出口: <b>' + esc(owner.exit_ip || owner.host) + '</b> (' + esc(owner.region) + ')</span>'
      + '<button type="button" class="btn-unbind" id="dquick-unbind" data-tag="' + esc(d.tag) + '" data-port="' + d.port + '">🌐 一键解绑（恢复母机直连）</button></div>'
    : '<div class="direct-bar">🌐 当前处于母机原生直连出网 (母机公网 IP: <b>' + esc(view.public_ip || '—') + '</b>)</div>';

  $('#dbody').innerHTML = unbindBar
    + '<dl class="kv">'
    + '<dt>出口</dt><dd><select id="dbind" data-tag="' + esc(d.tag) + '" data-port="' + d.port + '">'
    +   exitOptions(owner ? owner.host : '') + '</select></dd>'
    + '<dt>协议</dt><dd>' + esc(d.protocol) + '　' + esc(d.network || '')
    +   (d.tls && d.tls !== 'none' ? '　' + esc(d.tls) : '') + '</dd>'
    + '<dt>监听</dt><dd>' + esc(d.listen || '0.0.0.0') + '</dd>'
    + '</dl>'
    + (editable ? ('<div class="editbar">'
    +   '<label class="ef"><span>备注</span>'
    +     '<input id="dremark" type="text" value="' + esc(d.remark || '') + '"></label>'
    +   '<label class="ef"><span>端口</span>'
    +     '<input id="dport" type="text" inputmode="numeric" value="' + d.port + '"></label>'
    +   '<div class="port-recom-wrap" style="width:100%">'
    +     '<div class="port-recom-title">推荐常用端口（点击填入）：</div>'
    +     '<div class="port-chips" id="dport-chips"></div>'
    +     '<div class="port-status" id="dport-status"></div>'
    +   '</div>'
    +   '<label class="chk"><input type="checkbox" id="denable"'
    +     (d.enable === false ? '' : ' checked') + '> 启用</label>'
    +   '<span class="spacer"></span>'
    +   '<button class="primary" id="dsave">保存</button>'
    + '</div>'
    + '<div class="chead"><h3>客户端</h3><span class="count">'
    +   (d.clients || []).length + ' 个</span><span class="spacer"></span>'
    +   '<button id="dcadd">' + ICON.plus + '添加</button></div>'
    + (clients || '<div class="empty">没有客户端</div>'))
    : '<div class="hint">这个节点由 xray-cf-lite 管，端口、UUID 和分享链接都去它那边改。这里只决定它走哪条出口。</div>');

  if(editable){
    renderPortRecommendations('dport-chips', 'dport-status', 'dport', 'dsave', d.port);
  }
}

// 未绑定区的出口下拉，选中即绑
document.addEventListener('change', async e => {
  const sel = e.target.closest('.obind');
  if(!sel) return;
  sel.disabled = true;
  const host = sel.value || 'direct';
  try{
    await api('/api/xui/bind?tag=' + encodeURIComponent(sel.dataset.tag)
      + '&port=' + encodeURIComponent(sel.dataset.port || '')
      + '&host=' + encodeURIComponent(host), {method:'POST'});
    toast(sel.value ? '已绑定出口' : '已恢复母机原生直连');
    await poll();
  }catch(err){ toast(err.message, true); sel.disabled = false; }
});

// 出口下拉改动即生效。绑定按 inboundTag 走，host 传空表示解绑回直连。
document.addEventListener('change', async e => {
  const sel = e.target.closest('#dbind');
  if(!sel) return;
  sel.disabled = true;
  const host = sel.value || 'direct';
  try{
    await api('/api/xui/bind?tag=' + encodeURIComponent(sel.dataset.tag)
      + '&port=' + encodeURIComponent(sel.dataset.port || '')
      + '&host=' + encodeURIComponent(host), {method:'POST'});
    toast(sel.value ? '已绑定出口' : '已解绑（恢复母机原生直连）');
    await poll();
    if(curDetail) await openDetail(curDetail.id);
  }catch(err){ toast(err.message, true); }
  sel.disabled = false;
});

document.addEventListener('click', async e => {
  // 一键解绑恢复母机直连 (详情弹窗)
  const ub = e.target.closest('#dquick-unbind');
  if(ub){
    ub.disabled = true;
    try{
      await api('/api/xui/bind?tag=' + encodeURIComponent(ub.dataset.tag)
        + '&port=' + encodeURIComponent(ub.dataset.port || '')
        + '&host=direct', {method:'POST'});
      toast('已解除绑定，恢复母机原生直连');
      await poll();
      if(curDetail) await openDetail(curDetail.id);
    }catch(err){ toast(err.message, true); ub.disabled = false; }
    return;
  }

  // 快速解绑按钮 (出口卡片上的入站胶囊)
  const unbindBtn = e.target.closest('[data-unbind-tag]');
  if(unbindBtn){
    e.stopPropagation();
    unbindBtn.disabled = true;
    const tag = unbindBtn.dataset.unbindTag;
    const port = unbindBtn.dataset.unbindPort;
    try{
      await api('/api/xui/bind?tag=' + encodeURIComponent(tag)
        + '&port=' + encodeURIComponent(port || '')
        + '&host=direct', {method:'POST'});
      toast('已解除绑定，恢复母机原生直连');
      await poll();
      if(curDetail && $('#detail').classList.contains('open')) await openDetail(curDetail.id);
    }catch(err){
      toast(err.message, true);
      unbindBtn.disabled = false;
    }
    return;
  }

  // 打开全节点大厅
  if(e.target.closest('#openExplorerBtn') || e.target.closest('#openExplorerBtn2') || e.target.closest('#openExplorerFromDirect') || e.target.closest('.btn-start-core-exit') || e.target.closest('#openExplorerFromPipeline')){
    openNodeExplorer();
    return;
  }

  // 大厅中开启出口
  const sh = e.target.closest('[data-start-host]');
  if(sh){
    sh.disabled = true;
    const host = sh.dataset.startHost;
    try{
      await api('/api/nodes/start?host=' + encodeURIComponent(host), {method:'POST'});
      toast('正在开启出口隧道...');
      closeModal('nodeExplorer');
      poll();
    }catch(err){
      toast(err.message, true);
      sh.disabled = false;
    }
    return;
  }

  // 大厅中挂载到 443
  const mh = e.target.closest('[data-mount-host]');
  if(mh){
    mh.disabled = true;
    const host = mh.dataset.mountHost;
    const port = mh.dataset.mountPort || '443';
    try{
      await api('/api/nodes/start?host=' + encodeURIComponent(host) + '&bind_port=' + encodeURIComponent(port), {method:'POST'});
      toast('正在启动住宅节点并挂载到 :' + port + ' ...');
      closeModal('nodeExplorer');
      poll();
    }catch(err){
      toast(err.message, true);
      mh.disabled = false;
    }
    return;
  }

  // 大厅排序切换
  const sbtn = e.target.closest('.nex-sort-btn');
  if(sbtn){
    document.querySelectorAll('.nex-sort-btn').forEach(b => b.classList.remove('active'));
    sbtn.classList.add('active');
    nexSort = sbtn.dataset.sort;
    loadNodeExplorer();
    return;
  }

  // 大厅类型过滤切换
  const tchip = e.target.closest('.nex-filter-chip[data-type]');
  if(tchip){
    document.querySelectorAll('.nex-filter-chip[data-type]').forEach(b => b.classList.remove('active'));
    tchip.classList.add('active');
    nexType = tchip.dataset.type;
    loadNodeExplorer();
    return;
  }

  // 大厅地区过滤切换
  const rchip = e.target.closest('.nex-filter-chip[data-nreg]');
  if(rchip){
    document.querySelectorAll('.nex-filter-chip[data-nreg]').forEach(b => b.classList.remove('active'));
    rchip.classList.add('active');
    nexRegion = rchip.dataset.nreg;
    loadNodeExplorer();
    return;
  }

  const link = e.target.closest('[data-detail]');
  if(link) return openDetail(link.dataset.detail);

  if(e.target.closest('#dsave')){
    const btn = e.target.closest('#dsave');
    btn.disabled = true;
    const q = new URLSearchParams({
      id: curDetail.id,
      port: ($('#dport').value || '').trim(),
      remark: ($('#dremark').value || '').trim(),
      enable: $('#denable').checked ? '1' : '0',
    });
    try{
      await api('/api/panel/inbound/update?' + q, {method:'POST'});
      toast('已保存');
      await openDetail(curDetail.id);
      poll();
    }catch(err){ toast(err.message, true); btn.disabled = false; }
    return;
  }

  const add = e.target.closest('#dcadd');
  if(add){
    add.disabled = true;
    try{
      await api('/api/panel/client/add?id=' + curDetail.id, {method:'POST'});
      toast('已添加客户端');
      await openDetail(curDetail.id);
    }catch(err){ toast(err.message, true); add.disabled = false; }
    return;
  }

  const del = e.target.closest('[data-cdel]');
  if(del){
    if(!(await showConfirm('确定要删除客户端 ' + del.dataset.cdel + '？其连接将立即失效。', {danger:true, title:'删除客户端', okText:'确认删除'}))) return;
    del.disabled = true;
    try{
      await api('/api/panel/client/del?id=' + curDetail.id
        + '&email=' + encodeURIComponent(del.dataset.cdel), {method:'POST'});
      toast('已删除');
      await openDetail(curDetail.id);
    }catch(err){ toast(err.message, true); del.disabled = false; }
    return;
  }

  const reset = e.target.closest('[data-creset]');
  if(reset){
    if(!(await showConfirm('确定要重置 ' + reset.dataset.creset + ' 的凭据？已分发的旧链接将立即失效。', {danger:true, title:'重置客户端凭据', okText:'确认重置'}))) return;
    reset.disabled = true;
    try{
      await api('/api/panel/client/reset?id=' + curDetail.id
        + '&email=' + encodeURIComponent(reset.dataset.creset), {method:'POST'});
      toast('已重置');
      await openDetail(curDetail.id);
    }catch(err){ toast(err.message, true); reset.disabled = false; }
    return;
  }

  // 详情弹窗里删掉当前这个入站
  const dd = e.target.closest('#ddel');
  if(dd && curDetail){
    const name = (curDetail.remark || curDetail.protocol || '节点') + ' :' + curDetail.port;
    if(!(await showConfirm('确定要删除入站 ' + name + '？它的所有客户端链接都将失效，且不可撤销。', {danger:true, title:'删除入站', okText:'确认删除'}))) return;
    dd.disabled = true;
    try{
      await api('/api/xui/delete?ids=' + curDetail.id, {method:'POST'});
      toast('已删除 ' + name);
      curDetail = null;
      closeModal('detail');
      poll();
    }catch(err){ toast(err.message, true); dd.disabled = false; }
  }
});

// Node Explorer 控制器
let nexSort = 'quality';
let nexRegion = '';
let nexType = '';
let nexNodes = [];
let nexLoading = false;

async function openNodeExplorer(){
  openModal('nodeExplorer');
  renderExplorerRegions();
  loadNodeExplorer();
}

let nexKwTimer;
document.addEventListener('input', e => {
  if(e.target && e.target.id === 'nex-kw'){
    clearTimeout(nexKwTimer);
    nexKwTimer = setTimeout(() => {
      loadNodeExplorer();
    }, 250);
  }
});

async function loadNodeExplorer(){
  const listEl = $('#nex-list');
  const countEl = $('#nex-count');
  listEl.innerHTML = '<div class="empty">正在加载全部节点并评估 IP 纯净度画像…</div>';
  nexLoading = true;
  try{
    const q = new URLSearchParams({
      sort: nexSort,
    });
    if(nexRegion) q.set('region', nexRegion);
    if(nexType) q.set('type', nexType);
    const kw = ($('#nex-kw') ? $('#nex-kw').value : '').trim();
    if(kw) q.set('kw', kw);

    const res = await api('/api/nodes?' + q.toString());
    nexNodes = res.nodes || [];
    countEl.textContent = '共 ' + nexNodes.length + ' 个节点';
    const navCnt = $('#navNodeCount');
    if(navCnt && nexNodes.length > 0) navCnt.textContent = '(' + nexNodes.length + ')';
    renderNodeExplorerList(nexNodes);
  }catch(err){
    listEl.innerHTML = '<div class="empty">加载节点失败: ' + esc(err.message) + '</div>';
    countEl.textContent = '加载失败';
  }
  nexLoading = false;
}

function renderExplorerRegions(){
  const regBox = $('#nex-regions');
  if(!regBox || regBox.children.length > 0) return;
  const commonRegions = [
    {code:'', name:'全部地区'},
    {code:'JP', name:'🇯🇵 日本'},
    {code:'US', name:'🇺🇸 美国'},
    {code:'KR', name:'🇰🇷 韩国'},
    {code:'TW', name:'🇹🇼 台湾'},
    {code:'HK', name:'🇭🇰 香港'},
    {code:'SG', name:'🇸🇬 新加坡'},
    {code:'GB', name:'🇬🇧 英国'},
    {code:'DE', name:'🇩🇪 德国'},
  ];
  regBox.innerHTML = commonRegions.map(r =>
    '<button type="button" class="nex-filter-chip ' + (nexRegion === r.code ? 'active' : '') + '" data-nreg="' + r.code + '">'
    + esc(r.name) + '</button>'
  ).join('');
}

function renderNodeExplorerList(nodes){
  const listEl = $('#nex-list');
  if(!nodes || !nodes.length){
    listEl.innerHTML = '<div class="empty">未找到符合条件的节点</div>';
    return;
  }
  const primPort = getPrimaryInboundPort();

  listEl.innerHTML = nodes.map(n => {
    const q = n.quality || {};
    let qBadgeClass = 'dc';
    if(q.type === 'residential') qBadgeClass = 'res';
    else if(q.type === 'mobile') qBadgeClass = 'mob';

    const ispText = esc(q.isp || q.org || n.hostname);
    const asnText = q.asn ? ' · ' + esc(q.asn) : '';
    const cc = (n.country_code || 'KR').toUpperCase();

    const actBtns = n.running
      ? '<span class="tag-running"><span class="dot-indicator"></span> 运行中 (槽位 ' + n.slot + ')</span>'
        + '<button type="button" class="btn-mount-443" data-mount-host="' + esc(n.hostname) + '" data-mount-port="' + primPort + '" title="将已运行的出口挂载到 :' + primPort + ' 节点"><svg class="icon-xs" viewBox="0 0 24 24"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg><span>挂载到 ' + primPort + '</span></button>'
      : '<button type="button" class="btn-start-exit" data-start-host="' + esc(n.hostname) + '"><svg class="icon-xs" viewBox="0 0 24 24"><line x1="12" x2="12" y1="5" y2="19"/><line x1="5" x2="19" y1="12" y2="12"/></svg><span>开启出口</span></button>'
        + '<button type="button" class="btn-mount-443" data-mount-host="' + esc(n.hostname) + '" data-mount-port="' + primPort + '" title="开启此出口并立刻将 :' + primPort + ' 节点挂载至此"><svg class="icon-xs" viewBox="0 0 24 24"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg><span>挂载到 ' + primPort + '</span></button>';

    return '<div class="ncard">'
      + '<div class="ncard-flag"><span class="country-tag">' + cc + '</span></div>'
      + '<div class="ncard-info">'
      +   '<div class="ncard-title-row">'
      +     '<span class="ncard-host">' + esc(n.country || n.country_code) + ' · ' + esc(n.hostname) + '</span>'
      +     '<span class="ncard-ip mono">' + esc(n.ip) + '</span>'
      +     '<span class="qbadge ' + qBadgeClass + '">' + esc(q.type_label || '节点') + ' · ' + (q.score || 0) + '分</span>'
      +   '</div>'
      +   '<div class="ncard-meta-row">'
      +     '<span>' + ispText + asnText + '</span>'
      +     '<span>速度: ' + n.speed_mbps.toFixed(1) + ' Mbps</span>'
      +     '<span>延迟: ' + n.ping + ' ms</span>'
      +     '<span>' + n.sessions + ' 会话</span>'
      +   '</div>'
      + '</div>'
      + '<div class="ncard-acts">' + actBtns + '</div>'
      + '</div>';
  }).join('');
}

function flagEmoji(cc){
  if(!cc || cc.length !== 2) return '🌐';
  const codePoints = cc.toUpperCase().split('').map(c => 127397 + c.charCodeAt(0));
  return String.fromCodePoint(...codePoints);
}

document.addEventListener('click', e => {
  const c = e.target.closest('[data-copy]');
  if(c) copy(c.dataset.copy);
});

// ---- SOCKS5 凭据 ----
let curCred = null;

function socksURL(host, port, user, pass){
  if(!user) return 'socks5://' + host + ':' + port;
  return 'socks5://' + user + ':' + pass + '@' + host + ':' + port;
}

// SOCKS5 端口监听在母机（跑 fanout 的这台服务器）上，客户端要连的是母机的
// 公网 IPv4，流量再从出口 IP 出去。出口 IP 是"出去以后"的地址，不能当连接地址。
// public_ip 是后端探测到的母机公网地址；探测不到才退回访问面板用的主机名。
function credHost(e){
  return view.public_ip || location.hostname || e.host;
}

function openCred(slot){
  const e = view.exits.find(x => x.slot === slot);
  if(!e){ toast('这个出口不在了', true); return; }
  curCred = {slot: slot, port: e.port, host: credHost(e), exit_ip: e.exit_ip || e.host};
  $('#crtitle').textContent = (e.region || '—') + ' · :' + e.port;
  $('#cruser').value = e.socks_user || '';
  $('#crpass').value = e.socks_pass || '';
  $('#crhostport').textContent = curCred.host + ':' + curCred.port;
  $('#crexitip').textContent = (e.exit_ip || '—') + (e.country ? ' (' + e.country + ')' : '');
  refreshCredURL();
  openModal('credbox');
}

function refreshCredURL(){
  if(!curCred) return;
  $('#crurl').textContent = socksURL(curCred.host, curCred.port,
    $('#cruser').value.trim(), $('#crpass').value.trim());
}
$('#cruser').oninput = refreshCredURL;
$('#crpass').oninput = refreshCredURL;

$('#crrand').onclick = () => {
  // 客户端和服务端都要能识别，只用无歧义、无需转义的字符
  const abc = 'abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789';
  const gen = n => Array.from(crypto.getRandomValues(new Uint8Array(n)))
    .map(v => abc[v % abc.length]).join('');
  $('#cruser').value = 'fo' + gen(6);
  $('#crpass').value = gen(14);
  refreshCredURL();
};

$('#crcopy').onclick = () => { copy($('#crurl').textContent); };
const cpHostBtn = $('#crcopyhost');
if(cpHostBtn) cpHostBtn.onclick = () => { if(curCred) copy(curCred.host + ':' + curCred.port); };
const cpExitBtn = $('#crcopyip');
if(cpExitBtn) cpExitBtn.onclick = () => { if(curCred && curCred.exit_ip) copy(curCred.exit_ip); };

$('#crsave').onclick = async e => {
  if(!curCred) return;
  const btn = e.target; btn.disabled = true;
  const q = new URLSearchParams({
    slot: curCred.slot,
    user: $('#cruser').value.trim(),
    pass: $('#crpass').value.trim(),
  });
  try{
    const r = await api('/api/cred?' + q, {method:'POST'});
    $('#cruser').value = r.user;
    $('#crpass').value = r.pass;
    refreshCredURL();
    toast('已保存，立即生效');
    poll();
  }catch(err){ toast(err.message, true); }
  btn.disabled = false;
};

// ---- 导出与客户端订阅 ----
let exExportData = null;
let exActiveFilter = 'all';
let exActiveTab = 'cards';

function getFilteredExportData(){
  if(!exExportData || !exExportData.items) return {items: [], links: []};
  let items = exExportData.items;
  if(exActiveFilter === 'direct'){
    items = items.filter(it => it.is_direct);
  } else if(exActiveFilter === 'exit'){
    items = items.filter(it => !it.is_direct);
  }
  let links = [];
  items.forEach(it => {
    if(it.links && it.links.length) {
      links.push(...it.links);
    }
  });
  return {items, links};
}

function updateExportViews(){
  const {items, links} = getFilteredExportData();
  const countEl = $('#excount');
  const boxEl = $('#exbox');
  const wrap = $('#exCardsWrap');

  if(boxEl) boxEl.value = links.join('\n');
  if(countEl){
    let filterLabel = '全部';
    if(exActiveFilter === 'direct') filterLabel = '仅直连';
    else if(exActiveFilter === 'exit') filterLabel = '仅出口';
    countEl.textContent = links.length + ' 条 (' + filterLabel + ')';
  }
  if(!wrap) return;
  if(!items.length){
    wrap.innerHTML = '<div class="empty">该分类下没有可用节点</div>';
    return;
  }
  wrap.innerHTML = items.map(it => {
    const isDir = it.is_direct;
    const badge = isDir
      ? '<span class="ex-pill direct">原生直连</span>'
      : '<span class="ex-pill exit">出口 ' + esc(it.exit_region || '—') + (it.exit_ip ? ' · ' + esc(it.exit_ip) : '') + '</span>';
    const linksList = it.links || [];
    let linksHtml = '';
    if(!linksList.length){
      linksHtml = '<div class="ex-card-link-preview">(暂无节点链接)</div>';
    } else if(linksList.length === 1){
      const link = linksList[0];
      linksHtml = '<div style="display:flex;align-items:center;gap:8px">'
        + '<div class="ex-card-link-preview" title="' + esc(link) + '">' + esc(link) + '</div>'
        + '<button type="button" class="btn-copy-one" data-copy="' + esc(link) + '">复制</button>'
        + '</div>';
    } else {
      const copyAllCardLinks = linksList.join('\n');
      linksHtml = '<div class="ex-card-sublinks">'
        + '<div style="display:flex;align-items:center;justify-content:space-between;gap:8px;margin-bottom:6px;flex-wrap:wrap">'
        +   '<span style="font-size:11px;color:var(--text-muted)">共 ' + linksList.length + ' 个客户端凭据:</span>'
        +   '<button type="button" class="btn-copy-one" data-copy="' + esc(copyAllCardLinks) + '">复制全部 ' + linksList.length + ' 条链接</button>'
        + '</div>'
        + linksList.map((lnk, idx) => {
            const client = (it.clients && it.clients[idx]) ? it.clients[idx] : null;
            const label = client ? (client.email || ('凭据 #' + (idx + 1))) : ('凭据 #' + (idx + 1));
            return '<div class="ex-sublink-row">'
              + '<span class="ex-sublink-badge">' + esc(label) + '</span>'
              + '<span class="ex-card-link-preview" title="' + esc(lnk) + '">' + esc(lnk) + '</span>'
              + '<button type="button" class="btn-mini" data-copy="' + esc(lnk) + '">复制</button>'
              + '</div>';
          }).join('')
        + '</div>';
    }
    return '<div class="ex-card">'
      + '<div class="ex-card-top">'
      +   badge
      +   '<div class="ex-card-title">' + esc(it.remark || it.protocol) + ' · ' + esc(it.protocol) + ' :' + it.port + '</div>'
      + '</div>'
      + '<div class="ex-card-info">' + linksHtml + '</div>'
      + '</div>';
  }).join('');
}

function switchExTab(tab){
  exActiveTab = tab;
  const tabCards = $('#exTabCards');
  const tabText = $('#exTabText');
  const cardsWrap = $('#exCardsWrap');
  const textWrap = $('#exTextWrap');
  const filters = $('#exFilters');
  if(tab === 'cards'){
    if(tabCards) tabCards.classList.add('active');
    if(tabText) tabText.classList.remove('active');
    if(cardsWrap) cardsWrap.style.display = 'flex';
    if(textWrap) textWrap.style.display = 'none';
    if(filters) filters.style.display = 'inline-flex';
  } else {
    if(tabCards) tabCards.classList.remove('active');
    if(tabText) tabText.classList.add('active');
    if(cardsWrap) cardsWrap.style.display = 'none';
    if(textWrap) textWrap.style.display = 'block';
    if(filters) filters.style.display = 'none';
  }
}

const tabCardsBtn = $('#exTabCards');
if(tabCardsBtn) tabCardsBtn.onclick = () => switchExTab('cards');
const tabTextBtn = $('#exTabText');
if(tabTextBtn) tabTextBtn.onclick = () => switchExTab('text');

document.querySelectorAll('[data-exfilter]').forEach(btn => {
  btn.onclick = () => {
    document.querySelectorAll('[data-exfilter]').forEach(b => b.classList.remove('active'));
    btn.classList.add('active');
    exActiveFilter = btn.dataset.exfilter;
    updateExportViews();
  };
});

$('#exportAll').onclick = async () => {
  const ids = [];
  (view.direct || []).forEach(i => ids.push(i.id));
  (view.exits || []).forEach(x => (x.inbounds || []).forEach(i => ids.push(i.id)));
  if(!ids.length){ toast('还没有节点可导出', true); return; }

  let bp = (location.pathname || '').replace(/\/+$/, '');
  let domain = (curSettings && curSettings.domain) || (view && view.domain) || '';
  let subUrl = '';
  if (domain) {
    let mode = (curSettings && curSettings.ssl_mode) || (view && view.ssl_mode) || '';
    if (mode === 'caddy') {
      subUrl = 'https://' + domain + bp + '/sub';
    } else {
      let isTLS = (curSettings && (curSettings.is_tls || curSettings.ssl_mode === 'custom')) ||
        (view && view.is_tls) || location.protocol === 'https:';
      let scheme = isTLS ? 'https://' : 'http://';
      let port = (curSettings && curSettings.port) || (location.port ? parseInt(location.port, 10) : (isTLS ? 443 : 80));
      let portStr = '';
      if (isTLS) {
        if (port && port !== 443) portStr = ':' + port;
      } else {
        if (port && port !== 80) portStr = ':' + port;
      }
      subUrl = scheme + domain + portStr + bp + '/sub';
    }
  } else {
    subUrl = location.origin + bp + '/sub';
  }
  if(view && view.sub_token){
    subUrl += '?token=' + encodeURIComponent(view.sub_token);
  }
  const subEl = $('#subUrlInput');
  if(subEl) subEl.value = subUrl;

  $('#exbox').value = '读取中…';
  $('#excount').textContent = '';
  const cardsWrap = $('#exCardsWrap');
  if(cardsWrap) cardsWrap.innerHTML = '<div class="empty">正在加载节点链接明细…</div>';
  openModal('export');
  switchExTab('cards');

  try{
    const d = await api('/api/xui/links?ids=' + ids.join(','));
    exExportData = d;
    updateExportViews();
  }catch(err){
    $('#exbox').value = '导出失败: ' + err.message;
    if(cardsWrap) cardsWrap.innerHTML = '<div class="empty">导出失败: ' + esc(err.message) + '</div>';
  }
};
$('#copyall').onclick = () => {
  const v = $('#exbox').value;
  if(v){
    copy(v);
    let label = '全部';
    if(exActiveFilter === 'direct') label = '直连';
    else if(exActiveFilter === 'exit') label = '出口';
    toast('已复制 ' + label + ' 节点链接');
  }
};
const copySubBtn = $('#copySubUrlBtn');
if(copySubBtn) copySubBtn.onclick = () => {
  const v = $('#subUrlInput').value;
  if(v){ copy(v); toast('已复制客户端 Base64 订阅链接'); }
};

// ---- 设置：改密码 / 改路径 / 改端口 / 改本地监听 ----
let curSettings = null;
let curBackend = null;

// 后端切换：把本机能用的模式列出来，装了的可选，没装的置灰并说明原因
async function loadBackendModes(){
  const sel = $('#setBackend');
  const hint = $('#setBackendHint');
  try{
    const m = await api('/api/panel/mode');
    curBackend = m.mode || '';
    sel.innerHTML = '<option value="">自动（按本机装了什么挑）</option>'
      + (m.modes || []).map(x =>
          '<option value="' + esc(x.mode) + '"' + (x.available ? '' : ' disabled')
          + '>' + esc(x.label) + (x.available ? '' : '（没装）') + '</option>').join('');
    sel.value = curBackend;
    const bad = (m.modes || []).filter(x => !x.available);
    hint.textContent = m.describe
      ? ('当前：' + m.describe + (bad.length ? '。灰掉的是本机没装的。' : ''))
      : '节点从哪来。装了 3x-ui 或 xray-cf-lite 就能直接接管，都没有就用自建。';
  }catch(err){
    sel.innerHTML = '<option value="">读取失败</option>';
    hint.textContent = err.message;
  }
}

$('#settingsBtn').onclick = async () => {
  $('#setPw').value = '';
  $('#setPath').value = '';
  $('#setPathHint').textContent = '读取中…';
  openModal('settings');
  loadBackendModes();
  try{
    const s = await api('/api/settings');
    curSettings = s;
    $('#setPath').value = (s.base_path || '').replace(/^\//, '');
    $('#setPort').value = s.port || '';
    $('#setListen').value = s.listen_addr || '0.0.0.0';
    $('#setDomain').value = s.domain || '';
    if (s.ssl_mode === 'caddy') {
      $('#setSSLMode').value = 'caddy';
    } else if (s.ssl_mode === 'custom' || (s.cert_file && s.key_file && s.ssl_mode !== 'none')) {
      $('#setSSLMode').value = 'custom';
    } else {
      $('#setSSLMode').value = 'none';
    }
    $('#setCertFile').value = s.cert_file || '';
    $('#setKeyFile').value = s.key_file || '';
    $('#setCertStatus').textContent = '';
    toggleSSLCertWrap();
    $('#setPathHint').textContent = '界面挂在这个路径下，扫端口的探不到。只能用字母数字和 - _。';
    $('#updCur').textContent = s.version || '-';
    $('#updLatest').textContent = '';
    $('#updNotes').hidden = true;
    $('#updApply').hidden = true;
    $('#updCheck').disabled = false;
    $('#updCheck').textContent = '检查更新';
  }catch(err){ $('#setPathHint').textContent = '读取失败: ' + err.message; }
};

function toggleSSLCertWrap(){
  const isCustom = $('#setSSLMode').value === 'custom';
  $('#setSSLCertWrap').style.display = isCustom ? 'block' : 'none';
}
$('#setSSLMode').onchange = toggleSSLCertWrap;

$('#setCertCheck').onclick = async e => {
  e.target.disabled = true;
  $('#setCertStatus').textContent = '检测中…';
  try{
    const cert = $('#setCertFile').value.trim();
    const key = $('#setKeyFile').value.trim();
    if(!cert || !key){
      $('#setCertStatus').innerHTML = '<span style="color:var(--bad)">请先填写证书与私钥路径</span>';
      e.target.disabled = false;
      return;
    }
    const r = await api('/api/ssl/inspect', {
      method: 'POST',
      headers: {'Content-Type':'application/json'},
      body: JSON.stringify({cert_file: cert, key_file: key})
    });
    const info = (r.issuer || r.subject_cn || '有效') + '，剩余 ' + r.days_left + ' 天';
    if(r.is_expired){
      $('#setCertStatus').innerHTML = '<span style="color:var(--bad)">⚠️ 证书已过期 (' + esc(info) + ')</span>';
    } else if(r.days_left <= 7){
      $('#setCertStatus').innerHTML = '<span style="color:var(--warn)">⚠️ 即将到期 (' + esc(info) + ')</span>';
    } else {
      $('#setCertStatus').innerHTML = '<span style="color:var(--ok)">✓ 有效 (' + esc(info) + ')</span>';
    }
  }catch(err){
    $('#setCertStatus').innerHTML = '<span style="color:var(--bad)">✗ ' + esc(err.message) + '</span>';
  }
  e.target.disabled = false;
};

// 检查更新：问后端 GitHub 最新版，有新版就亮出更新按钮和更新内容
$('#updCheck').onclick = async e => {
  e.target.disabled = true;
  e.target.textContent = '检查中…';
  try{
    const u = await api('/api/update/check');
    $('#updCur').textContent = u.current || '-';
    if(u.has_update){
      $('#updLatest').textContent = '有新版本 ' + u.latest;
      $('#updApplyVer').textContent = u.latest;
      $('#updApply').hidden = false;
      $('#updNotes').textContent = u.notes || '（这个版本没写更新说明）';
      $('#updNotes').hidden = false;
    } else {
      $('#updLatest').textContent = '已是最新';
      $('#updApply').hidden = true;
      $('#updNotes').hidden = true;
    }
  }catch(err){ toast(err.message, true); }
  e.target.disabled = false;
  e.target.textContent = '检查更新';
};

// 一键更新：后端下载替换二进制并重启服务，进程重启期间界面会短暂断连
$('#updApply').onclick = async e => {
  const ver = $('#updApplyVer').textContent;
  if(!(await showConfirm('更新到 ' + ver + '？服务会重启，界面会短暂断开。', {title:'确认系统更新', okText:'立即更新'}))) return;
  e.target.disabled = true;
  e.target.textContent = '更新中…';
  try{
    const r = await api('/api/update/apply', {method:'POST'});
    if(r.restarting){
      $('#updNotes').textContent = '已下载新版本，服务正在重启，几秒后刷新页面即可。';
      $('#updNotes').hidden = false;
      toast('更新中，服务重启后刷新页面');
      // 给服务重启留点时间再自动刷新
      setTimeout(() => location.reload(), 6000);
    } else {
      toast(r.message || '已是最新版');
      e.target.disabled = false;
      e.target.textContent = '更新到 ' + $('#updApplyVer').textContent;
    }
  }catch(err){
    toast(err.message, true);
    e.target.disabled = false;
    e.target.textContent = '更新到 ' + $('#updApplyVer').textContent;
  }
};

// 端口/监听地址/协议/域名变了要提示用户之后从新地址进；密码/路径可原地生效
function nextURL(port, listen, path, domain, isTLS, sslMode){
  if(sslMode === 'caddy' && domain){
    return 'https://' + domain + (path ? '/' + path : '') + '/';
  }
  const scheme = isTLS ? 'https:' : (location.protocol || 'http:');
  const host = domain || ((listen && listen !== '0.0.0.0') ? listen : location.hostname);
  let portPart = '';
  if(isTLS){
    if(port !== 443) portPart = ':' + port;
  } else {
    if(port !== 80) portPart = ':' + port;
  }
  return scheme + '//' + host + portPart + (path ? '/' + path : '') + '/';
}

$('#setSave').onclick = async e => {
  e.target.disabled = true;
  const body = {};
  const pw = $('#setPw').value.trim();
  if(pw) body.password = pw;
  body.base_path = $('#setPath').value.trim();
  const port = parseInt($('#setPort').value.trim(), 10);
  if(port) body.port = port;
  body.listen_addr = $('#setListen').value;
  const domain = $('#setDomain').value.trim();
  body.domain = domain;
  const sslMode = $('#setSSLMode').value;
  body.ssl_mode = sslMode;
  if(sslMode === 'custom'){
    body.cert_file = $('#setCertFile').value.trim();
    body.key_file = $('#setKeyFile').value.trim();
  } else {
    body.cert_file = '';
    body.key_file = '';
  }

  const willBeTLS = (sslMode === 'custom' && body.cert_file && body.key_file);
  const wasTLS = curSettings && (curSettings.is_tls || curSettings.ssl_mode === 'custom');
  const tlsChanged = Boolean(willBeTLS) !== Boolean(wasTLS);
  const domainChanged = curSettings && (domain !== (curSettings.domain || ''));
  const modeChanged = curSettings && (sslMode !== (curSettings.ssl_mode || 'none'));

  const portChanged = curSettings && (port !== curSettings.port
    || body.listen_addr !== (curSettings.listen_addr || '0.0.0.0'));

  try{
    // 后端和其它设置分属两个接口，先切后端：切失败就别继续，免得用户以为整单都生效了
    const backend = $('#setBackend').value;
    if(curBackend !== null && backend !== curBackend){
      const r = await api('/api/panel/mode', {
        method:'POST',
        headers:{'Content-Type':'application/json'},
        body: JSON.stringify({mode: backend}),
      });
      curBackend = r.mode || '';
      $('#setBackendHint').textContent = '当前：' + (r.describe || r.kind || '已切换');
    }
    await api('/api/settings', {
      method:'POST',
      headers:{'Content-Type':'application/json'},
      body: JSON.stringify(body),
    });
    if(portChanged || tlsChanged || domainChanged || modeChanged){
      const url = nextURL(port, body.listen_addr, body.base_path, domain, willBeTLS, sslMode);
      $('#setPortHint').innerHTML = '监听或地址已切换，请从新地址打开：<a href="' + esc(url) + '" target="_blank">' + esc(url) + '</a>';
      toast('监听或地址已切换，用新地址重新打开');
      // 端口/协议变了当前连接可能断开，不自动跳转，让用户看清新地址
    } else {
      toast('已保存');
      // 路径可能变了，重新加载到新路径下
      const np = body.base_path;
      const cur = (curSettings && curSettings.base_path || '').replace(/^\//, '');
      if(np !== cur){
        location.href = location.protocol + '//' + location.host
          + (np ? '/' + np : '') + '/';
        return;
      }
      closeModal('settings');
      poll();
    }
  }catch(err){ toast(err.message, true); }
  e.target.disabled = false;
};

async function doLogout(){
  if(!(await showConfirm('确定要退出当前管理控制台？', {title:'退出登录', okText:'退出登录'}))) return;
  try{ await api('/api/logout', {method:'POST'}); }catch(e){}
  location.href = 'login';
}
const loBtn = $('#logoutBtn');
if(loBtn) loBtn.onclick = doLogout;
const setLoBtn = $('#settingsLogoutBtn');
if(setLoBtn) setLoBtn.onclick = doLogout;

const nexKwInput = $('#nex-kw');
const nexKwClear = $('#nex-kw-clear');
if(nexKwInput && nexKwClear){
  nexKwInput.addEventListener('input', () => {
    nexKwClear.style.display = nexKwInput.value ? 'inline-block' : 'none';
  });
  nexKwClear.onclick = () => {
    nexKwInput.value = '';
    nexKwClear.style.display = 'none';
    loadNodeExplorer();
  };
}

poll();
setInterval(poll, 3000);
api('/api/nodes?sort=quality').then(r => {
  if(r && r.nodes && $('#navNodeCount')) $('#navNodeCount').textContent = '(' + r.nodes.length + ')';
}).catch(()=>{});
</script>
</body>
</html>`
