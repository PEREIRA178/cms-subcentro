# Task 08 — Fix Auth: login real PocketBase + user CRUD

**Depends on:** Task 05 (admin shell)
**Estimated complexity:** media — PocketBase ya tiene auth, ajustar JWT cookie + user CRUD en templ

---

## Estado del repositorio al inicio de este task

```
Module: cms-plazareal
Auth collection: PocketBase "users" con campos role (superadmin/director/admin/editor/viewer)
Cookie: debe llamarse "pr_token" con issuer "cms-plazareal"
JWT middleware: internal/middleware/auth.go ya existe pero puede tener issuer/cookie incorrectos
Admin user CRUD: puede existir en HTML antiguo → migrar a templ
```

---

## Objetivo

Corregir el sistema de autenticación para que:
1. El cookie se llame `pr_token` y el JWT issuer sea `cms-plazareal`
2. El handler de login use `FindAuthRecordByEmail` + `ValidatePassword` de PocketBase
3. El admin tenga CRUD completo de usuarios (listar, crear, editar rol, eliminar) en templ

---

## Archivos a tocar

| Acción | Archivo |
|--------|---------|
| Verificar/Modificar | `internal/middleware/auth.go` — issuer + cookie name |
| Verificar/Modificar | `internal/handlers/admin/handlers.go` — Login, Logout, UsersXxx handlers |
| Crear | `internal/view/pages/admin/users_page.templ` |
| Crear | `internal/view/pages/admin/user_form.templ` |
| Modificar | `cmd/server/main.go` — rutas usuarios |

---

## Implementación

- [ ] **Step 1: Auditar el middleware de auth**

```bash
grep -n "cookie\|Cookie\|pr_token\|csl_token\|issuer\|Issuer\|jwt\|JWT\|ValidatePassword\|FindAuthRecord" internal/middleware/auth.go
```

```bash
grep -n "cookie\|Cookie\|pr_token\|csl_token\|jwt\|JWT\|Login\|Logout" internal/handlers/admin/handlers.go | head -30
```

Verificar:
1. ¿El cookie se llama `pr_token`? Si es `csl_token` u otro, renombrar.
2. ¿El issuer del JWT es `cms-plazareal`? Si es `cms-csl` u otro, corregir.
3. ¿El handler de login usa `pb.FindAuthRecordByEmail` y `record.ValidatePassword`?

- [ ] **Step 2: Corregir cookie name e issuer en middleware**

Si el middleware tiene el nombre incorrecto, buscar y reemplazar en `internal/middleware/auth.go`:

```go
// Reemplazar: c.Cookies("csl_token") por:
token := c.Cookies("pr_token")

// Verificar claims del JWT:
claims := &jwt.RegisteredClaims{}
_, err := jwt.ParseWithClaims(token, claims, keyFunc)
// ...
if claims.Issuer != "cms-plazareal" {
    return c.Redirect("/admin/login")
}
```

- [ ] **Step 3: Corregir handler Login en handlers.go**

El handler de login debe seguir este patrón:

```go
// Login — GET /admin/login (formulario) y POST /admin/login (procesar)
func LoginPage(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return helpers.Render(c, adminView.Login(""))
	}
}

func LoginSubmit(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		email := c.FormValue("email")
		password := c.FormValue("password")

		record, err := pb.FindAuthRecordByEmail("users", email)
		if err != nil || !record.ValidatePassword(password) {
			return helpers.Render(c, adminView.Login("Credenciales inválidas"))
		}

		// Construir JWT
		claims := jwt.RegisteredClaims{
			Subject:   record.Id,
			Issuer:    "cms-plazareal",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, err := token.SignedString([]byte(cfg.JWTSecret))
		if err != nil {
			return c.Status(500).SendString("token error")
		}

		c.Cookie(&fiber.Cookie{
			Name:     "pr_token",
			Value:    tokenStr,
			HTTPOnly: true,
			Secure:   cfg.Env == "production",
			SameSite: "Lax",
			MaxAge:   86400,
		})

		return c.Redirect("/admin/dashboard")
	}
}

func Logout(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.ClearCookie("pr_token")
		return c.Redirect("/admin/login")
	}
}
```

