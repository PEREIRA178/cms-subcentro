# Task 09 — Smoke Test de producción y deploy a Fly.io

**Depends on:** Tasks 01–08 completados
**Estimated complexity:** baja — verificación y deploy

---

## Estado del repositorio al inicio de este task

```
Module: cms-plazareal
Todas las features implementadas (Tasks 01–08)
Binario compilado localmente sin errores
Entorno Fly.io: app cms-plazareal ya existe
```

**Credenciales de deploy:**
- App: `cms-plazareal` en Fly.io
- URL producción: `https://cms-plazareal.fly.dev`
- Configuración: `fly.toml` en raíz del repo

---

## Objetivo

Verificar que todo el sistema funciona end-to-end localmente antes de desplegar, hacer el deploy a Fly.io, y confirmar que producción responde correctamente.

---

## Implementación

- [ ] **Step 1: Build limpio local**

```bash
go build -o /tmp/cms-plazareal-test ./cmd/server/main.go
echo "Build OK: $?"
```

Esperado: `Build OK: 0`

Si hay errores de compilación, revisar el log y corregir antes de continuar.

- [ ] **Step 2: Verificar que no quedan referencias a proyectos anteriores**

```bash
grep -r "jcp\|Colegio San Lorenzo\|colegiosanlorenzo\|Per laborem\|jcp-gestioninmobiliaria\|csl_token\|jcp-secret" \
  --include="*.go" --include="*.html" \
  -i -l
```

Esperado: sin resultados. Si aparecen archivos, corregirlos antes de continuar.

- [ ] **Step 3: Verificar módulo Go correcto**

```bash
head -1 go.mod
```

Esperado: `module cms-plazareal`

- [ ] **Step 4: Iniciar servidor local y verificar rutas**

```bash
go run cmd/server/main.go &
SERVER_PID=$!
sleep 3
```

Luego verificar cada ruta:

```bash
# Páginas públicas
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/ | grep -q "200" && echo "✅ /" || echo "❌ /"
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/index.html | grep -q "200" && echo "✅ /index.html" || echo "❌ /index.html"
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/buscador-tiendas.html | grep -q "200" && echo "✅ /buscador-tiendas.html" || echo "❌ /buscador-tiendas.html"
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/locales.html | grep -q "200" && echo "✅ /locales.html" || echo "❌ /locales.html"
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/eventos.html | grep -q "200" && echo "✅ /eventos.html" || echo "❌ /eventos.html"
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/noticias.html | grep -q "200" && echo "✅ /noticias.html" || echo "❌ /noticias.html"

# Fragments HTMX
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/fragments/tiendas | grep -q "200" && echo "✅ /fragments/tiendas" || echo "❌ /fragments/tiendas"
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/fragments/locales-disponibles | grep -q "200" && echo "✅ /fragments/locales-disponibles" || echo "❌ /fragments/locales-disponibles"
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/fragments/eventos-public | grep -q "200" && echo "✅ /fragments/eventos-public" || echo "❌ /fragments/eventos-public"

# Admin (debe redirigir a login sin cookie)
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/admin | grep -qE "200|302|303" && echo "✅ /admin" || echo "❌ /admin"
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/admin/login | grep -q "200" && echo "✅ /admin/login" || echo "❌ /admin/login"

# Detener servidor
kill $SERVER_PID 2>/dev/null
```

Todos deben mostrar ✅. Corregir cualquier ❌ antes de continuar.

- [ ] **Step 5: Test de reserva end-to-end (manual en browser)**

Iniciar el servidor:
```bash
go run cmd/server/main.go
```

En un browser:
1. Abrir `http://localhost:3000/eventos.html`
2. Si hay eventos publicados: click en "Reservar lugar" → debe abrir el `<dialog>` modal con el formulario
3. Completar: Nombre, Email válido → Submit
4. Esperado: mensaje verde "¡Reserva recibida! Te contactaremos a la brevedad."
5. Abrir `http://localhost:3000/admin/login` → login con `admin@plazareal.cl` / `plazareal2026admin!`
6. Ir a `/admin/reservas` → debe aparecer la reserva con status "pendiente"
7. Click "Confirmar" → status debe cambiar a "confirmada"

Detener el servidor con Ctrl+C.

