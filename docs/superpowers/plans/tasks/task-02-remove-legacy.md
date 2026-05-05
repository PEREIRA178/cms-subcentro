# Task 02 — Eliminar artefactos heredados (JCP + Colegio San Lorenzo)

**Depends on:** Task 01 (módulo renombrado a `cms-plazareal`)
**Estimated complexity:** media — modifica 4 archivos Go grandes

---

## Estado del repositorio al inicio de este task

```
Module: cms-plazareal (ya renombrado por Task 01)
```

El proyecto fue reciclado de dos proyectos anteriores y contiene:

**Archivos a eliminar:**
- `web/propiedades.html` — página de inmuebles (inmobiliaria)
- `internal/templates/web/propiedad.html` — template detalle inmueble
- `internal/templates/admin/pages/propiedades.html` — admin de inmuebles
- `internal/handlers/fragments/propiedades.go` — fragments handler inmuebles
- `jcp-gestioninmobiliaria.service` — archivo systemd del proyecto anterior
- `DESIGN_AUDIT.md` — auditoría de diseño vs realtor.com, no aplica a Plaza Real
- `package.json` — dependencia `@openai/codex`, residuo
- `package-lock.json` — residuo

**Archivos a modificar:**

`cmd/server/main.go` — tiene rutas de propiedades y logs con nombre JCP:
```go
// Líneas a eliminar (aproximadas):
frag.Get("/propiedades-destacadas", fragments.PropiedadesDestacadas(cfg, pb))
frag.Get("/propiedades-page", fragments.PropiedadesPage(cfg, pb))
app.Get("/propiedades/:key", web.PropiedadHandler(cfg, pb))
app.Get("/propiedades.html", web.PageHandler(cfg, "propiedades"))
// Bloque admin propiedades (~líneas 190-196)
adm.Get("/propiedades", admin.PropiedadesList(cfg, pb))
// ... y 5 rutas más de propiedades

// Logs a corregir:
log.Printf("🏢 JCP Gestión Inmobiliaria en http://localhost:%s", port)
```

`internal/handlers/web/handlers.go` — contiene `PropiedadHandler` (función de ~180 líneas) y el RSS Feed con contenido de Colegio San Lorenzo.

`internal/handlers/admin/handlers.go` — contiene stubs `PropiedadesList`, `PropiedadForm`, `PropiedadCreate`, `PropiedadEdit`, `PropiedadUpdate`, `PropiedadDelete`, `PropiedadToggleStatus`.

`internal/auth/collections.go` — contiene:
- Colección `propiedades` (inmobiliaria)
- `seedPropiedades` (6 propiedades de muestra chilenas)
- `seedContentBlocks` con eventos/noticias del Colegio San Lorenzo
- `SeedDevicesAndPlaylists` con referencias a "Colegio San Lorenzo" y URLs de `colegiosanlorenzo.cl`
- Default superadmin con email `admin@jcp-gestioninmobiliaria.cl`

---

## Objetivo

Eliminar todos los archivos y código relacionados con el módulo de propiedades inmobiliarias y el Colegio San Lorenzo, y reemplazar el seed data con contenido de Plaza Real.

---

## Archivos a tocar

| Acción | Archivo |
|--------|---------|
| Eliminar | `web/propiedades.html` |
| Eliminar | `internal/templates/web/propiedad.html` |
| Eliminar | `internal/templates/admin/pages/propiedades.html` |
| Eliminar | `internal/handlers/fragments/propiedades.go` |
| Eliminar | `jcp-gestioninmobiliaria.service` |
| Eliminar | `DESIGN_AUDIT.md` |
| Eliminar | `package.json`, `package-lock.json` |
| Modificar | `cmd/server/main.go` |
| Modificar | `internal/handlers/web/handlers.go` |
| Modificar | `internal/handlers/admin/handlers.go` |
| Modificar | `internal/auth/collections.go` |

---

## Implementación

- [ ] **Step 1: Eliminar archivos**

```bash
rm web/propiedades.html \
   internal/templates/web/propiedad.html \
   internal/templates/admin/pages/propiedades.html \
   internal/handlers/fragments/propiedades.go \
   jcp-gestioninmobiliaria.service \
   DESIGN_AUDIT.md \
   package.json \
   package-lock.json
```

