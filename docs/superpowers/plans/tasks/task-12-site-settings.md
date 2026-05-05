# Task 12 — Site Settings: hero background home + fondo buscador

**Depends on:** Task 05 (admin shell), Task 03 (R2 upload widget)
**Estimated complexity:** baja — colección key-value simple + admin form + lecturas en páginas públicas

---

## Estado del repositorio al inicio de este task

```
Module: cms-plazareal
layout/base.templ: existe (Task 02)
Index y Tiendas pages: existen como templ (Task 04)
R2 upload widget: existe (Task 03)
Colección site_settings: no existe aún
```

---

## Objetivo

Crear una colección `site_settings` (key-value store) que permita al admin cambiar desde el panel:
1. **`hero_bg_url`** — imagen de fondo del hero en la página principal (index)
2. **`search_bg_url`** — imagen de fondo del hero en la página buscador de tiendas

Agregar una sección "Ajustes" en el admin con un formulario simple (no tabla) que usa el upload widget de R2.

---

## Archivos a tocar

| Acción | Archivo |
|--------|---------|
| Modificar | `internal/auth/collections.go` — agregar colección site_settings |
| Crear | `internal/view/pages/admin/settings_page.templ` |
| Modificar | `internal/handlers/admin/handlers.go` — agregar SettingsPage + SettingsUpdate |
| Modificar | `internal/handlers/web/handlers.go` — IndexHandler y TiendasHandler leen settings |
| Modificar | `internal/view/pages/public/index.templ` — acepta HeroBgURL param |
| Modificar | `internal/view/pages/public/tiendas.templ` — acepta SearchBgURL param |
| Modificar | `cmd/server/main.go` — registrar rutas admin ajustes |

---

## Implementación

- [ ] **Step 1: Crear colección site_settings en collections.go**

```bash
grep -n "site_settings\|SiteSettings" internal/auth/collections.go
```

Si no existe, agregar usando el mismo patrón de los otros collections:

```go
// site_settings — key-value store para configuración global del sitio
// Campos:
//   key   string  unique, e.g. "hero_bg_url"
//   value string  el valor (URL de imagen, texto, etc.)
```

```go
&migrate.CreateCollection{
    Collection: &core.Collection{
        Name: "site_settings",
        Type: core.CollectionTypeBase,
        Fields: core.FieldsList{
            &core.TextField{Name: "key", Required: true},
            &core.TextField{Name: "value"},
        },
    },
},
```

También agregar seed data para los dos settings iniciales. Buscar la función de seed y agregar:

```go
// Seed site_settings con valores vacíos (se configuran desde el admin)
settingsToSeed := []struct{ key, value string }{
    {"hero_bg_url", ""},
    {"search_bg_url", ""},
}
for _, s := range settingsToSeed {
    existing, _ := pb.FindRecordsByFilter("site_settings", "key='"+s.key+"'", "", 1, 0)
    if len(existing) > 0 {
        continue
    }
    col, err := pb.FindCollectionByNameOrId("site_settings")
    if err != nil { continue }
    r := core.NewRecord(col)
    r.Set("key", s.key)
    r.Set("value", s.value)
    _ = pb.Save(r)
}
```

- [ ] **Step 2: Crear helper para leer site_settings**

Agregar en `internal/helpers/settings.go`:

```go
package helpers

import (
    pocketbase "github.com/pocketbase/pocketbase"
)

// GetSetting retorna el valor de una clave en site_settings.
// Retorna "" si la clave no existe o hay error.
func GetSetting(pb *pocketbase.PocketBase, key string) string {
    records, err := pb.FindRecordsByFilter("site_settings", "key='"+key+"'", "", 1, 0)
    if err != nil || len(records) == 0 {
        return ""
    }
    return records[0].GetString("value")
}

// SetSetting actualiza el valor de una clave en site_settings.
func SetSetting(pb *pocketbase.PocketBase, key, value string) error {
    records, _ := pb.FindRecordsByFilter("site_settings", "key='"+key+"'", "", 1, 0)
    if len(records) == 0 {
        return nil // no existe el seed — no crear dinámicamente
    }
    r := records[0]
    r.Set("value", value)
    return pb.Save(r)
}
```

- [ ] **Step 3: Crear internal/view/pages/admin/settings_page.templ**

