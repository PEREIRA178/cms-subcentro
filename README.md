# JCP Gestión Inmobiliaria

Sistema de gestión de propiedades (venta y arriendo) con estética inspirada en [realtor.com](https://www.realtor.com).

**Stack:** Go 1.23 + Fiber v2 + PocketBase (embebido) + HTMX + Material Design 3 + Cloudflare R2 + WebSockets

> Este repositorio **parte desde una copia del stack funcional** de `colegiosanlorenzo-v2` (CSM propio con PocketBase, rutas Fiber, fragments HTMX, auth JWT, realtime). Sobre esa base se añadió el módulo `propiedades` modelado sobre el patrón `noticias/comunicados` original.

---

## Qué cambió respecto al template escolar

| Área | Template (CSL) | Este proyecto (JCP) |
|---|---|---|
| Módulo nuevo | `content_blocks` (eventos/noticias) | **`propiedades`** (nueva colección PB con 30+ campos) |
| Handler público | `NoticiaHandler` | **`PropiedadHandler`** (`/propiedades/:key` — slug o id) |
| Fragments HTMX | `Noticias`, `Comunicados` | **`PropiedadesPage`** (filtros, sort, infinite scroll), **`PropiedadesDestacadas`** |
| Página pública | `noticias.html`, `comunicados.html` | **`web/propiedades.html`** (search bar realtor-style) |
| Template detalle | `internal/templates/web/noticia.html` | **`internal/templates/web/propiedad.html`** (galería + sticky contact card) |
| Tokens MD3 | Granate San Lorenzo `#9B1230` | Rojo realtor `#C41A2B` + navy `#324B6E` |

Archivos inalterados (se mantienen funcionales como referencia y base del admin): `comunicados.html`, `noticias.html`, `cepad.html`, `ceal.html`, etc. Pueden eliminarse en el siguiente PR si ya no se necesitan.

---

## Arquitectura

```
┌──────────────────────────────────────────────┐
│            CLIENTES (Browsers)               │
└────────┬───────────┬──────────────┬──────────┘
         │ HTTP      │ HTMX         │ WebSocket
┌────────▼───────────▼──────────────▼──────────┐
│          Go + Fiber v2 (API Gateway)         │
│  Auth MW | HTMX Fragments | WS Hub | Static  │
└────────┬───────────┬──────────────┬──────────┘
         │           │              │
┌────────▼───────────▼──────────────▼──────────┐
│           PocketBase (embebido)              │
│  users · propiedades · content_blocks · ...  │
└────────┬───────────┬──────────────┬──────────┘
         │           │              │
    ┌────▼───┐  ┌────▼────┐   ┌────▼──────┐
    │ SQLite │  │ CF R2   │   │ Twilio    │
    │pb_data │  │ media   │   │ WhatsApp  │
    └────────┘  └─────────┘   └───────────┘
```

---

## Requisitos

- Go 1.23+
- PocketBase (embebido)
- Cuenta Cloudflare R2 (para fotos de propiedades) — opcional en dev
- Cuenta Twilio (WhatsApp para corredores) — opcional en dev

---

## Inicio rápido

```bash
git clone https://github.com/PEREIRA178/jcp-gestioninmobiliaria.git
cd jcp-gestioninmobiliaria
go mod tidy

# Ejecutar
go run cmd/server/main.go

# Compilar
go build -o jcp-gestioninmobiliaria cmd/server/main.go
```

### URLs

- **Web pública:** http://localhost:3000
- **Propiedades:** http://localhost:3000/propiedades.html
- **Dashboard admin:** http://localhost:3000/admin
- **PocketBase admin:** http://localhost:8090/_/

### Credenciales por defecto

- Email: `admin@jcp-gestioninmobiliaria.cl`
- Password: `jcp2026admin!`

> ⚠️ Cambiar inmediatamente en producción.

---

## Colección `propiedades` (PocketBase)

Definida en [`internal/auth/collections.go`](internal/auth/collections.go) (bloque "10. PROPIEDADES").

Campos:

| Campo | Tipo | Notas |
|---|---|---|
| `titulo`, `slug`, `descripcion` | text/editor | `slug` se usa en URL pública |
| `operacion` | text | `VENTA` \| `ARRIENDO` |
| `tipo` | text | `CASA` \| `DEPARTAMENTO` \| `TERRENO` \| `PARCELA` \| `LOCAL` \| `OFICINA` \| `BODEGA` |
| `direccion`, `comuna`, `region` | text | — |
| `precio_uf`, `precio_clp` | number | Se muestran ambos si existen |
| `dormitorios`, `banos`, `estacionamientos` | number | — |
| `superficie_util`, `superficie_total` | number | m² |
| `ano_construccion`, `estado_propiedad` | number/text | `nueva`/`usada`/`a_estrenar` |
| `amenidades` | text | CSV: `piscina,jardin,quincho,...` |
| `status` | text | `borrador`/`publicado`/`reservada`/`vendida`/`arrendada` |
| `destacada`, `oportunidad` | bool | Badges visuales |
| `cover_image`, `gallery` | text | URL(s) de R2 |
| `tour_url` | text | 360°/Matterport |
| `lat`, `lng` | number | Para futuro mapa Leaflet |
| `contacto_whatsapp` | text | Abre `wa.me/...` desde el detalle |

Seed inicial: 6 propiedades chilenas en `seedPropiedades` dentro de `internal/auth/collections.go`.

---

## Rutas nuevas (JCP)

### Público
| Método | Ruta | Descripción |
|---|---|---|
| GET | `/propiedades.html` | Grilla con buscador + filtros (realtor-style) |
| GET | `/propiedades/:key` | Detalle de propiedad (acepta slug o id) |

### HTMX Fragments
| Método | Ruta | Descripción |
|---|---|---|
| GET | `/fragments/propiedades-destacadas` | Grilla de `destacada=true` para home |
| GET | `/fragments/propiedades-page` | Grilla filtrable paginada (htmx infinite scroll) |

Parámetros aceptados por `/fragments/propiedades-page`:
`operacion`, `tipo`, `comuna`, `dormitorios`, `precio_min`, `precio_max`, `q`, `sort`, `page`.

### Existentes del template escolar (aún funcionales)
Ver `cmd/server/main.go` para rutas de `comunicados`, `noticias`, admin dashboard, devices, websockets. Conservadas como referencia del patrón HTMX.

---

## Patrón HTMX — cómo funciona el módulo propiedades

1. `web/propiedades.html` carga la página (estática) con fuentes, estilos y navegación.
2. El `<div id="prop-results" hx-get="/fragments/propiedades-page" hx-trigger="load">` dispara al cargar.
3. El server renderiza las tarjetas en Go (ver `renderPropCard` en `internal/handlers/fragments/propiedades.go`) y devuelve HTML.
4. Al cambiar filtros (chips, select, input), HTMX re-hace el fetch y reemplaza `#prop-results`.
5. Botón "Cargar más" hace `hx-swap="beforeend"` apendizando la siguiente página.

**Todo el estado vive en el servidor.** No hay framework JS ni build step en cliente.

---

## Auditoría de diseño vs realtor.com

Ver [`DESIGN_AUDIT.md`](DESIGN_AUDIT.md) para el checklist completo (lo que ya se replicó, lo que falta, deuda de accesibilidad, y roadmap).

---

## Producción

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w" \
  -o jcp-gestioninmobiliaria cmd/server/main.go

sudo cp jcp-gestioninmobiliaria.service /etc/systemd/system/
sudo systemctl enable jcp-gestioninmobiliaria
sudo systemctl start jcp-gestioninmobiliaria
```

---

## Licencia

Propiedad de JCP Gestión Inmobiliaria. Todos los derechos reservados.
