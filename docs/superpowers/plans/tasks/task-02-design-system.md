# Task 02 — Design System: Liquid Glass CSS + layout/base.templ + layout/admin.templ

**Depends on:** Task 00
**Estimated complexity:** media-alta — define toda la identidad visual del proyecto

---

## Estado del repositorio al inicio de este task

```
Module: cms-plazareal
templ + templUI: instalados (Task 00)
internal/view/layout/: directorio vacío
web/static/css/: no existe (o solo favicon/logo)
```

---

## Objetivo

Crear el design system completo de Plaza Real con estética **Liquid Glass** (Apple WWDC 2025): superficies traslúcidas, blur, fondos neutros oscuros, refracción de luz. Paleta fija: rojo `#d71055`, azul `#06a0e0`, lima `#acc60d`. Fonts: Montserrat 900italic (headings) + Geist (body). Material Symbols Outlined (icons).

Dos layouts templ compartidos:
- `internal/view/layout/base.templ` — páginas públicas
- `internal/view/layout/admin.templ` — admin shell con sidebar Liquid Glass

---

## Archivos a crear

| Acción | Archivo |
|--------|---------|
| Crear | `web/static/css/app.css` — public pages CSS |
| Crear | `web/static/css/admin.css` — admin CSS |
| Crear | `internal/view/layout/base.templ` |
| Crear | `internal/view/layout/admin.templ` |

---

## Implementación

- [ ] **Step 1: Crear web/static/css/app.css**

```bash
mkdir -p web/static/css
```

Contenido completo de `web/static/css/app.css`:

