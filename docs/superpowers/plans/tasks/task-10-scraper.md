# Task 10 — Scraper: CLI para sitio antiguo → JSON importable

**Depends on:** Task 00 (foundation — module name correcto)
**Estimated complexity:** baja-media — CLI independiente, no afecta el servidor

---

## Estado del repositorio al inicio de este task

```
Module: cms-plazareal
Colección tiendas: gal ∈ {"placa-comercial", "torre-flamenco"}
Archivo de referencia: locales_torre_flamenco_actualizado (3).json (en raíz del repo)
Scraper previo: puede existir en cmd/scraper/ con valores de gal incorrectos
```

---

## Objetivo

Crear/actualizar un CLI scraper que genera JSON con el formato exacto esperado por la colección `tiendas`. Los valores de `gal` deben ser `"placa-comercial"` o `"torre-flamenco"` (slug format, lowercase-hyphenated). El JSON resultante debe poder importarse directamente a PocketBase via su API de importación o via un comando de seeding.

---

## Archivos a tocar

| Acción | Archivo |
|--------|---------|
| Crear/Reemplazar | `cmd/scraper/main.go` |
| Crear | `cmd/scraper/scraper.go` — lógica de scraping |
| Crear | `cmd/scraper/transform.go` — normalización de datos |

---

## Formato JSON esperado (output del scraper)

Basado en el archivo `locales_torre_flamenco_actualizado (3).json`:

```json
[
  {
    "nombre": "Nombre Tienda",
    "slug": "nombre-tienda",
    "cat": "gastronomia",
    "gal": "torre-flamenco",
    "local": "101",
    "logo": "https://...",
    "tags": ["tag1", "tag2"],
    "descripcion": "Descripción corta",
    "about": "Texto largo opcional",
    "about2": "",
    "pay": "Efectivo, Tarjeta",
    "photos": ["https://..."],
    "whatsapp": "+56912345678",
    "telefono": "+5652000000",
    "rating": "4.5",
    "horario_lv": "10:00 - 21:00",
    "horario_sab": "10:00 - 21:00",
    "horario_dom": "12:00 - 20:00",
    "status": "publicado",
    "destacada": false
  }
]
```

**Valores válidos para `gal`:**
- `"placa-comercial"` — Galería Placa Comercial
- `"torre-flamenco"` — Galería Torre Flamenco

**Valores válidos para `cat`:** Libre, pero mantener consistencia. Ejemplos del JSON de referencia.

---

## Implementación

- [ ] **Step 1: Revisar el JSON de referencia y el scraper existente**

```bash
ls cmd/scraper/ 2>/dev/null && echo "existe" || echo "no existe"
```

```bash
head -50 "locales_torre_flamenco_actualizado (3).json" 2>/dev/null || echo "archivo no encontrado"
```

Si el archivo JSON de referencia existe, identificar su estructura exacta para usarla como guía del formato de output.

Si el scraper ya existe, verificar:
```bash
grep -n "gal\|galeria\|placa\|flamenco\|torre" cmd/scraper/*.go 2>/dev/null | head -20
```

- [ ] **Step 2: Crear cmd/scraper/main.go**

```go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	inputFile := flag.String("input", "", "Archivo JSON de entrada (formato legacy)")
	outputFile := flag.String("output", "tiendas_import.json", "Archivo JSON de salida")
	galeria := flag.String("gal", "torre-flamenco", "Galería: placa-comercial | torre-flamenco")
	flag.Parse()

	// Validar galería
	if *galeria != "placa-comercial" && *galeria != "torre-flamenco" {
		fmt.Fprintf(os.Stderr, "Error: gal debe ser 'placa-comercial' o 'torre-flamenco'\n")
		os.Exit(1)
	}

	var tiendas []Tienda

	if *inputFile != "" {
		// Modo transformación: leer JSON existente y normalizar
		data, err := os.ReadFile(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error leyendo archivo: %v\n", err)
			os.Exit(1)
		}
		tiendas, err = transformJSON(data, *galeria)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error transformando JSON: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Fprintln(os.Stderr, "Uso: scraper -input <archivo.json> -gal placa-comercial|torre-flamenco")
		os.Exit(1)
	}

	out, err := json.MarshalIndent(tiendas, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error serializando JSON: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*outputFile, out, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error escribiendo output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ %d tiendas escritas en %s\n", len(tiendas), *outputFile)
}
```

