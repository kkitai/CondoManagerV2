package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/kkitai/CondoManagerV2/app/internal/config"
	"github.com/kkitai/CondoManagerV2/app/internal/database"
	"github.com/kkitai/CondoManagerV2/app/internal/handler"
	"github.com/kkitai/CondoManagerV2/app/internal/mailer"
	"github.com/kkitai/CondoManagerV2/app/internal/middleware"
	"github.com/kkitai/CondoManagerV2/app/internal/repository"
	"github.com/kkitai/CondoManagerV2/app/internal/service"
)

func main() {
	_ = godotenv.Load()

	logger, err := buildLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync() //nolint:errcheck

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	ctx := context.Background()
	db, err := database.NewPool(ctx, cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// repositories
	sessionRepo := repository.NewSessionRepository(db)
	userRepo := repository.NewUserRepository(db)
	invitationRepo := repository.NewInvitationRepository(db)

	// services
	sessionTTL := time.Duration(cfg.App.SessionMaxAge) * time.Second
	authSvc := service.NewAuthService(userRepo, sessionRepo, sessionTTL)
	userSvc := service.NewUserService(userRepo)
	mail := mailer.New(cfg.SMTP)
	invitationSvc := service.NewInvitationService(userRepo, invitationRepo, mail, cfg.App.BaseURL)

	// handlers
	renderer := handler.NewRenderer("app/templates")
	healthHandler := handler.NewHealthHandler(db)
	authHandler := handler.NewAuthHandler(renderer, authSvc, sessionTTL)
	userHandler := handler.NewUserHandler(renderer, userSvc, invitationSvc)
	invitationHandler := handler.NewInvitationHandler(renderer, invitationSvc)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logging(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.CORS(cfg.App.AllowedOrigins))
	r.Use(chiMiddleware.Compress(5))

	r.Get("/health", healthHandler.Check)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("app/static"))))

	// auth routes (no auth required)
	r.Get("/login", authHandler.ShowLogin)
	r.Post("/login", authHandler.Login)
	r.Post("/logout", authHandler.Logout)

	// invitation routes (no auth required)
	r.Get("/invite/{token}", invitationHandler.ShowAcceptForm)
	r.Post("/invite/{token}", invitationHandler.AcceptInvitation)

	// authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(authSvc))

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		})
		r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/users", http.StatusSeeOther)
		})

		// user management (admin only)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAdmin)

			r.Get("/users", userHandler.List)
			r.Get("/users/new", userHandler.New)
			r.Post("/users", userHandler.Create)
			r.Get("/users/export", userHandler.Export)
			r.Get("/users/{id}", userHandler.Show)
			r.Get("/users/{id}/edit", userHandler.Edit)
			r.Put("/users/{id}", userHandler.Update)
			r.Post("/users/{id}", userHandler.Update) // form method override fallback
			r.Put("/users/{id}/status", userHandler.UpdateStatus)
			r.Post("/users/{id}/invite", userHandler.SendInvitation)
		})
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("server starting", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	<-quit
	logger.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", zap.Error(err))
	}
	logger.Info("server stopped")
}

func buildLogger() (*zap.Logger, error) {
	if os.Getenv("APP_ENV") == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
