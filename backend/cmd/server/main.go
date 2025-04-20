package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Sosokker/todolist-backend/internal/api"
	"github.com/Sosokker/todolist-backend/internal/config"
	"github.com/Sosokker/todolist-backend/internal/repository"
	"github.com/Sosokker/todolist-backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	configPath := flag.String("config", ".", "Path to the config directory or file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	logger := setupLogger(cfg.Log)
	slog.SetDefault(logger)

	logger.Info("Starting Todolist Backend Service", "version", "1.2.0")
	logger.Debug("Configuration loaded", "config", cfg)

	pool, err := repository.NewConnectionPool(cfg.Database)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := runMigrations(cfg.Database.URL); err != nil {
		logger.Error("Database migration failed", "error", err)
		os.Exit(1)
	}

	repoRegistry := repository.NewRepositoryRegistry(pool)

	var storageService service.FileStorageService
	switch cfg.Storage.Type {
	case "local":
		storageService, err = service.NewLocalStorageService(cfg.Storage.Local, logger)
	case "gcs":
		storageService, err = service.NewGCStorageService(cfg.Storage.GCS, logger)
	default:
		err = fmt.Errorf("unsupported storage type: %s", cfg.Storage.Type)
	}
	if err != nil {
		logger.Error("Failed to initialize storage service", "error", err, "type", cfg.Storage.Type)
		os.Exit(1)
	}

	authService := service.NewAuthService(repoRegistry.UserRepo, cfg)
	userService := service.NewUserService(repoRegistry.UserRepo)
	tagService := service.NewTagService(repoRegistry.TagRepo)
	subtaskService := service.NewSubtaskService(repoRegistry.SubtaskRepo)
	todoService := service.NewTodoService(repoRegistry.TodoRepo, tagService, subtaskService, storageService)

	services := &service.ServiceRegistry{
		Auth:    authService,
		User:    userService,
		Tag:     tagService,
		Todo:    todoService,
		Subtask: subtaskService,
		Storage: storageService,
	}

	apiHandler := api.NewApiHandler(services, cfg, logger)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(NewStructuredLogger(logger))
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://your-frontend-domain.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route(cfg.Server.BasePath, func(subr chi.Router) {
		subr.Post("/auth/signup", apiHandler.SignupUserApi)
		subr.Post("/auth/login", apiHandler.LoginUserApi)
		subr.Get("/auth/google/login", apiHandler.InitiateGoogleLogin)
		subr.Get("/auth/google/callback", apiHandler.HandleGoogleCallback)

		subr.Group(func(prot chi.Router) {
			prot.Use(api.AuthMiddleware(authService, cfg))
			api.HandlerFromMux(apiHandler, prot)
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			http.Error(w, "Health check failed: DB ping error", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		logger.Info("Server starting", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("Server exited gracefully")
}

func setupLogger(cfg config.LogConfig) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug,
	}

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

func runMigrations(databaseURL string) error {
	if databaseURL == "" {
		return errors.New("database URL is required for migrations")
	}
	migrationPath := "file://migrations"

	m, err := migrate.New(migrationPath, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		sourceErr, dbErr := m.Close()
		slog.Error("Migration close errors", "source_error", sourceErr, "db_error", dbErr)
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		slog.Info("No new migrations to apply")
	} else {
		slog.Info("Database migrations applied successfully")
	}

	sourceErr, dbErr := m.Close()
	if sourceErr != nil {
		slog.Error("Error closing migration source", "error", sourceErr)
	}
	if dbErr != nil {
		slog.Error("Error closing migration database connection", "error", dbErr)
	}

	return nil
}

func NewStructuredLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			reqLogger := logger.With(
				slog.String("proto", r.Proto),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)

			defer func() {
				reqLogger.Info("request completed",
					slog.Int("status", ww.Status()),
					slog.Duration("latency", time.Since(start)),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
