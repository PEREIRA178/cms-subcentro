# Task 04 — Public Pages: index, tiendas, tienda-detail, noticias, comunicados → templ

**Depends on:** Task 02 (design system + layout/base.templ)
**Estimated complexity:** alta — migrar 5 páginas + crear 2 nuevas + server-side routing para tienda-detail

---

## Estado del repositorio al inicio de este task

```
Module: cms-plazareal
layout/base.templ: existe y compila (Task 02)
web/*.html: existen como archivos estáticos servidos con c.SendFile()
Rutas actuales: /buscador-tiendas.html → SendFile, /tienda-individual.html → SendFile (con JS leyendo URL params)
Nuevo patrón: rutas sin .html extension, páginas renderizadas con templ
```

---

## Objetivo

Migrar las páginas públicas de `SendFile(.html)` a handlers Go que usan `helpers.Render(c, component)`. Crear las páginas como templ components que usan `layout.Base`. Agregar rutas limpias (sin `.html`). El detalle de tienda pasa de JS-con-URL-params a server-side rendering en `/tiendas/:slug`.

Los fragmentos HTMX (hero, tiendas grid, noticias cards, etc.) se migran en Task 05 (fragments). Esta task solo cubre las páginas completas.

---

## Archivos a crear/modificar

| Acción | Archivo |
|--------|---------|
| Crear | `internal/view/pages/public/index.templ` |
| Crear | `internal/view/pages/public/tiendas.templ` |
| Crear | `internal/view/pages/public/tienda_detail.templ` |
| Crear | `internal/view/pages/public/noticias.templ` |
| Crear | `internal/view/pages/public/comunicados.templ` |
| Crear | `internal/view/pages/public/locales.templ` (placeholder — contenido en Task 07) |
| Crear | `internal/view/pages/public/promociones.templ` (placeholder — contenido en Task 07) |
| Modificar | `internal/handlers/web/handlers.go` — reemplazar SendFile con Render(templ) |
| Modificar | `cmd/server/main.go` — actualizar rutas |

---

## Implementación

- [ ] **Step 1: Crear internal/view/pages/public/index.templ**

```bash
mkdir -p internal/view/pages/public
```

Contenido de `internal/view/pages/public/index.templ`:

