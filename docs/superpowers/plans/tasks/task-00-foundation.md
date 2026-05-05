# Task 00 — Foundation: módulo + templ + templUI + render helper

**Depends on:** nada — es el primer task
**Estimated complexity:** baja-media — setup + find-replace masivo

---

## Estado del repositorio al inicio de este task

```
Module: jcp-gestioninmobiliaria  ← incorrecto
templ: NO instalado
templUI: NO instalado
go.mod: sin a-h/templ
Archivos .templ: ninguno
```

---

## Objetivo

Renombrar el módulo Go a `cms-plazareal`, instalar `a-h/templ` + templUI CLI, crear el render helper y el Makefile, sin tocar ninguna lógica de negocio.

---

## Archivos a tocar

| Acción | Archivo |
|--------|---------|
| Modificar | `go.mod` — cambiar module name |
| Modificar | todos `*.go` — actualizar import paths |
| Crear | `internal/helpers/render.go` |
| Crear | `Makefile` |
| Modificar | `go.mod`, `go.sum` — nueva dep `a-h/templ` |

---

## Implementación

- [ ] **Step 1: Renombrar módulo en go.mod**

```bash
sed -i '' 's|module jcp-gestioninmobiliaria|module cms-plazareal|' go.mod
head -1 go.mod
# Esperado: module cms-plazareal
```

- [ ] **Step 2: Actualizar todos los import paths en archivos Go**

```bash
find . -name "*.go" -not -path "*/vendor/*" -not -path "*/.git/*" | \
  xargs sed -i '' 's|jcp-gestioninmobiliaria/|cms-plazareal/|g'
```

Verificar que no quedan referencias:
```bash
grep -r "jcp-gestioninmobiliaria" --include="*.go" . | wc -l
# Esperado: 0
```

- [ ] **Step 3: Instalar templ CLI**

```bash
go install github.com/a-h/templ/cmd/templ@latest
templ version
# Esperado: algo como "templ version: v0.2.x"
```

- [ ] **Step 4: Agregar a-h/templ como dependencia Go**

```bash
go get github.com/a-h/templ
grep "a-h/templ" go.mod
# Esperado: github.com/a-h/templ v0.2.x
```

- [ ] **Step 5: Instalar templUI CLI**

```bash
go install github.com/axzilla/templui/cmd/templui@latest
templui version
# Esperado: versión impresa sin error
```

- [ ] **Step 6: Init templUI en el proyecto**

```bash
cd /Users/matiaspereira/cms-plazareal
templui init
```

Esto crea `templui.config.json` y copia los assets CSS de templUI a `web/static/`.

Verificar:
```bash
ls templui.config.json
ls web/static/
```

- [ ] **Step 7: Agregar componentes templUI necesarios**

```bash
templui add button input textarea select badge dialog dropdown toast tabs
```

Esperado: componentes creados en `internal/ui/` (o el path que configure templUI).

Verificar:
```bash
ls internal/ui/ 2>/dev/null || ls components/ 2>/dev/null
```

Anotar el path real donde templUI creó los componentes — se usará en los tasks siguientes.

- [ ] **Step 8: Crear internal/helpers/render.go**

```bash
mkdir -p internal/helpers
```

Contenido completo de `internal/helpers/render.go`:

```go
package helpers

import (
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
)

// Render writes a templ component to the Fiber response.
// Use this in every handler instead of c.SendFile or strings.Builder.
func Render(c *fiber.Ctx, component templ.Component) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return component.Render(c.UserContext(), c.Response().BodyWriter())
}
```

- [ ] **Step 9: Crear Makefile**

Contenido completo de `Makefile`:

```makefile
.PHONY: generate build dev clean

generate:
	templ generate

build: generate
	go build ./...

dev: generate
	go run cmd/server/main.go

clean:
	find . -name "*_templ.go" -delete
	find . -name "*.templ.txt" -delete
```

- [ ] **Step 10: Crear directorio de views**

```bash
mkdir -p internal/view/layout
mkdir -p internal/view/components
mkdir -p internal/view/pages/public
mkdir -p internal/view/pages/admin
mkdir -p internal/view/fragments
```

Crear un placeholder para que Go no se queje del package vacío:

```bash
cat > internal/view/layout/.gitkeep << 'EOF'
EOF
```

- [ ] **Step 11: Verificar que el proyecto compila**

```bash
go mod tidy
go build ./...
# Esperado: sin errores de compilación
```

Si hay errores de import, verificar con:
```bash
grep -r "jcp-gestioninmobiliaria" --include="*.go" .
```

Corregir cualquier path que quedó con el nombre anterior.

- [ ] **Step 12: Commit**

```bash
git add go.mod go.sum Makefile internal/helpers/render.go \
        internal/view/ templui.config.json
git add -u  # staged changes a .go files
git commit -m "chore: rename module to cms-plazareal, add a-h/templ + templUI, render helper"
```

---

## Estado del repositorio al finalizar este task

- `go.mod` dice `module cms-plazareal`
- Todos los `.go` usan `cms-plazareal/...` como import base
- `templ` CLI instalado y en PATH
- `a-h/templ` en `go.mod`
- `templUI` CLI instalado
- `internal/helpers/render.go` existe con la función `Render`
- `Makefile` con targets `generate`, `build`, `dev`
- `internal/view/` directory structure creada
- `go build ./...` pasa sin errores
