# Task 03 — Documentación: README, config y .gitignore

**Depends on:** Task 01, Task 02
**Estimated complexity:** baja — solo edición de texto y config

---

## Estado del repositorio al inicio de este task

```
Module: cms-plazareal
Artefactos JCP/escuela: eliminados (Task 02)
```

Archivos que necesitan corrección:
- `README.md` — habla de "JCP Gestión Inmobiliaria", instrucciones de un inmobiliario
- `DESIGN_SYSTEM.md` — tiene tokens de color de proyecto anterior
- `internal/config/config.go` — defaults apuntan a JCP: `AdminEmail: "admin@jcp-gestioninmobiliaria.cl"`, `SiteName: "JCP Gestión Inmobiliaria"`, `R2BucketName: "jcp-media"`, `JWTSecret: "jcp-secret-change-me-in-production"`, `PBAdmin: "admin@jcp.cl"`
- `.gitignore` — no incluye `node_modules/` ni `package-lock.json`

---

## Objetivo

Reescribir README para Plaza Real, actualizar config.go con defaults correctos, agregar entradas faltantes al .gitignore, y actualizar DESIGN_SYSTEM.md para quitar referencias al proyecto anterior.

---

## Archivos a tocar

| Acción | Archivo |
|--------|---------|
| Reescribir | `README.md` |
| Modificar | `internal/config/config.go` |
| Modificar | `DESIGN_SYSTEM.md` |
| Modificar | `.gitignore` |

---

## Implementación

- [ ] **Step 1: Reescribir README.md**

Reemplazar el contenido completo de `README.md` con:

```markdown
# Plaza Real CMS

Sistema de gestión de contenido para el centro comercial **Plaza Real**.

**Stack:** Go 1.23 · Fiber v2 · PocketBase (embebido) · HTMX · Cloudflare R2 · Fly.io

## Módulos públicos

| Módulo | Ruta | Descripción |
|--------|------|-------------|
| Inicio | `/` | Landing con hero, tiendas destacadas, noticias |
| Tiendas | `/buscador-tiendas.html` | Directorio de ~100 tiendas con búsqueda HTMX |
| Noticias | `/noticias.html` | Noticias y comunicados del mall |
| Locales disponibles | `/locales.html` | Espacios en arriendo publicados por el mall |
| Eventos | `/eventos.html` | Eventos del mall con formulario de reserva |

## Inicio rápido

```bash
git clone https://github.com/PEREIRA178/cms-plazareal.git
cd cms-plazareal
go run cmd/server/main.go
```

- Servidor: `http://localhost:3000`
- PocketBase Admin: `http://localhost:8090/_/`
- Admin CMS: `http://localhost:3000/admin`

## Credenciales por defecto

| Campo | Valor |
|-------|-------|
| Email | `admin@plazareal.cl` |
| Contraseña | `plazareal2026admin!` |

> Cambiar en producción via variables de entorno `ADMIN_EMAIL` y `ADMIN_PASSWORD`.

## Variables de entorno

| Variable | Default | Descripción |
|----------|---------|-------------|
| `PORT` | `3000` | Puerto HTTP |
| `ENV` | `development` | `development` o `production` |
| `BASE_URL` | `http://localhost:3000` | URL pública del sitio |
| `ADMIN_EMAIL` | `admin@plazareal.cl` | Email de acceso al admin |
| `ADMIN_PASSWORD` | `plazareal2026admin!` | Contraseña admin |
| `JWT_SECRET` | *(ver código)* | Secreto JWT — **cambiar en producción** |
| `R2_ACCOUNT_ID` | — | Account ID Cloudflare R2 |
| `R2_ACCESS_KEY_ID` | — | Access key R2 |
| `R2_SECRET_ACCESS_KEY` | — | Secret key R2 |
| `R2_BUCKET_NAME` | `plazareal-media` | Bucket de imágenes/videos |
| `R2_PUBLIC_URL` | — | URL pública del bucket |

## Despliegue (Fly.io)

```bash
fly deploy
```

La app corre en: `https://cms-plazareal.fly.dev`

## Importación masiva de tiendas

El admin en `/admin/tiendas` acepta importación JSON masiva. Formato requerido:

