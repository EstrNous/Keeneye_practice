package worker

import (
	"context"
	"log/slog"
	"time"

	"keeneye_practice/app/internal/config"
	"keeneye_practice/app/internal/domain"
)

type ExpirationWorker struct {
	repo     domain.RegistrationRepository
	interval time.Duration
	batch    int32
}

func NewExpirationWorker(repo domain.RegistrationRepository, cfg *config.Config) *ExpirationWorker {
	return &ExpirationWorker{
		repo:     repo,
		interval: cfg.ExpirationWorkerInterval,
		batch:    100,
	}
}

func (w *ExpirationWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *ExpirationWorker) tick(ctx context.Context) {
	ids, err := w.repo.RevokeExpired(ctx, w.batch)
	if err != nil {
		slog.ErrorContext(ctx, "expiration worker failed", "error", err)
		return
	}
	if len(ids) > 0 {
		slog.InfoContext(ctx, "revoked expired registration requests", "count", len(ids))
	}
}