- [ ] **Step 3: Crear cmd/scraper/transform.go**

```go
package main

import (
	"encoding/json"
	"regexp"
	"strings"
	"unicode"
)

// Tienda es el formato de output para importar en la colección tiendas de PocketBase.
type Tienda struct {
	Nombre      string   `json:"nombre"`
	Slug        string   `json:"slug"`
	Cat         string   `json:"cat"`
	Gal         string   `json:"gal"`
	Local       string   `json:"local"`
	Logo        string   `json:"logo"`
	Tags        []string `json:"tags"`
	Descripcion string   `json:"descripcion"`
	About       string   `json:"about"`
	About2      string   `json:"about2"`
	Pay         string   `json:"pay"`
	Photos      []string `json:"photos"`
	Whatsapp    string   `json:"whatsapp"`
	Telefono    string   `json:"telefono"`
	Rating      string   `json:"rating"`
	HorarioLV   string   `json:"horario_lv"`
	HorarioSab  string   `json:"horario_sab"`
	HorarioDom  string   `json:"horario_dom"`
	Status      string   `json:"status"`
	Destacada   bool     `json:"destacada"`
}

// LegacyTienda es el formato que puede venir del JSON de referencia o del sitio antiguo.
// Los campos son flexibles: usar interface{} para los que pueden ser string o array.
type LegacyTienda map[string]interface{}

func transformJSON(data []byte, galeria string) ([]Tienda, error) {
	var raw []LegacyTienda
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	result := make([]Tienda, 0, len(raw))
	for _, legacy := range raw {
		t := Tienda{
			Gal:    galeria,
			Status: "publicado",
		}

		if v, ok := legacy["nombre"].(string); ok { t.Nombre = v }
		if v, ok := legacy["name"].(string); ok && t.Nombre == "" { t.Nombre = v }
		t.Slug = slugify(t.Nombre)

		if v, ok := legacy["cat"].(string); ok { t.Cat = v }
		if v, ok := legacy["categoria"].(string); ok && t.Cat == "" { t.Cat = v }
		t.Cat = strings.ToLower(t.Cat)

		if v, ok := legacy["local"].(string); ok { t.Local = v }
		if v, ok := legacy["local_number"].(string); ok && t.Local == "" { t.Local = v }

		if v, ok := legacy["logo"].(string); ok { t.Logo = v }
		if v, ok := legacy["logo_url"].(string); ok && t.Logo == "" { t.Logo = v }

		t.Tags = extractStringSlice(legacy, "tags")
		t.Photos = extractStringSlice(legacy, "photos", "fotos", "images")

		if v, ok := legacy["descripcion"].(string); ok { t.Descripcion = v }
		if v, ok := legacy["description"].(string); ok && t.Descripcion == "" { t.Descripcion = v }

		if v, ok := legacy["about"].(string); ok { t.About = v }
		if v, ok := legacy["about2"].(string); ok { t.About2 = v }
		if v, ok := legacy["pay"].(string); ok { t.Pay = v }
		if v, ok := legacy["pago"].(string); ok && t.Pay == "" { t.Pay = v }

		if v, ok := legacy["whatsapp"].(string); ok { t.Whatsapp = v }
		if v, ok := legacy["telefono"].(string); ok { t.Telefono = v }
		if v, ok := legacy["phone"].(string); ok && t.Telefono == "" { t.Telefono = v }
		if v, ok := legacy["rating"].(string); ok { t.Rating = v }

		if v, ok := legacy["horario_lv"].(string); ok { t.HorarioLV = v }
		if v, ok := legacy["horario_sab"].(string); ok { t.HorarioSab = v }
		if v, ok := legacy["horario_dom"].(string); ok { t.HorarioDom = v }

		if v, ok := legacy["destacada"].(bool); ok { t.Destacada = v }
		if v, ok := legacy["featured"].(bool); ok && !t.Destacada { t.Destacada = v }

		// Normalizar gal del JSON original — puede venir con valores incorrectos
		if v, ok := legacy["gal"].(string); ok && v != "" {
			normalized := normalizeGal(v)
			if normalized != "" {
				t.Gal = normalized
			}
		}

		if t.Nombre == "" {
			continue // skip registros sin nombre
		}
		result = append(result, t)
	}
	return result, nil
}

var nonAlphanumRe = regexp.MustCompile(`[^a-z0-9\-]`)
var multiDashRe = regexp.MustCompile(`-+`)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) { return r }
		return '-'
	}, s)
	s = multiDashRe.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func normalizeGal(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if strings.Contains(v, "placa") || strings.Contains(v, "comercial") {
		return "placa-comercial"
	}
	if strings.Contains(v, "flamenco") || strings.Contains(v, "torre") {
		return "torre-flamenco"
	}
	return ""
}

func extractStringSlice(m map[string]interface{}, keys ...string) []string {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			switch val := v.(type) {
			case []interface{}:
				result := make([]string, 0, len(val))
				for _, item := range val {
					if s, ok := item.(string); ok && s != "" {
						result = append(result, s)
					}
				}
				return result
			case string:
				if val != "" {
					return []string{val}
				}
			}
		}
	}
	return []string{}
}
```