```templ
package admin

import (
    "cms-plazareal/internal/view/components"
    "cms-plazareal/internal/view/layout"
)

type SettingsPageData struct {
    HeroBgURL   string
    SearchBgURL string
    SuccessMsg  string
    ErrorMsg    string
}

templ SettingsPage(d SettingsPageData) {
    @layout.Admin("Ajustes del sitio", "ajustes", settingsPageBody(d))
}

templ settingsPageBody(d SettingsPageData) {
    <div class="admin-page">
        <div class="page-header">
            <h1 class="page-title">Ajustes del sitio</h1>
        </div>
        if d.SuccessMsg != "" {
            <div class="alert alert-success">{ d.SuccessMsg }</div>
        }
        if d.ErrorMsg != "" {
            <div class="alert alert-danger">{ d.ErrorMsg }</div>
        }
        <form
            hx-post="/admin/settings"
            hx-target="#settings-form-area"
            hx-swap="outerHTML"
            enctype="multipart/form-data"
            class="settings-form"
        >
            <div id="settings-form-area" class="card glass-card" style="padding:2rem; display:flex; flex-direction:column; gap:2rem;">
                <section>
                    <h2 class="section-title">Página Principal</h2>
                    <p class="text-muted text-sm">Imagen de fondo del hero en la página de inicio.</p>
                    @components.UploadField("hero_bg_url", d.HeroBgURL, "Fondo hero — Inicio")
                </section>
                <section>
                    <h2 class="section-title">Buscador de Tiendas</h2>
                    <p class="text-muted text-sm">Imagen de fondo del hero en la página /tiendas.</p>
                    @components.UploadField("search_bg_url", d.SearchBgURL, "Fondo hero — Buscador de tiendas")
                </section>
                <div style="display:flex; justify-content:flex-end;">
                    <button type="submit" class="btn btn-primary">
                        <span class="material-symbols-outlined">save</span>
                        Guardar ajustes
                    </button>
                </div>
            </div>
        </form>
        @components.UploadFieldScript()
    </div>
}
```

**Nota:** El `hx-target="#settings-form-area"` + `hx-swap="outerHTML"` permite que el handler POST devuelva solo el `settingsPageBody` interno con un mensaje de éxito, sin recargar la página completa.

- [ ] **Step 4: Agregar handlers en internal/handlers/admin/handlers.go**

```go
// SettingsPageHandler — GET /admin/settings
func SettingsPageHandler(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
    return func(c *fiber.Ctx) error {
        data := adminView.SettingsPageData{
            HeroBgURL:   helpers.GetSetting(pb, "hero_bg_url"),
            SearchBgURL: helpers.GetSetting(pb, "search_bg_url"),
        }
        content := adminView.SettingsPage(data)
        if c.Get("HX-Request") == "true" {
            return helpers.Render(c, content)
        }
        return helpers.Render(c, layout.Admin("Ajustes del sitio", "ajustes", content))
    }
}

// SettingsUpdate — POST /admin/settings
func SettingsUpdate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
    return func(c *fiber.Ctx) error {
        heroBg := c.FormValue("hero_bg_url")
        searchBg := c.FormValue("search_bg_url")

        var errMsg string
        if err := helpers.SetSetting(pb, "hero_bg_url", heroBg); err != nil {
            errMsg = "Error guardando hero_bg_url: " + err.Error()
        }
        if err := helpers.SetSetting(pb, "search_bg_url", searchBg); err != nil {
            errMsg = "Error guardando search_bg_url: " + err.Error()
        }

        successMsg := ""
        if errMsg == "" {
            successMsg = "Ajustes guardados correctamente."
        }

        // Devolver solo el interior del form (target es #settings-form-area)
        return helpers.Render(c, adminView.SettingsPage(adminView.SettingsPageData{
            HeroBgURL:   heroBg,
            SearchBgURL: searchBg,
            SuccessMsg:  successMsg,
            ErrorMsg:    errMsg,
        }))
    }
}
```

- [ ] **Step 5: Actualizar IndexHandler para leer hero_bg_url**

En `internal/handlers/web/handlers.go`, el handler `IndexHandler` (o equivalente para `/`) debe:

1. Leer `helpers.GetSetting(pb, "hero_bg_url")`
2. Pasarlo al templ component

Buscar la firma actual:
```bash
grep -n "func Index\|func Home\|Index()" internal/handlers/web/handlers.go | head -5
```

Actualizar el handler para incluir la lectura del setting:

