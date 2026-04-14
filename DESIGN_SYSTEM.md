# CSL Design System — MD3 Expressive 2026

## Seed Color
`#A51535` → Colegio San Lorenzo institutional red

---

## Color Tokens

### Primary Ramp
| Token | Value | Use |
|---|---|---|
| `--md-primary` | `#9B1230` | Main actions, nav CTA, active states |
| `--md-on-primary` | `#fff` | Text on primary |
| `--md-primary-container` | `#F5E0E4` | Tonal backgrounds, chips, badges |
| `--md-on-primary-container` | `#5C0A1E` | Text on primary container |
| `--md-primary-dark` | `#6E0C22` | Hover/pressed states |
| `--md-primary-mid` | `#BE3356` | Decorative accents |
| `--md-primary-faint` | `#FDF0F2` | Ultra-light selection bg |

### Secondary
| Token | Value |
|---|---|
| `--md-secondary` | `#9B9EA4` |
| `--md-secondary-container` | `#E8E8EC` |
| `--md-on-secondary-container` | `#3A3C42` |

### Tertiary
| Token | Value |
|---|---|
| `--md-tertiary` | `#8B6914` |
| `--md-tertiary-container` | `#FFF0C8` |
| `--md-on-tertiary-container` | `#2C1F00` |

### Surface System
| Token | Value | Use |
|---|---|---|
| `--md-surface` | `#FAF8F7` | Page background |
| `--md-surface-variant` | `#F2EFF0` | Borders, dividers |
| `--md-surface-dim` | `#EDE9EA` | Dashboard background |
| `--md-surface-container-low` | `#F7F4F5` | Subtle cards |
| `--md-surface-container` | `#F2EEEF` | Medium cards |
| `--md-surface-container-high` | `#ECE8E9` | Emphasized cards |
| `--md-surface-bright` | `#FFFFFF` | Card backgrounds |

### Text
| Token | Value |
|---|---|
| `--md-on-surface` | `#1C1B1A` |
| `--md-on-surface-variant` | `#4D4B4C` |
| `--md-outline` | `#7E7B7C` |
| `--md-outline-variant` | `#CEC9CA` |

### Inverse
| Token | Value |
|---|---|
| `--md-inverse-surface` | `#312F30` |
| `--md-inverse-on-surface` | `#F5F0F1` |

---

## Glassmorphism Tokens
| Token | Value |
|---|---|
| `--glass-bg` | `rgba(250,248,247,0.72)` |
| `--glass-blur` | `blur(20px) saturate(180%)` |
| `--glass-border` | `rgba(155,18,48,0.10)` |
| `--glass-shadow` | `0 8px 32px rgba(155,18,48,0.08)` |

**Usage:** Nav bar, floating cards, mobile menu overlay.

---

## Elevation (Box Shadows)
| Token | Value |
|---|---|
| `--elev-1` | `0 1px 2px rgba(28,27,26,.06), 0 1px 3px rgba(28,27,26,.04)` |
| `--elev-2` | `0 2px 6px rgba(28,27,26,.08), 0 1px 2px rgba(28,27,26,.04)` |
| `--elev-3` | `0 4px 12px rgba(28,27,26,.10), 0 2px 4px rgba(28,27,26,.06)` |
| `--elev-4` | `0 8px 24px rgba(28,27,26,.12), 0 2px 6px rgba(28,27,26,.06)` |
| `--elev-5` | `0 16px 40px rgba(28,27,26,.14), 0 4px 8px rgba(28,27,26,.06)` |

---

## Border Radius
| Token | Value | Use |
|---|---|---|
| `--r-xs` | `4px` | Small chips |
| `--r-sm` | `12px` | Cards, inputs |
| `--r-md` | `16px` | Medium cards |
| `--r-lg` | `24px` | Large cards, sections |
| `--r-xl` | `32px` | Hero cards |
| `--r-2xl` | `40px` | Feature cards |
| `--r-full` | `9999px` | Pills, buttons |

---

## Typography

### Fonts
- **Display:** `'DM Serif Display', Georgia, serif`
- **Body:** `'DM Sans', system-ui, sans-serif`