- [ ] **Step 4: Compilar y verificar**

```bash
go build ./cmd/scraper/
```

Si hay errores de compilación, corregir imports en `main.go`:
```go
import (
    "encoding/json"
    "flag"
    "fmt"
    "os"
)
```

- [ ] **Step 5: Test con el JSON de referencia**

Si el archivo de referencia existe:
```bash
go run cmd/scraper/main.go \
  -input "locales_torre_flamenco_actualizado (3).json" \
  -gal torre-flamenco \
  -output tiendas_torre_flamenco.json
```

Verificar el output:
```bash
# Contar registros
python3 -c "import json; d=json.load(open('tiendas_torre_flamenco.json')); print(f'{len(d)} tiendas')"

# Verificar que gal es correcto
python3 -c "import json; d=json.load(open('tiendas_torre_flamenco.json')); gals=set(t['gal'] for t in d); print('gal values:', gals)"
# Esperado: gal values: {'torre-flamenco'}
```

- [ ] **Step 6: Agregar instrucciones de import en el README del scraper**

Crear `cmd/scraper/README.md` con instrucciones mínimas:

```markdown
# Scraper / Importador

Transforma JSON de tiendas al formato de la colección `tiendas` de cms-plazareal.

## Uso

```bash
go run cmd/scraper/main.go \
  -input datos_originales.json \
  -gal placa-comercial \
  -output tiendas_import.json
```

## Import a PocketBase

Una vez generado `tiendas_import.json`, importar via API:

```bash
# Requiere token de admin de PocketBase
PB_TOKEN="..." 
for row in $(cat tiendas_import.json | jq -c '.[]'); do
  curl -s -X POST http://localhost:8090/api/collections/tiendas/records \
    -H "Authorization: $PB_TOKEN" \
    -H "Content-Type: application/json" \
    -d "$row"
done
```
```

- [ ] **Step 7: Commit**

```bash
git add cmd/scraper/main.go cmd/scraper/transform.go cmd/scraper/README.md
git commit -m "feat: add scraper CLI — transforms legacy JSON to tiendas format with correct gal values"
```

---

## Estado del repositorio al finalizar este task

- `go run cmd/scraper/main.go -input ... -gal placa-comercial|torre-flamenco` funciona
- Output JSON tiene `gal` ∈ `{"placa-comercial", "torre-flamenco"}` siempre
- `go build ./...` pasa sin errores
- El servidor (`cmd/server`) no fue modificado — el scraper es completamente independiente
