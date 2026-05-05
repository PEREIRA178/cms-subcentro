# Task 07 — Fix Auth: login real con PocketBase y usuarios funcionales

**Depends on:** Task 01, Task 02, Task 03
**Estimated complexity:** media — modifica auth flow crítico

---

## Estado del repositorio al inicio de este task

```
Module: cms-plazareal
Config defaults: ya apuntan a plazareal (Task 03)
```

**Problema crítico de autenticación:**

El sistema tiene una colección `users` en PocketBase (auth collection con email + password hash), pero el login **nunca la usa**. El handler `LoginSubmit` actual solo compara contra credenciales hardcodeadas en config:

```go
// CÓDIGO ACTUAL (incorrecto):
if email == cfg.AdminEmail && password == cfg.AdminPassword {
    token, err := auth.GenerateToken(cfg, "admin-id", email, "superadmin", "Administrador")
    // ...
}
```

Esto significa que solo existe un usuario posible y nunca se crea una sesión real ligada a un registro de PocketBase.

**Otros problemas:**

1. Cookie se llama `csl_token` (del Colegio San Lorenzo) → debe ser `pr_token`
2. JWT issuer dice `"jcp-gestioninmobiliaria"` → debe decir `"cms-plazareal"`
3. `UserList` sirve el template estático sin cargar usuarios reales de PocketBase
4. `UserCreate`, `UserUpdate`, `UserDelete` son stubs que devuelven un toast vacío sin persistir nada

**Archivos afectados y estado actual:**

`internal/auth/jwt.go`:
```go
Issuer: "jcp-gestioninmobiliaria",  // ← cambiar
```

`internal/middleware/auth.go`:
```go
tokenStr = c.Cookies("csl_token")   // ← cambiar a pr_token
c.ClearCookie("csl_token")          // ← cambiar a pr_token
```

`internal/handlers/admin/handlers.go` — funciones afectadas:
- `LoginSubmit(cfg *config.Config)` — solo acepta cfg, no pb
- `UserList(cfg *config.Config)` — solo sirve template estático
- `UserCreate(cfg *config.Config)` — stub
- `UserUpdate(cfg *config.Config)` — stub
- `UserDelete(cfg *config.Config)` — stub

`cmd/server/main.go` — rutas de usuarios:
```go
app.Post("/admin/login", admin.LoginSubmit(cfg))
adm.Get("/users", ..., admin.UserList(cfg))
adm.Post("/users", ..., admin.UserCreate(cfg))
adm.Put("/users/:id", ..., admin.UserUpdate(cfg))
adm.Delete("/users/:id", ..., admin.UserDelete(cfg))
```
Todas necesitan recibir `pb` además de `cfg`.

---

## Objetivo

Conectar el login y el CRUD de usuarios a PocketBase real. Renombrar la cookie. Corregir el JWT issuer.

---

## Archivos a tocar

| Archivo | Cambio |
|---------|--------|
| `internal/auth/jwt.go` | Cambiar issuer |
| `internal/middleware/auth.go` | Renombrar cookie `csl_token` → `pr_token` |
| `internal/handlers/admin/handlers.go` | Reemplazar `LoginSubmit`, `UserList`, `UserCreate`, `UserUpdate`, `UserDelete` |
| `cmd/server/main.go` | Actualizar firmas de rutas de usuarios y login |

---

## Implementación

- [ ] **Step 1: Corregir JWT issuer en internal/auth/jwt.go**

Buscar la línea:
```bash
grep -n "Issuer" internal/auth/jwt.go
```

Cambiar:
```go
Issuer: "jcp-gestioninmobiliaria",
// Por:
Issuer: "cms-plazareal",
```

- [ ] **Step 2: Renombrar cookie en internal/middleware/auth.go**

El nombre `csl_token` aparece exactamente 2 veces. Cambiar ambas:
```go
// Cambiar:
tokenStr = c.Cookies("csl_token")
c.ClearCookie("csl_token")
// Por:
tokenStr = c.Cookies("pr_token")
c.ClearCookie("pr_token")
```

