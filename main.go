package main

import (
	"log/slog"
	"net/http"
	"os"

	"git.inkyquill.net/inky/writer/app"
)

func main() {
	addr := ":8080"
	if value := os.Getenv("WRITER_ADDR"); value != "" {
		addr = value
	}

	server := &http.Server{
		Addr:    addr,
		Handler: app.New(&app.Dependencies{}),
	}

	slog.Info("starting writer", "addr", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