**Verificar imports necesarios:**
```go
import (
    "time"
    "github.com/golang-jwt/jwt/v4"
)
```

Verificar qué librería JWT usa el proyecto:
```bash
grep -rn "golang-jwt\|dgrijalva\|jwt" go.mod
```

Ajustar el import según el resultado.

- [ ] **Step 4: Crear internal/view/pages/admin/users_page.templ**

```templ
package admin

import "cms-plazareal/internal/view/layout"

type UserRow struct {
	ID       string
	Email    string
	Name     string
	Role     string
}

type UsersPageData struct {
	Rows       []UserRow
	CurrentUID string // ID del usuario logueado (para no poder eliminarse a sí mismo)
}

templ UsersPage(d UsersPageData) {
	@layout.Admin("Usuarios", "usuarios", usersPageBody(d))
}

templ usersPageBody(d UsersPageData) {
	<div class="admin-page">
		<div class="page-header">
			<h1 class="page-title">Usuarios</h1>
			<button
				class="btn btn-primary"
				hx-get="/admin/users/new"
				hx-target="#modal-container"
				hx-swap="innerHTML"
			>
				<span class="material-symbols-outlined">person_add</span>
				Nuevo Usuario
			</button>
		</div>
		<div class="table-card">
			<table class="admin-table">
				<thead>
					<tr>
						<th>Email</th>
						<th>Nombre</th>
						<th>Rol</th>
						<th></th>
					</tr>
				</thead>
				<tbody id="users-tbody">
					for _, row := range d.Rows {
						@UserTableRow(row, d.CurrentUID)
					}
				</tbody>
			</table>
		</div>
		<div id="modal-container"></div>
	</div>
}

templ UserTableRow(u UserRow, currentUID string) {
	<tr id={ "user-row-" + u.ID }>
		<td>{ u.Email }</td>
		<td class="font-medium">{ u.Name }</td>
		<td><span class="badge badge-role">{ u.Role }</span></td>
		<td class="row-actions">
			<button class="btn-icon"
				hx-get={ "/admin/users/" + u.ID + "/edit" }
				hx-target="#modal-container"
				hx-swap="innerHTML"
				title="Editar">
				<span class="material-symbols-outlined">edit</span>
			</button>
			if u.ID != currentUID {
				<button class="btn-icon btn-danger"
					hx-delete={ "/admin/users/" + u.ID }
					hx-target={ "#user-row-" + u.ID }
					hx-swap="outerHTML"
					hx-confirm={ "¿Eliminar usuario " + u.Email + "?" }
					title="Eliminar">
					<span class="material-symbols-outlined">person_remove</span>
				</button>
			}
		</td>
	</tr>
}
```

- [ ] **Step 5: Crear internal/view/pages/admin/user_form.templ**

```templ
package admin

type UserFormData struct {
	ID       string
	Email    string
	Name     string
	Role     string
	ErrorMsg string
}

templ UserForm(d UserFormData) {
	<div class="modal-overlay" hx-get="/admin/empty" hx-trigger="click[target===this]" hx-target="#modal-container" hx-swap="innerHTML">
		<div class="modal-card" @click.stop="">
			<div class="modal-header">
				<h2 class="modal-title">
					if d.ID == "" { Nuevo Usuario } else { Editar Usuario }
				</h2>
				<button class="btn-icon modal-close" hx-get="/admin/empty" hx-target="#modal-container" hx-swap="innerHTML">
					<span class="material-symbols-outlined">close</span>
				</button>
			</div>
			if d.ErrorMsg != "" {
				<div class="alert alert-danger">{ d.ErrorMsg }</div>
			}
			<form
				if d.ID == "" { hx-post="/admin/users" } else { hx-put={ "/admin/users/" + d.ID } }
				hx-target={ "#user-row-" + d.ID }
				hx-swap="outerHTML"
				class="modal-form"
			>
				<div class="form-group">
					<label class="form-label">Email *</label>
					<input type="email" name="email" class="form-input" value={ d.Email } required/>
				</div>
				<div class="form-group">
					<label class="form-label">Nombre</label>
					<input type="text" name="name" class="form-input" value={ d.Name }/>
				</div>
				if d.ID == "" {
					<div class="form-group">
						<label class="form-label">Contraseña *</label>
						<input type="password" name="password" class="form-input" required minlength="8"/>
					</div>
					<div class="form-group">
						<label class="form-label">Confirmar Contraseña *</label>
						<input type="password" name="passwordConfirm" class="form-input" required minlength="8"/>
					</div>
				}
				<div class="form-group">
					<label class="form-label">Rol *</label>
					<select name="role" class="form-select" required>
						<option value="viewer" selected?={ d.Role == "viewer" }>Viewer</option>
						<option value="editor" selected?={ d.Role == "editor" || d.Role == "" }>Editor</option>
						<option value="admin" selected?={ d.Role == "admin" }>Admin</option>
						<option value="director" selected?={ d.Role == "director" }>Director</option>
						<option value="superadmin" selected?={ d.Role == "superadmin" }>Superadmin</option>
					</select>
				</div>
				<div class="modal-footer">
					<button type="button" class="btn btn-ghost" hx-get="/admin/empty" hx-target="#modal-container" hx-swap="innerHTML">Cancelar</button>
					<button type="submit" class="btn btn-primary">Guardar</button>
				</div>
			</form>
		</div>
	</div>
}
```