- [ ] **Step 3: Reemplazar LoginSubmit en handlers/admin/handlers.go**

Buscar la función actual:
```bash
grep -n "func LoginSubmit\|func Logout" internal/handlers/admin/handlers.go
```

**Reemplazar la función `LoginSubmit` completa** con la siguiente implementación que autentica contra PocketBase. La firma cambia de `LoginSubmit(cfg *config.Config)` a `LoginSubmit(cfg *config.Config, pb *pocketbase.PocketBase)`:

```go
func LoginSubmit(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		email := strings.TrimSpace(c.FormValue("email"))
		password := c.FormValue("password")
		remember := c.FormValue("remember") == "on"

		if email == "" || password == "" {
			return c.Status(fiber.StatusBadRequest).SendString(
				`<div class="toast toast-error">Email y contraseña requeridos</div>`,
			)
		}

		// Buscar usuario en PocketBase por email
		record, err := pb.FindAuthRecordByEmail("users", email)
		if err != nil || !record.ValidatePassword(password) {
			return c.Status(fiber.StatusUnauthorized).SendString(
				`<div class="toast toast-error">Credenciales incorrectas</div>`,
			)
		}

		// Verificar que el usuario esté activo
		if !record.GetBool("activo") {
			return c.Status(fiber.StatusForbidden).SendString(
				`<div class="toast toast-error">Usuario desactivado. Contactar al administrador.</div>`,
			)
		}

		role := record.GetString("role")
		nombre := record.GetString("nombre")
		if nombre == "" {
			nombre = email
		}

		token, err := auth.GenerateToken(cfg, record.Id, email, role, nombre)
		if err != nil {
			return c.Status(500).SendString(`<div class="toast toast-error">Error generando sesión</div>`)
		}

		expiry := 24 * time.Hour
		if remember {
			expiry = 72 * time.Hour
		}
		c.Cookie(&fiber.Cookie{
			Name:     "pr_token",
			Value:    token,
			Expires:  time.Now().Add(expiry),
			HTTPOnly: true,
			Secure:   cfg.Env == "production",
			SameSite: "Lax",
			Path:     "/",
		})
		c.Set("HX-Redirect", "/admin/dashboard")
		return c.SendString("")
	}
}
```

**Reemplazar también `Logout`** para que use el nuevo nombre de cookie:
```go
func Logout() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Cookie(&fiber.Cookie{
			Name:    "pr_token",
			Value:   "",
			Expires: time.Now().Add(-time.Hour),
			Path:    "/",
		})
		c.Set("HX-Redirect", "/admin/login")
		return c.SendString("")
	}
}
```

- [ ] **Step 4: Reemplazar UserList en handlers/admin/handlers.go**

La función actual solo llama `c.SendFile(...)`. Reemplazar completamente:

```go
func UserList(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		records, _ := pb.FindRecordsByFilter("users", "", "nombre", 200, 0)

		type userRow struct {
			ID       string
			Nombre   string
			Email    string
			Role     string
			Telefono string
			Activo   string
		}
		var rows []userRow
		for _, r := range records {
			activo := "No"
			if r.GetBool("activo") {
				activo = "Sí"
			}
			rows = append(rows, userRow{
				ID:       r.Id,
				Nombre:   r.GetString("nombre"),
				Email:    r.GetString("email"),
				Role:     r.GetString("role"),
				Telefono: r.GetString("telefono"),
				Activo:   activo,
			})
		}

		tmpl, err := template.ParseFiles("./internal/templates/admin/pages/users.html")
		if err != nil {
			return c.Status(500).SendString("Template error: " + err.Error())
		}
		c.Set("Content-Type", "text/html; charset=utf-8")
		return tmpl.Execute(c, map[string]any{"Users": rows})
	}
}
```

- [ ] **Step 5: Reemplazar UserCreate**

