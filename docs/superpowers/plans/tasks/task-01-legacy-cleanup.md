# Task 01 — Legacy Cleanup: propiedades + ollama + categorías + seed data

**Depends on:** Task 00
**Estimated complexity:** media — múltiples archivos, sin nueva lógica

---

## Estado del repositorio al inicio de este task

```
Module: cms-plazareal (Task 00 done)
Archivos legacy presentes:
  - web/propiedades.html
  - internal/handlers/fragments/propiedades.go
  - internal/services/ollama/ollama.go
  - jcp-gestioninmobiliaria.service
  - DESIGN_AUDIT.md
  - package.json, package-lock.json
  - internal/templates/admin/pages/propiedades.html
  - internal/templates/admin/pages/whatsapp-logs.html (revisar)
Content_blocks categories: EVENTO, INFORMACIÓN, EMERGENCIA, NOTICIA (incorrecto)
Categorías correctas Plaza Real: NOTICIA, COMUNICADO, PROMOCION
Seed data: incluye contenido de escuela ("Per laborem ad lucem") y devices WEB1/T001/T002/P001/P002/P003
```

---

## Objetivo

Eliminar todos los artefactos de proyectos anteriores (propiedades, Ollama), corregir las categorías de content_blocks a los valores reales de Plaza Real, limpiar seed data de la escuela, y quitar config innecesaria de Ollama/WhatsApp en esta fase.

---

## Archivos a tocar

| Acción | Archivo |
|--------|---------|
| Eliminar | `web/propiedades.html` |
| Eliminar | `internal/handlers/fragments/propiedades.go` |
| Eliminar | `internal/services/ollama/ollama.go` |
| Eliminar | `internal/services/ollama/` (directorio) |
| Eliminar | `jcp-gestioninmobiliaria.service` |
| Eliminar | `DESIGN_AUDIT.md` |
| Eliminar | `package.json`, `package-lock.json` |
| Modificar | `internal/auth/collections.go` — limpiar seed data escuela, corregir category enum hint |
| Modificar | `internal/config/config.go` — eliminar Ollama fields, actualizar defaults |
| Modificar | `cmd/server/main.go` — quitar rutas de propiedades + referencia a ollama |
| Modificar | `internal/handlers/admin/handlers.go` — quitar imports rotos |

---

## Implementación

- [ ] **Step 1: Eliminar archivos legacy**

```bash
rm -f web/propiedades.html
rm -f internal/handlers/fragments/propiedades.go
rm -rf internal/services/ollama/
rm -f "jcp-gestioninmobiliaria.service"
rm -f DESIGN_AUDIT.md
rm -f package.json package-lock.json
```

Verificar:
```bash
ls web/propiedades.html 2>/dev/null && echo "ERROR: aún existe" || echo "✅ eliminado"
ls internal/services/ollama/ 2>/dev/null && echo "ERROR: aún existe" || echo "✅ eliminado"
```

- [ ] **Step 2: Eliminar template admin de propiedades**

```bash
rm -f internal/templates/admin/pages/propiedades.html
```

- [ ] **Step 3: Actualizar internal/config/config.go**

Buscar los campos de Ollama:
```bash
grep -n "Ollama\|ollama\|OllamaURL\|OllamaModel" internal/config/config.go
```

Eliminar del struct `Config` las líneas:
```go
// Eliminar estas líneas del struct:
OllamaURL   string
OllamaModel string
```

Eliminar del bloque `Load()`:
```go
// Eliminar estas líneas:
OllamaURL:   getEnv("OLLAMA_URL", "http://localhost:11434"),
OllamaModel: getEnv("OLLAMA_MODEL", "llama3.2"),
```

Actualizar los defaults de JCP → Plaza Real:
```go
// Cambiar:
PBAdmin:       getEnv("PB_ADMIN_EMAIL", "admin@jcp.cl"),
AdminEmail:    getEnv("ADMIN_EMAIL", "admin@jcp-gestioninmobiliaria.cl"),
AdminPassword: getEnv("ADMIN_PASSWORD", "jcp2026admin!"),
JWTSecret:     getEnv("JWT_SECRET", "jcp-secret-change-me-in-production"),
R2BucketName:  getEnv("R2_BUCKET_NAME", "jcp-media"),

// Por:
PBAdmin:       getEnv("PB_ADMIN_EMAIL", "admin@plazareal.cl"),
AdminEmail:    getEnv("ADMIN_EMAIL", "admin@plazareal.cl"),
AdminPassword: getEnv("ADMIN_PASSWORD", "plazareal2026admin!"),
JWTSecret:     getEnv("JWT_SECRET", "pr-secret-change-in-production"),
R2BucketName:  getEnv("R2_BUCKET_NAME", "plazareal-media"),
```

- [ ] **Step 4: Quitar rutas de propiedades en cmd/server/main.go**

