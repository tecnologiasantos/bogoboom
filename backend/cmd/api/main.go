// Command api is the Bogoboom HTTP API.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"bogoboom-backend/internal/adapters/handler"
	"bogoboom-backend/internal/adapters/middleware"
	"bogoboom-backend/internal/adapters/repository"
	"bogoboom-backend/internal/config"
	"bogoboom-backend/internal/core/domain"
	"bogoboom-backend/internal/core/service"
)

func main() {
	_ = godotenv.Load() // optional .env for local development

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	trustedProxies, err := config.LoadTrustedProxies()
	if err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	if len(trustedProxies) == 0 {
		log.Println("Trusted proxies: none (client IP is the connection address)")
	} else {
		log.Printf("Trusted proxies: %s (client IP from %s)", strings.Join(trustedProxies, ", "), middleware.ClientIPHeader)
	}
	origins, err := handler.NewOriginPolicy(cfg.Server.FrontendURL)
	if err != nil {
		log.Fatalf("invalid FRONTEND_URL: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool, err := config.NewDatabasePool(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("Connected to PostgreSQL")

	app := buildApp(cfg, pool, trustedProxies, origins)

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Println("Shutting down...")
		cancel()
		if err := app.Shutdown(); err != nil {
			log.Printf("Shutdown error: %v", err)
		}
	}()

	addr := ":" + cfg.Server.Port
	log.Printf("Bogoboom API listening on %s (%s)", addr, cfg.Env)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// buildApp wires repositories, services, handlers and routes. It does not connect to anything, so a
// test can call it with a nil pool and read the route table.
func buildApp(cfg *config.Config, pool *pgxpool.Pool, trustedProxies []string, origins *handler.OriginPolicy) *fiber.App {
	users := repository.NewUserRepository(pool)
	newcomers := repository.NewNewcomerRepository(pool)
	companies := repository.NewCompanyRepository(pool)
	agencies := repository.NewAgencyRepository(pool)
	props := repository.NewPropertyRepository(pool)
	favs := repository.NewFavoriteRepository(pool)
	notes := repository.NewNotificationRepository(pool)
	inquiries := repository.NewInquiryRepository(pool)
	audit := repository.NewAuditRepository(pool)

	registration := service.NewRegistrationService(users)
	newcomerSvc := service.NewNewcomerService(newcomers, companies, props, notes, audit)
	leadSvc := service.NewLeadService(agencies, newcomers, props, notes, audit)
	propSvc := service.NewPropertyService(agencies, props, newcomers, notes, favs)
	publicSvc := service.NewPublicService(inquiries, agencies, audit)

	account := handler.NewAccountHandler(registration, users)
	newcomerH := handler.NewNewcomerHandler(newcomerSvc, propSvc)
	agencyH := handler.NewAgencyHandler(leadSvc, propSvc)
	publicH := handler.NewPublicHandler(publicSvc)

	auth0 := middleware.NewAuth0Middleware(middleware.Auth0Config{Domain: cfg.Auth0.Domain, Audience: cfg.Auth0.Audience})
	resolver := middleware.NewUserResolver(users)

	fiberCfg := fiber.Config{
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		BodyLimit:    64 * 1024, // JSON forms only: no uploads in this API
		ErrorHandler: handler.GlobalErrorHandler,
	}
	middleware.ConfigureClientIP(&fiberCfg, trustedProxies)
	app := fiber.New(fiberCfg)

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(middleware.SecurityHeaders())
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: origins.Allows,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy", "timestamp": time.Now().UTC()})
	})

	api := app.Group("/api/v1")
	api.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Bogoboom API v1", "version": "1.0.0"})
	})

	// Public forms: strict per-IP limits.
	forms := middleware.NewRateLimiter(10, time.Minute)
	api.Post("/partner-inquiries", forms, publicH.PartnerInquiry)
	api.Post("/leads", forms, publicH.Lead)

	// A valid token is enough to register (the user does not exist yet).
	reg := api.Group("/register", auth0.Protect(), middleware.NewRateLimiter(10, time.Minute))
	reg.Post("/newcomer", account.RegisterNewcomer)
	reg.Post("/hr", account.RegisterHR)
	reg.Post("/agency", account.RegisterAgency)

	// Everything else needs a token AND a registered, active user.
	protected := api.Group("", auth0.Protect(), resolver.Resolve(), middleware.NewRateLimiter(240, time.Minute))
	protected.Get("/me", account.Me)

	// Newcomer routes have no common prefix, so the role guard is attached per route. (A Group with
	// middleware applies it to every route registered after it under the same prefix.)
	newc := middleware.RolesAllowed(domain.RoleNewcomer)
	protected.Get("/newcomers/me", newc, newcomerH.Me)
	protected.Put("/newcomers/me", newc, newcomerH.Update)
	protected.Get("/properties", newc, newcomerH.SearchProperties)
	protected.Get("/favorites", newc, newcomerH.Favorites)
	protected.Put("/favorites/:id", newc, newcomerH.AddFavorite)
	protected.Delete("/favorites/:id", newc, newcomerH.RemoveFavorite)

	hr := protected.Group("/hr", middleware.RolesAllowed(domain.RoleHR))
	hr.Post("/newcomers", newcomerH.Invite)
	hr.Get("/newcomers", newcomerH.ListForHR)

	ag := protected.Group("/agency", middleware.RolesAllowed(domain.RoleAgency))
	ag.Get("/account", agencyH.Account)
	ag.Get("/leads", agencyH.Leads)
	ag.Post("/leads/:id/unlock", middleware.NewRateLimiter(30, time.Minute), agencyH.Unlock)
	ag.Get("/notifications", agencyH.Notifications)
	ag.Post("/notifications/:id/read", agencyH.MarkRead)
	ag.Get("/properties", agencyH.ListProperties)
	ag.Post("/properties", agencyH.CreateProperty)
	ag.Put("/properties/:id", agencyH.UpdateProperty)
	ag.Delete("/properties/:id", agencyH.DeleteProperty)

	adm := protected.Group("/admin", middleware.RolesAllowed(domain.RoleAdmin))
	adm.Get("/partner-inquiries", publicH.ListInquiries)
	adm.Put("/agencies/:id/verify", publicH.VerifyAgency)
	adm.Put("/agencies/:id/plan", publicH.SetPlan)

	return app
}
