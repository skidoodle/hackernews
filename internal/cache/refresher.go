package cache

import (
	"context"
	"log/slog"
	"time"
)

type IDListFetcher interface {
	GetStoryIDs(ctx context.Context, storyType string) ([]int, error)
}

type Refresher struct {
	client   IDListFetcher
	logger   *slog.Logger
	interval time.Duration
	stop     chan struct{}
}

func NewRefresher(client IDListFetcher, logger *slog.Logger, interval time.Duration) *Refresher {
	return &Refresher{
		client:   client,
		logger:   logger,
		interval: interval,
		stop:     make(chan struct{}),
	}
}

func (r *Refresher) Start() {
	r.logger.Info("starting background ID list cache refresher", "interval", r.interval)
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	r.refresh()

	for {
		select {
		case <-ticker.C:
			r.refresh()
		case <-r.stop:
			r.logger.Info("stopping background cache refresher")
			return
		}
	}
}

func (r *Refresher) Stop() {
	close(r.stop)
}

func (r *Refresher) refresh() {
	r.logger.Info("performing background ID list cache refresh")
	storyTypes := []string{"top", "new", "ask", "show", "job"}
	ctx := context.Background()

	for _, storyType := range storyTypes {
		if _, err := r.client.GetStoryIDs(ctx, storyType); err != nil {
			r.logger.Error("failed to refresh ID list cache", "type", storyType, "error", err)
		}
	}
}
