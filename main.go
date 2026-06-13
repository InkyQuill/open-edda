package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"git.inkyquill.net/inky/writer/agent"
	"git.inkyquill.net/inky/writer/app"
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

	db, err := store.Open(dbPath)
	if err != nil {
		return nil, func() {}, err
	}
	cleanup := func() { _ = db.Close() }

	if err := migrate(db, migrationsPath); err != nil {
		cleanup()
		return nil, func() {}, err
	}
	if err := ensurePlaceholderAuthor(db); err != nil {
		cleanup()
		return nil, func() {}, err
	}
	if err := validateStaticPath(staticPath); err != nil {
		cleanup()
		return nil, func() {}, err
	}

	projectService := project.NewService(db)
	agentService := agent.NewService(db, projectService, nil)
	skillService := skill.NewService(db)
	agentService.SetSkillService(skillService)

	return &app.Dependencies{
		ProjectService: projectService,
		AgentService:   agentService,
		SkillService:   skillService,
		StaticFS:       os.DirFS(staticPath),
	}, cleanup, nil
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

func ensurePlaceholderAuthor(db *sql.DB) error {
	_, err := db.Exec(`
		INSERT OR IGNORE INTO authors (id, email, password_hash, created_at)
		VALUES (?, ?, ?, ?)
	`, "author-1", "author@example.invalid", "placeholder", time.Now().UTC().Format(time.RFC3339Nano))
	return err
}

func getenvDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