```templ
package public

import "cms-plazareal/internal/view/layout"

templ Index() {
	@layout.Base("Inicio", "inicio") {
		<!-- HERO SECTION -->
		<section style="padding-top:68px">
			<div id="hero-carousel"
				hx-get="/fragments/hero"
				hx-trigger="load"
				hx-swap="innerHTML">
				<div style="min-height:600px;background:var(--ink-2)"></div>
			</div>
		</section>

		<!-- TIENDAS DESTACADAS -->
		<section class="section" style="background:var(--surface)">
			<div class="container">
				<div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:40px">
					<div>
						<span style="font-size:.68rem;font-weight:700;letter-spacing:3px;text-transform:uppercase;color:var(--red)">Directorio</span>
						<h2 class="display" style="font-size:clamp(1.8rem,4vw,2.8rem);margin-top:6px">Tiendas <em>destacadas</em></h2>
					</div>
					<a href="/buscador-tiendas" class="btn-outline" style="white-space:nowrap">
						Ver todas <span class="material-symbols-outlined" style="font-size:16px">arrow_forward</span>
					</a>
				</div>
				<div id="tiendas-destacadas"
					hx-get="/fragments/tiendas-destacadas"
					hx-trigger="load"
					hx-swap="innerHTML">
					<div style="display:flex;gap:16px">
						<div style="width:200px;height:240px;border-radius:16px;background:var(--surface-2)"></div>
						<div style="width:200px;height:240px;border-radius:16px;background:var(--surface-2)"></div>
						<div style="width:200px;height:240px;border-radius:16px;background:var(--surface-2)"></div>
					</div>
				</div>
			</div>
		</section>

		<!-- TIENDAS MARQUEE -->
		<section style="background:var(--ink-2);overflow:hidden;padding:24px 0;border-top:1px solid rgba(255,255,255,.06);border-bottom:1px solid rgba(255,255,255,.06)">
			<div id="tiendas-marquee"
				hx-get="/fragments/tiendas-marquee"
				hx-trigger="load"
				hx-swap="innerHTML">
			</div>
		</section>

		<!-- NOTICIAS Y COMUNICADOS -->
		<section class="section" style="background:var(--white)">
			<div class="container">
				<div style="margin-bottom:40px">
					<span style="font-size:.68rem;font-weight:700;letter-spacing:3px;text-transform:uppercase;color:var(--red)">Actualidad</span>
					<h2 class="display" style="font-size:clamp(1.8rem,4vw,2.8rem);margin-top:6px">Noticias y <em>comunicados</em></h2>
				</div>
				<div id="noticias-home"
					hx-get="/fragments/noticias"
					hx-trigger="load"
					hx-swap="innerHTML">
				</div>
			</div>
		</section>

		<!-- EVENTOS / PROMOCIONES -->
		<section class="section" style="background:var(--surface)">
			<div class="container">
				<div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:40px">
					<div>
						<span style="font-size:.68rem;font-weight:700;letter-spacing:3px;text-transform:uppercase;color:var(--red)">Agenda</span>
						<h2 class="display" style="font-size:clamp(1.8rem,4vw,2.8rem);margin-top:6px">Eventos y <em>promociones</em></h2>
					</div>
					<a href="/promociones" class="btn-outline" style="white-space:nowrap">
						Ver todos <span class="material-symbols-outlined" style="font-size:16px">arrow_forward</span>
					</a>
				</div>
				<div id="promos-home"
					hx-get="/fragments/promos"
					hx-trigger="load"
					hx-swap="innerHTML">
				</div>
			</div>
		</section>

		<!-- MAPA -->
		<section style="background:var(--ink-2);padding:80px 0">
			<div class="container">
				<div class="glass-card-dark" style="display:grid;grid-template-columns:1fr 1fr;overflow:hidden">
					<div style="padding:48px 44px">
						<span style="font-size:.68rem;font-weight:700;letter-spacing:3px;text-transform:uppercase;color:var(--blue)">Ubicación</span>
						<h3 class="display" style="font-size:1.8rem;color:#fff;margin-top:8px;margin-bottom:16px">Plaza Real <em>Copiapó</em></h3>
						<p style="color:rgba(255,255,255,.55);line-height:1.7;margin-bottom:24px">Av. Copayapu 441, Copiapó, Atacama.<br/>Lunes a Domingo 10:00 – 21:00</p>
						<a href="https://maps.google.com/?q=Plaza+Real+Copiapó" target="_blank" class="btn-primary">
							<span class="material-symbols-outlined" style="font-size:16px">directions</span> Cómo llegar
						</a>
					</div>
					<div style="height:360px">
						<iframe
							src="https://maps.google.com/maps?q=Plaza+Real+Copiap%C3%B3&output=embed"
							width="100%" height="100%"
							style="border:0;filter:grayscale(20%)"
							allowfullscreen=""
							loading="lazy">
						</iframe>
					</div>
				</div>
			</div>
		</section>
	}
}
```

- [ ] **Step 2: Crear internal/view/pages/public/tiendas.templ**