- [ ] **Step 2: Limpiar cmd/server/main.go**

Buscar y eliminar estas líneas/bloques (usar las búsquedas para encontrar las líneas exactas):

```bash
grep -n "propiedades\|JCP Gestión" cmd/server/main.go
```

Eliminar:
1. La línea: `app.Get("/propiedades.html", web.PageHandler(cfg, "propiedades"))`
2. El comentario y las dos líneas: `// Real-estate fragments (JCP Gestión Inmobiliaria)` + `frag.Get("/propiedades-destacadas", ...)` + `frag.Get("/propiedades-page", ...)`
3. La línea: `app.Get("/propiedades/:key", web.PropiedadHandler(cfg, pb))`
4. El bloque comentado `// Propiedades (kept for backwards compat)` y sus 7 rutas siguientes
5. El log con "JCP Gestión Inmobiliaria" → reemplazar por:

```go
log.Printf("🏢 Plaza Real CMS en http://localhost:%s", port)
log.Printf("📊 Dashboard: http://localhost:%s/admin", port)
log.Printf("🔧 PocketBase Admin: http://localhost:8090/_/")
```

- [ ] **Step 3: Limpiar internal/handlers/web/handlers.go**

3a. Eliminar el struct `propiedadData` (buscar con `grep -n "propiedadData" internal/handlers/web/handlers.go`).

3b. Eliminar la función `PropiedadHandler` completa y sus helpers `splitAndTrim`, `featBox`, `formatCLPString`, `onlyDigits`, `urlEscape`.

3c. Reemplazar `RSSFeed` para quitar referencias al Colegio San Lorenzo. Encontrar la función con:
```bash
grep -n "RSSFeed\|San Lorenzo\|colegiosanlorenzo" internal/handlers/web/handlers.go
```

