package main

import (
	"database/sql"
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
	if value := os.Getenv("WRITER_ADDR"); value != "" {
		addr = value
	}

	deps, cleanup, err := buildDependencies()
	if err != nil {
		slog.Error("initialize writer", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	server := &http.Server{
		Addr:    addr,
		Handler: app.New(deps),
	}

	slog.Info("starting writer", "addr", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func buildDependencies() (*app.Dependencies, func(), error) {
	dbPath := getenvDefault("WRITER_DB_PATH", "writer.db")
	migrationsPath := getenvDefault("WRITER_MIGRATIONS_PATH", "migrations")
	staticPath := getenvDefault("WRITER_STATIC_PATH", "frontend/dist")
	jwtSecret, err := jwtSecret()
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
	projectService := project.NewService(db)
	agentService := agent.NewService(db, projectService, nil)
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

func jwtSecret() (string, error) {
	if secret := os.Getenv("WRITER_JWT_SECRET"); secret != "" {
		if err := auth.ValidateSecret(secret); err != nil {
			return "", fmt.Errorf("WRITER_JWT_SECRET: %w", err)
		}
		return secret, nil
	}
	if secret := os.Getenv("WRITER_SECRET"); secret != "" {
		if err := auth.ValidateSecret(secret); err != nil {
			return "", fmt.Errorf("WRITER_SECRET: %w", err)
		}
		return secret, nil
	}
	return "", fmt.Errorf("JWT secret is not configured; set WRITER_JWT_SECRET or WRITER_SECRET")
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

func getenvDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
