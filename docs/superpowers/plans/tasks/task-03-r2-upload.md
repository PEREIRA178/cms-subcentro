# Task 03 — R2 Upload Widget: endpoint + templ component

**Depends on:** Task 00
**Estimated complexity:** media — R2 ya está configurado, agregar endpoint + widget reutilizable

---

## Estado del repositorio al inicio de este task

```
Module: cms-plazareal
R2 service: internal/services/r2/r2.go ya existe
Config: R2AccountID, R2AccessKey, R2SecretKey, R2BucketName, R2PublicURL en config.go
Uso actual: tiendas guardan URLs externas (strings) que vienen de R2
Problema: el admin no tiene UI para subir imágenes → el admin debe poder subir y autocompletar el campo URL
```

---

## Objetivo

Crear un endpoint `POST /admin/upload` que recibe un archivo, lo sube a R2 y devuelve `{"url": "https://..."}`. Crear un templ component `UploadField` reutilizable que wrappea un `<input type="file">` con drag-drop, preview y autocompletado del campo URL oculto.

---

## Archivos a tocar

| Acción | Archivo |
|--------|---------|
| Crear | `internal/view/components/upload_field.templ` |
| Modificar | `internal/handlers/admin/handlers.go` — agregar `UploadFile` handler |
| Modificar | `cmd/server/main.go` — registrar ruta `/admin/upload` |

---

## Implementación

- [ ] **Step 1: Revisar el servicio R2 existente**

```bash
cat internal/services/r2/r2.go | head -60
```

Identificar:
1. El nombre de la función de upload (probablemente `Upload` o `UploadFile`)
2. La firma: qué recibe (reader + key + contentType) y qué devuelve

Si la función no existe o tiene firma diferente a la esperada, ajustar el handler en Step 3.

La función esperada en `r2.go`:
```go
func (c *Client) Upload(key string, body io.Reader, contentType string) (string, error)
// Retorna: URL pública del objeto subido
```

- [ ] **Step 2: Agregar UploadFile handler en internal/handlers/admin/handlers.go**

Agregar al final del archivo:

```go
// UploadFile recibe multipart/form-data con campo "file",
// lo sube a R2 y devuelve {"url": "https://..."} como JSON.
// POST /admin/upload
func UploadFile(cfg *config.Config, r2Client *r2.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fh, err := c.FormFile("file")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "no file provided"})
		}

		// 10 MB max
		const maxSize = 10 << 20
		if fh.Size > maxSize {
			return c.Status(400).JSON(fiber.Map{"error": "file too large (max 10 MB)"})
		}

		f, err := fh.Open()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "cannot open file"})
		}
		defer f.Close()

		// Construir key único: uploads/YYYY-MM/timestamp-filename
		key := fmt.Sprintf("uploads/%s/%d-%s",
			time.Now().Format("2006-01"),
			time.Now().UnixMilli(),
			sanitizeFilename(fh.Filename),
		)

		contentType := fh.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		url, err := r2Client.Upload(key, f, contentType)
		if err != nil {
			log.Printf("R2 upload error: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "upload failed"})
		}

		return c.JSON(fiber.Map{"url": url})
	}
}

func sanitizeFilename(name string) string {
	name = strings.ToLower(name)
	r := strings.NewReplacer(" ", "-", "(", "", ")", "", "[", "", "]", "")
	return r.Replace(name)
}
```

**Verificar imports necesarios en handlers.go:**
```bash
grep -n '"fmt"\|"time"\|"log"\|"strings"\|"cms-plazareal/internal/services/r2"' internal/handlers/admin/handlers.go
```

Si `r2` no está importado, agregar al bloque `import`:
```go
"cms-plazareal/internal/services/r2"
```

- [ ] **Step 3: Registrar ruta en cmd/server/main.go**

Buscar cómo se instancia el cliente R2:
```bash
grep -n "r2\.\|r2Client\|R2\|NewClient" cmd/server/main.go | head -10
```

Si el cliente R2 ya se instancia en main.go, agregar la ruta en el grupo admin:
```go
adm.Post("/upload", middleware.RoleRequired("superadmin", "director", "admin", "editor"), admin.UploadFile(cfg, r2Client))
```

Si el cliente R2 NO se instancia aún, agregar antes de definir las rutas:
```go
r2Client := r2.NewClient(cfg.R2AccountID, cfg.R2AccessKey, cfg.R2SecretKey, cfg.R2BucketName, cfg.R2PublicURL)
```

(Verificar el constructor real en `internal/services/r2/r2.go`.)

- [ ] **Step 4: Crear internal/view/components/upload_field.templ**

```bash
mkdir -p internal/view/components
```

Contenido completo de `internal/view/components/upload_field.templ`:

```templ
package components

// UploadField renders a drag-drop file upload widget.
// fieldName: the hidden input name that will receive the final URL (e.g. "logo")
// currentURL: the existing URL to pre-populate (empty on create)
// label: label shown above the widget
templ UploadField(fieldName string, currentURL string, label string) {
	<div class="upload-field" x-data={ "uploadWidget('" + fieldName + "')" }>
		<label class="form-label">{ label }</label>
		if currentURL != "" {
			<div class="upload-preview" x-show="!previewURL">
				<img :src="''" src={ currentURL } alt="Preview actual" class="upload-img-preview"/>
				<span class="upload-preview-label">Imagen actual</span>
			</div>
		}
		<div class="upload-preview" x-show="previewURL">
			<img :src="previewURL" alt="Preview nueva" class="upload-img-preview"/>
		</div>
		<div
			class="upload-dropzone"
			:class="{ 'dragover': isDragging, 'uploading': isUploading }"
			@dragover.prevent="isDragging = true"
			@dragleave.prevent="isDragging = false"
			@drop.prevent="handleDrop($event)"
			@click="$refs.fileInput.click()"
		>
			<span class="material-symbols-outlined upload-icon" x-show="!isUploading">upload</span>
			<span class="material-symbols-outlined upload-icon loading-spin" x-show="isUploading">progress_activity</span>
			<p x-show="!isUploading">Arrastra una imagen o <strong>haz clic</strong></p>
			<p x-show="isUploading">Subiendo...</p>
			<p class="upload-hint">PNG, JPG, WEBP — máx. 10 MB</p>
			<input
				type="file"
				x-ref="fileInput"
				accept="image/*"
				style="display:none"
				@change="handleFileChange($event)"
			/>
		</div>
		<input type="hidden" :name="fieldName" :value="urlValue" x-init="urlValue = currentVal"/>
		<p x-show="errorMsg" class="upload-error" x-text="errorMsg"></p>
		<p x-show="urlValue && !errorMsg" class="upload-success">
			<span class="material-symbols-outlined" style="font-size:14px">check_circle</span>
			Imagen lista
		</p>
	</div>
}

// UploadFieldScript should be included once per page that uses UploadField.
templ UploadFieldScript() {
	<script>
	function uploadWidget(fieldName) {
		return {
			fieldName,
			currentVal: document.querySelector(`[name="${fieldName}"]`)?.value || '',
			urlValue: '',
			previewURL: '',
			isDragging: false,
			isUploading: false,
			errorMsg: '',

			handleDrop(e) {
				this.isDragging = false;
				const file = e.dataTransfer.files[0];
				if (file) this.upload(file);
			},
			handleFileChange(e) {
				const file = e.target.files[0];
				if (file) this.upload(file);
			},
			async upload(file) {
				this.isUploading = true;
				this.errorMsg = '';
				const fd = new FormData();
				fd.append('file', file);
				try {
					const res = await fetch('/admin/upload', { method: 'POST', body: fd });
					const data = await res.json();
					if (!res.ok) throw new Error(data.error || 'Upload failed');
					this.urlValue = data.url;
					this.previewURL = data.url;
				} catch(e) {
					this.errorMsg = e.message;
				} finally {
					this.isUploading = false;
				}
			}
		};
	}
	</script>
	<style>
	.upload-field { display: flex; flex-direction: column; gap: 8px; }
	.upload-dropzone {
		border: 2px dashed var(--border-bright);
		border-radius: var(--r-md);
		padding: 24px 16px;
		text-align: center;
		cursor: pointer;
		transition: all .15s;
		color: var(--text-muted);
	}
	.upload-dropzone:hover, .upload-dropzone.dragover {
		border-color: var(--red);
		background: rgba(215,16,85,0.04);
		color: var(--text);
	}
	.upload-dropzone.uploading { opacity: .6; pointer-events: none; }
	.upload-icon { font-size: 28px; margin-bottom: 6px; }
	.upload-hint { font-size: .72rem; margin-top: 4px; }
	.upload-img-preview { max-height: 120px; border-radius: var(--r-md); object-fit: cover; margin-bottom: 4px; }
	.upload-preview { display: flex; flex-direction: column; align-items: flex-start; gap: 4px; }
	.upload-preview-label { font-size: .72rem; color: var(--text-muted); }
	.upload-error { font-size: .78rem; color: var(--danger); display: flex; align-items: center; gap: 4px; }
	.upload-success { font-size: .78rem; color: var(--success); display: flex; align-items: center; gap: 4px; }
	</style>
}
```

- [ ] **Step 5: Ejecutar templ generate**

```bash
make generate
```

Verificar que se generó `upload_field_templ.go`:
```bash
ls internal/view/components/
# Esperado: upload_field.templ upload_field_templ.go
```

- [ ] **Step 6: Compilar**

```bash
go build ./...
```

Si hay error en `UploadFile` por importación de `r2`, verificar la firma exacta del constructor y upload method:
```bash
grep -n "func New\|func.*Upload\|type Client" internal/services/r2/r2.go
```

Ajustar el handler para que coincida con la API real del servicio.

- [ ] **Step 7: Test manual**

```bash
go run cmd/server/main.go
```

```bash
# Test del endpoint con curl (requiere R2 configurado en env o vars de desarrollo)
curl -X POST http://localhost:3000/admin/upload \
  -H "Cookie: pr_token=TOKEN_VALIDO" \
  -F "file=@/path/to/test.jpg"
# Esperado: {"url":"https://...r2.dev/uploads/2026-05/..."}
```

Si R2 no está configurado en dev, el endpoint retorna `{"error":"upload failed"}` — eso es correcto. El widget solo falla al subir, no bloquea el formulario.

- [ ] **Step 8: Commit**

```bash
git add internal/view/components/upload_field.templ \
        internal/view/components/upload_field_templ.go \
        internal/handlers/admin/handlers.go \
        cmd/server/main.go
git commit -m "feat: add R2 upload endpoint + UploadField templ component with drag-drop"
```

---

## Uso del UploadField en formularios admin

En cualquier formulario admin que necesite un campo de imagen:

```templ
// En el form templ:
@components.UploadField("logo", store.Logo, "Logo de la tienda")
@components.UploadField("image_url", "", "Imagen del evento")

// En el <head> de la página (solo una vez):
@components.UploadFieldScript()
```

El campo oculto `<input name="logo" value="URL">` se autocompleta después del upload y se envía con el form normalmente.

---

## Estado del repositorio al finalizar este task

- `POST /admin/upload` acepta archivos y retorna JSON con URL de R2
- `UploadField` templ component reutilizable con drag-drop, preview, Alpine.js
- `UploadFieldScript` inyecta el JS/CSS necesario
- `go build ./...` pasa sin errores