- [ ] **Step 6: Test de auth real (manual en browser)**

1. Abrir `http://localhost:3000/admin/login`
2. Ingresar credenciales incorrectas → debe mostrar error "Credenciales incorrectas" (no redirigir)
3. Ingresar `admin@plazareal.cl` / `plazareal2026admin!` → debe redirigir a `/admin/dashboard`
4. Verificar que el dashboard muestra stats reales (tiendas publicadas, etc.)
5. Ir a `/admin/users` → debe listar usuarios de PocketBase
6. Logout → debe redirigir a `/admin/login`

- [ ] **Step 7: Verificar fly.toml apunta al build correcto**

```bash
grep -E "build|dockerfile|app\s*=" fly.toml | head -10
```

Si hay una sección `[build]` con `dockerfile`, verificar que el Dockerfile existe:
```bash
ls Dockerfile 2>/dev/null && echo "Dockerfile OK" || echo "Dockerfile ausente"
```

Si no hay Dockerfile y fly.toml usa buildpacks, verificar:
```bash
grep "builder\|buildpacks" fly.toml
```

- [ ] **Step 8: Verificar variables de entorno en Fly.io**

```bash
fly secrets list
```

Variables requeridas en producción:
- `JWT_SECRET` — secreto real (no el default de desarrollo)
- `ADMIN_EMAIL` — `admin@plazareal.cl`
- `ADMIN_PASSWORD` — contraseña segura
- `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET_NAME`, `R2_PUBLIC_URL` — solo si R2 está configurado

Si falta `JWT_SECRET`:
```bash
fly secrets set JWT_SECRET="$(openssl rand -hex 32)"
```

- [ ] **Step 9: Deploy**

```bash
fly deploy
```

Esperado: `v<N> deployed successfully`

Si falla el deploy, revisar el log:
```bash
fly logs
```

- [ ] **Step 10: Smoke test en producción**

```bash
BASE="https://cms-plazareal.fly.dev"

curl -s -o /dev/null -w "%{http_code}" "$BASE/" | grep -q "200" && echo "✅ $BASE/" || echo "❌ $BASE/"
curl -s -o /dev/null -w "%{http_code}" "$BASE/buscador-tiendas.html" | grep -q "200" && echo "✅ /buscador-tiendas.html" || echo "❌ /buscador-tiendas.html"
curl -s -o /dev/null -w "%{http_code}" "$BASE/locales.html" | grep -q "200" && echo "✅ /locales.html" || echo "❌ /locales.html"
curl -s -o /dev/null -w "%{http_code}" "$BASE/eventos.html" | grep -q "200" && echo "✅ /eventos.html" || echo "❌ /eventos.html"
curl -s -o /dev/null -w "%{http_code}" "$BASE/noticias.html" | grep -q "200" && echo "✅ /noticias.html" || echo "❌ /noticias.html"
curl -s -o /dev/null -w "%{http_code}" "$BASE/admin/login" | grep -q "200" && echo "✅ /admin/login" || echo "❌ /admin/login"
```

Todos deben mostrar ✅.

- [ ] **Step 11: Verificar en browser producción**

En `https://cms-plazareal.fly.dev`:
- Home carga sin errores JS en consola
- Tiendas existentes aparecen en `/buscador-tiendas.html`
- `/locales.html` muestra empty state correcto (si no hay locales publicados)
- `/eventos.html` carga correctamente (empty state o eventos reales)
- `/noticias.html` muestra noticias de Plaza Real (no del colegio)
- Login en `/admin` funciona con las credenciales configuradas
- Dashboard muestra stats
- Sidebar admin incluye: Tiendas, Locales, Eventos, Noticias, Reservas, Usuarios

- [ ] **Step 12: Commit y tag de producción**

```bash
git add -A
git commit -m "chore: production smoke test passed — Plaza Real CMS v1.0 ready"
git tag v1.0.0
git push origin main --tags
```

---

## Estado del repositorio al finalizar este task

- Todos los sistemas operativos en `https://cms-plazareal.fly.dev`
- Sin referencias a proyectos anteriores (JCP, Colegio San Lorenzo)
- Auth real contra PocketBase
- Reservas funcionales end-to-end
- Tag `v1.0.0` en git
