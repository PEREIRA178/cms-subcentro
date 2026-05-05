# Task 01 — Renombrar módulo Go: jcp-gestioninmobiliaria → cms-plazareal

**Depends on:** ninguna (primera task)
**Estimated complexity:** mecánica / bajo riesgo

---

## Estado del repositorio al inicio de este task

```
Directorio raíz: /Users/matiaspereira/cms-plazareal
Rama activa: main
```

El módulo Go actualmente se llama `jcp-gestioninmobiliaria` en `go.mod` (línea 1) y en todos los imports internos de los archivos `.go`. Esto es un residuo del proyecto anterior. El repositorio GitHub ya se llama `cms-plazareal` pero el módulo Go no refleja ese nombre.

Archivos `.go` con imports internos afectados (todos tienen `"jcp-gestioninmobiliaria/internal/..."`):
- `cmd/server/main.go`
- `internal/handlers/admin/handlers.go`
- `internal/handlers/api/handlers.go`
- `internal/handlers/fragments/fragments.go`
- `internal/handlers/fragments/tiendas.go`
- `internal/handlers/fragments/propiedades.go`
- `internal/handlers/web/handlers.go`
- `internal/middleware/auth.go`
- `internal/auth/jwt.go`

---

## Objetivo

Cambiar el nombre del módulo Go de `jcp-gestioninmobiliaria` a `cms-plazareal` en `go.mod` y en todos los archivos `.go`, y verificar que el proyecto compila sin errores.

---

## Archivos a modificar

| Archivo | Cambio |
|---------|--------|
| `go.mod` línea 1 | `module jcp-gestioninmobiliaria` → `module cms-plazareal` |
| Todos los `*.go` con import interno | `"jcp-gestioninmobiliaria/` → `"cms-plazareal/` |

---

## Implementación

- [ ] **Step 1: Renombrar module en go.mod**

```bash
sed -i '' 's|module jcp-gestioninmobiliaria|module cms-plazareal|' go.mod
```

Verificar:
```bash
head -1 go.mod
# Esperado: module cms-plazareal
```

- [ ] **Step 2: Actualizar todos los imports internos**

```bash
find . -name "*.go" -not -path "./vendor/*" | xargs sed -i '' 's|"jcp-gestioninmobiliaria/|"cms-plazareal/|g'
```

- [ ] **Step 3: Verificar que no quedan referencias al nombre antiguo**

```bash
grep -r "jcp-gestioninmobiliaria" --include="*.go" .
# Esperado: sin output (ningún resultado)
```

- [ ] **Step 4: Compilar**

```bash
go build ./...
# Esperado: sin errores, sin output
```

Si hay errores de compilación en este paso, son por imports faltantes en archivos que pueden haberse saltado. Buscar con `grep -r "jcp-gestioninmobiliaria" .` y corregir manualmente.

- [ ] **Step 5: Commit**

```bash
git add go.mod $(find . -name "*.go" -not -path "./vendor/*")
git commit -m "refactor: rename Go module to cms-plazareal"
```

---

## Verificación final

```bash
head -1 go.mod
# → module cms-plazareal

grep -r "jcp-gestioninmobiliaria" --include="*.go" . | wc -l
# → 0

go build ./...
# → sin errores
```

---

## Estado del repositorio al finalizar este task

- `go.mod` línea 1: `module cms-plazareal`
- Todos los imports internos usan `"cms-plazareal/internal/..."`
- El proyecto compila limpio