```templ
package public

import "cms-plazareal/internal/view/layout"

templ Tiendas() {
	@layout.Base("Tiendas", "tiendas") {
		<div class="page-hero">
			<div class="page-hero-inner">
				<span class="page-hero-eyebrow">Plaza Real · Copiapó</span>
				<h1 class="page-hero-title">Directorio de <em>Tiendas</em></h1>
				<p class="page-hero-sub">Más de 100 tiendas, restaurantes y servicios. Busca por nombre o categoría.</p>
			</div>
		</div>
		<div class="section">
			<div class="container">
				<!-- Filtros -->
				<div style="display:flex;flex-wrap:wrap;gap:12px;align-items:center;margin-bottom:28px">
					<div style="position:relative;flex:1;min-width:240px">
						<span class="material-symbols-outlined" style="position:absolute;left:14px;top:50%;transform:translateY(-50%);color:var(--muted);pointer-events:none;font-size:18px">search</span>
						<input
							id="tienda-search"
							type="text"
							placeholder="Buscar tienda..."
							style="width:100%;padding:11px 16px 11px 42px;border-radius:var(--r-full);border:1.5px solid var(--border);background:var(--white);font-family:var(--font-body);font-size:.9rem;outline:none"
							hx-get="/fragments/tiendas"
							hx-trigger="keyup changed delay:300ms"
							hx-target="#tiendas-grid"
							hx-swap="innerHTML"
							name="q"
							hx-include="[name='cat'],[name='gal']"
						/>
					</div>
					<div style="display:flex;flex-wrap:wrap;gap:8px">
						<button class="chip active" hx-get="/fragments/tiendas" hx-target="#tiendas-grid" hx-swap="innerHTML" hx-include="[name='q'],[name='gal']" name="cat" value="">Todas</button>
						<button class="chip" hx-get="/fragments/tiendas" hx-target="#tiendas-grid" hx-swap="innerHTML" hx-include="[name='q'],[name='gal']" name="cat" value="tiendas">Moda</button>
						<button class="chip" hx-get="/fragments/tiendas" hx-target="#tiendas-grid" hx-swap="innerHTML" hx-include="[name='q'],[name='gal']" name="cat" value="restaurantes">Restaurantes</button>
						<button class="chip" hx-get="/fragments/tiendas" hx-target="#tiendas-grid" hx-swap="innerHTML" hx-include="[name='q'],[name='gal']" name="cat" value="farmacias">Farmacias</button>
						<button class="chip" hx-get="/fragments/tiendas" hx-target="#tiendas-grid" hx-swap="innerHTML" hx-include="[name='q'],[name='gal']" name="cat" value="salud">Salud</button>
						<button class="chip" hx-get="/fragments/tiendas" hx-target="#tiendas-grid" hx-swap="innerHTML" hx-include="[name='q'],[name='gal']" name="cat" value="tecnologia">Tecnología</button>
						<button class="chip" hx-get="/fragments/tiendas" hx-target="#tiendas-grid" hx-swap="innerHTML" hx-include="[name='q'],[name='gal']" name="cat" value="servicios">Servicios</button>
					</div>
					<!-- Galería filter -->
					<select name="gal"
						style="padding:10px 14px;border-radius:var(--r-full);border:1.5px solid var(--border);background:var(--white);font-family:var(--font-body);font-size:.84rem;outline:none"
						hx-get="/fragments/tiendas" hx-target="#tiendas-grid" hx-swap="innerHTML" hx-include="[name='q'],[name='cat']">
						<option value="">Todas las galerías</option>
						<option value="placa-comercial">Placa Comercial</option>
						<option value="torre-flamenco">Torre Flamenco</option>
					</select>
				</div>
				<!-- Grid -->
				<div id="tiendas-grid"
					hx-get="/fragments/tiendas"
					hx-trigger="load"
					hx-swap="innerHTML">
					<div style="display:flex;align-items:center;gap:8px;color:var(--muted);padding:40px 0">
						<span class="material-symbols-outlined loading-spin">progress_activity</span>
						Cargando tiendas...
					</div>
				</div>
			</div>
		</div>
	}
}
```

- [ ] **Step 3: Crear internal/view/pages/public/tienda_detail.templ**

