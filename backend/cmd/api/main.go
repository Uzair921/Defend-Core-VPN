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

	"vpn.local/backend/internal/auth"
	"vpn.local/backend/internal/cache"
	"vpn.local/backend/internal/config"
	"vpn.local/backend/internal/db"
	"vpn.local/backend/internal/middleware"
	"vpn.local/backend/internal/users"
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
