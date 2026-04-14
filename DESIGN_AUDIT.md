# DESIGN AUDIT — JCP Gestión Inmobiliaria vs realtor.com

**Fecha:** Abril 2026
**Base visual actual:** copiada desde `colegiosanlorenzo-v2` (Material Design 3 Expressive, tokens MD3, HTMX).
**Aspiración:** [realtor.com](https://www.realtor.com) — especialista en compradores de casa.

Este documento sirve como **checklist de auditoría de HTML/UX**: qué tiene realtor.com que deberíamos replicar, qué debemos adaptar de la base escolar, y qué cambios concretos (file/line) quedan pendientes para que `web/propiedades.html` y `internal/templates/web/propiedad.html` alcancen esa referencia.

---

## 1. Objetivo de producto

| Dimensión | Colegio San Lorenzo (base) | JCP Gestión Inmobiliaria (meta) | Referencia realtor.com |
|---|---|---|---|
| Usuario | Apoderados + comunidad escolar | Comprador / arrendatario / corredor | Home buyer |
| Acción principal | Ver comunicados | **Buscar propiedad → contactar** | Search → filter → save → contact agent |
| Unidad de contenido | Comunicado / noticia | **Propiedad** (casa/depto/terreno) | Listing |
| Fragmento clave HTMX | `/fragments/noticias` | `/fragments/propiedades-page` | (server render + client filters) |

---

## 2. Paleta y tokens — cambios aplicados

**Base escolar (granate):**
```css
--md-primary: #9B1230;   /* granate San Lorenzo */
--md-primary-container: #F5E0E4;
```

**JCP / realtor.com (rojo vibrante + navy):**
```css
--md-primary: #C41A2B;            /* rojo realtor */
--md-primary-dark: #900F1F;
--md-primary-accent: #F7931E;     /* naranja CTA / badge Oportunidad */
--md-secondary: #324B6E;          /* navy (filter bar, hero) */
--md-secondary-container: #E5ECF5;
--md-surface: #FAFAF9;            /* gris más frío */
--md-outline: #8B95A3;
--md-inverse-surface: #0F1419;    /* footer near-black */
```

**Ubicación:**
- `web/propiedades.html` línea ~12-30 (aplicado ✅)
- `internal/templates/web/propiedad.html` línea ~14-28 (aplicado ✅)
- `web/index.html` — **pendiente**: sigue con paleta granate escolar. Ver §9.

---

## 3. Bloques de la home que realtor.com tiene y que debemos tener

| # | Bloque realtor.com | Estado en JCP |
|---|---|---|
| 1 | Hero con barra de búsqueda grande (ciudad/código postal + tabs Buy/Rent/Sold) | ✅ Implementado en `web/propiedades.html` (`.search-bar` con tabs) |
| 2 | Filtros rápidos bajo el hero (dormitorios, precio, tipo) | ✅ Implementado como chips HTMX |
| 3 | Grid de "Recently listed / Popular homes" con tarjetas foto-primero | ✅ `.prop-grid` con `.prop-card` |
| 4 | Badges sobre la foto (New, Open House, Price Reduced) | ✅ `prop-badge-featured`, `prop-badge-deal`, `prop-badge-venta` |
| 5 | Heart/save icon en cada card | ✅ `.prop-fav` toggle (persistencia pendiente) |
| 6 | Precio grande + bedrooms / baths / sqft en línea | ✅ `.prop-price` + `.prop-feats` con SVG icons |
| 7 | Vista detalle con galería hero (1 grande + 4 thumbs) | ✅ `internal/templates/web/propiedad.html` `.gallery` CSS grid |
| 8 | Sticky contact card en detalle ("Contact agent") | ✅ `.contact-card` position sticky + form + WhatsApp |
| 9 | Mapa interactivo / Street View | ❌ **Pendiente** — ver §7 |
| 10 | Similar homes / "More like this" | ❌ **Pendiente** — ver §8 |
| 11 | Mortgage calculator / affordability | ❌ Fuera de alcance v1 — ver §10 |
| 12 | School ratings widget | N/A (no aplica) |

---

## 4. Tipografía

| Uso | Base escolar | realtor.com | JCP adoptado |
|---|---|---|---|
| Display (titulares) | `DM Serif Display` | Inter / Graphik (sans pesada) | `DM Serif Display` (mantenido por coherencia) |
| Body | `DM Sans` | Inter | `DM Sans` |
| Precios | serif grande | sans bold | **serif grande** → da un look más premium que realtor para propiedades chilenas, diferenciador |

**Decisión:** mantener DM Serif Display para precios y titulares. realtor.com usa sans-pesado para precios; nosotros apostamos por serif grande como sello premium boutique.

---

## 5. Cards — checklist de paridad

- [x] Foto 4/3 con `object-fit:cover`
- [x] Badges posicionados `top:12px left:12px` con backdrop-blur
- [x] Heart icon `top:12px right:12px` con `click.stopPropagation`
- [x] Precio grande (serif) — $/UF
- [x] 3–4 features con iconos SVG inline (cama, baño, m², auto)
- [x] Dirección truncada a 2 líneas con pin icon
- [x] Hover: `translateY(-4px)` + shadow + zoom suave de la foto
- [ ] **Carrusel de fotos dentro de la card** (realtor.com permite paginar sin entrar al detalle) — pendiente
- [ ] **"Price cut" / indicador de bajada** con flecha roja — pendiente (campo `precio_anterior` en collection)
- [ ] **Tiempo en mercado** ("New · 2 days on market") — pendiente (campo `publicado_en` ya existe, falta renderear)

---

## 6. Buscador tipo realtor (hero)

Realtor.com tiene:
- **Input único autocompletado** ("Enter a city, address, ZIP code, or neighborhood")
- **Tabs Buy / Rent / Sold / Home Value**
- **Dropdown de tipo** escondido detrás de "More filters"

Nosotros implementamos (✅ en `web/propiedades.html` líneas ~140-180):
- Tabs `Todas / Comprar / Arrendar`
- Input único con HTMX debounce 400ms
- Select de tipo visible (Casa/Depto/Parcela/etc.)
- Botón submit primary red

**Falta:** geocoding con autocompletado de comunas chilenas. Opción ligera: pre-cargar un JSON `web/static/comunas.json` y usar `<datalist>`.

---

## 7. Mapa — pendiente

Realtor.com tiene vista dividida lista + mapa interactivo con pins clickeables y filtros por área dibujada. Para v1:

**Opción ligera (recomendada):** Leaflet + OpenStreetMap en la vista detalle (`propiedad.html`) para mostrar la ubicación de una sola propiedad. Los campos `lat`, `lng` ya existen en la colección `propiedades`.

**Opción completa (v2):** añadir `/propiedades-mapa.html` con vista split-screen (Leaflet + grid scroll).

**Tracking:** crear issue "Añadir mapa Leaflet al detalle de propiedad".

---

## 8. "Similar homes" — pendiente

Al final del detalle realtor muestra 3-4 cards similares. Implementación sugerida:

```go
// internal/handlers/fragments/propiedades.go
// Nuevo: PropiedadesSimilares(id) - mismo tipo + comuna + ±30% precio
func PropiedadesSimilares(pb *pocketbase.PocketBase, id string) ([]propiedad, error) { ... }
```

Ruta: `/fragments/propiedades-similares/:id` → incluido vía `hx-get` al final de `propiedad.html`.

---

## 9. index.html — pendiente

Actualmente `web/index.html` sigue siendo la home del colegio (granate + Per laborem ad lucem + hero con logo SL). Rutas posibles:

**A.** Redirigir `/` → `/propiedades.html` (más rápido, suficiente para v1). En `cmd/server/main.go` línea 73, cambiar:
```go
app.Get("/", web.IndexHandler(cfg))
```
por:
```go
app.Get("/", func(c *fiber.Ctx) error { return c.Redirect("/propiedades.html") })
```

**B.** Reemplazar `web/index.html` por un home-marketing (hero + propiedades destacadas + "Cómo funciona" + CTA corredores). Usar `/fragments/propiedades-destacadas` (ya implementado en `fragments/propiedades.go`).

**Decisión para este PR:** dejar `index.html` intacto; agregar `/propiedades.html` como ruta funcional. Siguiente iteración: opción B.

---

## 10. Fuera de alcance (v1)

- Calculadora de crédito hipotecario (simulador UF → dividendo)
- Comparador de propiedades (checkbox en card + barra inferior)
- Saved searches con alertas por email
- Tour 3D tipo Matterport (campo `tour_url` ya existe, basta `<iframe>` en detalle cuando haya contenido)

---

## 11. Deuda de accesibilidad identificada

- [ ] Contraste del navy `#324B6E` sobre fondo oscuro del hero: verificar WCAG AA
- [ ] `aria-label` en todos los SVG icons de `.prop-feat` (actualmente decorativos sin label)
- [ ] Keyboard navigation en tabs de búsqueda (actualmente solo `click`, falta `role="tablist"`+flechas)
- [ ] `prefers-reduced-motion` para deshabilitar hover transform y reveal animations

---

## 12. Resumen ejecutivo del audit

**Logrado en esta iteración:**
1. Paleta migrada de granate escolar → rojo realtor + navy
2. Tarjeta de propiedad con paridad funcional a realtor (fotos, badges, heart, features, hover)
3. Hero con barra de búsqueda tabs+input+tipo+submit
4. Filter chips de dormitorios + sort dropdown (HTMX)
5. Detalle con galería 1+4 + sticky contact card + WhatsApp CTA
6. Seeds con 6 propiedades chilenas reales (Lo Barnechea, Ñuñoa, Olmué, Providencia, Las Condes, Viña)

**Faltante crítico para paridad realtor:**
- Mapa Leaflet en detalle
- Carrusel de fotos dentro de la card
- "Similar homes" al pie del detalle
- Autocomplete de comunas
- Indicador "New" / "Price cut" / "X días en mercado"

**Faltante no crítico (diferido):**
- Home marketing de reemplazo a `index.html`
- Simulador hipotecario
- Comparador de propiedades
- Saved searches / alertas