```go
func IndexHandler(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
    return func(c *fiber.Ctx) error {
        heroBgURL := helpers.GetSetting(pb, "hero_bg_url")
        return helpers.Render(c, publicView.Index(publicView.IndexData{
            HeroBgURL: heroBgURL,
        }))
    }
}
```

Actualizar el templ `internal/view/pages/public/index.templ` para aceptar el data struct:

```templ
type IndexData struct {
    HeroBgURL string
}

templ Index(d IndexData) {
    @layout.Base("Plaza Real — Copiapó", "inicio") {
        <section class="hero-section"
            if d.HeroBgURL != "" {
                style={ "background-image: url(" + d.HeroBgURL + ")" }
            }
        >
            // ... resto del hero
        </section>
        // ... resto de la página
    }
}
```

Si `Index()` actualmente no acepta params, cambiar la firma y actualizar la llamada en el handler.

- [ ] **Step 6: Actualizar TiendasHandler para leer search_bg_url**

En `internal/handlers/web/handlers.go`, el handler de `/tiendas` (buscador):

```go
func TiendasHandler(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
    return func(c *fiber.Ctx) error {
        searchBgURL := helpers.GetSetting(pb, "search_bg_url")
        return helpers.Render(c, publicView.Tiendas(publicView.TiendasData{
            SearchBgURL: searchBgURL,
        }))
    }
}
```

Actualizar `internal/view/pages/public/tiendas.templ`:

```templ
type TiendasData struct {
    SearchBgURL string
}

templ Tiendas(d TiendasData) {
    @layout.Base("Tiendas — Plaza Real", "tiendas") {
        <section class="hero-section hero-sm"
            if d.SearchBgURL != "" {
                style={ "background-image: url(" + d.SearchBgURL + ")" }
            }
        >
            // buscador, filtros
        </section>
        // ...
    }
}
```

- [ ] **Step 7: Registrar rutas en cmd/server/main.go**

```go
// Ajustes
adm.Get("/settings", middleware.RoleRequired("superadmin", "director"), admin.SettingsPageHandler(cfg, pb))
adm.Post("/settings", middleware.RoleRequired("superadmin", "director"), admin.SettingsUpdate(cfg, pb))
```

También agregar "Ajustes" al sidebar en `internal/view/layout/admin.templ`. Buscar la lista de ítems del sidebar y agregar:
```templ
@sidebarItem("/admin/settings", "settings", "tune", "Ajustes", activePage)
```

- [ ] **Step 8: Compilar y verificar**

```bash
make generate
go build ./...
```

- [ ] **Step 9: Test manual**

```bash
go run cmd/server/main.go
```

1. Ir a `http://localhost:3000/admin/settings`
2. Subir una imagen para "Fondo hero — Inicio" (requiere R2 configurado, o usar URL de prueba)
3. Guardar → mensaje de éxito, formulario se actualiza
4. Ir a `http://localhost:3000/` → hero con la imagen de fondo elegida
5. Subir imagen para buscador → guardar → verificar en `http://localhost:3000/tiendas`

Si R2 no está configurado en dev, usar el campo URL directamente en el campo oculto para probar:
```bash
# Simular guardado directo via curl
curl -X POST http://localhost:3000/admin/settings \
  -H "Cookie: pr_token=TOKEN_VALIDO" \
  -d "hero_bg_url=https://picsum.photos/1920/600&search_bg_url=https://picsum.photos/1920/400"
```

- [ ] **Step 10: Commit**

```bash
git add internal/auth/collections.go \
        internal/helpers/settings.go \
        internal/view/pages/admin/settings_page.templ \
        internal/view/pages/admin/settings_page_templ.go \
        internal/view/pages/public/index.templ \
        internal/view/pages/public/index_templ.go \
        internal/view/pages/public/tiendas.templ \
        internal/view/pages/public/tiendas_templ.go \
        internal/view/layout/admin.templ \
        internal/view/layout/admin_templ.go \
        internal/handlers/admin/handlers.go \
        internal/handlers/web/handlers.go \
        cmd/server/main.go
git commit -m "feat: site_settings collection + admin ajustes page for hero/search bg images"
```

---

## Estado del repositorio al finalizar este task

- `site_settings` collection con seed `hero_bg_url` + `search_bg_url`
- Admin `/admin/settings` permite cambiar ambas imágenes via R2 upload
- Homepage hero usa `hero_bg_url` si está configurada
- Buscador de tiendas usa `search_bg_url` si está configurada
- Sin imagen configurada → fondo negro (comportamiento previo intacto)
- `go build ./...` pasa sin errores