```templ
package public

import (
	"cms-plazareal/internal/view/layout"
	"fmt"
)

type TiendaDetailData struct {
	ID         string
	Nombre     string
	Cat        string
	Gal        string
	Local      string
	Logo       string
	Tags       []string
	Desc       string
	About      string
	About2     string
	Pay        string
	Photos     []string
	Whatsapp   string
	Telefono   string
	Rating     string
	HorarioLV  string
	HorarioSab string
	HorarioDom string
	Similar    []SimilarTienda
}

type SimilarTienda struct {
	Slug   string
	Nombre string
	Logo   string
	Cat    string
}

templ TiendaDetail(t TiendaDetailData) {
	@layout.Base(t.Nombre, "tiendas") {
		<!-- BREADCRUMB + HERO -->
		<div style="padding-top:68px;background:var(--ink-2)">
			<div class="container" style="padding-top:32px;padding-bottom:32px">
				<div style="font-size:.76rem;color:rgba(255,255,255,.45);margin-bottom:20px">
					<a href="/index.html" style="color:rgba(255,255,255,.45)">Inicio</a>
					<span style="margin:0 8px">/</span>
					<a href="/buscador-tiendas" style="color:rgba(255,255,255,.45)">Tiendas</a>
					<span style="margin:0 8px">/</span>
					<span style="color:rgba(255,255,255,.7)">{ t.Nombre }</span>
				</div>
				<div style="display:flex;align-items:flex-start;gap:28px;flex-wrap:wrap">
					if t.Logo != "" {
						<img src={ t.Logo } alt={ t.Nombre } style="width:90px;height:90px;border-radius:16px;object-fit:cover;border:1px solid rgba(255,255,255,.1)"/>
					} else {
						<div style="width:90px;height:90px;border-radius:16px;background:rgba(255,255,255,.06);display:flex;align-items:center;justify-content:center">
							<span class="material-symbols-outlined" style="font-size:36px;color:rgba(255,255,255,.3)">storefront</span>
						</div>
					}
					<div>
						<span class="badge badge-info" style="margin-bottom:10px">{ t.Cat }</span>
						<h1 class="display" style="font-size:clamp(1.8rem,4vw,3rem);color:#fff;margin-bottom:8px">{ t.Nombre }</h1>
						<p style="color:rgba(255,255,255,.55)">
							<span class="material-symbols-outlined" style="font-size:15px;vertical-align:-3px">location_on</span>
							{ t.Local }
							if t.Gal != "" {
								· { galLabel(t.Gal) }
							}
						</p>
					</div>
				</div>
			</div>
		</div>

		<div class="section">
			<div class="container">
				<div style="display:grid;grid-template-columns:2fr 1fr;gap:32px;align-items:start">
					<!-- MAIN CONTENT -->
					<div>
						if len(t.Photos) > 0 {
							<div style="margin-bottom:32px;display:grid;grid-template-columns:repeat(auto-fill,minmax(200px,1fr));gap:12px">
								for _, photo := range t.Photos {
									<img src={ photo } alt={ t.Nombre } style="border-radius:12px;height:160px;object-fit:cover;width:100%"/>
								}
							</div>
						}
						if t.About != "" {
							<div class="glass-card" style="padding:24px;margin-bottom:20px">
								<h3 style="font-size:.9rem;font-weight:700;margin-bottom:12px;color:var(--muted)">SOBRE LA TIENDA</h3>
								<p style="line-height:1.7;color:var(--ink)">{ t.About }</p>
							</div>
						}
						if t.About2 != "" {
							<div class="glass-card" style="padding:24px;margin-bottom:20px">
								<p style="line-height:1.7;color:var(--ink)">{ t.About2 }</p>
							</div>
						}
					</div>
					<!-- SIDEBAR INFO -->
					<div style="display:flex;flex-direction:column;gap:16px">
						<!-- Horarios -->
						if t.HorarioLV != "" {
							<div class="glass-card" style="padding:20px">
								<h3 style="font-size:.8rem;font-weight:700;letter-spacing:1px;text-transform:uppercase;color:var(--muted);margin-bottom:14px">Horarios</h3>
								<div style="display:flex;flex-direction:column;gap:8px;font-size:.85rem">
									<div style="display:flex;justify-content:space-between">
										<span style="color:var(--muted)">Lun–Vie</span>
										<span style="font-weight:600">{ t.HorarioLV }</span>
									</div>
									if t.HorarioSab != "" {
										<div style="display:flex;justify-content:space-between">
											<span style="color:var(--muted)">Sábado</span>
											<span style="font-weight:600">{ t.HorarioSab }</span>
										</div>
									}
									if t.HorarioDom != "" {
										<div style="display:flex;justify-content:space-between">
											<span style="color:var(--muted)">Domingo</span>
											<span style="font-weight:600">{ t.HorarioDom }</span>
										</div>
									}
								</div>
							</div>
						}
						<!-- Contacto -->
						if t.Whatsapp != "" || t.Telefono != "" {
							<div class="glass-card" style="padding:20px">
								<h3 style="font-size:.8rem;font-weight:700;letter-spacing:1px;text-transform:uppercase;color:var(--muted);margin-bottom:14px">Contacto</h3>
								if t.Whatsapp != "" {
									<a href={ templ.SafeURL("https://wa.me/" + t.Whatsapp) } target="_blank"
										class="btn-primary" style="width:100%;justify-content:center;margin-bottom:10px"
										hx-post="/api/leads" hx-vals={ fmt.Sprintf(`{"tipo":"whatsapp","tienda_id":"%s"}`, t.ID) } hx-swap="none">
										<span class="material-symbols-outlined" style="font-size:16px">chat</span>
										WhatsApp
									</a>
								}
								if t.Telefono != "" {
									<p style="font-size:.85rem;text-align:center;color:var(--muted)">{ t.Telefono }</p>
								}
							</div>
						}
						<!-- Tags -->
						if len(t.Tags) > 0 {
							<div class="glass-card" style="padding:20px">
								<h3 style="font-size:.8rem;font-weight:700;letter-spacing:1px;text-transform:uppercase;color:var(--muted);margin-bottom:12px">Tags</h3>
								<div style="display:flex;flex-wrap:wrap;gap:6px">
									for _, tag := range t.Tags {
										<span class="chip" style="font-size:.72rem;padding:5px 12px">{ tag }</span>
									}
								</div>
							</div>
						}
						<!-- Pay -->
						if t.Pay != "" {
							<div class="glass-card" style="padding:20px">
								<h3 style="font-size:.8rem;font-weight:700;letter-spacing:1px;text-transform:uppercase;color:var(--muted);margin-bottom:8px">Medios de pago</h3>
								<p style="font-size:.85rem;line-height:1.6">{ t.Pay }</p>
							</div>
						}
					</div>
				</div>

				<!-- TIENDAS SIMILARES -->
				if len(t.Similar) > 0 {
					<div style="margin-top:48px;border-top:1px solid var(--border);padding-top:40px">
						<h3 class="display" style="font-size:1.4rem;margin-bottom:24px">También te puede interesar</h3>
						<div style="display:grid;grid-template-columns:repeat(auto-fill,minmax(160px,1fr));gap:16px">
							for _, s := range t.Similar {
								<a href={ templ.SafeURL("/tiendas/" + s.Slug) } class="glass-card" style="padding:16px;text-align:center;display:flex;flex-direction:column;align-items:center;gap:10px;transition:transform .18s" onmouseover="this.style.transform='translateY(-3px)'" onmouseout="this.style.transform=''">
									if s.Logo != "" {
										<img src={ s.Logo } alt={ s.Nombre } style="width:56px;height:56px;border-radius:10px;object-fit:cover"/>
									} else {
										<div style="width:56px;height:56px;border-radius:10px;background:var(--surface-2);display:flex;align-items:center;justify-content:center">
											<span class="material-symbols-outlined" style="color:var(--muted)">storefront</span>
										</div>
									}
									<p style="font-size:.82rem;font-weight:600">{ s.Nombre }</p>
								</a>
							}
						</div>
					</div>
				}
			</div>
		</div>
	}
}

func galLabel(gal string) string {
	switch gal {
	case "placa-comercial":
		return "Placa Comercial"
	case "torre-flamenco":
		return "Torre Flamenco"
	}
	return gal
}
```

