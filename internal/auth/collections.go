package auth

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

// RegisterPBHooks sets up PocketBase collections and auth hooks.
func RegisterPBHooks(pb *pocketbase.PocketBase) {
	pb.OnServe().BindFunc(func(se *core.ServeEvent) error {
		log.Println("📦 PocketBase: Verificando colecciones...")
		if err := ensureCollections(se.App); err != nil {
			log.Printf("⚠️  Error creando colecciones: %v", err)
		}
		return se.Next()
	})
}

func ensureCollections(app core.App) error {
	// ── 1. USERS ──
	if _, err := app.FindCollectionByNameOrId("users"); err != nil {
		col := core.NewAuthCollection("users")
		col.Fields.Add(
			&core.TextField{Name: "role", Required: true},
			&core.TextField{Name: "nombre"},
			&core.TextField{Name: "telefono"},
			&core.TextField{Name: "rut"},
			&core.BoolField{Name: "activo"},
		)
		col.AuthToken.Duration = 259200
		col.PasswordAuth.Enabled = true
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'users' created")
	}

	// ── 2. MEDIA (biblioteca multimedia → R2) ──
	if _, err := app.FindCollectionByNameOrId("media"); err != nil {
		col := core.NewBaseCollection("media")
		col.Fields.Add(
			&core.TextField{Name: "filename", Required: true},
			&core.URLField{Name: "url_r2"},
			&core.TextField{Name: "type", Required: true}, // imagen|video|youtube|vimeo
			&core.NumberField{Name: "size"},
			&core.TextField{Name: "uploaded_by"},
			&core.TextField{Name: "status"}, // borrador|publicado|archivado
			&core.TextField{Name: "description"},
			&core.NumberField{Name: "duration_seconds"},
			&core.FileField{Name: "thumbnail", MaxSelect: 1, MaxSize: 5242880},
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'media' created")
	}

	// ── 3. CONTENT_BLOCKS (unified: eventos, noticias, comunicados) ──
	if _, err := app.FindCollectionByNameOrId("content_blocks"); err != nil {
		col := core.NewBaseCollection("content_blocks")
		col.Fields.Add(
			&core.TextField{Name: "title", Required: true},
			&core.EditorField{Name: "description"},
			// EMERGENCIA|REUNIÓN|INFORMACIÓN|ACADÉMICO|EVENTO|DEPORTIVO|NOTICIA
			&core.TextField{Name: "category"},
			&core.BoolField{Name: "urgency"},
			&core.DateField{Name: "date"},
			&core.BoolField{Name: "featured"},
			&core.TextField{Name: "status"}, // borrador|publicado|archivado
			&core.TextField{Name: "media_ids"}, // comma-separated media IDs
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'content_blocks' created")
	}

	// ── 4. MULTIMEDIA (legacy, kept for playlists) ──
	if _, err := app.FindCollectionByNameOrId("multimedia"); err != nil {
		col := core.NewBaseCollection("multimedia")
		col.Fields.Add(
			&core.TextField{Name: "filename", Required: true},
			&core.URLField{Name: "url_r2"},
			&core.TextField{Name: "type", Required: true},
			&core.NumberField{Name: "size"},
			&core.TextField{Name: "uploaded_by"},
			&core.TextField{Name: "estado"},
			&core.TextField{Name: "descripcion"},
			&core.NumberField{Name: "duracion_segundos"},
			&core.FileField{Name: "thumbnail", MaxSelect: 1, MaxSize: 5242880},
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'multimedia' created")
	}

	// ── 5. PLAYLISTS ──
	if _, err := app.FindCollectionByNameOrId("playlists"); err != nil {
		col := core.NewBaseCollection("playlists")
		col.Fields.Add(
			&core.TextField{Name: "name", Required: true},
			&core.TextField{Name: "description"},
			&core.TextField{Name: "status"},
			&core.TextField{Name: "created_by"},
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'playlists' created")
	}

	// ── 6. PLAYLIST_ITEMS ──
	if _, err := app.FindCollectionByNameOrId("playlist_items"); err != nil {
		col := core.NewBaseCollection("playlist_items")
		col.Fields.Add(
			&core.TextField{Name: "playlist_id", Required: true},
			&core.TextField{Name: "multimedia_id"},
			&core.TextField{Name: "tipo"},
			&core.NumberField{Name: "orden", Required: true},
			&core.NumberField{Name: "duracion_segundos"},
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'playlist_items' created")
	}

	// ── 7. DEVICES ──
	if _, err := app.FindCollectionByNameOrId("devices"); err != nil {
		col := core.NewBaseCollection("devices")
		col.Fields.Add(
			&core.TextField{Name: "name", Required: true},
			&core.TextField{Name: "type", Required: true},
			&core.TextField{Name: "code", Required: true},
			&core.TextField{Name: "layout"},
			&core.TextField{Name: "ubicacion"},
			&core.TextField{Name: "playlist_id"},
			&core.TextField{Name: "status"},
			&core.DateField{Name: "last_seen"},
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'devices' created")
	}

	// ── 8. FORM_RESPONSES ──
	if _, err := app.FindCollectionByNameOrId("form_responses"); err != nil {
		col := core.NewBaseCollection("form_responses")
		col.Fields.Add(
			&core.TextField{Name: "event_id"},
			&core.TextField{Name: "user_id"},
			&core.TextField{Name: "tipo"},
			&core.TextField{Name: "valor"},
			&core.TextField{Name: "mensaje"},
			&core.BoolField{Name: "leido"},
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'form_responses' created")
	}

	// ── 9. WHATSAPP_LOGS ──
	if _, err := app.FindCollectionByNameOrId("whatsapp_logs"); err != nil {
		col := core.NewBaseCollection("whatsapp_logs")
		col.Fields.Add(
			&core.TextField{Name: "event_id"},
			&core.TextField{Name: "phone"},
			&core.TextField{Name: "message_sid"},
			&core.TextField{Name: "status"},
			&core.TextField{Name: "direction"},
			&core.TextField{Name: "body"},
			&core.TextField{Name: "error_message"},
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'whatsapp_logs' created")
	}

	// ── 10. TIENDAS (store directory — Subcentro Las Condes) ──
	if _, err := app.FindCollectionByNameOrId("tiendas"); err != nil {
		col := core.NewBaseCollection("tiendas")
		col.Fields.Add(
			&core.TextField{Name: "nombre", Required: true},
			&core.TextField{Name: "slug"},
			// tiendas|restaurantes|farmacias|salud|tecnologia|servicios
			&core.TextField{Name: "cat", Required: true},
			// norte|sur
			&core.TextField{Name: "gal"},
			&core.TextField{Name: "local"},
			&core.TextField{Name: "logo"},
			&core.TextField{Name: "tags"},          // comma-separated
			&core.TextField{Name: "desc"},           // short description (hero)
			&core.EditorField{Name: "about"},        // about paragraph 1
			&core.TextField{Name: "about2"},         // about paragraph 2
			&core.TextField{Name: "pay"},            // medios de pago
			&core.TextField{Name: "photos"},         // comma-separated URLs (min 4)
			&core.TextField{Name: "similar"},        // comma-separated slugs
			&core.TextField{Name: "whatsapp"},
			&core.TextField{Name: "telefono"},
			&core.TextField{Name: "rating"},         // e.g. "4.7"
			&core.TextField{Name: "horario_lv"},     // Lun–Vie
			&core.TextField{Name: "horario_sab"},    // Sábado
			&core.TextField{Name: "horario_dom"},    // Domingo
			// borrador|publicado
			&core.TextField{Name: "status"},
			&core.BoolField{Name: "destacada"},
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'tiendas' created")
	}

	// ── 11. PROPIEDADES (real estate listings — JCP Gestión Inmobiliaria) ──
	if _, err := app.FindCollectionByNameOrId("propiedades"); err != nil {
		col := core.NewBaseCollection("propiedades")
		col.Fields.Add(
			&core.TextField{Name: "titulo", Required: true},
			&core.TextField{Name: "slug"},
			&core.EditorField{Name: "descripcion"},
			// VENTA | ARRIENDO
			&core.TextField{Name: "operacion", Required: true},
			// CASA | DEPARTAMENTO | TERRENO | PARCELA | LOCAL | OFICINA | BODEGA
			&core.TextField{Name: "tipo", Required: true},
			&core.TextField{Name: "direccion"},
			&core.TextField{Name: "comuna"},
			&core.TextField{Name: "region"},
			&core.NumberField{Name: "precio_uf"},
			&core.NumberField{Name: "precio_clp"},
			&core.NumberField{Name: "dormitorios"},
			&core.NumberField{Name: "banos"},
			&core.NumberField{Name: "estacionamientos"},
			&core.NumberField{Name: "superficie_util"},    // m²
			&core.NumberField{Name: "superficie_total"},   // m²
			&core.NumberField{Name: "ano_construccion"},
			&core.TextField{Name: "estado_propiedad"},     // nueva|usada|a_estrenar
			&core.TextField{Name: "amenidades"},           // comma-separated
			// borrador|publicado|reservada|vendida|arrendada
			&core.TextField{Name: "status"},
			&core.BoolField{Name: "destacada"},
			&core.BoolField{Name: "oportunidad"},
			&core.TextField{Name: "cover_image"},          // URL (R2)
			&core.TextField{Name: "gallery"},              // comma-separated URLs
			&core.TextField{Name: "tour_url"},             // 360°/video
			&core.NumberField{Name: "lat"},
			&core.NumberField{Name: "lng"},
			&core.DateField{Name: "publicado_en"},
			&core.TextField{Name: "corredor_id"},
			&core.TextField{Name: "contacto_whatsapp"},
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'propiedades' created")
	}

	// ── 11. Default superadmin ──
	users, _ := app.FindCollectionByNameOrId("users")
	if users != nil {
		records, err := app.FindRecordsByFilter(users, "role = 'superadmin'", "", 1, 0)
		if err != nil || len(records) == 0 {
			record := core.NewRecord(users)
			record.Set("email", "admin@jcp-gestioninmobiliaria.cl")
			record.Set("password", "jcp2026admin!")
			record.Set("passwordConfirm", "jcp2026admin!")
			record.Set("nombre", "Administrador")
			record.Set("role", "superadmin")
			record.Set("activo", true)
			record.Set("verified", true)
			if err := app.Save(record); err != nil {
				log.Printf("⚠️  Error creating superadmin: %v", err)
			} else {
				log.Println("  ✅ Default superadmin created")
			}
		}
	}

	// ── Seed tiendas demo ──
	if err := seedTiendas(app); err != nil {
		log.Printf("⚠️  Error seeding tiendas: %v", err)
	}

	// ── Seed propiedades demo ──
	if err := seedPropiedades(app); err != nil {
		log.Printf("⚠️  Error seeding propiedades: %v", err)
	}

	// ── 11. Seed content_blocks ──
	if err := seedContentBlocks(app); err != nil {
		log.Printf("⚠️  Error seeding content_blocks: %v", err)
	}

	// ── 12. Migrate content_blocks — add new fields if missing ──
	migrateContentBlocks(app)

	// ── 13. Migrate urgency: set urgency=true for all EMERGENCIA records ──
	migrateUrgencyFromCategory(app)

	// ── 14. Migrate content_blocks: add template field ──
	migrateContentBlocksTemplate(app)

	// ── 15. Migrate devices: add current_view field ──
	migrateDevicesCurrentView(app)

	// ── 16. Migrate playlist_items: add content_block_id field ──
	migratePlaylistItemsContentBlockID(app)

	// ── 17. Migrate multimedia: add start_time field ──
	migrateMultimediaStartTime(app)

	// ── 18. Seed devices, playlists and playlist items ──
	if err := SeedDevicesAndPlaylists(app); err != nil {
		log.Printf("⚠️  Error seeding devices/playlists: %v", err)
	}

	return nil
}

// migrateUrgencyFromCategory sets urgency=true for all EMERGENCIA records (idempotent).
func migrateUrgencyFromCategory(app core.App) {
	records, err := app.FindRecordsByFilter("content_blocks",
		"category = 'EMERGENCIA' && urgency = false", "", 1000, 0)
	if err != nil || len(records) == 0 {
		return
	}
	for _, r := range records {
		r.Set("urgency", true)
		if err := app.Save(r); err != nil {
			log.Printf("⚠️  urgency migration error for %s: %v", r.Id, err)
		}
	}
	log.Printf("  ✅ Migrated urgency=true for %d EMERGENCIA records", len(records))
}

// migrateContentBlocks adds fields introduced after initial collection creation.
func migrateContentBlocks(app core.App) {
	col, err := app.FindCollectionByNameOrId("content_blocks")
	if err != nil {
		return
	}
	changed := false
	for _, name := range []string{"pdf_url", "image_url", "body"} {
		if col.Fields.GetByName(name) == nil {
			col.Fields.Add(&core.TextField{Name: name})
			changed = true
			log.Printf("  ✅ content_blocks: added field '%s'", name)
		}
	}
	if changed {
		if err := app.Save(col); err != nil {
			log.Printf("⚠️  content_blocks migration error: %v", err)
		}
	}
}

type seedBlock struct {
	title       string
	description string
	category    string
	urgency     bool
	date        string
	featured    bool
}

func seedContentBlocks(app core.App) error {
	col, err := app.FindCollectionByNameOrId("content_blocks")
	if err != nil {
		return err
	}

	existing, _ := app.FindRecordsByFilter(col, "status = 'publicado'", "", 1, 0)
	if len(existing) > 0 {
		return nil // already seeded
	}

	events := []seedBlock{
		{
			title:       "⚠ Simulacro de Evacuación — Jueves 2 de abril",
			description: "Recordamos a toda la comunidad escolar que el jueves 2 de abril se realizará el simulacro de evacuación obligatorio a las 10:00 horas. Participación de todos los cursos.",
			category:    "EMERGENCIA",
			urgency:     true,
			date:        "2026-04-02 10:00:00",
			featured:    true,
		},
		{
			title:       "Reunión de Apoderados 7° Básico",
			description: "Se cita a los apoderados de 7° año básico a reunión del primer trimestre 2026. La reunión se realizará el 17 de abril a las 18:30 hrs en la sala del curso. Asistencia obligatoria.",
			category:    "REUNIÓN",
			urgency:     false,
			date:        "2026-04-17 18:30:00",
			featured:    true,
		},
		{
			title:       "Campeonato de Tenis Padre-Hijo",
			description: "Inscripciones abiertas para el campeonato de tenis padre-hijo, 3 de abril de 2026. Una iniciativa del Área Deportiva EDEX que une a las familias. Inscripciones en secretaría.",
			category:    "EVENTO",
			urgency:     false,
			date:        "2026-04-03 09:00:00",
			featured:    true,
		},
		{
			title:       "Inicio año escolar 2026 — Nuevas iniciativas pedagógicas",
			description: "El Colegio San Lorenzo inicia el año escolar 2026 con importantes cambios en su propuesta pedagógica, incorporando metodologías activas, trabajo por proyectos y herramientas tecnológicas en el aula.",
			category:    "ACADÉMICO",
			urgency:     false,
			date:        "2026-04-01 08:00:00",
			featured:    false,
		},
		{
			title:       "Sistema Digital Wellness — Comunicación por WhatsApp",
			description: "El colegio informa que toda la comunicación oficial con apoderados se realizará a través del sistema Digital Wellness. Los avisos llegan directamente por WhatsApp. Consulta en secretaría.",
			category:    "INFORMACIÓN",
			urgency:     false,
			date:        "2026-03-15 00:00:00",
			featured:    false,
		},
		{
			title:       "Reunión General Enseñanza Media — 6 de mayo",
			description: "Se cita a apoderados de 1° a 4° Medio a reunión general informativa del primer trimestre 2026. Información sobre evaluaciones integradoras y proceso PAES 2026. Miércoles 6 de mayo a las 19:00 hrs en el gimnasio.",
			category:    "REUNIÓN",
			urgency:     false,
			date:        "2026-05-06 19:00:00",
			featured:    false,
		},
		{
			title:       "Calendario de pruebas primer trimestre 2026 — CEAL",
			description: "Se informa a los apoderados que el calendario de pruebas del primer trimestre 2026 está disponible en la sección CEAL del sitio web. Incluye fechas de pruebas, integradoras y exámenes para todos los niveles.",
			category:    "ACADÉMICO",
			urgency:     false,
			date:        "2026-03-05 00:00:00",
			featured:    false,
		},
		{
			title:       "Lista de útiles escolares 2026 disponible en CEPAD",
			description: "Las listas de útiles escolares para todos los niveles del año 2026 ya están disponibles en la sección CEPAD del sitio web del colegio. Descarga la lista correspondiente al nivel de tu hijo/a.",
			category:    "INFORMACIÓN",
			urgency:     false,
			date:        "2026-03-01 00:00:00",
			featured:    false,
		},
	}

	news := []seedBlock{
		{
			title:       "Resultados SIMCE 2025 — Colegio San Lorenzo entre los mejores de Atacama",
			description: "El Colegio San Lorenzo obtuvo resultados destacados en las pruebas SIMCE de 4° y 8° básico, posicionándose entre los establecimientos de mejor rendimiento en la Región de Atacama.",
			category:    "NOTICIA",
			urgency:     false,
			date:        "2026-03-28 00:00:00",
			featured:    true,
		},
		{
			title:       "Equipo sub-14 clasifica al Campeonato Regional de Fútbol",
			description: "Nuestro equipo de fútbol sub-14 representará a Atacama en el campeonato regional 2026 tras ganar la etapa comunal con resultados históricos.",
			category:    "NOTICIA",
			urgency:     false,
			date:        "2026-03-25 00:00:00",
			featured:    true,
		},
		{
			title:       "Festival de Arte EDEX 2026 — Más de 200 estudiantes en escena",
			description: "Más de 200 estudiantes participaron en la muestra artística anual del programa EDEX, mostrando sus talentos en música, danza, teatro y artes visuales.",
			category:    "NOTICIA",
			urgency:     false,
			date:        "2026-03-20 00:00:00",
			featured:    false,
		},
		{
			title:       "Nuevo laboratorio de ciencias equipado con tecnología 2026",
			description: "El Colegio San Lorenzo inaugura su nuevo laboratorio de ciencias con equipamiento moderno, beneficiando a más de 400 estudiantes de enseñanza media.",
			category:    "NOTICIA",
			urgency:     false,
			date:        "2026-03-10 00:00:00",
			featured:    false,
		},
	}

	all := append(events, news...)
	for _, b := range all {
		r := core.NewRecord(col)
		r.Set("title", b.title)
		r.Set("description", b.description)
		r.Set("category", b.category)
		r.Set("urgency", b.urgency)
		r.Set("date", b.date)
		r.Set("featured", b.featured)
		r.Set("status", "publicado")
		if err := app.Save(r); err != nil {
			log.Printf("⚠️  seed block error: %v", err)
		}
	}
	log.Printf("  ✅ Seeded %d content_blocks", len(all))
	return nil
}

var _ = types.DateTime{}

// ── New migrations ─────────────────────────────────────────────────────────────

// migrateContentBlocksTemplate adds the 'template' field to content_blocks,
// enabling multiple slide layouts (e.g. "hero-classic", "hero-full-video", "alert-emergencia").
func migrateContentBlocksTemplate(app core.App) {
	col, err := app.FindCollectionByNameOrId("content_blocks")
	if err != nil || col.Fields.GetByName("template") != nil {
		return
	}
	col.Fields.Add(&core.TextField{Name: "template"})
	if err := app.Save(col); err != nil {
		log.Printf("⚠️  content_blocks template migration: %v", err)
		return
	}
	log.Println("  ✅ content_blocks: added field 'template'")
}

// migrateDevicesCurrentView adds the 'current_view' field to devices
// so the carousel can record which slide each device is displaying.
func migrateDevicesCurrentView(app core.App) {
	col, err := app.FindCollectionByNameOrId("devices")
	if err != nil || col.Fields.GetByName("current_view") != nil {
		return
	}
	col.Fields.Add(&core.TextField{Name: "current_view"})
	if err := app.Save(col); err != nil {
		log.Printf("⚠️  devices current_view migration: %v", err)
		return
	}
	log.Println("  ✅ devices: added field 'current_view'")
}

// migratePlaylistItemsContentBlockID adds 'content_block_id' to playlist_items
// so items of tipo="content_block" can reference a content_blocks record.
func migratePlaylistItemsContentBlockID(app core.App) {
	col, err := app.FindCollectionByNameOrId("playlist_items")
	if err != nil || col.Fields.GetByName("content_block_id") != nil {
		return
	}
	col.Fields.Add(&core.TextField{Name: "content_block_id"})
	if err := app.Save(col); err != nil {
		log.Printf("⚠️  playlist_items content_block_id migration: %v", err)
		return
	}
	log.Println("  ✅ playlist_items: added field 'content_block_id'")
}

// migrateMultimediaStartTime adds 'start_time' (seconds) to multimedia
// so video items can carry a seek position independent of the URL fragment.
func migrateMultimediaStartTime(app core.App) {
	col, err := app.FindCollectionByNameOrId("multimedia")
	if err != nil || col.Fields.GetByName("start_time") != nil {
		return
	}
	col.Fields.Add(&core.NumberField{Name: "start_time"})
	if err := app.Save(col); err != nil {
		log.Printf("⚠️  multimedia start_time migration: %v", err)
		return
	}
	log.Println("  ✅ multimedia: added field 'start_time'")
}

// ── Device & playlist seeder ───────────────────────────────────────────────────

// SeedDevicesAndPlaylists is idempotent: it skips entirely if a web_hero device
// already exists. On first run it creates:
//   - 1 web_hero device  ("Web Hero - Landing Pública")
//   - 2 vertical totems  (T001, T002)
//   - 3 horizontal screens (P001, P002, P003)
//   - 1 playlist "Hero Principal 2026" with 3 items:
//     slide 1 → content_block  (template: "hero-classic")
//     slide 2 → image          (placeholder URL)
//     slide 3 → video          (INFRA_INA20261-1.mp4, start 8.18 s)
//
// All devices are assigned to that playlist.
func SeedDevicesAndPlaylists(app core.App) error {
	// Idempotency: skip if any web_hero device already exists.
	existing, _ := app.FindRecordsByFilter("devices", "type = 'web_hero'", "", 1, 0)
	if len(existing) > 0 {
		return nil
	}

	// ── 1. Hero-classic content block ─────────────────────────────────────────
	cbCol, err := app.FindCollectionByNameOrId("content_blocks")
	if err != nil {
		return fmt.Errorf("content_blocks collection not found: %w", err)
	}
	cb := core.NewRecord(cbCol)
	cb.Set("title", "Per laborem ad lucem")
	cb.Set("description", "Formando generaciones con excelencia académica, valores humanos y el espíritu del norte de Chile.")
	cb.Set("category", "INFORMACIÓN")
	cb.Set("status", "publicado")
	cb.Set("template", "hero-classic")
	if err := app.Save(cb); err != nil {
		return fmt.Errorf("save hero content_block: %w", err)
	}

	// ── 2. Hero slide image multimedia ────────────────────────────────────────
	mmCol, err := app.FindCollectionByNameOrId("multimedia")
	if err != nil {
		return fmt.Errorf("multimedia collection not found: %w", err)
	}
	imgMM := core.NewRecord(mmCol)
	imgMM.Set("filename", "Comunicado-coloreate.png")
	imgMM.Set("url_r2", "https://i0.wp.com/colegiosanlorenzo.cl/wp-content/uploads/2026/03/Comunicado-coloreate.png?w=500&ssl=1")
	imgMM.Set("type", "imagen")
	imgMM.Set("estado", "publicado")
	if err := app.Save(imgMM); err != nil {
		return fmt.Errorf("save image multimedia: %w", err)
	}

	// ── 3. Video multimedia (start at 8.18 s) ─────────────────────────────────
	vidMM := core.NewRecord(mmCol)
	vidMM.Set("filename", "INFRA_INA20261-1.mp4")
	vidMM.Set("url_r2", "https://colegiosanlorenzo.cl/wp-content/uploads/2026/03/INFRA_INA20261-1.mp4")
	vidMM.Set("type", "video")
	vidMM.Set("estado", "publicado")
	vidMM.Set("start_time", 8.18)
	if err := app.Save(vidMM); err != nil {
		return fmt.Errorf("save video multimedia: %w", err)
	}

	// ── 4. Playlist ───────────────────────────────────────────────────────────
	plCol, err := app.FindCollectionByNameOrId("playlists")
	if err != nil {
		return fmt.Errorf("playlists collection not found: %w", err)
	}
	pl := core.NewRecord(plCol)
	pl.Set("name", "Hero Principal 2026")
	pl.Set("description", "Playlist principal para la landing pública del colegio")
	pl.Set("status", "activa")
	if err := app.Save(pl); err != nil {
		return fmt.Errorf("save playlist: %w", err)
	}

	// ── 5. Playlist items ─────────────────────────────────────────────────────
	piCol, err := app.FindCollectionByNameOrId("playlist_items")
	if err != nil {
		return fmt.Errorf("playlist_items collection not found: %w", err)
	}

	type piSeed struct {
		tipo     string
		cbID     string
		mmID     string
		orden    int
		duracion int
	}
	piItems := []piSeed{
		{tipo: "content_block", cbID: cb.Id, orden: 1, duracion: 10},
		{tipo: "image", mmID: imgMM.Id, orden: 2, duracion: 8},
		{tipo: "video", mmID: vidMM.Id, orden: 3, duracion: 30},
	}
	for _, it := range piItems {
		pi := core.NewRecord(piCol)
		pi.Set("playlist_id", pl.Id)
		pi.Set("tipo", it.tipo)
		pi.Set("orden", it.orden)
		pi.Set("duracion_segundos", it.duracion)
		if it.cbID != "" {
			pi.Set("content_block_id", it.cbID)
		}
		if it.mmID != "" {
			pi.Set("multimedia_id", it.mmID)
		}
		if err := app.Save(pi); err != nil {
			log.Printf("⚠️  seed playlist_item (orden %d): %v", it.orden, err)
		}
	}

	// ── 6. Devices — all assigned to the same playlist ────────────────────────
	devCol, err := app.FindCollectionByNameOrId("devices")
	if err != nil {
		return fmt.Errorf("devices collection not found: %w", err)
	}

	type devSeed struct {
		name      string
		dtype     string
		code      string
		ubicacion string
	}
	devItems := []devSeed{
		{"Web Hero - Landing Pública", "web_hero", "WEB1", "Sitio Web Público"},
		{"Totem Entrada Principal", "vertical", "T001", "Entrada Principal"},
		{"Totem Gimnasio", "vertical", "T002", "Gimnasio"},
		{"Pantalla Sala Profesores", "horizontal", "P001", "Sala de Profesores"},
		{"Pantalla Casino", "horizontal", "P002", "Casino"},
		{"Pantalla Patio Principal", "horizontal", "P003", "Patio Principal"},
	}
	for _, d := range devItems {
		dev := core.NewRecord(devCol)
		dev.Set("name", d.name)
		dev.Set("type", d.dtype)
		dev.Set("code", d.code)
		dev.Set("ubicacion", d.ubicacion)
		dev.Set("playlist_id", pl.Id)
		dev.Set("status", "activo")
		if err := app.Save(dev); err != nil {
			log.Printf("⚠️  seed device %s: %v", d.name, err)
		}
	}

	log.Printf("  ✅ SeedDevicesAndPlaylists: 6 devices + playlist '%s' + 3 items", pl.GetString("name"))
	return nil
}

// ── seedPropiedades inserts demo real-estate listings if collection is empty ──
func seedPropiedades(app core.App) error {
	col, err := app.FindCollectionByNameOrId("propiedades")
	if err != nil {
		return err
	}
	existing, _ := app.FindRecordsByFilter(col, "status = 'publicado'", "", 1, 0)
	if len(existing) > 0 {
		return nil // already seeded
	}

	type propSeed struct {
		titulo, slug, descripcion                  string
		operacion, tipo                            string
		direccion, comuna, region                  string
		precioUF, precioCLP                        float64
		dorm, banos, estac                         int
		supUtil, supTotal                          float64
		ano                                        int
		estado, amenidades, cover                  string
		destacada, oportunidad                     bool
	}

	now := types.NowDateTime()

	seeds := []propSeed{
		{
			titulo:      "Casa mediterránea con jardín — La Dehesa",
			slug:        "casa-mediterranea-la-dehesa",
			descripcion: "Amplia casa de 3 pisos con jardín, piscina y vista a la cordillera. Living comedor con doble altura, cocina equipada y quincho independiente. Barrio residencial tranquilo, cercano a colegios y supermercados.",
			operacion:   "VENTA", tipo: "CASA",
			direccion:   "Av. La Dehesa 1234",
			comuna:      "Lo Barnechea", region: "Metropolitana",
			precioUF:    18500, precioCLP: 720000000,
			dorm:        4, banos: 3, estac: 2,
			supUtil:     220, supTotal: 450, ano: 2016,
			estado:      "usada",
			amenidades:  "piscina,jardin,quincho,bodega,seguridad_24h",
			destacada:   true, oportunidad: false,
		},
		{
			titulo:      "Departamento moderno con vista al parque — Ñuñoa",
			slug:        "depto-moderno-nunoa",
			descripcion: "Departamento de 2 dormitorios y 2 baños con vista panorámica al Parque Bustamante. Edificio con gimnasio, piscina y sala multiuso. Excelente conectividad en metro Ñuñoa.",
			operacion:   "VENTA", tipo: "DEPARTAMENTO",
			direccion:   "Av. Irarrázaval 3400, Torre B, Dep. 1204",
			comuna:      "Ñuñoa", region: "Metropolitana",
			precioUF:    6400, precioCLP: 250000000,
			dorm:        2, banos: 2, estac: 1,
			supUtil:     72, supTotal: 85, ano: 2022,
			estado:      "a_estrenar",
			amenidades:  "gimnasio,piscina,sala_multiuso,conserjeria",
			destacada:   true, oportunidad: true,
		},
		{
			titulo:      "Parcela 5.000 m² con vista al valle — Olmué",
			slug:        "parcela-olmue-vista-valle",
			descripcion: "Terreno de 5.000 m² con árboles nativos, acceso a agua de pozo y rol propio. Ideal para casa de campo o proyecto inmobiliario. A 90 min de Santiago.",
			operacion:   "VENTA", tipo: "PARCELA",
			direccion:   "Camino Los Olmos Km 4",
			comuna:      "Olmué", region: "Valparaíso",
			precioUF:    3200, precioCLP: 124000000,
			dorm:        0, banos: 0, estac: 0,
			supUtil:     0, supTotal: 5000, ano: 0,
			estado:      "usada",
			amenidades:  "agua_pozo,electricidad,acceso_pavimentado",
			destacada:   false, oportunidad: true,
		},
		{
			titulo:      "Arriendo departamento amoblado — Providencia",
			slug:        "arriendo-depto-amoblado-providencia",
			descripcion: "Acogedor departamento de 1 dormitorio, completamente amoblado y equipado. Ubicación privilegiada a pasos del metro Pedro de Valdivia. Ideal para profesionales o parejas.",
			operacion:   "ARRIENDO", tipo: "DEPARTAMENTO",
			direccion:   "Av. Providencia 2015, Dep. 803",
			comuna:      "Providencia", region: "Metropolitana",
			precioUF:    21, precioCLP: 820000,
			dorm:        1, banos: 1, estac: 1,
			supUtil:     48, supTotal: 55, ano: 2019,
			estado:      "usada",
			amenidades:  "amoblado,equipado,gimnasio,conserjeria",
			destacada:   false, oportunidad: false,
		},
		{
			titulo:      "Oficina Class A — Las Condes",
			slug:        "oficina-class-a-las-condes",
			descripcion: "Oficina de 180 m² en edificio corporativo Class A. Piso 22 con vista 360°. Incluye 4 estacionamientos y 2 bodegas. Excelente conectividad.",
			operacion:   "ARRIENDO", tipo: "OFICINA",
			direccion:   "Av. Apoquindo 4500, Piso 22",
			comuna:      "Las Condes", region: "Metropolitana",
			precioUF:    75, precioCLP: 2900000,
			dorm:        0, banos: 2, estac: 4,
			supUtil:     180, supTotal: 180, ano: 2018,
			estado:      "usada",
			amenidades:  "seguridad_24h,aire_acondicionado,gimnasio,estacionamiento_visitas",
			destacada:   true, oportunidad: false,
		},
		{
			titulo:      "Local comercial esquina — Centro Viña",
			slug:        "local-comercial-vina-centro",
			descripcion: "Local comercial de 120 m² en esquina con alto flujo peatonal. Dos accesos, vitrina amplia. Ideal para retail, cafetería o servicios.",
			operacion:   "ARRIENDO", tipo: "LOCAL",
			direccion:   "Esq. Arlegui con Quillota",
			comuna:      "Viña del Mar", region: "Valparaíso",
			precioUF:    38, precioCLP: 1480000,
			dorm:        0, banos: 1, estac: 0,
			supUtil:     120, supTotal: 120, ano: 2005,
			estado:      "usada",
			amenidades:  "dos_accesos,vitrina_esquina,alto_flujo",
			destacada:   false, oportunidad: true,
		},
	}

	for _, s := range seeds {
		r := core.NewRecord(col)
		r.Set("titulo", s.titulo)
		r.Set("slug", s.slug)
		r.Set("descripcion", s.descripcion)
		r.Set("operacion", s.operacion)
		r.Set("tipo", s.tipo)
		r.Set("direccion", s.direccion)
		r.Set("comuna", s.comuna)
		r.Set("region", s.region)
		r.Set("precio_uf", s.precioUF)
		r.Set("precio_clp", s.precioCLP)
		r.Set("dormitorios", s.dorm)
		r.Set("banos", s.banos)
		r.Set("estacionamientos", s.estac)
		r.Set("superficie_util", s.supUtil)
		r.Set("superficie_total", s.supTotal)
		r.Set("ano_construccion", s.ano)
		r.Set("estado_propiedad", s.estado)
		r.Set("amenidades", s.amenidades)
		r.Set("status", "publicado")
		r.Set("destacada", s.destacada)
		r.Set("oportunidad", s.oportunidad)
		r.Set("publicado_en", now)
		if err := app.Save(r); err != nil {
			log.Printf("⚠️  seed propiedad %s: %v", s.titulo, err)
		}
	}
	log.Printf("  ✅ seedPropiedades: %d listings", len(seeds))
	return nil
}

// seedTiendas inserts demo store listings if collection is empty.
func seedTiendas(app core.App) error {
	col, err := app.FindCollectionByNameOrId("tiendas")
	if err != nil {
		return err
	}
	existing, _ := app.FindRecordsByFilter(col, "status = 'publicado'", "", 1, 0)
	if len(existing) > 0 {
		return nil
	}

	type tiendaSeed struct {
		nombre, slug, cat, gal, local, logo string
		tags, desc, about, about2, pay      string
		photos, similar                     string
		whatsapp, telefono, rating          string
		horarioLV, horarioSab, horarioDom   string
		destacada                           bool
	}

	seeds := []tiendaSeed{
		{
			nombre: "Starbucks", slug: "starbucks",
			cat: "restaurantes", gal: "norte", local: "Local 34",
			logo:       "https://logo.clearbit.com/starbucks.com",
			tags:       "Café,Frappuccino,WiFi,Snacks",
			desc:       "Tu café favorito con las mejores bebidas frías, calientes y Frappuccino.",
			about:      "Starbucks en Subcentro te ofrece la experiencia completa de la cadena más reconocida del mundo. Desde el icónico Frappuccino hasta bebidas de temporada y pastries artesanales.",
			about2:     "Local con zona de asientos cómoda y WiFi gratuito. Aceptamos la app Starbucks para acumular Stars.",
			pay:        "Efectivo · Tarjetas · App Starbucks · Mercado Pago",
			photos:     "https://images.unsplash.com/photo-1461023058943-07fcbe16d735?w=600&q=80,https://images.unsplash.com/photo-1559925393-8be0ec4767c8?w=400&q=80,https://images.unsplash.com/photo-1504674900247-0877df9cc836?w=400&q=80,https://images.unsplash.com/photo-1554118811-1e0d58224f24?w=400&q=80",
			similar:    "oakberry,krispy-kreme,dunkin,falafel-republic",
			whatsapp:   "56912345678", telefono: "+56 2 1234 5678",
			rating: "4.8", horarioLV: "9:00 – 21:00", horarioSab: "10:00 – 20:00", horarioDom: "Cerrado",
			destacada: true,
		},
		{
			nombre: "Adidas Combat Sports", slug: "adidas-combat",
			cat: "tiendas", gal: "norte", local: "Local 8",
			logo:       "https://subcentro.cl/wp-content/uploads/2026/03/Adidas-combat-100-1.jpg",
			tags:       "MMA,Boxeo,Judo,Equipamiento",
			desc:       "Tienda especializada en equipamiento y ropa de combate de la marca más icónica del deporte.",
			about:      "Adidas Combat Sports es la referencia para artes marciales en Las Condes. Guantes, protecciones, ropa técnica y calzado especializado de la línea Combat Sports.",
			about2:     "Desde boxeo y MMA hasta judo y taekwondo. Nuestros asesores son deportistas con experiencia.",
			pay:        "Efectivo · Tarjetas · Cuotas sin interés",
			photos:     "https://images.unsplash.com/photo-1517438984742-1262db08379e?w=600&q=80,https://images.unsplash.com/photo-1544367567-0f2fcb009e0b?w=400&q=80,https://images.unsplash.com/photo-1571019613454-1cb2f99b2d8b?w=400&q=80,https://images.unsplash.com/photo-1584466977773-e625c37cdd50?w=400&q=80",
			similar:    "starbucks,la-fete,tua,oakberry",
			whatsapp:   "56912345678", telefono: "+56 2 1234 5678",
			rating: "4.6", horarioLV: "9:00 – 21:00", horarioSab: "10:00 – 20:00", horarioDom: "Cerrado",
			destacada: true,
		},
		{
			nombre: "La Fête", slug: "la-fete",
			cat: "tiendas", gal: "norte", local: "Local 15",
			logo:       "https://subcentro.cl/wp-content/uploads/2025/05/la-fete-100-1.jpg",
			tags:       "Moda femenina,Accesorios,Francés",
			desc:       "Moda y accesorios con estilo francés. Elegancia y tendencia en cada colección.",
			about:      "La Fête trae el espíritu de la moda parisina con diseños únicos y materiales de calidad para la mujer moderna.",
			about2:     "Nueva colección de temporada disponible. Embalaje especial para regalos.",
			pay:        "Efectivo · Tarjetas · Débito",
			photos:     "https://images.unsplash.com/photo-1490481651871-ab68de25d43d?w=600&q=80,https://images.unsplash.com/photo-1515886657613-9f3515b0c78f?w=400&q=80,https://images.unsplash.com/photo-1469334031218-e382a71b716b?w=400&q=80,https://images.unsplash.com/photo-1541099649105-f69ad21f3246?w=400&q=80",
			similar:    "tua,starbucks,oakberry,adidas-combat",
			whatsapp:   "56912345678", telefono: "+56 2 1234 5678",
			rating: "4.6", horarioLV: "9:00 – 21:00", horarioSab: "10:00 – 20:00", horarioDom: "Cerrado",
			destacada: false,
		},
		{
			nombre: "TUA", slug: "tua",
			cat: "tiendas", gal: "sur", local: "Local 22",
			logo:       "https://subcentro.cl/wp-content/uploads/2025/08/TUA.png",
			tags:       "Accesorios,Moda,Mujer",
			desc:       "Accesorios y complementos de moda para la mujer contemporánea.",
			about:      "TUA es tu destino de moda en la Galería Sur. Bolsos, cinturones, joyería de fantasía y los complementos perfectos para cualquier look.",
			about2:     "Nuevas colecciones cada temporada. Atención personalizada y precios accesibles.",
			pay:        "Efectivo · Tarjetas · Débito",
			photos:     "https://images.unsplash.com/photo-1469334031218-e382a71b716b?w=600&q=80,https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=400&q=80,https://images.unsplash.com/photo-1472851294608-062f824d29cc?w=400&q=80,https://images.unsplash.com/photo-1483985988355-763728e1935b?w=400&q=80",
			similar:    "la-fete,starbucks,oakberry,adidas-combat",
			whatsapp:   "56912345678", telefono: "+56 2 1234 5678",
			rating: "4.5", horarioLV: "9:00 – 21:00", horarioSab: "10:00 – 20:00", horarioDom: "Cerrado",
			destacada: false,
		},
		{
			nombre: "OakBerry", slug: "oakberry",
			cat: "restaurantes", gal: "norte", local: "Local 27",
			logo:       "https://logo.clearbit.com/oakberry.com",
			tags:       "Açaí,Saludable,Vegano,Sin gluten",
			desc:       "Los mejores açaí bowls energizantes, personalizados con tus toppings favoritos.",
			about:      "OakBerry ofrece açaí premium sin azúcar añadida, personalizable con más de 30 toppings. Perfecto para desayuno o snack post-entrenamiento.",
			about2:     "100% sin gluten y apto para veganos. También contamos con smoothies y jugos naturales.",
			pay:        "Efectivo · Tarjetas · App OakBerry · Mercado Pago",
			photos:     "https://images.unsplash.com/photo-1511690743698-d9d85f2fbf38?w=600&q=80,https://images.unsplash.com/photo-1505252585461-04db1eb84625?w=400&q=80,https://images.unsplash.com/photo-1464305795204-6f5bbfc7fb81?w=400&q=80,https://images.unsplash.com/photo-1490474418585-ba9bad8fd0ea?w=400&q=80",
			similar:    "starbucks,krispy-kreme,falafel-republic,la-fete",
			whatsapp:   "56912345678", telefono: "+56 2 1234 5678",
			rating: "4.9", horarioLV: "9:00 – 21:00", horarioSab: "10:00 – 20:00", horarioDom: "Cerrado",
			destacada: true,
		},
		{
			nombre: "Falafel Republic", slug: "falafel-republic",
			cat: "restaurantes", gal: "sur", local: "Local 12",
			logo:       "https://subcentro.cl/wp-content/uploads/2025/02/LOGO-FALAFEL-100.jpg",
			tags:       "Árabe,Vegetariano,Vegano",
			desc:       "Auténtica cocina árabe: falafel crujiente, shawarma, hummus y mucho más.",
			about:      "Todo preparado con ingredientes frescos y recetas tradicionales. El falafel se hace diariamente con garbanzos seleccionados.",
			about2:     "Opciones vegetarianas y veganas. Menú incluye tabule, hummus artesanal y postres árabes.",
			pay:        "Efectivo · Tarjetas · Transbank",
			photos:     "https://images.unsplash.com/photo-1555626906-fcf10d6851b4?w=600&q=80,https://images.unsplash.com/photo-1547592180-85f173990554?w=400&q=80,https://images.unsplash.com/photo-1565299715199-866c917206bb?w=400&q=80,https://images.unsplash.com/photo-1529042410759-befb1204b468?w=400&q=80",
			similar:    "starbucks,oakberry,krispy-kreme,starbucks",
			whatsapp:   "56912345678", telefono: "+56 2 1234 5678",
			rating: "4.8", horarioLV: "9:00 – 21:00", horarioSab: "10:00 – 20:00", horarioDom: "Cerrado",
			destacada: true,
		},
		{
			nombre: "Krispy Kreme", slug: "krispy-kreme",
			cat: "restaurantes", gal: "sur", local: "Local 19",
			logo:       "https://logo.clearbit.com/krispykreme.com",
			tags:       "Donas,Café,Postres",
			desc:       "Las donas más famosas del mundo, recién horneadas. El letrero rojo lo dice todo.",
			about:      "Krispy Kreme en Subcentro trae la icónica Original Glazed junto a una completa carta de cafés y bebidas frías.",
			about2:     "Docenas disponibles para llevar y opciones de regalo. Cuando el letrero rojo está encendido, las donas acaban de salir del horno.",
			pay:        "Efectivo · Tarjetas · App Krispy Kreme",
			photos:     "https://images.unsplash.com/photo-1551024601-bec78aea704b?w=600&q=80,https://images.unsplash.com/photo-1499636136210-6f4ee915583e?w=400&q=80,https://images.unsplash.com/photo-1558961363-fa8fdf82db35?w=400&q=80,https://images.unsplash.com/photo-1629975913069-609d3fd9f6e2?w=400&q=80",
			similar:    "starbucks,oakberry,falafel-republic,la-fete",
			whatsapp:   "56912345678", telefono: "+56 2 1234 5678",
			rating: "4.7", horarioLV: "9:00 – 21:00", horarioSab: "10:00 – 20:00", horarioDom: "Cerrado",
			destacada: true,
		},
		{
			nombre: "Farmacias Ahumada", slug: "ahumada",
			cat: "farmacias", gal: "norte", local: "Local 5",
			logo:       "https://subcentro.cl/wp-content/uploads/2023/12/Farmacias-Ahumada_250px.png",
			tags:       "Farmacia,Salud,Medicamentos",
			desc:       "Tu farmacia de confianza con amplio stock de medicamentos y productos de salud.",
			about:      "Farmacias Ahumada ofrece atención farmacéutica profesional con la mayor variedad de medicamentos de venta libre y recetados.",
			about2:     "Programa de fidelidad con descuentos exclusivos. Entrega a domicilio disponible.",
			pay:        "Efectivo · Tarjetas · Débito · Web pay",
			photos:     "https://images.unsplash.com/photo-1584308666744-24d5c474f2ae?w=600&q=80,https://images.unsplash.com/photo-1576091160550-2173dba999ef?w=400&q=80,https://images.unsplash.com/photo-1585435557343-3b092031a831?w=400&q=80,https://images.unsplash.com/photo-1559757175-0eb30cd8c063?w=400&q=80",
			similar:    "cruz-verde,salcobrand,starbucks,oakberry",
			whatsapp:   "56912345678", telefono: "+56 2 1234 5678",
			rating: "4.4", horarioLV: "8:30 – 21:30", horarioSab: "9:00 – 21:00", horarioDom: "10:00 – 19:00",
			destacada: true,
		},
		{
			nombre: "Cruz Verde", slug: "cruz-verde",
			cat: "farmacias", gal: "sur", local: "Local 6",
			logo:       "https://subcentro.cl/wp-content/uploads/2023/12/Logo_Cruz_FondoBlanco_250px.png",
			tags:       "Farmacia,Salud,Dermocosméticos",
			desc:       "Farmacia Cruz Verde con atención personalizada y los mejores precios en salud.",
			about:      "Cruz Verde en Subcentro cuenta con todo lo que necesitas en medicamentos, dermocosméticos y productos de cuidado personal.",
			about2:     "Club Cruz Verde con beneficios exclusivos. Despacho rápido y seguro.",
			pay:        "Efectivo · Tarjetas · Débito",
			photos:     "https://images.unsplash.com/photo-1576671081837-49000212a370?w=600&q=80,https://images.unsplash.com/photo-1587854692152-cbe660dbde88?w=400&q=80,https://images.unsplash.com/photo-1559757148-5c350d0d3c56?w=400&q=80,https://images.unsplash.com/photo-1576671081837-49000212a370?w=400&q=80",
			similar:    "ahumada,salcobrand,starbucks,oakberry",
			whatsapp:   "56912345678", telefono: "+56 2 1234 5678",
			rating: "4.3", horarioLV: "8:30 – 21:30", horarioSab: "9:00 – 21:00", horarioDom: "10:00 – 19:00",
			destacada: false,
		},
		{
			nombre: "Salcobrand", slug: "salcobrand",
			cat: "farmacias", gal: "norte", local: "Local 7",
			logo:       "https://subcentro.cl/wp-content/uploads/2023/12/LOGO-SB_250px.png",
			tags:       "Farmacia,Salud,Belleza",
			desc:       "La farmacia con los mejores precios en medicamentos, belleza y cuidado personal.",
			about:      "Salcobrand combina farmacia tradicional con dermocosméticos de alta gama. Atención de químicos farmacéuticos en horario extendido.",
			about2:     "Programa SalcoBrand con puntos y descuentos. Más de 10.000 productos disponibles.",
			pay:        "Efectivo · Tarjetas · Débito · Transferencia",
			photos:     "https://images.unsplash.com/photo-1563213126-a4273aed2016?w=600&q=80,https://images.unsplash.com/photo-1576671081837-49000212a370?w=400&q=80,https://images.unsplash.com/photo-1587854692152-cbe660dbde88?w=400&q=80,https://images.unsplash.com/photo-1559757175-0eb30cd8c063?w=400&q=80",
			similar:    "ahumada,cruz-verde,starbucks,oakberry",
			whatsapp:   "56912345678", telefono: "+56 2 1234 5678",
			rating: "4.5", horarioLV: "8:30 – 21:30", horarioSab: "9:00 – 21:00", horarioDom: "10:00 – 19:00",
			destacada: false,
		},
		{
			nombre: "Subway", slug: "subway",
			cat: "restaurantes", gal: "sur", local: "Local 31",
			logo:       "https://logo.clearbit.com/subway.com",
			tags:       "Sándwiches,Saludable,Personalizado",
			desc:       "Sándwiches frescos y saludables hechos a tu gusto en el momento.",
			about:      "Subway te permite personalizar cada sándwich con ingredientes frescos. Elige tu pan, proteína, vegetales y salsas.",
			about2:     "Menús especiales del día y combos familiares. Opciones vegetarianas disponibles.",
			pay:        "Efectivo · Tarjetas · Débito · App Subway",
			photos:     "https://images.unsplash.com/photo-1509722747041-616f39b57569?w=600&q=80,https://images.unsplash.com/photo-1565299507177-b0ac66763828?w=400&q=80,https://images.unsplash.com/photo-1553909489-cd47e0907980?w=400&q=80,https://images.unsplash.com/photo-1540189549336-e6e99c3679fe?w=400&q=80",
			similar:    "falafel-republic,starbucks,oakberry,krispy-kreme",
			whatsapp:   "56912345678", telefono: "+56 2 1234 5678",
			rating: "4.3", horarioLV: "9:00 – 21:00", horarioSab: "10:00 – 20:00", horarioDom: "Cerrado",
			destacada: true,
		},
	}

	for _, s := range seeds {
		r := core.NewRecord(col)
		r.Set("nombre", s.nombre)
		r.Set("slug", s.slug)
		r.Set("cat", s.cat)
		r.Set("gal", s.gal)
		r.Set("local", s.local)
		r.Set("logo", s.logo)
		r.Set("tags", s.tags)
		r.Set("desc", s.desc)
		r.Set("about", s.about)
		r.Set("about2", s.about2)
		r.Set("pay", s.pay)
		r.Set("photos", s.photos)
		r.Set("similar", s.similar)
		r.Set("whatsapp", s.whatsapp)
		r.Set("telefono", s.telefono)
		r.Set("rating", s.rating)
		r.Set("horario_lv", s.horarioLV)
		r.Set("horario_sab", s.horarioSab)
		r.Set("horario_dom", s.horarioDom)
		r.Set("status", "publicado")
		r.Set("destacada", s.destacada)
		if err := app.Save(r); err != nil {
			log.Printf("⚠️  seed tienda %s: %v", s.nombre, err)
		}
	}
	log.Printf("  ✅ seedTiendas: %d stores", len(seeds))
	return nil
}