```go
func UserCreate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		col, err := pb.FindCollectionByNameOrId("users")
		if err != nil {
			return c.Status(500).SendString(`<div class="toast toast-error">Error interno</div>`)
		}

		email := strings.TrimSpace(c.FormValue("email"))
		password := c.FormValue("password")
		nombre := strings.TrimSpace(c.FormValue("nombre"))
		role := c.FormValue("role")

		if email == "" || password == "" || role == "" {
			return c.Status(400).SendString(`<div class="toast toast-error">Email, contraseña y rol son obligatorios</div>`)
		}
		// Validar roles permitidos
		validRoles := map[string]bool{"superadmin": true, "director": true, "admin": true, "editor": true}
		if !validRoles[role] {
			return c.Status(400).SendString(`<div class="toast toast-error">Rol inválido</div>`)
		}

		record := core.NewRecord(col)
		record.Set("email", email)
		record.Set("password", password)
		record.Set("passwordConfirm", password)
		record.Set("nombre", nombre)
		record.Set("role", role)
		record.Set("telefono", c.FormValue("telefono"))
		record.Set("activo", true)
		record.Set("verified", true)

		if err := pb.Save(record); err != nil {
			return c.Status(400).SendString(`<div class="toast toast-error">No se pudo crear el usuario (¿email duplicado?)</div>`)
		}

		c.Set("HX-Redirect", "/admin/users")
		return c.SendString(`<div class="toast toast-success">Usuario creado</div>`)
	}
}
```

- [ ] **Step 6: Reemplazar UserUpdate**