- [ ] **Step 4: Crear internal/view/pages/public/noticias.templ**

```templ
package public

import "cms-plazareal/internal/view/layout"

templ Noticias() {
	@layout.Base("Noticias", "noticias") {
		<div class="page-hero">
			<div class="page-hero-inner">
				<span class="page-hero-eyebrow">Plaza Real · Copiapó</span>
				<h1 class="page-hero-title">Noticias y <em>Prensa</em></h1>
				<p class="page-hero-sub">Actualidad, novedades y cobertura de prensa del centro comercial Plaza Real.</p>
			</div>
		</div>
		<div class="section">
			<div class="container">
				<div id="noticias-grid"
					hx-get="/fragments/noticias-page"
					hx-trigger="load"
					hx-swap="innerHTML">
					<div style="display:flex;align-items:center;gap:8px;color:var(--muted);padding:40px 0">
						<span class="material-symbols-outlined loading-spin">progress_activity</span>
						Cargando noticias...
					</div>
				</div>
			</div>
		</div>
	}
}

templ Comunicados() {
	@layout.Base("Comunicados", "comunicados") {
		<div class="page-hero">
			<div class="page-hero-inner">
				<span class="page-hero-eyebrow">Plaza Real · Comunidad</span>
				<h1 class="page-hero-title">Comunicados <em>Oficiales</em></h1>
				<p class="page-hero-sub">Información relevante para la comunidad de locatarios y visitantes.</p>
			</div>
		</div>
		<div class="section">
			<div class="container">
				<div id="comunicados-grid"
					hx-get="/fragments/comunicados-page"
					hx-trigger="load"
					hx-swap="innerHTML">
					<div style="display:flex;align-items:center;gap:8px;color:var(--muted);padding:40px 0">
						<span class="material-symbols-outlined loading-spin">progress_activity</span>
						Cargando comunicados...
					</div>
				</div>
			</div>
		</div>
	}
}
```

- [ ] **Step 5: Crear placeholders para páginas que se completarán en Task 07**

