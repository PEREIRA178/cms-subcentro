# Scraper CLI

Standalone CLI que transforma un JSON legacy de tiendas (del sitio antiguo de
plazareal.cl o un JSON curado a mano) al formato esperado por la colección
`tiendas` del CMS.

## Build

```bash
go build -o bin/scraper ./cmd/scraper/
```

## Uso

```bash
./bin/scraper -input legacy.json -gal torre-flamenco -output tiendas_import.json
```

### Flags

- `-input` (requerido): archivo JSON de entrada con un array de objetos legacy.
- `-gal` (default `torre-flamenco`): galería destino. Debe ser exactamente
  `placa-comercial` o `torre-flamenco`.
- `-output` (default `tiendas_import.json`): archivo JSON de salida.

El campo `gal` por tienda puede sobreescribirse desde el JSON legacy si trae un
valor reconocible (contiene `placa`/`comercial` o `flamenco`/`torre`); de lo
contrario se aplica el valor del flag `-gal`.

## Campos soportados (con fallbacks)

- `nombre` ← `nombre` | `name`
- `cat` ← `cat` | `categoria` (lowercased)
- `local` ← `local` | `local_number`
- `logo` ← `logo` | `logo_url`
- `descripcion` ← `descripcion` | `description`
- `pay` ← `pay` | `pago`
- `telefono` ← `telefono` | `phone`
- `photos` ← `photos` | `fotos` | `images`
- `tags` ← `tags`
- `destacada` ← `destacada` | `featured`

`status` se setea siempre a `"publicado"`. `slug` se genera con `slugify(nombre)`.
Tiendas sin `nombre` se descartan.

## Importar al CMS

```bash
curl -X POST https://cms.plazareal.cl/api/admin/import-tiendas \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  --data-binary @tiendas_import.json
```