- [ ] **Step 6: Agregar handlers de usuarios en handlers.go**

```go
// UsersList — GET /admin/users
func UsersList(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		currentUID := c.Locals("userID").(string) // seteado por el middleware de auth
		records, _ := pb.FindRecordsByFilter("users", "1=1", "email", 200, 0)
		rows := make([]adminView.UserRow, 0, len(records))
		for _, r := range records {
			rows = append(rows, adminView.UserRow{
				ID:    r.Id,
				Email: r.GetString("email"),
				Name:  r.GetString("name"),
				Role:  r.GetString("role"),
			})
		}
		content := adminView.UsersPage(adminView.UsersPageData{Rows: rows, CurrentUID: currentUID})
		if c.Get("HX-Request") == "true" {
			return helpers.Render(c, content)
		}
		return helpers.Render(c, layout.Admin("Usuarios", "usuarios", content))
	}
}

// UserNew — GET /admin/users/new
func UserNew(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return helpers.Render(c, adminView.UserForm(adminView.UserFormData{}))
	}
}

// UserEdit — GET /admin/users/:id/edit
func UserEdit(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r, err := pb.FindRecordById("users", c.Params("id"))
		if err != nil {
			return c.Status(404).SendString("not found")
		}
		return helpers.Render(c, adminView.UserForm(adminView.UserFormData{
			ID:    r.Id,
			Email: r.GetString("email"),
			Name:  r.GetString("name"),
			Role:  r.GetString("role"),
		}))
	}
}

// UserCreate — POST /admin/users
func UserCreate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		col, err := pb.FindCollectionByNameOrId("users")
		if err != nil {
			return c.Status(500).SendString("collection error")
		}
		r := core.NewRecord(col)
		r.Set("email", c.FormValue("email"))
		r.Set("name", c.FormValue("name"))
		r.Set("role", c.FormValue("role"))
		r.Set("password", c.FormValue("password"))
		r.Set("passwordConfirm", c.FormValue("passwordConfirm"))
		if err := pb.Save(r); err != nil {
			return helpers.Render(c, adminView.UserForm(adminView.UserFormData{ErrorMsg: err.Error()}))
		}
		c.Set("HX-Redirect", "/admin/users")
		return c.SendStatus(204)
	}
}

// UserUpdate — PUT /admin/users/:id
func UserUpdate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r, err := pb.FindRecordById("users", c.Params("id"))
		if err != nil {
			return c.Status(404).SendString("not found")
		}
		r.Set("email", c.FormValue("email"))
		r.Set("name", c.FormValue("name"))
		r.Set("role", c.FormValue("role"))
		if err := pb.Save(r); err != nil {
			return helpers.Render(c, adminView.UserForm(adminView.UserFormData{ID: r.Id, ErrorMsg: err.Error()}))
		}
		row := adminView.UserRow{
			ID:    r.Id,
			Email: r.GetString("email"),
			Name:  r.GetString("name"),
			Role:  r.GetString("role"),
		}
		c.Set("HX-Trigger", "closeModal")
		return helpers.Render(c, adminView.UserTableRow(row, ""))
	}
}

// UserDelete — DELETE /admin/users/:id
func UserDelete(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		currentUID := c.Locals("userID").(string)
		if c.Params("id") == currentUID {
			return c.Status(400).SendString("no puedes eliminarte a ti mismo")
		}
		r, err := pb.FindRecordById("users", c.Params("id"))
		if err != nil {
			return c.Status(404).SendString("not found")
		}
		if err := pb.Delete(r); err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.SendStatus(200)
	}
}
```