`internal/view/pages/public/locales.templ`:
```templ
package public

import "cms-plazareal/internal/view/layout"

templ Locales() {
	@layout.Base("Locales Disponibles", "locales") {
		<div class="page-hero">
			<div class="page-hero-inner">
				<span class="page-hero-eyebrow">Plaza Real · Arriendo</span>
				<h1 class="page-hero-title">Locales <em>Disponibles</em></h1>
				<p class="page-hero-sub">Espacios comerciales disponibles para arriendo en Plaza Real.</p>
			</div>
		</div>
		<div class="section">
			<div class="container">
				<div id="locales-grid"
					hx-get="/fragments/locales-disponibles"
					hx-trigger="load"
					hx-swap="innerHTML">
					<div style="display:flex;align-items:center;gap:8px;color:var(--muted);padding:40px 0">
						<span class="material-symbols-outlined loading-spin">progress_activity</span>
						Cargando locales...
					</div>
				</div>
			</div>
		</div>
	}
}

templ Promociones() {
	@layout.Base("Eventos y Promociones", "promociones") {
		<div class="page-hero">
			<div class="page-hero-inner">
				<span class="page-hero-eyebrow">Plaza Real · Agenda</span>
				<h1 class="page-hero-title">Eventos y <em>Promociones</em></h1>
				<p class="page-hero-sub">Descubre eventos, ofertas y campañas especiales del mall y sus tiendas.</p>
			</div>
		</div>
		<div class="section">
			<div class="container">
				<div id="promos-grid"
					hx-get="/fragments/promos-page"
					hx-trigger="load"
					hx-swap="innerHTML">
					<div style="display:flex;align-items:center;gap:8px;color:var(--muted);padding:40px 0">
						<span class="material-symbols-outlined loading-spin">progress_activity</span>
						Cargando promociones...
					</div>
				</div>
			</div>
		</div>
		<!-- Modal reserva -->
		<dialog id="reserva-modal" style="border:none;border-radius:20px;padding:0;width:min(540px,95vw);background:var(--white);box-shadow:0 24px 80px rgba(0,0,0,.28)">
			<div id="reserva-modal-inner" style="padding:32px"></div>
		</dialog>
		<script>
			document.getElementById('reserva-modal').addEventListener('click', function(e) {
				if (e.target === this) this.close();
			});
		</script>
	}
}
```

- [ ] **Step 6: Actualizar internal/handlers/web/handlers.go**

Reemplazar el package web con la nueva implementación que usa `helpers.Render`:

```go
package web

import (
	"strings"

	"cms-plazareal/internal/config"
	"cms-plazareal/internal/helpers"
	"cms-plazareal/internal/view/pages/public"

	"github.com/gofiber/fiber/v2"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func IndexHandler(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return helpers.Render(c, public.Index())
	}
}

func TiendasPageHandler(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return helpers.Render(c, public.Tiendas())
	}
}

func TiendaDetailHandler(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		slug := c.Params("slug")
		if slug == "" {
			return c.Redirect("/buscador-tiendas")
		}

		records, err := pb.FindRecordsByFilter("tiendas", "slug = '"+sanitizeSlug(slug)+"'", "", 1, 0)
		if err != nil || len(records) == 0 {
			return c.Status(404).SendString("Tienda no encontrada")
		}
		r := records[0]

		data := public.TiendaDetailData{
			ID:         r.Id,
			Nombre:     r.GetString("nombre"),
			Cat:        r.GetString("cat"),
			Gal:        r.GetString("gal"),
			Local:      r.GetString("local"),
			Logo:       r.GetString("logo"),
			Tags:       splitCSV(r.GetString("tags")),
			Desc:       r.GetString("desc"),
			About:      r.GetString("about"),
			About2:     r.GetString("about2"),
			Pay:        r.GetString("pay"),
			Photos:     splitCSV(r.GetString("photos")),
			Whatsapp:   r.GetString("whatsapp"),
			Telefono:   r.GetString("telefono"),
			Rating:     r.GetString("rating"),
			HorarioLV:  r.GetString("horario_lv"),
			HorarioSab: r.GetString("horario_sab"),
			HorarioDom: r.GetString("horario_dom"),
		}

		// Load similar tiendas
		simSlugs := splitCSV(r.GetString("similar"))
		for _, s := range simSlugs {
			if s == "" {
				continue
			}
			simRecs, err := pb.FindRecordsByFilter("tiendas", "slug = '"+sanitizeSlug(s)+"'", "", 1, 0)
			if err != nil || len(simRecs) == 0 {
				continue
			}
			sr := simRecs[0]
			data.Similar = append(data.Similar, public.SimilarTienda{
				Slug:   sr.GetString("slug"),
				Nombre: sr.GetString("nombre"),
				Logo:   sr.GetString("logo"),
				Cat:    sr.GetString("cat"),
			})
		}

		return helpers.Render(c, public.TiendaDetail(data))
	}
}

func NoticiasPageHandler(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return helpers.Render(c, public.Noticias())
	}
}

func ComunicadosPageHandler(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return helpers.Render(c, public.Comunicados())
	}
}

func LocalesPageHandler(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return helpers.Render(c, public.Locales())
	}
}

func PromocionesPageHandler(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return helpers.Render(c, public.Promociones())
	}
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	var result []string
	for _, v := range strings.Split(s, ",") {
		if t := strings.TrimSpace(v); t != "" {
			result = append(result, t)
		}
	}
	return result
}

func sanitizeSlug(s string) string {
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "\\", "")
	return s
}

// DeviceDisplay sirve el modo kiosk para un dispositivo horizontal/screen
func DeviceDisplay(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		code := c.Params("code")
		if code == "" {
			return c.Status(404).SendString("Código de dispositivo inválido")
		}
		return c.SendFile("./internal/templates/devices/display.html")
	}
}

// TotemDisplay sirve el modo kiosk para un totem vertical
func TotemDisplay(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendFile("./internal/templates/devices/totem.html")
	}
}
```

