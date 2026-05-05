# Task 04 — Scraper de tiendas del sitio antiguo

**Depends on:** Task 01 (módulo renombrado)
**Estimated complexity:** media — nuevo archivo Go + ajuste de selectores HTML

---

## Estado del repositorio al inicio de este task

```
Module: cms-plazareal
Directorio cmd/: solo existe cmd/server/main.go
```

El mall Plaza Real tiene ~100 tiendas pero solo 27 están cargadas en PocketBase. Las restantes deben obtenerse del sitio web antiguo del mall.

**IMPORTANTE:** Antes de ejecutar el scraper, el agente debe confirmar con el usuario la URL exacta del sitio antiguo. Preguntar: "¿Cuál es la URL del sitio antiguo de Plaza Real? Por ejemplo: `https://plazareal.cl/tiendas`"

El scraper produce un JSON en el formato exacto que acepta el botón "Importar JSON" del admin (`/admin/tiendas`). El formato fue establecido por el archivo `locales_torre_flamenco_actualizado.json` que el usuario proporcionó como referencia.

**Formato de salida requerido** — cada objeto JSON debe tener exactamente estos campos:
```json
{
  "nombre": "Nombre Tienda",
  "slug": "nombre-tienda",
  "cat": "tiendas",
  "local": "Local 10",
  "gal": "plaza-real",
  "logo": "",
  "tags": "Tag1, Tag2",
  "desc": "",
  "about": "",
  "about2": "",
  "pay": "false",
  "photos": "",
  "similar": "",
  "whatsapp": "",
  "telefono": "",
  "rating": "",
  "horario_lv": "",
  "horario_sab": "",
  "horario_dom": "",
  "status": "publicado",
  "destacada": "false"
}
```

---

## Objetivo

Crear `cmd/scraper/main.go` — una herramienta CLI que scrapea el sitio antiguo de Plaza Real y produce un JSON importable directamente via `/admin/tiendas`.

---

## Archivos a crear

| Acción | Archivo |
|--------|---------|
| Crear | `cmd/scraper/main.go` |
| Modificar | `go.mod` + `go.sum` (nueva dependencia) |

---

## Implementación

- [ ] **Step 1: Crear cmd/scraper/main.go**

Crear el directorio y archivo:
```bash
mkdir -p cmd/scraper
```

Contenido completo de `cmd/scraper/main.go`:

