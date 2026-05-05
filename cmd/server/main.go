package main

import (
	"log"
	"os"

	"cms-plazareal/internal/auth"
	"cms-plazareal/internal/config"
	apiHandlers "cms-plazareal/internal/handlers/api"
	"cms-plazareal/internal/handlers/admin"
	"cms-plazareal/internal/handlers/fragments"
	"cms-plazareal/internal/handlers/web"
	"cms-plazareal/internal/handlers/ws"
	"cms-plazareal/internal/middleware"
	"cms-plazareal/internal/realtime"
	"cms-plazareal/internal/services/r2"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	gows "github.com/gofiber/websocket/v2"
	"github.com/pocketbase/pocketbase"
)

func main() {
	cfg := config.Load()

	pb := pocketbase.New()
	auth.RegisterPBHooks(pb)
	realtime.RegisterPBHooks(pb)

	r2Client, err := r2.New(cfg)
	if err != nil {
		log.Printf("R2 client init failed (uploads disabled): %v", err)
	}

	go func() {
		if err := pb.Start(); err != nil {
			log.Fatalf("PocketBase failed: %v", err)
		}
	}()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			log.Printf("ERROR [%d] %s %s: %v", code, c.Method(), c.Path(), err)
			if c.Get("HX-Request") == "true" {
				return c.Status(code).SendString(`<div class="toast toast-error">Error interno</div>`)
			}
			return c.Status(code).SendString("Error interno del servidor")
		},
		BodyLimit: 50 * 1024 * 1024,
	})

	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} ${method} ${path} (${latency})\n",
		TimeFormat: "15:04:05",
	}))
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, HX-Request, HX-Trigger",
	}))

	app.Static("/static", "./web/static", fiber.Static{
		Compress:      true,
		CacheDuration: cfg.StaticCacheDuration,
	})

	hub := realtime.NewHub()
	go hub.Run()
	realtime.SetHubInstance(hub)

	// ── PUBLIC WEB ──
	app.Get("/", web.IndexHandler(cfg))
	app.Get("/index.html", web.IndexHandler(cfg))
	app.Get("/buscador-tiendas", web.TiendasPageHandler(cfg))
	app.Get("/buscador-tiendas.html", web.TiendasPageHandler(cfg))
	app.Get("/tiendas/:slug", web.TiendaDetailHandler(cfg, pb))
	app.Get("/noticias", web.NoticiasPageHandler(cfg))
	app.Get("/noticias.html", web.NoticiasPageHandler(cfg))
	app.Get("/comunicados", web.ComunicadosPageHandler(cfg))
	app.Get("/locales", web.LocalesPageHandler(cfg))
	app.Get("/locales.html", web.LocalesPageHandler(cfg))
	app.Get("/promociones", web.PromocionesPageHandler(cfg))
	app.Get("/eventos.html", web.PromocionesPageHandler(cfg))

	// ── HTMX FRAGMENTS ──
	frag := app.Group("/fragments")
	frag.Get("/hero", fragments.HeroCarousel(cfg, pb))
	frag.Get("/eventos", fragments.Eventos(cfg, pb))
	frag.Get("/noticias", fragments.Noticias(cfg, pb))
	frag.Get("/comunicados", fragments.Comunicados(cfg, pb))
	frag.Get("/blog", fragments.Blog(cfg, pb))
	frag.Get("/noticias-page", fragments.NoticiasPage(cfg, pb))
	// Tiendas fragments (Subcentro)
	frag.Get("/tiendas", fragments.TiendasPage(cfg, pb))
	frag.Get("/tiendas-destacadas", fragments.TiendasDestacadas(cfg, pb))
	frag.Get("/tiendas-marquee", fragments.TiendasMarquee(cfg, pb))
	frag.Get("/tienda/:key", fragments.TiendaDetail(cfg, pb))

	app.Get("/noticias/:id", web.NoticiaHandler(cfg, pb))
	app.Get("/rss.xml", web.RSSFeed(cfg))

	// ── PUBLIC API ──
	api := app.Group("/api")
	api.Get("/devices/:code/playlist", apiHandlers.DevicePlaylist(cfg, pb))
	api.Get("/events/upcoming", apiHandlers.UpcomingEvents(cfg, pb))

	// ── DEVICE / WS ──
	app.Get("/display/:code", web.DeviceDisplay(cfg))
	app.Get("/totem/:code", web.TotemDisplay(cfg))
	app.Use("/ws", func(c *fiber.Ctx) error {
		if gows.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws/device/:code", gows.New(ws.DeviceSocket(hub)))
	app.Get("/ws/web", gows.New(ws.WebSocket(hub)))

	// ── ADMIN ──
	app.Get("/admin/login", admin.LoginPage(cfg))
	app.Post("/admin/login", admin.LoginSubmit(cfg))
	app.Post("/admin/logout", admin.Logout())

	adm := app.Group("/admin", middleware.AuthRequired(cfg))

	adm.Get("/", admin.Dashboard(cfg))
	adm.Get("/dashboard", admin.Dashboard(cfg))
	adm.Get("/dashboard/stats", admin.DashboardStats(cfg, pb))

	// Empty endpoint used to safely close modals via HTMX
	adm.Get("/empty", func(c *fiber.Ctx) error {
		return c.SendString("")
	})

	// R2 upload (drag-drop image upload endpoint used by UploadField widget)
	adm.Post("/upload", middleware.RoleRequired("superadmin", "director", "admin", "editor"), admin.UploadFile(cfg, r2Client))

	// Multimedia
	adm.Get("/multimedia", admin.MultimediaList(cfg, pb))
	adm.Get("/multimedia/new", admin.MultimediaForm(cfg))
	adm.Post("/multimedia", admin.MultimediaCreate(cfg, pb))
	adm.Get("/multimedia/:id/edit", admin.MultimediaEdit(cfg, pb))
	adm.Put("/multimedia/:id", admin.MultimediaUpdate(cfg, pb))
	adm.Delete("/multimedia/:id", admin.MultimediaDelete(cfg, pb))

	// Events (content_blocks excl. NOTICIA)
	adm.Get("/events", admin.EventsList(cfg, pb))
	adm.Get("/events/new", admin.EventForm(cfg))
	adm.Post("/events", admin.EventCreate(cfg, pb))
	adm.Get("/events/:id/edit", admin.EventEdit(cfg, pb))
	adm.Put("/events/:id", admin.EventUpdate(cfg, pb))
	adm.Delete("/events/:id", admin.EventDelete(cfg, pb))
	adm.Post("/events/:id/publish", admin.EventPublish(cfg, pb))

	// News (content_blocks category=NOTICIA)
	adm.Get("/news", admin.NewsList(cfg, pb))
	adm.Get("/news/new", admin.NewsForm(cfg))
	adm.Post("/news", admin.NewsCreate(cfg, pb))
	adm.Get("/news/:id/edit", admin.NewsEdit(cfg, pb))
	adm.Put("/news/:id", admin.NewsUpdate(cfg, pb))
	adm.Delete("/news/:id", admin.NewsDelete(cfg, pb))

	// Playlists
	adm.Get("/playlists", admin.PlaylistList(cfg, pb))
	adm.Get("/playlists/new", admin.PlaylistForm(cfg, pb))
	adm.Post("/playlists", admin.PlaylistCreate(cfg, pb))
	adm.Get("/playlists/:id/edit", admin.PlaylistEdit(cfg, pb))
	adm.Put("/playlists/:id", admin.PlaylistUpdate(cfg, pb))
	adm.Delete("/playlists/:id", admin.PlaylistDelete(cfg, pb))
	adm.Post("/playlists/:id/items/reorder", admin.PlaylistReorder(cfg, pb))

	// Devices
	adm.Get("/devices", admin.DeviceList(cfg, pb))
	adm.Get("/devices/new", admin.DeviceForm(cfg, pb))
	adm.Post("/devices", admin.DeviceCreate(cfg, pb))
	adm.Get("/devices/:id/edit", admin.DeviceEdit(cfg, pb))
	adm.Put("/devices/:id", admin.DeviceUpdate(cfg, pb))
	adm.Delete("/devices/:id", admin.DeviceDelete(cfg, pb))
	adm.Post("/devices/:id/assign-playlist", admin.DeviceAssignPlaylist(cfg, pb))

	// Users
	adm.Get("/users", middleware.RoleRequired("superadmin", "director"), admin.UserList(cfg))
	adm.Post("/users", middleware.RoleRequired("superadmin", "director"), admin.UserCreate(cfg))
	adm.Put("/users/:id", middleware.RoleRequired("superadmin", "director"), admin.UserUpdate(cfg))
	adm.Delete("/users/:id", middleware.RoleRequired("superadmin"), admin.UserDelete(cfg))

	adm.Get("/whatsapp-logs", admin.WhatsAppLogs(cfg))

	// Tiendas
	adm.Get("/tiendas", admin.TiendasList(cfg, pb))
	adm.Get("/tiendas/new", admin.TiendaForm(cfg))
	adm.Post("/tiendas", admin.TiendaCreate(cfg, pb))
	adm.Get("/tiendas/:id/edit", admin.TiendaEdit(cfg, pb))
	adm.Put("/tiendas/:id", admin.TiendaUpdate(cfg, pb))
	adm.Delete("/tiendas/:id", admin.TiendaDelete(cfg, pb))
	adm.Post("/tiendas/:id/publish", admin.TiendaToggleStatus(cfg, pb))

	app.Post("/webhook/whatsapp", web.WhatsAppWebhook(cfg))

	port := cfg.Port
	if port == "" {
		port = "3000"
	}
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	log.Printf("🏢 Plaza Real CMS en http://localhost:%s", port)
	log.Printf("📊 Dashboard: http://localhost:%s/admin", port)
	log.Printf("🔧 PocketBase Admin: http://localhost:8090/_/")

	log.Fatal(app.Listen(":" + port))
}