```json
[
  {
    "nombre": "Nombre Tienda",
    "slug": "nombre-tienda",
    "cat": "tiendas",
    "local": "Local 10",
    "gal": "norte",
    "logo": "https://...",
    "tags": "Tag1, Tag2",
    "desc": "Descripción corta",
    "about": "Texto largo",
    "about2": "",
    "pay": "Efectivo · Tarjetas",
    "photos": "https://url1,https://url2",
    "similar": "slug1,slug2",
    "whatsapp": "56912345678",
    "telefono": "+56 9 1234 5678",
    "rating": "4.5",
    "horario_lv": "9:00 – 21:00",
    "horario_sab": "10:00 – 20:00",
    "horario_dom": "Cerrado",
    "status": "publicado",
    "destacada": "false"
  }
]
```

Para generar este JSON desde el sitio antiguo, usar el scraper:
```bash
go run cmd/scraper/main.go --url "https://OLD_SITE_URL" --gal "plaza-real" --out tiendas.json
```

## Scraper de tiendas

```bash
go run cmd/scraper/main.go --url "URL_SITIO_ANTIGUO" --gal "norte" --out tiendas_norte.json
# Luego importar via /admin/tiendas → "Importar JSON"
```
```

- [ ] **Step 2: Actualizar internal/config/config.go**

Buscar y reemplazar los valores por defecto incorrectos. Ejecutar primero:
```bash
grep -n "jcp\|csl\|Colegio\|JCP\|San Lorenzo" internal/config/config.go
```

Cambios a hacer en el bloque de defaults dentro de la función `Load()`:

```go
// Cambiar estas líneas:
PBAdmin:       getEnv("PB_ADMIN_EMAIL", "admin@jcp.cl"),
AdminEmail:    getEnv("ADMIN_EMAIL", "admin@jcp-gestioninmobiliaria.cl"),
AdminPassword: getEnv("ADMIN_PASSWORD", "jcp2026admin!"),
JWTSecret:     getEnv("JWT_SECRET", "jcp-secret-change-me-in-production"),
R2BucketName:  getEnv("R2_BUCKET_NAME", "jcp-media"),
SiteName:      "JCP Gestión Inmobiliaria",

// Por:
PBAdmin:       getEnv("PB_ADMIN_EMAIL", "admin@plazareal.cl"),
AdminEmail:    getEnv("ADMIN_EMAIL", "admin@plazareal.cl"),
AdminPassword: getEnv("ADMIN_PASSWORD", "plazareal2026admin!"),
JWTSecret:     getEnv("JWT_SECRET", "pr-secret-change-in-production"),
R2BucketName:  getEnv("R2_BUCKET_NAME", "plazareal-media"),
SiteName:      "Plaza Real CMS",
```

- [ ] **Step 3: Actualizar DESIGN_SYSTEM.md**

Buscar referencias al proyecto anterior:
```bash
grep -n "JCP\|jcp\|CSL\|csl\|realtor\|9B1230\|San Lorenzo" DESIGN_SYSTEM.md
```

Reemplazar:
- Cualquier referencia a "JCP", "CSL", "Colegio San Lorenzo", "realtor.com" por "Plaza Real"
- Color `#9B1230` (granate San Lorenzo) → `#d60d52` (rojo Plaza Real)
- Título del documento si dice algo distinto a "Plaza Real Design System"

- [ ] **Step 4: Agregar entradas a .gitignore**

Verificar el estado actual:
```bash
cat .gitignore
```

Agregar al final si no están presentes:
```
node_modules/
package-lock.json
.DS_Store
cms-plazareal
```

(El último es el binario compilado que `go build` genera en el directorio raíz.)

- [ ] **Step 5: Verificar compilación (config cambió)**

```bash
go build ./...
# Esperado: sin errores
```

- [ ] **Step 6: Commit**

```bash
git add README.md DESIGN_SYSTEM.md internal/config/config.go .gitignore
git commit -m "docs: rewrite README for Plaza Real, fix config defaults, update .gitignore"
```

---

## Estado del repositorio al finalizar este task

- `README.md` documenta Plaza Real CMS correctamente
- `config.go` defaults apuntan todos a `plazareal`
- `.gitignore` incluye `node_modules/` y `.DS_Store`
- `DESIGN_SYSTEM.md` sin referencias a proyectos anteriores
