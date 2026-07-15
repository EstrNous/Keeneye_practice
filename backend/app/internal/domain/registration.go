package domain

import (
	"context"
	"io"

	"github.com/jackc/pgx/v5/pgtype"
)

type RegistrationRequest struct {
	ID          int32
	BatchID     string
	Email       string
	Fio         string
	Role        string
	GroupID     *int32
	GroupName   string
	InviteToken string
	Status      string
	ExpiresAt   pgtype.Timestamp
	CompletedAt pgtype.Timestamp
}

type RegistrationBatch struct {
	ID           string
	CreatedBy    int32
	TotalRows    int32
	SuccessCount int32
	ErrorCount   int32
}

type EmailRetryMessage struct {
	RequestID int32
	OutboxID  int32
	Email     string
	TokenRaw  string
}

type EmailRetryBackend interface {
	Enqueue(ctx context.Context, msg EmailRetryMessage) error
	Start(ctx context.Context) error
}

type RegistrationRepository interface {
	CreateBatch(ctx context.Context, batchID string, createdBy int32, totalRows int32) (*RegistrationBatch, error)
	UpdateBatchCounts(ctx context.Context, batchID string, success, failed int32) error
	GetBatch(ctx context.Context, batchID string) (*RegistrationBatch, error)
	ListRequestsByBatch(ctx context.Context, batchID string) ([]RegistrationRequest, error)

	CreateRequestWithOutbox(ctx context.Context, input CreateRegistrationInput) (requestID int32, outboxID int32, rawToken string, err error)
	CompleteRequestByToken(ctx context.Context, tokenHash, password, phone string) (int32, error)
	RevokeExpired(ctx context.Context, limit int32) ([]int32, error)

	MarkOutboxSent(ctx context.Context, outboxID int32) error
	MarkOutboxFailed(ctx context.Context, outboxID int32, errMsg string) error
	MarkOutboxDead(ctx context.Context, outboxID int32, errMsg string) error
	ListOutboxForRetry(ctx context.Context, maxAttempts, limit int32) ([]EmailOutboxItem, error)
	GetRequestByID(ctx context.Context, requestID int32) (*RegistrationRequest, error)
	CountUsersByEmail(ctx context.Context, email string) (int64, error)
	HasPendingRequest(ctx context.Context, email string) (bool, error)
	GetRequestByToken(ctx context.Context, tokenHash string) (*RegistrationRequest, error)
	GetGroupByName(ctx context.Context, name string) (*Group, error)
}

type CreateRegistrationInput struct {
	BatchID   string
	Email     string
	Fio       string
	Role      string
	GroupID   *int32
	GroupName string
	ExpiresAt pgtype.Timestamp
}

type EmailOutboxItem struct {
	OutboxID  int32
	RequestID int32
	Attempts  int32
}

type EmailResender interface {
	ResendEmail(ctx context.Context, msg EmailRetryMessage) error
}

type RegistrationService interface {
	ProcessBatchCSV(ctx context.Context, createdBy int32, r io.Reader) (*BatchUploadResult, error)
	GetBatchStatus(ctx context.Context, batchID string) (*BatchStatusResult, error)
	PreviewComplete(ctx context.Context, token string) (*CompletePreview, error)
	CompleteRegistration(ctx context.Context, token, password, phone string) (int32, error)
	SendRegistrationEmail(ctx context.Context, requestID, outboxID int32, email, rawToken string) error
	EmailResender
}

type BatchUploadResult struct {
	BatchID string
	Total   int
	Created int
	Failed  int
	Errors  []BatchRowError
}

type BatchRowError struct {
	Row     int
	Email   string
	Code    string
	Message string
}

type BatchStatusResult struct {
	Batch    RegistrationBatch
	Requests []RegistrationRequest
}

type CompletePreview struct {
	Email     string
	Fio       string
	Role      string
	GroupName string
}