```css
/* ═══════════════════════════════════════
   PLAZA REAL — PUBLIC CSS
   Design: Liquid Glass + Ultra Modern
   ═══════════════════════════════════════ */

/* ── TOKENS ── */
:root {
  /* Brand */
  --red: #d71055;
  --red-d: #a80940;
  --red-glass: rgba(215, 16, 85, 0.15);
  --blue: #06a0e0;
  --blue-glass: rgba(6, 160, 224, 0.15);
  --lime: #acc60d;
  --lime-glass: rgba(172, 198, 13, 0.12);

  /* Neutrals */
  --ink: #0a0a0a;
  --ink-2: #141414;
  --ink-3: #1e1e1e;
  --surface: #f6f5f2;
  --surface-2: #ebebea;
  --white: #ffffff;
  --muted: #6b6b6b;
  --border: #e0dfdc;
  --border-dark: rgba(255,255,255,0.10);

  /* Glass */
  --glass-bg: rgba(255,255,255,0.65);
  --glass-bg-dark: rgba(20,20,20,0.55);
  --glass-border: rgba(255,255,255,0.35);
  --glass-border-dark: rgba(255,255,255,0.12);
  --glass-blur: blur(24px) saturate(180%);
  --glass-shadow: 0 8px 40px rgba(0,0,0,0.12), 0 0 0 1px rgba(255,255,255,0.25);
  --glass-shadow-dark: 0 8px 40px rgba(0,0,0,0.45), 0 0 0 1px rgba(255,255,255,0.08);

  /* Typography */
  --font-display: 'Montserrat', sans-serif;
  --font-body: 'Geist', system-ui, sans-serif;

  /* Radius */
  --r-sm: 8px;
  --r-md: 14px;
  --r-lg: 20px;
  --r-xl: 28px;
  --r-full: 9999px;
}

/* ── RESET ── */
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
html { scroll-behavior: smooth; -webkit-font-smoothing: antialiased; }
body {
  font-family: var(--font-body);
  background: var(--surface);
  color: var(--ink);
  overflow-x: hidden;
}
img { max-width: 100%; display: block; }
a { text-decoration: none; color: inherit; }

/* ── TYPOGRAPHY ── */
.display { font-family: var(--font-display); font-weight: 900; letter-spacing: -2px; }
.display em { font-style: italic; color: var(--blue); }
h1,h2,h3,h4 { font-family: var(--font-display); }

/* ── GLASS CARD ── */
.glass-card {
  background: var(--glass-bg);
  backdrop-filter: var(--glass-blur);
  -webkit-backdrop-filter: var(--glass-blur);
  border: 1px solid var(--glass-border);
  box-shadow: var(--glass-shadow);
  border-radius: var(--r-lg);
}
.glass-card-dark {
  background: var(--glass-bg-dark);
  backdrop-filter: var(--glass-blur);
  -webkit-backdrop-filter: var(--glass-blur);
  border: 1px solid var(--glass-border-dark);
  box-shadow: var(--glass-shadow-dark);
  border-radius: var(--r-lg);
}

/* ── NAVBAR ── */
.navbar {
  position: fixed; top: 0; left: 0; right: 0; z-index: 200;
  height: 68px;
  background: rgba(10,10,10,0.82);
  backdrop-filter: blur(28px) saturate(200%);
  -webkit-backdrop-filter: blur(28px) saturate(200%);
  border-bottom: 1px solid var(--border-dark);
  padding: 0 40px;
  display: flex; align-items: center; justify-content: space-between;
}
.navbar-logo img { height: 40px; filter: brightness(0) invert(1); }
.navbar-links { display: flex; gap: 32px; align-items: center; }
.navbar-links a {
  color: rgba(255,255,255,.58);
  font-size: .86rem; font-weight: 500;
  transition: color .15s;
}
.navbar-links a:hover, .navbar-links a.active { color: #fff; }
.navbar-cta {
  background: var(--red) !important;
  color: #fff !important;
  padding: 9px 22px;
  border-radius: var(--r-full);
  font-weight: 700 !important;
  font-size: .83rem !important;
  transition: opacity .15s;
}
.navbar-cta:hover { opacity: .88; }
.navbar-burger { display: none; background: none; border: none; cursor: pointer; flex-direction: column; gap: 5px; padding: 4px; }
.navbar-burger span { display: block; width: 22px; height: 2px; background: #fff; border-radius: 2px; }
.mob-menu {
  display: none; position: fixed; top: 68px; left: 0; right: 0;
  background: rgba(10,10,10,.96);
  backdrop-filter: blur(24px);
  z-index: 199;
  padding: 16px 24px 28px;
  flex-direction: column; gap: 2px;
  border-bottom: 1px solid var(--border-dark);
}
.mob-menu a { color: rgba(255,255,255,.75); font-size: .97rem; font-weight: 500; padding: 12px 0; border-bottom: 1px solid rgba(255,255,255,.06); display: block; }
.mob-menu.open { display: flex; }

/* ── PAGE HERO ── */
.page-hero {
  background: linear-gradient(rgba(10,10,10,.72), rgba(10,10,10,.86)),
              url('https://plazareal.cl/wp-content/uploads/2025/12/plaza-real-hall-pantallas.jpeg');
  background-size: cover; background-position: center; background-attachment: fixed;
  padding: 130px 24px 72px;
  position: relative; overflow: hidden;
}
.page-hero::before {
  content: ''; position: absolute;
  width: 600px; height: 600px;
  background: radial-gradient(circle, var(--red) 0%, transparent 70%);
  top: -300px; right: -200px; opacity: .12; pointer-events: none;
}
.page-hero::after {
  content: ''; position: absolute;
  width: 400px; height: 400px;
  background: radial-gradient(circle, var(--blue) 0%, transparent 70%);
  bottom: -200px; left: -100px; opacity: .10; pointer-events: none;
}
.page-hero-inner { max-width: 1200px; margin: 0 auto; position: relative; z-index: 1; }
.page-hero-eyebrow {
  font-size: .68rem; font-weight: 700; letter-spacing: 3px; text-transform: uppercase;
  color: var(--red); margin-bottom: 10px; display: block;
}
.page-hero-title {
  font-family: var(--font-display);
  font-size: clamp(2.2rem, 5.5vw, 3.8rem);
  font-weight: 900; color: #fff;
  letter-spacing: -2px; line-height: 1.02;
  margin-bottom: 14px;
}
.page-hero-title em { font-style: italic; color: var(--blue); }
.page-hero-sub { color: rgba(255,255,255,.62); font-size: .97rem; max-width: 600px; line-height: 1.7; font-weight: 300; }

/* ── SECTION ── */
.container { max-width: 1200px; margin: 0 auto; padding: 0 24px; }
.section { padding: 64px 0 96px; }

/* ── BUTTONS ── */
.btn-primary {
  display: inline-flex; align-items: center; gap: 8px;
  background: var(--red); color: #fff;
  border: none; padding: 12px 28px;
  border-radius: var(--r-full);
  font-family: var(--font-body); font-size: .9rem; font-weight: 700;
  cursor: pointer; transition: opacity .15s;
}
.btn-primary:hover { opacity: .86; }
.btn-outline {
  display: inline-flex; align-items: center; gap: 8px;
  background: transparent; color: var(--muted);
  border: 1.5px solid var(--border);
  padding: 11px 24px; border-radius: var(--r-full);
  font-family: var(--font-body); font-size: .9rem; font-weight: 600;
  cursor: pointer; transition: all .15s;
}
.btn-outline:hover { border-color: var(--muted); color: var(--ink); }

/* ── CHIP / TAG ── */
.chip {
  display: inline-flex; align-items: center; gap: 6px;
  font-family: var(--font-body); font-size: .74rem; font-weight: 700;
  padding: 7px 16px; border-radius: var(--r-full);
  border: 1.5px solid var(--border); background: var(--white);
  color: var(--muted); cursor: pointer; transition: all .15s;
}
.chip:hover, .chip.active { border-color: var(--red); color: var(--red); background: var(--red-glass); }

/* ── FOOTER ── */
footer {
  background: var(--ink); color: rgba(255,255,255,.42);
  padding: 40px 24px; text-align: center;
  font-size: .8rem; line-height: 1.8;
}
footer a { color: rgba(255,255,255,.55); }
footer a:hover { color: #fff; }

/* ── ANIMATIONS ── */
@keyframes spin { to { transform: rotate(360deg); } }
@keyframes fadeUp { from { opacity: 0; transform: translateY(16px); } to { opacity: 1; transform: none; } }
.animate-fadeup { animation: fadeUp .4s ease forwards; }

/* ── LOADING ── */
.loading-spin { animation: spin 1s linear infinite; }
.htmx-indicator { opacity: 0; transition: opacity .2s; }
.htmx-request .htmx-indicator { opacity: 1; }

/* ── RESPONSIVE ── */
@media (max-width: 768px) {
  .navbar-links { display: none; }
  .navbar-burger { display: flex; }
}
```