```go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Tienda replica exactamente el schema del importador masivo del admin (/admin/tiendas → "Importar JSON").
// Todos los campos son string para compatibilidad con el parser del admin.
type Tienda struct {
	Nombre     string `json:"nombre"`
	Slug       string `json:"slug"`
	Cat        string `json:"cat"`
	Local      string `json:"local"`
	Gal        string `json:"gal"`
	Logo       string `json:"logo"`
	Tags       string `json:"tags"`
	Desc       string `json:"desc"`
	About      string `json:"about"`
	About2     string `json:"about2"`
	Pay        string `json:"pay"`
	Photos     string `json:"photos"`
	Similar    string `json:"similar"`
	Whatsapp   string `json:"whatsapp"`
	Telefono   string `json:"telefono"`
	Rating     string `json:"rating"`
	HorarioLV  string `json:"horario_lv"`
	HorarioSab string `json:"horario_sab"`
	HorarioDom string `json:"horario_dom"`
	Status     string `json:"status"`
	Destacada  string `json:"destacada"`
}

func main() {
	siteURL := flag.String("url", "", "URL del sitio antiguo (ej: https://plazareal.cl/tiendas)")
	outFile  := flag.String("out", "tiendas_scraped.json", "Archivo JSON de salida")
	galName  := flag.String("gal", "plaza-real", "Valor del campo gal (ej: norte, sur, flamenco)")
	flag.Parse()

	if *siteURL == "" {
		log.Fatal("--url es requerido. Ejemplo: go run cmd/scraper/main.go --url https://plazareal.cl/tiendas")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(*siteURL)
	if err != nil {
		log.Fatalf("GET %s: %v", *siteURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("HTTP %d al obtener %s", resp.StatusCode, *siteURL)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatalf("parse HTML: %v", err)
	}

	var tiendas []Tienda
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		// NOTA: Los selectores de clase deben ajustarse al HTML real del sitio antiguo.
		// Inspeccionar primero con: curl -s URL | grep -i 'class=' | sort -u | head -40
		// Las clases más comunes en sitios de malls chilenos: "store-item", "tienda-item", "local-item"
		if n.Type == html.ElementNode && (hasClass(n, "store-item") || hasClass(n, "tienda-item") || hasClass(n, "local-item")) {
			t := extractTienda(n, *galName)
			if t.Nombre != "" {
				tiendas = append(tiendas, t)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	if len(tiendas) == 0 {
		log.Println("⚠️  No se encontraron tiendas. Verificar los selectores CSS en extractTienda().")
		log.Printf("   Inspeccionar HTML con: curl -s '%s' | grep -i 'class=' | sort -u | head -40", *siteURL)
	}

	out, err := json.MarshalIndent(tiendas, "", "  ")
	if err != nil {
		log.Fatalf("marshal json: %v", err)
	}

	if err := os.WriteFile(*outFile, out, 0644); err != nil {
		log.Fatalf("write %s: %v", *outFile, err)
	}

	fmt.Printf("✅ %d tiendas → %s\n", len(tiendas), *outFile)
	fmt.Println("Para importar: /admin/tiendas → botón 'Importar JSON' → seleccionar el archivo")
}

func extractTienda(n *html.Node, gal string) Tienda {
	t := Tienda{
		Gal:       gal,
		Status:    "publicado",
		Destacada: "false",
		Pay:       "false",
	}
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch {
			// Ajustar estas clases al HTML real del sitio antiguo:
			case hasClass(n, "store-name"), hasClass(n, "tienda-nombre"), hasClass(n, "local-nombre"):
				t.Nombre = textContent(n)
				t.Slug = toSlug(t.Nombre)
			case hasClass(n, "store-local"), hasClass(n, "tienda-local"), hasClass(n, "local-numero"):
				t.Local = textContent(n)
			case hasClass(n, "store-category"), hasClass(n, "tienda-cat"), hasClass(n, "local-categoria"):
				t.Cat = mapCategory(textContent(n))
			case hasClass(n, "store-desc"), hasClass(n, "tienda-desc"):
				t.Desc = textContent(n)
			case hasClass(n, "store-phone"), hasClass(n, "tienda-telefono"):
				t.Telefono = textContent(n)
			case hasClass(n, "store-tags"), hasClass(n, "tienda-tags"):
				t.Tags = textContent(n)
			case n.Data == "img" && t.Logo == "":
				src := attr(n, "src")
				if src != "" && !strings.Contains(src, "placeholder") && !strings.Contains(src, "default") {
					t.Logo = src
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return t
}

func mapCategory(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	switch {
	case strings.Contains(raw, "restaurante") || strings.Contains(raw, "gastro") ||
		strings.Contains(raw, "comida") || strings.Contains(raw, "café") || strings.Contains(raw, "cafe"):
		return "restaurantes"
	case strings.Contains(raw, "farmacia"):
		return "farmacias"
	case strings.Contains(raw, "salud") || strings.Contains(raw, "médico") ||
		strings.Contains(raw, "clínica") || strings.Contains(raw, "dental"):
		return "salud"
	case strings.Contains(raw, "tecnolog") || strings.Contains(raw, "computaci"):
		return "tecnologia"
	case strings.Contains(raw, "servicio") || strings.Contains(raw, "banco") ||
		strings.Contains(raw, "financi"):
		return "servicios"
	default:
		return "tiendas"
	}
}

func toSlug(s string) string {
	s = strings.ToLower(s)
	r := strings.NewReplacer(
		" ", "-", "á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u",
		"ü", "u", "ñ", "n", "&", "y", ".", "", ",", "", "(", "", ")", "",
		"/", "-", "'", "", "\"", "",
	)
	slug := r.Replace(s)
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	return strings.Trim(slug, "-")
}

func hasClass(n *html.Node, class string) bool {
	for _, a := range n.Attr {
		if a.Key == "class" {
			for _, c := range strings.Fields(a.Val) {
				if c == class {
					return true
				}
			}
		}
	}
	return false
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func textContent(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.TrimSpace(b.String())
}
```

- [ ] **Step 2: Agregar dependencia golang.org/x/net**

```bash
go get golang.org/x/net/html
go mod tidy
```

Verificar que se agregó:
```bash
grep "golang.org/x/net" go.mod
# Esperado: golang.org/x/net v0.x.x
```

- [ ] **Step 3: Compilar el scraper**

```bash
go build ./cmd/scraper/
# Esperado: sin errores
```

- [ ] **Step 4: Inspeccionar el HTML del sitio antiguo (requiere URL del usuario)**

**⚠️ Pausar aquí si no se tiene la URL del sitio antiguo. Preguntar al usuario.**

Una vez confirmada la URL:
```bash
curl -s "URL_DEL_SITIO" | grep -i 'class=' | tr '"' '\n' | grep -i 'class\|store\|tienda\|local\|item' | sort -u | head -30
```

Con la salida de ese comando, identificar las clases CSS reales y actualizar los `hasClass(n, "...")` en `extractTienda` para que coincidan.

- [ ] **Step 5: Ejecutar el scraper**

```bash
go run cmd/scraper/main.go --url "URL_DEL_SITIO" --gal "plaza-real" --out tiendas_scraped.json
```

Verificar el resultado:
```bash
python3 -c "
import json
d = json.load(open('tiendas_scraped.json'))
print(f'{len(d)} tiendas encontradas')
for t in d[:3]:
    print(f'  - {t[\"nombre\"]} | {t[\"cat\"]} | {t[\"local\"]}')
"
```

Si la cantidad es 0, ajustar los selectores CSS (ver Step 4) y repetir.

- [ ] **Step 6: Commit**

```bash
git add cmd/scraper/main.go go.mod go.sum
git commit -m "feat: add store scraper CLI (cmd/scraper) — produces JSON for bulk admin import"
```

---

## Instrucciones de uso post-commit

1. Ejecutar el scraper para obtener `tiendas_scraped.json`
2. Abrir `/admin/tiendas` en el browser
3. Click en botón **"Importar JSON"**
4. Seleccionar `tiendas_scraped.json`
5. Confirmar — las tiendas se cargan directamente

---

## Estado del repositorio al finalizar este task

- `cmd/scraper/main.go` existe y compila
- `go.mod` incluye `golang.org/x/net`
- El scraper genera JSON compatible con el importador del admin