Reemplazar el cuerpo de `RSSFeed` por:
```go
func RSSFeed(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		now := time.Now().Format(time.RFC1123Z)
		rss := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title>Plaza Real — Noticias y Eventos</title>
    <link>%s</link>
    <description>Noticias, eventos y novedades del centro comercial Plaza Real</description>
    <language>es-cl</language>
    <lastBuildDate>%s</lastBuildDate>
    <atom:link href="%s/rss.xml" rel="self" type="application/rss+xml"/>
  </channel>
</rss>`, cfg.BaseURL, now, cfg.BaseURL)
		c.Set("Content-Type", "application/rss+xml; charset=utf-8")
		return c.SendString(rss)
	}
}
```

3d. Verificar que el archivo `handlers.go` sigue compilando — puede que ya no necesite el import `"github.com/pocketbase/pocketbase/core"` si se eliminó `PropiedadHandler`. Revisar y limpiar imports sin usar.

- [ ] **Step 4: Limpiar internal/handlers/admin/handlers.go**

Buscar y eliminar las funciones:
```bash
grep -n "func Propiedad\|func PropiedadesList" internal/handlers/admin/handlers.go
```

Eliminar completamente: `PropiedadesList`, `PropiedadForm`, `PropiedadCreate`, `PropiedadEdit`, `PropiedadUpdate`, `PropiedadDelete`, `PropiedadToggleStatus` y la función helper `setPropiedadFields` si existe.

- [ ] **Step 5: Limpiar internal/auth/collections.go**

5a. Eliminar el bloque de colección `propiedades`. Buscar:
```bash
grep -n "11. PROPIEDADES\|propiedades" internal/auth/collections.go | head -20
```
Eliminar el bloque completo `if _, err := app.FindCollectionByNameOrId("propiedades"); err != nil { ... }`.

5b. Eliminar la llamada `seedPropiedades`:
```bash
grep -n "seedPropiedades" internal/auth/collections.go
```
Eliminar la línea `if err := seedPropiedades(app); err != nil { ... }`.

5c. Eliminar la función `seedPropiedades` completa (es la función más larga del archivo, ~140 líneas con 6 propiedades de muestra).

5d. Reemplazar el contenido de `seedContentBlocks` — la función actualmente tiene eventos y noticias del Colegio San Lorenzo. Reemplazar los slices `events` y `news` con:

```go
events := []seedBlock{
    {
        title:       "Gran Apertura Temporada Otoño 2026",
        description: "Te invitamos a celebrar la nueva temporada con descuentos exclusivos en todas nuestras tiendas. ¡Imperdible!",
        category:    "EVENTO",
        urgency:     false,
        date:        "2026-05-15 11:00:00",
        featured:    true,
    },
    {
        title:       "Horarios especiales Fiestas Patrias",
        description: "Durante el 17, 18 y 19 de septiembre Plaza Real tendrá horarios especiales. Consulta la información de tu tienda favorita.",
        category:    "INFORMACIÓN",
        urgency:     false,
        date:        "2026-09-15 00:00:00",
        featured:    true,
    },
}
news := []seedBlock{
    {
        title:       "Plaza Real renueva su food court con nuevas propuestas gastronómicas",
        description: "El centro comercial suma nuevas opciones de alimentación para todos los gustos y presupuestos.",
        category:    "NOTICIA",
        urgency:     false,
        date:        "2026-05-01 00:00:00",
        featured:    true,
    },
}
```

5e. En `SeedDevicesAndPlaylists`, reemplazar el contenido hardcodeado del Colegio San Lorenzo:
```bash
grep -n "Per laborem\|Colegio San Lorenzo\|colegiosanlorenzo\|INFRA_INA\|Comunicado-coloreate" internal/auth/collections.go
```

Reemplazar:
```go
// Cambiar:
cb.Set("title", "Per laborem ad lucem")
cb.Set("description", "Formando generaciones con excelencia académica...")
// Por:
cb.Set("title", "Bienvenidos a Plaza Real")
cb.Set("description", "Tu centro comercial de referencia. Tiendas, gastronomía y servicios en un solo lugar.")
```

Para los multimedia seeds con URLs de `colegiosanlorenzo.cl`, reemplazar las URLs por strings vacíos:
```go
// imgMM: cambiar url_r2 de la URL de wordpress a ""
imgMM.Set("url_r2", "")
imgMM.Set("filename", "hero-placeholder.png")

// vidMM: cambiar url_r2 de la URL del video de la escuela a ""
vidMM.Set("url_r2", "")
vidMM.Set("filename", "hero-video-placeholder.mp4")
vidMM.Set("start_time", 0)
```

5f. Cambiar el email del superadmin por defecto:
```bash
grep -n "admin@jcp-gestioninmobiliaria\|jcp2026admin" internal/auth/collections.go
```
Reemplazar:
```go
record.Set("email", "admin@plazareal.cl")
record.Set("password", "plazareal2026admin!")
record.Set("passwordConfirm", "plazareal2026admin!")
```

- [ ] **Step 6: Verificar compilación**

```bash
go build ./...
# Esperado: sin errores
```

Si hay errores de "undefined: X" es porque alguna función de propiedades fue referenciada en un lugar que no se limpió. Buscar con:
```bash
grep -rn "Propiedad\|propiedades" --include="*.go" . | grep -v "_test.go" | grep -v "locales_disponibles"
```

- [ ] **Step 7: Verificar que no quedan referencias al proyecto anterior**

```bash
grep -ri "jcp\|Colegio San Lorenzo\|colegiosanlorenzo\|Per laborem" \
  --include="*.go" --include="*.html" . | grep -v ".git"
# Esperado: sin resultados relevantes
```

- [ ] **Step 8: Commit**

```bash
git add -A
git commit -m "chore: remove all JCP/school artifacts, clean seed data for Plaza Real"
```

---

## Estado del repositorio al finalizar este task

- No existen archivos `propiedades.go`, `propiedades.html`, `propiedad.html`
- No existen rutas `/propiedades*` en `main.go`
- El RSS Feed devuelve contenido de Plaza Real
- El seed data de `content_blocks` tiene eventos/noticias de Plaza Real
- El email del superadmin es `admin@plazareal.cl`
- `go build ./...` pasa sin errores