- [ ] **Step 2: Crear web/static/css/admin.css**

Contenido completo de `web/static/css/admin.css`:

```css
/* ═══════════════════════════════════════
   PLAZA REAL — ADMIN CSS
   Design: Liquid Glass Dark + Material
   ═══════════════════════════════════════ */

:root {
  /* Brand */
  --red: #d71055;
  --red-d: #a80940;
  --blue: #06a0e0;
  --lime: #acc60d;

  /* Admin surface tokens */
  --bg: #0d0d0d;
  --surface: #161616;
  --surface-2: #1f1f1f;
  --surface-3: #272727;
  --border: rgba(255,255,255,0.08);
  --border-bright: rgba(255,255,255,0.16);

  /* Glass */
  --glass-sidebar: rgba(18,18,18,0.80);
  --glass-blur: blur(24px) saturate(160%);
  --glass-card: rgba(255,255,255,0.035);

  /* Text */
  --text: #f0f0f0;
  --text-muted: rgba(255,255,255,0.42);
  --text-mid: rgba(255,255,255,0.68);

  /* Status */
  --success: #2db37a;
  --success-bg: rgba(45,179,122,0.12);
  --warning: #f59e0b;
  --warning-bg: rgba(245,158,11,0.12);
  --danger: #ef4444;
  --danger-bg: rgba(239,68,68,0.12);
  --info: var(--blue);
  --info-bg: rgba(6,160,224,0.12);

  /* Typography */
  --font-display: 'Montserrat', sans-serif;
  --font-body: 'Geist', system-ui, sans-serif;

  /* Sidebar */
  --sidebar-w: 240px;

  /* Radius */
  --r-sm: 6px;
  --r-md: 10px;
  --r-lg: 16px;
  --r-xl: 20px;
  --r-full: 9999px;
}

/* ── RESET ── */
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
html { -webkit-font-smoothing: antialiased; }
body.admin-body {
  font-family: var(--font-body);
  background: var(--bg);
  color: var(--text);
  display: flex;
  min-height: 100vh;
}

/* ── SIDEBAR ── */
.sidebar {
  width: var(--sidebar-w);
  flex-shrink: 0;
  background: var(--glass-sidebar);
  backdrop-filter: var(--glass-blur);
  -webkit-backdrop-filter: var(--glass-blur);
  border-right: 1px solid var(--border);
  position: fixed; top: 0; left: 0; bottom: 0;
  display: flex; flex-direction: column;
  z-index: 100;
  padding: 24px 12px;
}
.sidebar-logo {
  padding: 4px 12px 24px;
  border-bottom: 1px solid var(--border);
  margin-bottom: 16px;
}
.sidebar-logo img { height: 34px; filter: brightness(0) invert(1); }
.sidebar-section-label {
  font-size: .62rem; font-weight: 700; letter-spacing: 2px;
  text-transform: uppercase; color: var(--text-muted);
  padding: 12px 12px 6px;
}
.sidebar-link {
  display: flex; align-items: center; gap: 10px;
  padding: 9px 12px; border-radius: var(--r-md);
  font-size: .84rem; font-weight: 500;
  color: var(--text-muted);
  transition: all .15s; cursor: pointer;
  text-decoration: none;
}
.sidebar-link:hover { background: var(--glass-card); color: var(--text); }
.sidebar-link.active {
  background: rgba(215, 16, 85, 0.15);
  color: var(--red);
}
.sidebar-link .material-symbols-outlined { font-size: 18px; }
.sidebar-footer {
  margin-top: auto;
  border-top: 1px solid var(--border);
  padding-top: 12px;
}

/* ── MAIN CONTENT ── */
.admin-main {
  margin-left: var(--sidebar-w);
  flex: 1;
  display: flex; flex-direction: column;
  min-height: 100vh;
}

/* ── TOPBAR ── */
.topbar {
  height: 60px;
  background: rgba(13,13,13,0.85);
  backdrop-filter: blur(20px);
  border-bottom: 1px solid var(--border);
  padding: 0 28px;
  display: flex; align-items: center; justify-content: space-between;
  position: sticky; top: 0; z-index: 50;
}
.topbar-title {
  font-family: var(--font-display);
  font-size: 1rem; font-weight: 800;
  color: var(--text);
}
.topbar-actions { display: flex; gap: 10px; align-items: center; }

/* ── CONTENT AREA ── */
.admin-content { padding: 28px; }

/* ── CARD ── */
.card {
  background: var(--glass-card);
  border: 1px solid var(--border);
  border-radius: var(--r-lg);
  overflow: hidden;
}
.card-header {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border);
  display: flex; align-items: center; justify-content: space-between;
}
.card-title { font-size: .9rem; font-weight: 700; color: var(--text); }

/* ── TABLE ── */
table { width: 100%; border-collapse: collapse; font-size: .83rem; }
thead th {
  padding: 10px 16px;
  text-align: left;
  font-size: .68rem; font-weight: 700; letter-spacing: 1px; text-transform: uppercase;
  color: var(--text-muted);
  border-bottom: 1px solid var(--border);
}
tbody td {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
  color: var(--text-mid);
  vertical-align: middle;
}
tbody tr:last-child td { border-bottom: none; }
tbody tr:hover td { background: rgba(255,255,255,.02); }
.empty-state-cell { text-align: center; padding: 48px 16px !important; color: var(--text-muted) !important; }

/* ── BADGES ── */
.badge {
  display: inline-flex; align-items: center;
  padding: 3px 10px; border-radius: var(--r-full);
  font-size: .68rem; font-weight: 700; letter-spacing: .5px;
}
.badge-success { background: var(--success-bg); color: var(--success); }
.badge-warning { background: var(--warning-bg); color: var(--warning); }
.badge-danger  { background: var(--danger-bg);  color: var(--danger);  }
.badge-info    { background: var(--info-bg);    color: var(--info);    }
.badge-neutral { background: rgba(255,255,255,.06); color: var(--text-muted); }
.badge-publicado  { background: var(--success-bg); color: var(--success); }
.badge-borrador   { background: var(--warning-bg); color: var(--warning); }
.badge-pendiente  { background: var(--warning-bg); color: var(--warning); }
.badge-confirmada { background: var(--success-bg); color: var(--success); }
.badge-cancelada  { background: var(--danger-bg);  color: var(--danger);  }
.badge-NOTICIA    { background: var(--info-bg); color: var(--info); }
.badge-COMUNICADO { background: rgba(172,198,13,0.12); color: var(--lime); }
.badge-PROMOCION  { background: rgba(215,16,85,0.12); color: var(--red); }
.badge-superadmin { background: rgba(215,16,85,0.15); color: var(--red); }
.badge-director   { background: rgba(6,160,224,0.15); color: var(--blue); }
.badge-admin      { background: rgba(45,179,122,0.12); color: var(--success); }
.badge-editor     { background: rgba(255,255,255,.06); color: var(--text-muted); }

/* ── BUTTONS (admin) ── */
.topbar-btn {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 7px 16px; border-radius: var(--r-md);
  font-family: var(--font-body); font-size: .8rem; font-weight: 600;
  cursor: pointer; border: 1px solid var(--border);
  background: var(--glass-card); color: var(--text-mid);
  transition: all .15s;
}
.topbar-btn:hover { background: var(--surface-2); color: var(--text); }
.topbar-btn-primary {
  background: var(--red); color: #fff; border-color: transparent;
}
.topbar-btn-primary:hover { opacity: .88; }
.topbar-btn-outline { background: transparent; }
.btn-icon {
  display: inline-flex; align-items: center; justify-content: center;
  width: 30px; height: 30px; border-radius: var(--r-sm);
  background: transparent; border: none; cursor: pointer;
  color: var(--text-muted); transition: all .15s;
}
.btn-icon:hover { background: var(--surface-3); color: var(--text); }
.btn-icon.btn-danger:hover { background: var(--danger-bg); color: var(--danger); }
.btn-icon .material-symbols-outlined { font-size: 16px; }

/* ── FORMS (admin modals) ── */
.form-group { margin-bottom: 16px; display: flex; flex-direction: column; gap: 6px; }
.form-group label { font-size: .76rem; font-weight: 600; color: var(--text-muted); }
.form-group input,
.form-group textarea,
.form-group select {
  padding: 9px 12px;
  background: var(--surface-2);
  border: 1px solid var(--border-bright);
  border-radius: var(--r-md);
  font-family: var(--font-body); font-size: .88rem;
  color: var(--text);
  outline: none; transition: border-color .15s;
  width: 100%;
}
.form-group input:focus,
.form-group textarea:focus,
.form-group select:focus { border-color: var(--red); }
.form-group textarea { resize: vertical; min-height: 90px; }
.form-group select option { background: var(--surface-2); }

/* ── TOAST ── */
.toast {
  padding: 12px 18px; border-radius: var(--r-md);
  font-size: .84rem; font-weight: 600;
  display: inline-flex; align-items: center; gap: 8px;
}
.toast-success { background: var(--success-bg); color: var(--success); }
.toast-error { background: var(--danger-bg); color: var(--danger); }
.toast-info { background: var(--info-bg); color: var(--info); }

/* ── MODAL ── */
dialog.admin-modal {
  background: var(--surface);
  border: 1px solid var(--border-bright);
  border-radius: var(--r-xl);
  color: var(--text);
  padding: 0;
  width: min(560px, 95vw);
  box-shadow: 0 32px 80px rgba(0,0,0,0.6);
}
dialog.admin-modal::backdrop { background: rgba(0,0,0,0.7); backdrop-filter: blur(4px); }
.modal-header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--border);
  display: flex; align-items: center; justify-content: space-between;
}
.modal-title { font-size: .95rem; font-weight: 800; }
.modal-body { padding: 20px 24px; }
.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid var(--border);
  display: flex; gap: 10px; justify-content: flex-end;
}

/* ── STAT CARDS ── */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 28px;
}
.stat-card {
  background: var(--glass-card);
  border: 1px solid var(--border);
  border-radius: var(--r-lg);
  padding: 20px 24px;
  display: flex; flex-direction: column; gap: 8px;
}
.stat-card-icon {
  width: 36px; height: 36px; border-radius: var(--r-md);
  display: flex; align-items: center; justify-content: center;
  font-size: 18px;
}
.stat-card-value { font-size: 1.8rem; font-weight: 900; font-family: var(--font-display); }
.stat-card-label { font-size: .75rem; color: var(--text-muted); font-weight: 500; }

/* ── ANIMATIONS ── */
@keyframes spin { to { transform: rotate(360deg); } }
.loading-spin { animation: spin 1s linear infinite; }
.htmx-indicator { opacity: 0; transition: opacity .2s; }
.htmx-request .htmx-indicator { opacity: 1; }

/* ── LOGIN PAGE ── */
.login-bg {
  min-height: 100vh;
  background: var(--bg);
  display: flex; align-items: center; justify-content: center;
  position: relative; overflow: hidden;
}
.login-bg::before {
  content: ''; position: absolute;
  width: 800px; height: 800px;
  background: radial-gradient(circle, var(--red) 0%, transparent 65%);
  top: -400px; left: -200px; opacity: .07; pointer-events: none;
}
.login-bg::after {
  content: ''; position: absolute;
  width: 600px; height: 600px;
  background: radial-gradient(circle, var(--blue) 0%, transparent 65%);
  bottom: -300px; right: -150px; opacity: .07; pointer-events: none;
}
.login-card {
  width: min(420px, 95vw);
  background: var(--surface);
  border: 1px solid var(--border-bright);
  border-radius: var(--r-xl);
  padding: 40px;
  box-shadow: 0 40px 120px rgba(0,0,0,0.5);
  position: relative; z-index: 1;
}
```