### Scale
| Class | Size | Weight | Line Height | Use |
|---|---|---|---|---|
| `.display-l` | `clamp(48px, 7vw, 88px)` | 400 | 1.03 | Hero headlines |
| `.display-m` | `clamp(36px, 5.5vw, 64px)` | 400 | 1.07 | Page titles |
| `.display-s` | `clamp(30px, 4vw, 52px)` | 400 | 1.1 | Sub-headlines |
| `.headline-l` | `clamp(26px, 3.5vw, 40px)` | 400 | 1.15 | Section titles |
| `.headline-m` | `clamp(20px, 2.8vw, 32px)` | 400 | 1.2 | Card titles |
| `.body-l` | `17px` | 400 | 1.75 | Body text |
| `.label-primary` | `11px` | 500 | — | Eyebrows, labels |

---

## Spacing Scale
| Token | Value |
|---|---|
| `--sp-4` | `4px` |
| `--sp-8` | `8px` |
| `--sp-12` | `12px` |
| `--sp-16` | `16px` |
| `--sp-20` | `20px` |
| `--sp-24` | `24px` |
| `--sp-32` | `32px` |
| `--sp-40` | `40px` |
| `--sp-48` | `48px` |
| `--sp-64` | `64px` |
| `--sp-80` | `80px` |
| `--sp-96` | `96px` |

---

## Animation Easings
| Token | Value | Use |
|---|---|---|
| `--ease-express` | `cubic-bezier(0.05, 0.7, 0.1, 1.0)` | Snappy interactions |
| `--ease-standard` | `cubic-bezier(0.2, 0.0, 0.0, 1.0)` | Smooth transitions |

---

## Layout Constants
| Token | Value |
|---|---|
| `--max-w` | `1200px` |
| `--nav-h` | `72px` (60px mobile) |
| `--sidebar-w` | `260px` (admin) |
| `--topbar-h` | `64px` (admin) |

---

## Component Patterns

### Nav (Glassmorphism)
```css
.nav {
  position: sticky; top: 0; z-index: 200;
  height: var(--nav-h);
  background: var(--glass-bg);
  backdrop-filter: var(--glass-blur);
  border-bottom: 1px solid var(--glass-border);
  box-shadow: var(--glass-shadow);
}
```

### Button — Filled
```css
.btn-filled {
  font-size: 14px; font-weight: 500;
  padding: 13px 28px; border-radius: var(--r-full);
  background: var(--md-primary); color: var(--md-on-primary);
  box-shadow: 0 2px 8px rgba(155,18,48,0.30);
}
```

### Card
```css
.card {
  background: var(--md-surface-bright);
  border-radius: var(--r-lg);
  border: 1px solid var(--md-outline-variant);
  transition: box-shadow 200ms, transform 200ms;
}
.card:hover { box-shadow: var(--elev-2); transform: translateY(-2px); }
```

### Badge
```css
.badge {
  padding: 4px 10px; border-radius: var(--r-full);
  font-size: 11px; font-weight: 500;
}
```

---

## Device Layouts (CSS Grid)

| Layout | CSS | Use |
|---|---|---|
| `1col` | `grid-template-columns: 1fr` | Full-screen media |
| `2col-50` | `grid-template-columns: 1fr 1fr` | Media + Events |
| `3col-33` | `grid-template-columns: 1fr 1fr 1fr` | Multi-content |

---

## HTMX Patterns

### Fragment loading
```html
<div hx-get="/fragments/eventos" hx-trigger="load" hx-swap="outerHTML">
  <!-- Loading placeholder -->
</div>
```

### Realtime refresh via WebSocket
```javascript
ws.onmessage = function(e) {
  const msg = JSON.parse(e.data);
  if (msg.type === 'refresh_web') {
    htmx.ajax('GET', '/fragments/eventos', {target: '#eventos', swap: 'outerHTML'});
  }
};
```

---

## RBAC Roles
| Role | Permissions |
|---|---|
| `superadmin` | Everything + user management + delete |
| `director` | Everything except delete users |
| `secretaria` | Events, news, multimedia, playlists |
| `profesor` | View events, create news (own) |
| `invitado` | View-only dashboard |