**Nota:** El middleware debe setear `c.Locals("userID", claims.Subject)` para que el handler pueda leer el ID del usuario actual. Verificar que el middleware existente lo hace:
```bash
grep -n "Locals\|userID\|Subject" internal/middleware/auth.go
```

Si no lo hace, agregar en el middleware:
```go
c.Locals("userID", claims.Subject)
return c.Next()
```

- [ ] **Step 7: Registrar rutas en cmd/server/main.go**

Verificar rutas de login existentes:
```bash
grep -n "login\|Login\|logout\|Logout\|users\|Users" cmd/server/main.go | head -20
```

Asegurarse que existan estas rutas (agregar si faltan):
```go
// Auth — fuera del grupo protegido
app.Get("/admin/login", admin.LoginPage(cfg))
app.Post("/admin/login", admin.LoginSubmit(cfg, pb))
app.Get("/admin/logout", admin.Logout(cfg))

// Usuarios — dentro del grupo adm con RoleRequired superadmin
adm.Get("/users", middleware.RoleRequired("superadmin"), admin.UsersList(cfg, pb))
adm.Get("/users/new", middleware.RoleRequired("superadmin"), admin.UserNew(cfg))
adm.Get("/users/:id/edit", middleware.RoleRequired("superadmin"), admin.UserEdit(cfg, pb))
adm.Post("/users", middleware.RoleRequired("superadmin"), admin.UserCreate(cfg, pb))
adm.Put("/users/:id", middleware.RoleRequired("superadmin"), admin.UserUpdate(cfg, pb))
adm.Delete("/users/:id", middleware.RoleRequired("superadmin"), admin.UserDelete(cfg, pb))
```

- [ ] **Step 8: Compilar y verificar**

```bash
make generate
go build ./...
```

Si hay error en `c.Locals("userID").(string)` por type assertion: verificar que el middleware sí setea el valor como string. Si setea como `interface{}` que puede ser nil, usar:
```go
currentUID, _ := c.Locals("userID").(string)
```

- [ ] **Step 9: Test manual**

```bash
go run cmd/server/main.go
```

1. Ir a `http://localhost:3000/admin/login` — debe mostrar formulario de login
2. Login con credenciales válidas → redirige a `/admin/dashboard`
3. Verificar cookie `pr_token` en DevTools → Application → Cookies
4. `http://localhost:3000/admin/users` — lista de usuarios (solo para superadmin)
5. Crear usuario, editar rol, intentar eliminar usuario actual (debe mostrar error)
6. Logout → cookie limpiada, redirige a `/admin/login`

- [ ] **Step 10: Commit**

```bash
git add internal/middleware/auth.go \
        internal/view/pages/admin/users_page.templ \
        internal/view/pages/admin/users_page_templ.go \
        internal/view/pages/admin/user_form.templ \
        internal/view/pages/admin/user_form_templ.go \
        internal/handlers/admin/handlers.go \
        cmd/server/main.go
git commit -m "fix: auth cookie pr_token + issuer cms-plazareal + user CRUD admin templ"
```

---

## Estado del repositorio al finalizar este task

- Cookie `pr_token` + issuer `cms-plazareal` correctos en todo el flujo
- Login usa `FindAuthRecordByEmail` + `ValidatePassword` de PocketBase
- Admin CRUD de usuarios solo accesible por `superadmin`
- No se puede eliminar el usuario logueado actualmente
- `go build ./...` pasa sin errores