- [ ] **Step 3: Crear internal/view/layout/base.templ**

```bash
touch internal/view/layout/base.templ
```

Contenido completo:

```templ
package layout

templ Base(title string, activePage string) {
	<!DOCTYPE html>
	<html lang="es">
	<head>
		<meta charset="UTF-8"/>
		<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
		<title>{ title } — Plaza Real Copiapó</title>
		<meta name="description" content="Centro comercial Plaza Real en Copiapó. Tiendas, restaurantes, servicios y más."/>
		<link rel="icon" type="image/svg+xml" href="/static/favicon.svg"/>
		<link rel="preconnect" href="https://fonts.googleapis.com"/>
		<link href="https://fonts.googleapis.com/css2?family=Montserrat:ital,wght@0,400;0,700;0,800;0,900;1,900&family=Geist:wght@300;400;500;600;700&display=swap" rel="stylesheet"/>
		<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200"/>
		<link rel="stylesheet" href="/static/css/app.css"/>
		<script src="https://unpkg.com/htmx.org@1.9.12" defer></script>
		<script src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js" defer></script>
	</head>
	<body>
		@Navbar(activePage)
		{ children... }
		@Footer()
		<script>
			document.querySelector('.navbar-burger')?.addEventListener('click', function() {
				document.querySelector('.mob-menu')?.classList.toggle('open');
			});
		</script>
	</body>
	</html>
}

templ Navbar(activePage string) {
	<nav class="navbar">
		<a href="/index.html" class="navbar-logo">
			<img src="https://plazareal.cl/wp-content/uploads/2025/12/logo-mall-plaza-real-2024@3x.png" alt="Plaza Real"/>
		</a>
		<div class="navbar-links">
			<a href="/index.html" class={ "navbar-link", templ.KV("active", activePage == "inicio") }>Inicio</a>
			<a href="/buscador-tiendas" class={ "navbar-link", templ.KV("active", activePage == "tiendas") }>Tiendas</a>
			<a href="/locales" class={ "navbar-link", templ.KV("active", activePage == "locales") }>Locales</a>
			<a href="/promociones" class={ "navbar-link", templ.KV("active", activePage == "promociones") }>Eventos</a>
			<a href="/noticias" class={ "navbar-link", templ.KV("active", activePage == "noticias") }>Noticias</a>
			<a href="/admin" class="navbar-cta">Admin</a>
		</div>
		<button class="navbar-burger" aria-label="Menú">
			<span></span><span></span><span></span>
		</button>
	</nav>
	<div class="mob-menu">
		<a href="/index.html">Inicio</a>
		<a href="/buscador-tiendas">Tiendas</a>
		<a href="/locales">Locales Disponibles</a>
		<a href="/promociones">Eventos & Promociones</a>
		<a href="/noticias">Noticias</a>
		<a href="/admin">Admin</a>
	</div>
}

templ Footer() {
	<footer>
		<p>&copy; 2026 Plaza Real Copiapó · <a href="/admin">Admin</a></p>
	</footer>
}
```