Buscar:
```bash
grep -n "propiedades\|Propiedades\|ollama\|Ollama\|DevicePlaylist\|DeviceDisplay\|TotemDisplay" cmd/server/main.go | head -20
```

Eliminar las rutas:
```go
// Eliminar estas líneas:
app.Get("/propiedades.html", web.PageHandler(cfg, "propiedades"))
frag.Get("/propiedades-destacadas", fragments.PropiedadesDestacadas(cfg, pb))
frag.Get("/propiedades-page", fragments.PropiedadesPage(cfg, pb))
```

Si existe alguna ruta de Ollama, eliminarla también.

- [ ] **Step 5: Corregir categorías en collections.go**

Buscar la colección `content_blocks`:
```bash
grep -n "EVENTO\|INFORMACIÓN\|EMERGENCIA\|NOTICIA\|COMUNICADO\|PROMOCION\|category" internal/auth/collections.go | head -20
```

La colección `content_blocks` tiene un campo `category` de tipo `TextField`. Agregar una validación de valores permitidos. Buscar dónde se define el campo `category` en `content_blocks` y reemplazar:

```go
// Si el campo category es solo TextField sin validación, cambiar a:
&core.SelectField{
    Name:   "category",
    Values: []string{"NOTICIA", "COMUNICADO", "PROMOCION"},
},
```

Si ya existe como `TextField`, buscar la línea exacta y reemplazarla.

**Nota importante:** El campo `urgency` actualmente se setea cuando `category == "EMERGENCIA"` — con las nuevas categorías ya no tiene sentido. Revisar y eliminar el campo `urgency` de la colección si existe:

```bash
grep -n "urgency" internal/auth/collections.go
```

Si existe, eliminar la línea `&core.BoolField{Name: "urgency"}` de la definición de content_blocks.

- [ ] **Step 6: Limpiar seed data de escuela en collections.go**

Buscar el seed de content_blocks con contenido de la escuela:
```bash
grep -n "Per laborem\|Formando\|escuela\|colegio\|alumno\|laboratorio" internal/auth/collections.go
```

Reemplazar el seed de content_blocks del colegio con uno de Plaza Real:

```go
// Contenido de seed para Plaza Real (reemplaza el de la escuela):
cb.Set("title", "Plaza Real — Tu centro comercial en Copiapó")
cb.Set("description", "Más de 100 locales comerciales, restaurantes y servicios. Visítanos de lunes a domingo.")
cb.Set("category", "NOTICIA")
cb.Set("status", "publicado")
cb.Set("featured", true)
```

- [ ] **Step 7: Actualizar SeedDevicesAndPlaylists para totem Plaza Real**

Buscar el seed de devices:
```bash
grep -n "Web Hero - Landing\|Totem Entrada\|WEB1\|T001\|T002\|P001\|P002\|P003" internal/auth/collections.go | head -10
```

El seed crea 6 devices (1 web_hero + 2 totems + 3 pantallas). Plaza Real tiene 1 totem touch. Actualizar el seed para que cree solo lo necesario:

```go
devItems := []devSeed{
    {"Web Hero — Landing Pública", "web_hero", "WEB1", "Sitio Web Público"},
    {"Totem Touch — Entrada Principal", "vertical", "TOTEM1", "Entrada Principal"},
}
```

Eliminar las líneas de P001, P002, P003 del slice `devItems`.

- [ ] **Step 8: Quitar referencias a Ollama en handlers**

```bash
grep -rn "ollama\|Ollama\|OllamaURL\|OllamaModel" --include="*.go" .
```

Para cada archivo que aparezca, eliminar las líneas que referencian Ollama (imports, llamadas, etc.).

- [ ] **Step 9: Verificar compilación**

```bash
go build ./...
```

Si hay errores de "undefined: fragments.PropiedadesDestacadas" u otras referencias a funciones eliminadas, corregirlos en `cmd/server/main.go` o donde corresponda.

- [ ] **Step 10: Verificar que no quedan referencias a escuela/JCP**

```bash
grep -r "jcp\|Colegio San Lorenzo\|colegiosanlorenzo\|Per laborem\|Formando genera\|OllamaURL\|OllamaModel\|propiedades" \
  --include="*.go" --include="*.html" -i -l
```

Esperado: solo puede aparecer en archivos de plan `.md` — no en código.

- [ ] **Step 11: Commit**

```bash
git add -A
git commit -m "chore: remove propiedades/ollama legacy, fix content_block categories to NOTICIA/COMUNICADO/PROMOCION, Plaza Real seed data"
```

---

## Estado del repositorio al finalizar este task

- Sin referencias a propiedades, Ollama ni escuela en código Go/HTML
- `content_blocks.category` acepta: `NOTICIA`, `COMUNICADO`, `PROMOCION`
- Config defaults apuntan a `plazareal.cl`
- Seed de devices: solo WEB1 + TOTEM1
- `go build ./...` pasa sin errores