```go
func UserUpdate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		record, err := pb.FindRecordById("users", c.Params("id"))
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">Usuario no encontrado</div>`)
		}

		if nombre := strings.TrimSpace(c.FormValue("nombre")); nombre != "" {
			record.Set("nombre", nombre)
		}
		if role := c.FormValue("role"); role != "" {
			record.Set("role", role)
		}
		if tel := c.FormValue("telefono"); tel != "" {
			record.Set("telefono", tel)
		}
		activoVal := c.FormValue("activo")
		if activoVal != "" {
			record.Set("activo", activoVal == "true" || activoVal == "1" || activoVal == "on")
		}
		if pw := c.FormValue("password"); pw != "" {
			record.Set("password", pw)
			record.Set("passwordConfirm", pw)
		}

		if err := pb.Save(record); err != nil {
			return c.Status(500).SendString(`<div class="toast toast-error">Error actualizando usuario</div>`)
		}
		c.Set("HX-Redirect", "/admin/users")
		return c.SendString(`<div class="toast toast-success">Usuario actualizado</div>`)
	}
}
```

- [ ] **Step 7: Reemplazar UserDelete**

```go
func UserDelete(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Prevenir auto-borrado
		currentUserID, _ := c.Locals("user_id").(string)
		if c.Params("id") == currentUserID {
			return c.Status(400).SendString(`<div class="toast toast-error">No puedes eliminarte a ti mismo</div>`)
		}

		record, err := pb.FindRecordById("users", c.Params("id"))
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">Usuario no encontrado</div>`)
		}
		if err := pb.Delete(record); err != nil {
			return c.Status(500).SendString(`<div class="toast toast-error">Error eliminando usuario</div>`)
		}
		return c.SendStatus(200)
	}
}
```

- [ ] **Step 8: Actualizar imports en handlers/admin/handlers.go**

Verificar que los imports incluyan `"github.com/pocketbase/pocketbase/core"`. Si no está, agregarlo:
```bash
grep -n '"github.com/pocketbase/pocketbase/core"' internal/handlers/admin/handlers.go
```

Si no aparece, buscar el bloque `import (` y agregar la línea.

- [ ] **Step 9: Actualizar firmas en cmd/server/main.go**

Buscar las rutas actuales:
```bash
grep -n "LoginSubmit\|UserList\|UserCreate\|UserUpdate\|UserDelete" cmd/server/main.go
```

Cambiar las firmas para incluir `pb`:
```go
// Cambiar:
app.Post("/admin/login", admin.LoginSubmit(cfg))
adm.Get("/users", middleware.RoleRequired("superadmin", "director"), admin.UserList(cfg))
adm.Post("/users", middleware.RoleRequired("superadmin", "director"), admin.UserCreate(cfg))
adm.Put("/users/:id", middleware.RoleRequired("superadmin", "director"), admin.UserUpdate(cfg))
adm.Delete("/users/:id", middleware.RoleRequired("superadmin"), admin.UserDelete(cfg))

// Por:
app.Post("/admin/login", admin.LoginSubmit(cfg, pb))
adm.Get("/users", middleware.RoleRequired("superadmin", "director"), admin.UserList(cfg, pb))
adm.Post("/users", middleware.RoleRequired("superadmin", "director"), admin.UserCreate(cfg, pb))
adm.Put("/users/:id", middleware.RoleRequired("superadmin", "director"), admin.UserUpdate(cfg, pb))
adm.Delete("/users/:id", middleware.RoleRequired("superadmin"), admin.UserDelete(cfg, pb))
```

- [ ] **Step 10: Actualizar internal/templates/admin/pages/users.html**

El template actualmente muestra una tabla vacía o con datos de muestra hardcodeados. Debe iterar sobre los usuarios de PocketBase. Abrir el archivo y buscar la sección de tabla:

```bash
grep -n "tbody\|Usuario\|usuario\|email\|role" internal/templates/admin/pages/users.html | head -20
```

Reemplazar la sección `<tbody>` para usar template data:
```html
<tbody id="users-tbody">
  {{range .Users}}
  <tr>
    <td>{{.Nombre}}</td>
    <td>{{.Email}}</td>
    <td><span class="badge badge-{{.Role}}">{{.Role}}</span></td>
    <td>{{.Telefono}}</td>
    <td>{{.Activo}}</td>
    <td>
      <button class="btn-icon"
        hx-get="/admin/users/{{.ID}}/edit"
        hx-target="#modal-container"
        hx-swap="innerHTML"
        title="Editar">
        <span class="material-symbols-outlined">edit</span>
      </button>
      <button class="btn-icon btn-danger"
        hx-delete="/admin/users/{{.ID}}"
        hx-confirm="¿Eliminar usuario {{.Nombre}}?"
        hx-target="closest tr"
        hx-swap="outerHTML swap:0.3s"
        title="Eliminar">
        <span class="material-symbols-outlined">delete</span>
      </button>
    </td>
  </tr>
  {{else}}
  <tr><td colspan="6" class="empty-state-cell">No hay usuarios registrados.</td></tr>
  {{end}}
</tbody>
```

- [ ] **Step 11: Compilar**

```bash
go build ./...
# Esperado: sin errores
```

- [ ] **Step 12: Test manual**

```bash
go run cmd/server/main.go
```

1. Abrir `http://localhost:3000/admin/login`
2. Ingresar `admin@plazareal.cl` / `plazareal2026admin!`
3. Debe redirigir a `/admin/dashboard` sin error
4. Abrir `/admin/users` — debe listar el superadmin
5. Intentar crear un usuario con email duplicado — debe mostrar error

- [ ] **Step 13: Commit**

```bash
git add internal/auth/jwt.go internal/middleware/auth.go \
        internal/handlers/admin/handlers.go \
        internal/templates/admin/pages/users.html \
        cmd/server/main.go
git commit -m "fix: real PocketBase auth login, functional user CRUD, rename cookie pr_token"
```

---

## Estado del repositorio al finalizar este task

- Login autentica contra la colección `users` de PocketBase
- `UserList` muestra usuarios reales
- `UserCreate/Update/Delete` persisten en PocketBase
- Cookie: `pr_token`
- JWT issuer: `cms-plazareal`
- `go build ./...` pasa sin errores