**Nota:** Si hay funciones adicionales en el `web/handlers.go` actual que no están arriba (ej. `NoticiaIndividual`, `PropiedadIndividual`), eliminar las de propiedades y mantener las de noticias si son necesarias.

- [ ] **Step 7: Actualizar rutas en cmd/server/main.go**

Buscar las rutas públicas actuales:
```bash
grep -n "propiedades\|noticias\|tienda-individual\|buscador\|SendFile" cmd/server/main.go | head -20
```

Reemplazar las rutas públicas:
```go
// RUTAS PÚBLICAS — reemplazar las existentes con:
app.Get("/", web.IndexHandler(cfg))
app.Get("/index.html", web.IndexHandler(cfg))

app.Get("/buscador-tiendas", web.TiendasPageHandler(cfg))
app.Get("/buscador-tiendas.html", web.TiendasPageHandler(cfg)) // backward compat

app.Get("/tiendas/:slug", web.TiendaDetailHandler(cfg, pb))

app.Get("/noticias", web.NoticiasPageHandler(cfg))
app.Get("/noticias.html", web.NoticiasPageHandler(cfg)) // backward compat

app.Get("/comunicados", web.ComunicadosPageHandler(cfg))

app.Get("/locales", web.LocalesPageHandler(cfg))
app.Get("/locales.html", web.LocalesPageHandler(cfg)) // backward compat

app.Get("/promociones", web.PromocionesPageHandler(cfg))
app.Get("/eventos.html", web.PromocionesPageHandler(cfg)) // backward compat

// Eliminar estas rutas (propiedades):
// app.Get("/propiedades.html", ...)
// app.Get("/tienda-individual.html", ...) ← reemplazado por /tiendas/:slug
```

- [ ] **Step 8: templ generate + compilar**

```bash
make generate
go build ./...
```

Si hay errores de tipos no encontrados, verificar que los structs `TiendaDetailData` y `SimilarTienda` están en el mismo package que la función que los usa.

- [ ] **Step 9: Test manual básico**

```bash
go run cmd/server/main.go
```

Verificar en browser:
- `http://localhost:3000/` → renderiza Index con hero placeholder y secciones
- `http://localhost:3000/buscador-tiendas` → renderiza Tiendas con buscador
- `http://localhost:3000/noticias` → renderiza Noticias
- `http://localhost:3000/tiendas/mcdonalds` (si existe slug) → renderiza TiendaDetail
- Sin errores 500 en consola

- [ ] **Step 10: Commit**

```bash
git add internal/view/pages/public/ internal/handlers/web/handlers.go cmd/server/main.go
git commit -m "feat: migrate public pages to templ — index, tiendas, tienda-detail SSR, noticias, comunicados, locales, promociones"
```

---

## Estado del repositorio al finalizar este task

- Todas las páginas públicas renderizadas con templ usando `layout.Base`
- Ruta server-side `/tiendas/:slug` con datos reales de PocketBase
- Rutas limpias sin `.html` (con backward-compat para las .html existentes)
- `web/propiedades.html` y rutas de propiedades eliminadas
- `go build ./...` pasa sin errores

