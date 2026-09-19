package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"defendcore-vpn/internal/auth"
	"defendcore-vpn/internal/cache"
	"defendcore-vpn/internal/config"
	"defendcore-vpn/internal/db"
	"defendcore-vpn/internal/devices"
	"defendcore-vpn/internal/middleware"
	"defendcore-vpn/internal/organizations"
	"defendcore-vpn/internal/policy"
	"defendcore-vpn/internal/services"
	"defendcore-vpn/internal/users"
	"defendcore-vpn/internal/vpnctl"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Database
	database, err := db.Connect(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}
	defer database.Close()
	log.Println("database connected")

	if err := database.RunMigrations(ctx); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}
	log.Println("migrations complete")

	// Redis
	cacheClient, err := cache.Connect(ctx, cfg.Redis)
	if err != nil {
		log.Fatalf("redis connect failed: %v", err)
	}
	defer cacheClient.Close()
	log.Println("redis connected")

	// Wire dependencies
	usersRepo := users.NewRepository(database.Pool)
	refreshRepo := auth.NewRefreshRepository(database.Pool)
	jwtSvc := auth.NewJWTService(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	authSvc := auth.NewService(usersRepo, refreshRepo, jwtSvc)
	authHandler := auth.NewHandler(authSvc)

	// Devices
	devicesRepo := devices.NewRepository(database.Pool)
	devicesSvc := devices.NewService(devicesRepo, os.Getenv("VPN_SERVER_PUBLIC_KEY"))
	devicesHandler := devices.NewHandler(devicesSvc, envOr("VPN_PUBLIC_HOST", "127.0.0.1"), cfg.VPN.Port)

	// Access policies
	policyRepo := policy.NewRepository(database.Pool)
	policyHandler := policy.NewHandler(policyRepo)


        // Multi-VPN Services
        servicesRepo := services.NewRepository(database.Pool)
        servicesSvc := services.NewService(servicesRepo)
        servicesHandler := services.NewHandler(servicesSvc)
	// VPN control plane
	vpnctlRepo := vpnctl.NewRepository(database.Pool)
	vpnctlHandler := vpnctl.NewHandler(vpnctlRepo, cfg.VPN.APIKey)

// Organizations (SaaS multi-tenant)
orgsRepo := organizations.NewRepository(database.Pool)
orgsSvc := organizations.NewService(orgsRepo)
orgsHandler := organizations.NewHandler(orgsSvc)

	// Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))

	// Public
	r.Get("/api/v1/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"defendcore-vpn-api","version":"0.1.0"}`))
	})

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)

		// Protected
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtSvc))
			r.Get("/me", authHandler.Me)
			r.Post("/logout", authHandler.Logout)
		})
	})

	r.Route("/api/v1/devices", func(r chi.Router) {
		r.Use(middleware.Auth(jwtSvc))
		r.Post("/", devicesHandler.Create)
		r.Get("/", devicesHandler.List)
		r.Get("/{id}", devicesHandler.Get)
		r.Delete("/{id}", devicesHandler.Delete)
	})

	// Access Policies (admin only — TODO: add role check)
	r.Route("/api/v1/admin/policies", func(r chi.Router) {
		r.Use(middleware.Auth(jwtSvc))
		r.Post("/", policyHandler.Create)
		r.Get("/user/{user_id}", policyHandler.ListByUser)
		r.Get("/{id}", policyHandler.Get)
		r.Patch("/{id}", policyHandler.Update)
		r.Delete("/{id}", policyHandler.Delete)
	})
        // =====================================================
        // Multi-VPN Services
        // =====================================================

        // Service Types (public — read-only catalog)
        r.Route("/api/v1/vpn/service-types", func(r chi.Router) {
                r.Get("/", servicesHandler.ListServiceTypes)
                r.Get("/{code}", servicesHandler.GetServiceType)
        })

        // Services (admin)
        r.Route("/api/v1/admin/vpn/services", func(r chi.Router) {
                r.Use(middleware.Auth(jwtSvc))
                r.Post("/", servicesHandler.CreateService)
                r.Get("/", servicesHandler.ListServices)
                r.Get("/{id}", servicesHandler.GetService)
                r.Patch("/{id}", servicesHandler.UpdateService)
                r.Delete("/{id}", servicesHandler.DeleteService)

                // User assignments
                r.Post("/{id}/users", servicesHandler.AssignUser)
                r.Post("/{id}/users/bulk", servicesHandler.BulkAssignUsers)
                r.Get("/{id}/users", servicesHandler.ListServiceUsers)
                r.Delete("/{id}/users/{user_id}", servicesHandler.UnassignUser)
        })

        // Services (client)
        r.Route("/api/v1/vpn/services", func(r chi.Router) {
                r.Use(middleware.Auth(jwtSvc))
                r.Get("/", servicesHandler.ListMyServices)
                r.Get("/{id}/config", servicesHandler.GetClientConfig)
        })