- [ ] **Step 4: Crear internal/view/layout/admin.templ**

```templ
package layout

import "github.com/a-h/templ"

templ Admin(title string, activePage string, content templ.Component) {
	<!DOCTYPE html>
	<html lang="es">
	<head>
		<meta charset="UTF-8"/>
		<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
		<title>{ title } — Plaza Real Admin</title>
		<link rel="icon" type="image/svg+xml" href="/static/favicon.svg"/>
		<link rel="preconnect" href="https://fonts.googleapis.com"/>
		<link href="https://fonts.googleapis.com/css2?family=Montserrat:ital,wght@0,700;0,800;0,900&family=Geist:wght@300;400;500;600;700&display=swap" rel="stylesheet"/>
		<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200"/>
		<link rel="stylesheet" href="/static/css/admin.css"/>
		<script src="https://unpkg.com/htmx.org@1.9.12"></script>
		<script src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js" defer></script>
	</head>
	<body class="admin-body">
		@AdminSidebar(activePage)
		<div class="admin-main">
			<div class="topbar">
				<span class="topbar-title">{ title }</span>
				<div class="topbar-actions" id="topbar-actions"></div>
			</div>
			<div id="content" class="admin-content">
				@content
			</div>
		</div>
		<div id="modal-container"></div>
	</body>
	</html>
}

templ AdminSidebar(activePage string) {
	<aside class="sidebar">
		<div class="sidebar-logo">
			<img src="https://plazareal.cl/wp-content/uploads/2025/12/logo-mall-plaza-real-2024@3x.png" alt="Plaza Real"/>
		</div>
		<span class="sidebar-section-label">Contenido</span>
		<a href="/admin/tiendas" class={ "sidebar-link", templ.KV("active", activePage == "tiendas") }>
			<span class="material-symbols-outlined">storefront</span>Tiendas
		</a>
		<a href="/admin/locales" class={ "sidebar-link", templ.KV("active", activePage == "locales") }>
			<span class="material-symbols-outlined">real_estate_agent</span>Locales
		</a>
		<a href="/admin/content?cat=PROMOCION" class={ "sidebar-link", templ.KV("active", activePage == "promociones") }>
			<span class="material-symbols-outlined">celebration</span>Promociones
		</a>
		<a href="/admin/content?cat=NOTICIA" class={ "sidebar-link", templ.KV("active", activePage == "noticias") }>
			<span class="material-symbols-outlined">newspaper</span>Noticias
		</a>
		<a href="/admin/content?cat=COMUNICADO" class={ "sidebar-link", templ.KV("active", activePage == "comunicados") }>
			<span class="material-symbols-outlined">campaign</span>Comunicados
		</a>
		<a href="/admin/reservas" class={ "sidebar-link", templ.KV("active", activePage == "reservas") }>
			<span class="material-symbols-outlined">event_seat</span>Reservas
		</a>
		<span class="sidebar-section-label">Media & Dispositivos</span>
		<a href="/admin/multimedia" class={ "sidebar-link", templ.KV("active", activePage == "multimedia") }>
			<span class="material-symbols-outlined">perm_media</span>Multimedia
		</a>
		<a href="/admin/devices" class={ "sidebar-link", templ.KV("active", activePage == "devices") }>
			<span class="material-symbols-outlined">tv</span>Dispositivos
		</a>
		<span class="sidebar-section-label">Administración</span>
		<a href="/admin/users" class={ "sidebar-link", templ.KV("active", activePage == "users") }>
			<span class="material-symbols-outlined">manage_accounts</span>Usuarios
		</a>
		<a href="/admin/reports" class={ "sidebar-link", templ.KV("active", activePage == "reports") }>
			<span class="material-symbols-outlined">bar_chart</span>Informes
		</a>
		<div class="sidebar-footer">
			<form hx-post="/admin/logout" hx-swap="none">
				<button type="submit" class="sidebar-link" style="width:100%;border:none;cursor:pointer">
					<span class="material-symbols-outlined">logout</span>Cerrar sesión
				</button>
			</form>
		</div>
	</aside>
}
```

- [ ] **Step 5: Ejecutar templ generate**

```bash
make generate
# Equivale a: templ generate
```

Esperado: sin errores. Se generan `base_templ.go` y `admin_templ.go` en `internal/view/layout/`.

Verificar:
```bash
ls internal/view/layout/
# Esperado: admin.templ admin_templ.go base.templ base_templ.go
```

- [ ] **Step 6: Verificar compilación**

```bash
go build ./...
# Esperado: sin errores
```

- [ ] **Step 7: Commit**

```bash
git add web/static/css/ internal/view/ Makefile
git commit -m "feat: add Liquid Glass design system — app.css, admin.css, base.templ, admin.templ"
```

---

## Estado del repositorio al finalizar este task

- `web/static/css/app.css` con design system Liquid Glass público
- `web/static/css/admin.css` con design system admin dark + Liquid Glass
- `internal/view/layout/base.templ` compilado (existe `base_templ.go`)
- `internal/view/layout/admin.templ` compilado (existe `admin_templ.go`)
- Ambos layouts usan paleta Plaza Real: rojo `#d71055`, azul `#06a0e0`, lima `#acc60d`
- `go build ./...` pasa sin errores
