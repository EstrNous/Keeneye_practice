package worker

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"keeneye_practice/app/internal/config"
	"keeneye_practice/app/internal/domain"
)

type EmailSender interface {
	ResendEmail(ctx context.Context, msg domain.EmailRetryMessage) error
}

type EmailRetryWorker struct {
	repo        domain.RegistrationRepository
	sender      EmailSender
	cfg         *config.Config
	interval    time.Duration
	failRate    float64
	maxAttempts int32
	batchLimit  int32
	rng         *rand.Rand
}

func NewEmailRetryWorker(repo domain.RegistrationRepository, sender EmailSender, cfg *config.Config) *EmailRetryWorker {
	return &EmailRetryWorker{
		repo:        repo,
		sender:      sender,
		cfg:         cfg,
		interval:    cfg.EmailRetryWorkerInterval,
		failRate:    cfg.EmailRetrySimulateFailRate,
		maxAttempts: cfg.EmailRetryMaxAttempts,
		batchLimit:  20,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (w *EmailRetryWorker) Enqueue(ctx context.Context, msg domain.EmailRetryMessage) error {
	// worker backend processes outbox via polling; no-op enqueue
	return nil
}

func (w *EmailRetryWorker) tick(ctx context.Context) {
	items, err := w.repo.ListOutboxForRetry(ctx, w.maxAttempts, w.batchLimit)
	if err != nil {
		slog.ErrorContext(ctx, "email retry worker list failed", "error", err)
		return
	}

	for _, item := range items {
		w.processItem(ctx, item)
	}
}

func (w *EmailRetryWorker) processItem(ctx context.Context, item domain.EmailOutboxItem) {
	req, err := w.repo.GetRequestByID(ctx, item.RequestID)
	if err != nil {
		slog.ErrorContext(ctx, "email retry get request failed", "request_id", item.RequestID, "error", err)
		return
	}
	if req.Status != "pending" {
		return
	}

	if w.failRate > 0 && w.rng.Float64() < w.failRate {
		errMsg := "simulated random send failure"
		slog.WarnContext(ctx, "simulated email retry failure", "outbox_id", item.OutboxID)
		if item.Attempts+1 >= w.maxAttempts {
			_ = w.repo.MarkOutboxDead(ctx, item.OutboxID, errMsg)
		} else {
			_ = w.repo.MarkOutboxFailed(ctx, item.OutboxID, errMsg)
		}
		return
	}

	msg := domain.EmailRetryMessage{
		RequestID: item.RequestID,
		OutboxID:  item.OutboxID,
		Email:     req.Email,
		TokenRaw:  req.InviteToken,
	}

	if err := w.sender.ResendEmail(ctx, msg); err != nil {
		slog.ErrorContext(ctx, "email retry send failed", "outbox_id", item.OutboxID, "error", err)
		if item.Attempts+1 >= w.maxAttempts {
			_ = w.repo.MarkOutboxDead(ctx, item.OutboxID, err.Error())
		} else {
			_ = w.repo.MarkOutboxFailed(ctx, item.OutboxID, err.Error())
		}
		return
	}

	slog.InfoContext(ctx, "email retry sent", "outbox_id", item.OutboxID, "request_id", item.RequestID)
}

func (w *EmailRetryWorker) Start(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

// TickForTest exposes one retry iteration for unit tests.
func (w *EmailRetryWorker) TickForTest(ctx context.Context) {
	w.tick(ctx)
}
