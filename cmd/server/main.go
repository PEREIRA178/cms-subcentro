package main

import (
	"log"
	"os"

	"cms-plazareal/internal/auth"
	"cms-plazareal/internal/config"
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

	// ── Analytics: record public page views ──
	// Skips /admin, /frag, /fragments, /static, /api/, /ws, and any 4xx/5xx.
	// Fail-open: never blocks or fails a request.
	app.Use(middleware.TrackPageView(cfg, pb))

	hub := realtime.NewHub()
	go hub.Run()
	realtime.SetHubInstance(hub)

	// ── PUBLIC WEB ──
	app.Get("/", web.IndexHandler(cfg, pb))
	app.Get("/index.html", web.IndexHandler(cfg, pb))
	app.Get("/buscador-tiendas", web.TiendasPageHandler(cfg, pb))
	app.Get("/buscador-tiendas.html", web.TiendasPageHandler(cfg, pb))
	app.Get("/tiendas/:slug", web.TiendaDetailHandler(cfg, pb))
	app.Get("/noticias", web.NoticiasPageHandler(cfg))
	app.Get("/noticias.html", web.NoticiasPageHandler(cfg))
	app.Get("/comunicados", web.ComunicadosPageHandler(cfg))
	app.Get("/locales", web.LocalesPageHandler(cfg, pb))
	app.Get("/locales-disponibles", web.LocalesPageHandler(cfg, pb))
	app.Get("/promociones", web.PromocionesPageHandler(cfg))

	// ── HTMX FRAGMENTS ──
	frag := app.Group("/fragments")
	frag.Get("/hero", fragments.HeroCarousel(cfg, pb))
	frag.Get("/eventos", fragments.Eventos(cfg, pb))
	// Content blocks (NOTICIA / COMUNICADO / PROMOCION) — templ-rendered cards
	frag.Get("/noticias", fragments.NoticiasCards(cfg, pb))
	frag.Get("/noticias-page", fragments.NoticiasCards(cfg, pb))
	frag.Get("/comunicados", fragments.ComunicadosCards(cfg, pb))
	frag.Get("/comunicados-cards", fragments.ComunicadosCards(cfg, pb))
	frag.Get("/comunicados-page", fragments.ComunicadosCards(cfg, pb))
	frag.Get("/promos", fragments.PromosCards(cfg, pb))
	frag.Get("/promos-page", fragments.PromosCards(cfg, pb))
	// Locales disponibles fragment (used by homepage embed and the public page)
	frag.Get("/locales-disponibles", fragments.LocalesCards(cfg, pb))
	frag.Get("/locales-cards", fragments.LocalesCards(cfg, pb))
	frag.Get("/blog", fragments.Blog(cfg, pb))
	// Tiendas fragments (Subcentro)
	frag.Get("/tiendas", fragments.TiendasPage(cfg, pb))
	frag.Get("/tiendas-destacadas", fragments.TiendasDestacadas(cfg, pb))
	frag.Get("/tiendas-marquee", fragments.TiendasMarquee(cfg, pb))
	frag.Get("/tienda/:key", fragments.TiendaDetail(cfg, pb))

	app.Get("/rss.xml", web.RSSFeed(cfg))

	// ── WS ──
	app.Use("/ws", func(c *fiber.Ctx) error {
		if gows.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws/web", gows.New(ws.WebSocket(hub)))

	// ── ADMIN ──
	app.Get("/admin/login", admin.LoginPage(cfg))
	app.Post("/admin/login", admin.LoginSubmit(cfg, pb))
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

	// Content blocks (NOTICIA / COMUNICADO / PROMOCION) — templ CRUD
	adm.Get("/noticias", admin.ContentList(cfg, pb, "NOTICIA"))
	adm.Get("/comunicados", admin.ContentList(cfg, pb, "COMUNICADO"))
	adm.Get("/promociones", admin.ContentList(cfg, pb, "PROMOCION"))
	adm.Get("/content", func(c *fiber.Ctx) error {
		cat := c.Query("cat", "NOTICIA")
		return admin.ContentList(cfg, pb, cat)(c)
	})
	adm.Get("/content/new", admin.ContentNew(cfg))
	adm.Get("/content/:id/edit", admin.ContentEdit(cfg, pb))
	adm.Post("/content", admin.ContentCreate(cfg, pb))
	adm.Put("/content/:id", admin.ContentUpdate(cfg, pb))
	adm.Delete("/content/:id", admin.ContentDelete(cfg, pb))

	// Users
	adm.Get("/users", middleware.RoleRequired("superadmin"), admin.UsersList(cfg, pb))
	adm.Get("/users/new", middleware.RoleRequired("superadmin"), admin.UserNew(cfg))
	adm.Get("/users/:id/edit", middleware.RoleRequired("superadmin"), admin.UserEdit(cfg, pb))
	adm.Post("/users", middleware.RoleRequired("superadmin"), admin.UserCreate(cfg, pb))
	adm.Put("/users/:id", middleware.RoleRequired("superadmin"), admin.UserUpdate(cfg, pb))
	adm.Delete("/users/:id", middleware.RoleRequired("superadmin"), admin.UserDelete(cfg, pb))

	// Reports / Analytics
	adm.Get("/reports", middleware.RoleRequired("superadmin", "director"), admin.ReportsPageHandler(cfg, pb))
	adm.Get("/reports/export", middleware.RoleRequired("superadmin", "director"), admin.ReportsExport(cfg, pb))

	// Site Settings
	adm.Get("/settings", middleware.RoleRequired("superadmin", "director"), admin.SettingsPageHandler(cfg, pb))
	adm.Post("/settings", middleware.RoleRequired("superadmin", "director"), admin.SettingsUpdate(cfg, pb))

	// Tiendas
	adm.Get("/tiendas", admin.TiendasList(cfg, pb))
	adm.Get("/tiendas/new", admin.TiendaForm(cfg))
	adm.Post("/tiendas", admin.TiendaCreate(cfg, pb))
	adm.Get("/tiendas/:id/edit", admin.TiendaEdit(cfg, pb))
	adm.Put("/tiendas/:id", admin.TiendaUpdate(cfg, pb))
	adm.Delete("/tiendas/:id", admin.TiendaDelete(cfg, pb))
	adm.Post("/tiendas/:id/publish", admin.TiendaToggleStatus(cfg, pb))

	// Locales disponibles (CRUD)
	adm.Get("/locales", admin.LocalesList(cfg, pb))
	adm.Get("/locales/new", admin.LocalNew(cfg))
	adm.Post("/locales", admin.LocalCreate(cfg, pb))
	adm.Get("/locales/:id/edit", admin.LocalEdit(cfg, pb))
	adm.Put("/locales/:id", admin.LocalUpdate(cfg, pb))
	adm.Delete("/locales/:id", admin.LocalDelete(cfg, pb))

	// Reservas (admin: list, change estado, delete)
	adm.Get("/reservas", admin.ReservasList(cfg, pb))
	adm.Post("/reservas/:id/estado", admin.ReservaUpdateEstado(cfg, pb))
	adm.Delete("/reservas/:id", admin.ReservaDelete(cfg, pb))

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
