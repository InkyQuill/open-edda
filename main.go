package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"git.inkyquill.net/inky/writer/agent"
	"git.inkyquill.net/inky/writer/app"
	"git.inkyquill.net/inky/writer/auth"
	"git.inkyquill.net/inky/writer/project"
	"git.inkyquill.net/inky/writer/skill"
	"git.inkyquill.net/inky/writer/store"
	"github.com/pressly/goose/v3"
)

func main() {
	addr := ":8080"
	if value := firstEnv("OPEN_EDDA_ADDR", "WRITER_ADDR"); value != "" {
		addr = value
	}

	deps, cleanup, err := buildDependencies()
	if err != nil {
		slog.Error("initialize open edda", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	server := &http.Server{
		Addr:    addr,
		Handler: app.New(deps),
	}

	slog.Info("starting open edda", "addr", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func buildDependencies() (*app.Dependencies, func(), error) {
	dbPath := getenvDefault("OPEN_EDDA_DB_PATH", "edda.db", "WRITER_DB_PATH")
	migrationsPath := getenvDefault("OPEN_EDDA_MIGRATIONS_PATH", "migrations", "WRITER_MIGRATIONS_PATH")
	staticPath := getenvDefault("OPEN_EDDA_STATIC_PATH", "frontend/dist", "WRITER_STATIC_PATH")
	jwtSecret, err := jwtSecret()
	if err != nil {
		return nil, func() {}, err
	}
	apiKeyEncryptionSecret, err := apiKeyEncryptionSecret()
	if err != nil {
		return nil, func() {}, err
	}

	db, err := store.Open(dbPath)
	if err != nil {
		return nil, func() {}, err
	}
	cleanup := func() { _ = db.Close() }

	if err := migrate(db, migrationsPath); err != nil {
		cleanup()
		return nil, func() {}, err
	}
	if err := validateStaticPath(staticPath); err != nil {
		cleanup()
		return nil, func() {}, err
	}

	authService := auth.NewService(db, jwtSecret)
	if err := bootstrapSingleUser(authService); err != nil {
		cleanup()
		return nil, func() {}, err
	}
	projectService := project.NewService(db)
	agentService := agent.NewService(db, projectService, nil)
	agentService.SetEncryptionSecret(apiKeyEncryptionSecret)
	skillService := skill.NewService(db)
	agentService.SetSkillService(skillService)

	return &app.Dependencies{
		AuthService:    authService,
		ProjectService: projectService,
		AgentService:   agentService,
		SkillService:   skillService,
		StaticFS:       os.DirFS(staticPath),
	}, cleanup, nil
}

func bootstrapSingleUser(authService *auth.Service) error {
	email := firstEnv("OPEN_EDDA_BOOTSTRAP_EMAIL", "WRITER_BOOTSTRAP_EMAIL")
	password := firstEnv("OPEN_EDDA_BOOTSTRAP_PASSWORD", "WRITER_BOOTSTRAP_PASSWORD")
	if email == "" && password == "" {
		return nil
	}
	if email == "" || password == "" {
		return fmt.Errorf("single-user bootstrap requires both OPEN_EDDA_BOOTSTRAP_EMAIL and OPEN_EDDA_BOOTSTRAP_PASSWORD")
	}

	if _, err := authService.Register(context.Background(), email, password); err != nil {
		if errors.Is(err, auth.ErrEmailTaken) {
			return nil
		}
		return fmt.Errorf("bootstrap single user: %w", err)
	}
	return nil
}

func jwtSecret() (string, error) {
	if secret := firstEnv("OPEN_EDDA_JWT_SECRET", "WRITER_JWT_SECRET"); secret != "" {
		if err := auth.ValidateSecret(secret); err != nil {
			return "", fmt.Errorf("JWT secret: %w", err)
		}
		return secret, nil
	}
	if secret := firstEnv("OPEN_EDDA_SECRET", "WRITER_SECRET"); secret != "" {
		if err := auth.ValidateSecret(secret); err != nil {
			return "", fmt.Errorf("application secret: %w", err)
		}
		return secret, nil
	}
	return "", fmt.Errorf("JWT secret is not configured; set OPEN_EDDA_JWT_SECRET or OPEN_EDDA_SECRET")
}

func apiKeyEncryptionSecret() (string, error) {
	if secret := firstEnv("OPEN_EDDA_API_KEY_ENCRYPTION_SECRET", "WRITER_API_KEY_ENCRYPTION_SECRET"); secret != "" {
		if err := auth.ValidateSecret(secret); err != nil {
			return "", fmt.Errorf("API key encryption secret: %w", err)
		}
		return secret, nil
	}
	return "", fmt.Errorf("API key encryption secret is not configured; set OPEN_EDDA_API_KEY_ENCRYPTION_SECRET or WRITER_API_KEY_ENCRYPTION_SECRET")
}

func validateStaticPath(staticPath string) error {
	info, err := os.Stat(filepath.Join(staticPath, "index.html"))
	if err != nil {
		return fmt.Errorf("frontend build missing at %q: %w", staticPath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("frontend index at %q is a directory", staticPath)
	}
	return nil
}

func migrate(db *sql.DB, migrationsPath string) error {
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}
	return goose.Up(db, migrationsPath)
}

func getenvDefault(name string, fallback string, legacyNames ...string) string {
	if value := firstEnv(append([]string{name}, legacyNames...)...); value != "" {
		return value
	}
	return fallback
}

func firstEnv(names ...string) string {
	for _, name := range names {
		if value := os.Getenv(name); value != "" {
			return value
		}
	}
	return ""
}
