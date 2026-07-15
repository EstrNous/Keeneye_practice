package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"keeneye_practice/app/internal/config"
	"keeneye_practice/app/internal/domain"

	kafkago "github.com/segmentio/kafka-go"
)

type EmailSender interface {
	ResendEmail(ctx context.Context, msg domain.EmailRetryMessage) error
}

type EmailRetryBackend struct {
	writer      *kafkago.Writer
	reader      *kafkago.Reader
	repo        domain.RegistrationRepository
	sender      EmailSender
	failRate    float64
	maxAttempts int32
	rng         *rand.Rand
}

func NewEmailRetryBackend(cfg *config.Config, repo domain.RegistrationRepository, sender EmailSender) *EmailRetryBackend {
	brokers := cfg.KafkaBrokers
	if len(brokers) == 0 {
		brokers = []string{"kafka:9092"}
	}

	return &EmailRetryBackend{
		writer: &kafkago.Writer{
			Addr:     kafkago.TCP(brokers...),
			Topic:    cfg.KafkaEmailTopic,
			Balancer: &kafkago.LeastBytes{},
		},
		reader: kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:  brokers,
			Topic:    cfg.KafkaEmailTopic,
			GroupID:  cfg.KafkaConsumerGroup,
			MinBytes: 1,
			MaxBytes: 10e6,
		}),
		repo:        repo,
		sender:      sender,
		failRate:    cfg.EmailRetrySimulateFailRate,
		maxAttempts: cfg.EmailRetryMaxAttempts,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (b *EmailRetryBackend) Enqueue(ctx context.Context, msg domain.EmailRetryMessage) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return b.writer.WriteMessages(ctx, kafkago.Message{Value: payload})
}

func (b *EmailRetryBackend) Start(ctx context.Context) error {
	defer func() {
		_ = b.reader.Close()
		_ = b.writer.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		m, err := b.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			slog.ErrorContext(ctx, "kafka fetch failed", "error", err)
			time.Sleep(time.Second)
			continue
		}

		if err := b.handleMessage(ctx, m); err != nil {
			slog.WarnContext(ctx, "kafka message not committed", "error", err)
			continue
		}

		if err := b.reader.CommitMessages(ctx, m); err != nil {
			slog.ErrorContext(ctx, "kafka commit failed", "error", err)
		}
	}
}

func (b *EmailRetryBackend) handleMessage(ctx context.Context, m kafkago.Message) error {
	var msg domain.EmailRetryMessage
	if err := json.Unmarshal(m.Value, &msg); err != nil {
		slog.ErrorContext(ctx, "invalid kafka payload", "error", err)
		return nil
	}

	req, err := b.repo.GetRequestByID(ctx, msg.RequestID)
	if err != nil {
		return err
	}
	if req.Status != "pending" {
		return nil
	}

	if b.failRate > 0 && b.rng.Float64() < b.failRate {
		errMsg := "simulated random kafka processing failure"
		slog.WarnContext(ctx, "simulated kafka retry failure", "outbox_id", msg.OutboxID)
		_ = b.repo.MarkOutboxFailed(ctx, msg.OutboxID, errMsg)
		return fmt.Errorf("%s", errMsg)
	}

	if msg.TokenRaw == "" {
		msg.TokenRaw = req.InviteToken
	}
	if msg.Email == "" {
		msg.Email = req.Email
	}

	if err := b.sender.ResendEmail(ctx, msg); err != nil {
		attempts := int32(0)
		items, listErr := b.repo.ListOutboxForRetry(ctx, b.maxAttempts+1, 1)
		if listErr == nil {
			for _, item := range items {
				if item.OutboxID == msg.OutboxID {
					attempts = item.Attempts
				}
			}
		}
		if attempts+1 >= b.maxAttempts {
			_ = b.repo.MarkOutboxDead(ctx, msg.OutboxID, err.Error())
		} else {
			_ = b.repo.MarkOutboxFailed(ctx, msg.OutboxID, err.Error())
		}
		return err
	}

	slog.InfoContext(ctx, "kafka email retry sent", "outbox_id", msg.OutboxID)
	return nil
}