// =====================================================
// SaaS Multi-Tenant — SuperAdmin
// =====================================================
r.Route("/api/v1/superadmin", func(r chi.Router) {
r.Use(middleware.Auth(jwtSvc))
r.Use(middleware.SuperAdmin(usersRepo))

r.Post("/organizations", orgsHandler.CreateOrganization)
r.Get("/organizations", orgsHandler.ListOrganizations)
r.Get("/organizations/{id}", orgsHandler.GetOrganization)
r.Patch("/organizations/{id}", orgsHandler.UpdateOrganization)
r.Delete("/organizations/{id}", orgsHandler.DeleteOrganization)

r.Post("/organizations/{id}/users", orgsHandler.AddOrganizationUser)
r.Get("/organizations/{id}/users", orgsHandler.ListOrganizationUsers)
r.Delete("/organizations/{id}/users/{user_id}", orgsHandler.RemoveOrganizationUser)

r.Post("/organizations/{id}/subscriptions", orgsHandler.CreateSubscription)
r.Get("/organizations/{id}/subscriptions", orgsHandler.ListSubscriptions)
r.Patch("/subscriptions/{id}", orgsHandler.UpdateSubscriptionStatus)
r.Delete("/subscriptions/{id}", orgsHandler.DeleteSubscription)

r.Post("/organizations/{id}/invoices", orgsHandler.CreateInvoice)
r.Get("/organizations/{id}/invoices", orgsHandler.ListInvoices)
r.Post("/invoices/{id}/paid", orgsHandler.MarkInvoicePaid)
})

// =====================================================
// SaaS Multi-Tenant — Org Admin (self-service)
// =====================================================
r.Route("/api/v1/org/me", func(r chi.Router) {
r.Use(middleware.Auth(jwtSvc))
r.Use(middleware.Tenant(orgsRepo))

r.Get("/", orgsHandler.GetMyOrganization)
r.Get("/users", orgsHandler.ListMyUsers)
r.Get("/subscriptions", orgsHandler.ListMySubscriptions)
r.Get("/invoices", orgsHandler.ListMyInvoices)
})

// VPN control plane — server-side (API key auth)
	r.Route("/api/v1/vpn", func(r chi.Router) {
		// Server registration + heartbeat (called by VPN server)
		r.Post("/servers/register", vpnctlHandler.RegisterServer)
		r.Post("/servers/heartbeat", vpnctlHandler.Heartbeat)

		// Session lifecycle (called by VPN server)
		r.Post("/sessions", vpnctlHandler.StartSession)
		r.Patch("/sessions/{id}", vpnctlHandler.UpdateSession)
		r.Post("/sessions/{id}/end", vpnctlHandler.EndSession)

		// User-facing (JWT auth)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtSvc))
			r.Get("/servers", vpnctlHandler.ListServers)
			r.Get("/sessions", vpnctlHandler.ListSessions)
			r.Get("/sessions/{id}", vpnctlHandler.GetSession)
		})

		// Policy fetch (API key auth — called by VPN server)
		r.With(vpnctlHandler.APIKeyAuth).Get("/policies/{user_id}", policyHandler.GetEffectivePolicy)
	})

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("DefendCore VPN API starting on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown failed: %v", err)
	}
	log.Println("server stopped")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
