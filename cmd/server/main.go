package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hackernews/internal/cache"
	"hackernews/internal/config"
	"hackernews/internal/handler"
	"hackernews/internal/hn"
	"hackernews/internal/view"
)

//go:embed all:web
var webFS embed.FS

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(logger); err != nil {
		logger.Error("application startup error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg := config.New()

	templateSubFS, err := fs.Sub(webFS, "web/template")
	if err != nil {
		return fmt.Errorf("failed to create sub-filesystem for templates: %w", err)
	}

	templateCache, err := view.NewTemplateCache(templateSubFS)
	if err != nil {
		return fmt.Errorf("failed to create template cache: %w", err)
	}

	staticSubFS, err := fs.Sub(webFS, "web/static")
	if err != nil {
		return fmt.Errorf("failed to create sub-filesystem for static assets: %w", err)
	}

	hnClient := hn.NewClient(logger, cfg)

	refresher := cache.NewRefresher(hnClient, logger, 90*time.Second)
	go refresher.Start()

	app := &handler.App{
		Logger:        logger,
		Config:        cfg,
		HackerNews:    hnClient,
		TemplateCache: templateCache,
		StaticFS:      staticSubFS,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      app.Routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		logger.Info("shutting down server", "signal", s.String())
		refresher.Stop()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	logger.Info("starting server", "addr", srv.Addr)

	err = srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}

	err = <-shutdownError
	if err != nil {
		return fmt.Errorf("error during shutdown: %w", err)
	}

	logger.Info("server stopped")
	return nil
}
