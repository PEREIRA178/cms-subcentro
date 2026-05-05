# Task 11 — Smoke Test + Deploy Fly.io

**Depends on:** Tasks 01–10 (todo completo)
**Estimated complexity:** baja — verificar compilación, variables de entorno, deploy

---

## Estado del repositorio al inicio de este task

```
Module: cms-plazareal
Todos los tasks 00-10 completados
fly.toml: existe (modificado en el repo)
Secrets: R2_*, JWT_SECRET, ADMIN_EMAIL, ADMIN_PASSWORD deben estar en Fly.io
```

---

## Objetivo

Verificar que el servidor compila y arranca correctamente. Ejecutar smoke tests locales. Hacer deploy a Fly.io y verificar que la app funciona en producción.

---

## Implementación

- [ ] **Step 1: Compilación limpia**

```bash
make generate
go build ./...
```

Esperado: sin errores ni warnings. Si hay errores, corregirlos antes de continuar.

- [ ] **Step 2: Verificar que no quedan referencias legacy**

```bash
grep -r "jcp\|colegiosanlorenzo\|Per laborem\|OllamaURL\|OllamaModel\|csl_token\|propiedades" \
  --include="*.go" --include="*.html" --include="*.templ" -l
```

Esperado: sin resultados (o solo en archivos `.md` de planes).

- [ ] **Step 3: Smoke test local**

```bash
go run cmd/server/main.go &
sleep 2
SERVER_PID=$!
```

Verificar rutas críticas:
```bash
# Página home
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/
# Esperado: 200

# Tiendas
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/tiendas
# Esperado: 200

# Locales disponibles
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/locales-disponibles
# Esperado: 200

# Admin login (sin cookie = redirect)
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/admin/dashboard
# Esperado: 302 (redirect a /admin/login)

# Login page
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/admin/login
# Esperado: 200

# Fragmentos
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/frag/tiendas-grid
# Esperado: 200

curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/frag/noticias-cards
# Esperado: 200
```

```bash
kill $SERVER_PID 2>/dev/null
```

- [ ] **Step 4: Verificar fly.toml**

```bash
cat fly.toml
```

Verificar que `app` coincide con el nombre de la app en Fly.io. Verificar sección `[build]` y `[[services]]`. Verificar que el puerto interno coincide con el puerto de Fiber (por defecto 3000):

```toml
[[services]]
  internal_port = 3000
  protocol = "tcp"
```

Si el puerto es diferente en `cmd/server/main.go`:
```bash
grep -n "Listen\|PORT\|port" cmd/server/main.go | head -10
```

Ajustar `internal_port` en `fly.toml` según el puerto real.

- [ ] **Step 5: Verificar secrets en Fly.io**

```bash
fly secrets list
```

Verificar que existen:
- `R2_ACCOUNT_ID`
- `R2_ACCESS_KEY`
- `R2_SECRET_KEY`
- `R2_BUCKET_NAME`
- `R2_PUBLIC_URL`
- `JWT_SECRET`
- `ADMIN_EMAIL`
- `ADMIN_PASSWORD`

Si falta alguno, setear:
```bash
fly secrets set NOMBRE_SECRET=valor
```

- [ ] **Step 6: Verificar que el volume de PocketBase está configurado**

```bash
fly volumes list
```

Esperado: al menos un volumen montado en `/data` o `/pb_data` (donde PocketBase guarda la DB).

Verificar en `fly.toml` que existe:
```toml
[mounts]
  source = "pb_data"
  destination = "/pb_data"
```

Y que `cmd/server/main.go` usa la ruta correcta para el directorio de PocketBase:
```bash
grep -n "pb_data\|/data\|DataDir\|Dir" cmd/server/main.go | head -10
```

Si el DataDir usa una variable de entorno o flag, verificar que está configurado en `fly.toml` o como secret.

- [ ] **Step 7: Deploy**

```bash
fly deploy
```

El comando hace build de la imagen Docker, push y deploy. Monitorear el output — si hay errores de build, corregir.

Verificar que el deploy terminó:
```bash
fly status
```

Esperado: `running`.

- [ ] **Step 8: Smoke test en producción**

Obtener la URL de la app:
```bash
fly info | grep Hostname
```

```bash
APP_URL="https://cms-plazareal.fly.dev"

# Home
curl -s -o /dev/null -w "%{http_code}" $APP_URL/
# Esperado: 200

# Tiendas
curl -s -o /dev/null -w "%{http_code}" $APP_URL/tiendas
# Esperado: 200

# Admin login (sin cookie)
curl -s -o /dev/null -w "%{http_code}" $APP_URL/admin/dashboard
# Esperado: 302

# Admin login page
curl -s -o /dev/null -w "%{http_code}" $APP_URL/admin/login
# Esperado: 200
```

- [ ] **Step 9: Test de login en producción**

Abrir en browser: `https://cms-plazareal.fly.dev/admin/login`

1. Login con las credenciales de `ADMIN_EMAIL` + `ADMIN_PASSWORD`
2. Verificar que redirige a `/admin/dashboard`
3. Navegar a Tiendas, Noticias, Locales, Reservas — verificar que cargan
4. Verificar que el cookie `pr_token` existe en DevTools

- [ ] **Step 10: Verificar logs si hay errores**

```bash
fly logs
```

Buscar errores de:
- `panic:` — error crítico de Go
- `connection refused` — PocketBase no arrancó
- `no such table` — migración no corrió

Si PocketBase no tiene la DB inicializada en el volumen de producción, puede ser necesario correr el servidor una vez para que ejecute `initCollections()`.

- [ ] **Step 11: Commit final y tag**

```bash
git add fly.toml
git status
# Verificar que no hay cambios sin commitear
git commit -m "chore: finalize fly.toml for production deploy" --allow-empty
git tag v1.0.0
```

---

## Estado del repositorio al finalizar este task

- `cms-plazareal.fly.dev` sirve todas las rutas públicas y el admin
- Login funciona con `pr_token` cookie
- PocketBase DB inicializada en el volumen de Fly.io
- Secrets configurados en Fly.io
- Tag `v1.0.0` en git