---

## Amendments (agregados después del diseño inicial)

### A1: Hero background por tienda en TiendaDetailData

Agregar campo `HeroBg string` al struct `TiendaDetailData`. El handler lo lee de `r.GetString("hero_bg")`.

La sección hero de `tienda_detail.templ` debe usar el valor:
```templ
<section class="tienda-hero"
  if d.HeroBg != "" {
    style={ "background-image: url(" + d.HeroBg + ")" }
  }
>
```

Sin `HeroBg`, el hero queda con `background: #0d0d0d` (fondo negro — comportamiento actual).

### A2: Chip de horario conectado a horas reales

Eliminar el chip estático `"Abierto ahora"`. En su lugar, calcular el estado en el handler server-side:

```go
// En handlers/web/handlers.go — dentro de TiendaDetailHandler:
isOpen, statusLabel := computeOpenStatus(
    record.GetString("horario_lv"),
    record.GetString("horario_sab"),
    record.GetString("horario_dom"),
    record.GetString("status_horario"), // campo extra: "normal" | "solo-reserva" | "cerrado-temporal"
)
```

Función auxiliar (agregar en `handlers/web/handlers.go`):

```go
import "time"

// computeOpenStatus retorna (isOpen bool, label string) según el día/hora actual.
// horarioStr tiene formato "HH:MM - HH:MM"; string vacío = sin horario declarado.
// statusOverride: "solo-reserva" | "cerrado-temporal" sobreescriben el cálculo.
func computeOpenStatus(lv, sab, dom, statusOverride string) (bool, string) {
    if statusOverride == "cerrado-temporal" {
        return false, "Cerrado temporalmente"
    }
    if statusOverride == "solo-reserva" {
        return false, "Solo con reserva"
    }

    now := time.Now()
    var horario string
    switch now.Weekday() {
    case time.Saturday:
        horario = sab
    case time.Sunday:
        horario = dom
    default:
        horario = lv
    }

    if horario == "" {
        return false, "Sin horario"
    }

    parts := strings.SplitN(horario, " - ", 2)
    if len(parts) != 2 {
        return false, "Sin horario"
    }

    parseHM := func(s string) (int, int, bool) {
        var h, m int
        _, err := fmt.Sscanf(strings.TrimSpace(s), "%d:%02d", &h, &m)
        return h, m, err == nil
    }

    oh, om, ok1 := parseHM(parts[0])
    ch, cm, ok2 := parseHM(parts[1])
    if !ok1 || !ok2 {
        return false, "Sin horario"
    }

    nowMin := now.Hour()*60 + now.Minute()
    openMin := oh*60 + om
    closeMin := ch*60 + cm

    if nowMin >= openMin && nowMin < closeMin {
        return true, "Abierto ahora"
    }
    return false, "Cerrado"
}
```

`TiendaDetailData` añadir:
```go
IsOpen      bool
StatusLabel string // "Abierto ahora" | "Cerrado" | "Solo con reserva" | "Cerrado temporalmente"
```

En el templ, reemplazar el chip estático por:
```templ
<span class={
    "status-chip",
    templ.KV("status-open", d.IsOpen),
    templ.KV("status-closed", !d.IsOpen),
}>
    { d.StatusLabel }
</span>
```

CSS en `app.css`:
```css
.status-open  { background: rgba(44,190,80,0.18); color: #2cbe50; border: 1px solid rgba(44,190,80,0.35); }
.status-closed { background: rgba(215,16,85,0.15); color: #d71055; border: 1px solid rgba(215,16,85,0.30); }
```

También agregar `status_horario` como campo `select` en la colección `tiendas` (en `collections.go`):
```go
&core.SelectField{Name: "status_horario", Values: []string{"normal", "solo-reserva", "cerrado-temporal"}},
```

Y agregarlo al form de tienda en **Task 05** (ver Amendment A3 en task-05).

### A3: Eliminar valoraciones Google

Eliminar de `TiendaDetailData` los campos `Rating string` y `RatingCount string` (o equivalentes).

Eliminar del templ el bloque que renderiza estrellas y número de valoraciones. Buscar con:
```bash
grep -n "rating\|Rating\|star\|Star\|valorac" internal/view/pages/public/tienda_detail.templ
```

Eliminar esas líneas del templ. No reemplazar por nada — simplemente no mostrar valoraciones.
