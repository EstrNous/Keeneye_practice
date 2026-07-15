package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"keeneye_practice/app/internal/apperrors"
	"keeneye_practice/app/internal/config"
	"keeneye_practice/app/internal/csvparser"
	"keeneye_practice/app/internal/domain"
	"keeneye_practice/app/internal/mail"
	"keeneye_practice/app/internal/tokenutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type registrationService struct {
	repo         domain.RegistrationRepository
	mailer       mail.Mailer
	retryBackend domain.EmailRetryBackend
	cfg          *config.Config
}

func NewRegistrationService(
	repo domain.RegistrationRepository,
	mailer mail.Mailer,
	retryBackend domain.EmailRetryBackend,
	cfg *config.Config,
) domain.RegistrationService {
	return &registrationService{
		repo:         repo,
		mailer:       mailer,
		retryBackend: retryBackend,
		cfg:          cfg,
	}
}

func (s *registrationService) ProcessBatchCSV(ctx context.Context, createdBy int32, r io.Reader) (*domain.BatchUploadResult, error) {
	rows, parseErrors, err := csvparser.Parse(r)
	if err != nil {
		return nil, apperrors.NewValidation(err.Error())
	}

	batchID := uuid.New().String()
	total := len(rows) + len(parseErrors)
	if _, err := s.repo.CreateBatch(ctx, batchID, createdBy, int32(total)); err != nil {
		return nil, err
	}

	result := &domain.BatchUploadResult{
		BatchID: batchID,
		Total:   total,
		Errors:  make([]domain.BatchRowError, 0, len(parseErrors)),
	}

	for _, pe := range parseErrors {
		result.Errors = append(result.Errors, domain.BatchRowError{
			Row: pe.Row, Email: pe.Email, Code: pe.Code, Message: pe.Message,
		})
		result.Failed++
	}

	expiresAt := pgtype.Timestamp{Time: time.Now().Add(s.cfg.RegistrationLinkTTL), Valid: true}

	for _, row := range rows {
		rowErr := s.processRow(ctx, batchID, row, expiresAt)
		if rowErr != nil {
			result.Errors = append(result.Errors, *rowErr)
			result.Failed++
			continue
		}
		result.Created++
	}

	if err := s.repo.UpdateBatchCounts(ctx, batchID, int32(result.Created), int32(result.Failed)); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *registrationService) processRow(ctx context.Context, batchID string, row csvparser.Row, expiresAt pgtype.Timestamp) *domain.BatchRowError {
	count, err := s.repo.CountUsersByEmail(ctx, row.Email)
	if err != nil {
		return &domain.BatchRowError{Row: row.RowNumber, Email: row.Email, Code: "internal_error", Message: err.Error()}
	}
	if count > 0 {
		return &domain.BatchRowError{Row: row.RowNumber, Email: row.Email, Code: "conflict", Message: "email already registered"}
	}

	pending, err := s.repo.HasPendingRequest(ctx, row.Email)
	if err != nil {
		return &domain.BatchRowError{Row: row.RowNumber, Email: row.Email, Code: "internal_error", Message: err.Error()}
	}
	if pending {
		return &domain.BatchRowError{Row: row.RowNumber, Email: row.Email, Code: "conflict", Message: "pending registration request already exists"}
	}

	var groupID *int32
	if row.Role == "student" {
		group, err := s.repo.GetGroupByName(ctx, row.GroupName)
		if err != nil {
			if errors.Is(err, apperrors.ErrNotFound) {
				return &domain.BatchRowError{Row: row.RowNumber, Email: row.Email, Code: "not_found", Message: fmt.Sprintf("group %q not found", row.GroupName)}
			}
			return &domain.BatchRowError{Row: row.RowNumber, Email: row.Email, Code: "internal_error", Message: err.Error()}
		}
		groupID = &group.ID
	}

	requestID, outboxID, rawToken, err := s.repo.CreateRequestWithOutbox(ctx, domain.CreateRegistrationInput{
		BatchID:   batchID,
		Email:     row.Email,
		Fio:       row.FIO,
		Role:      row.Role,
		GroupID:   groupID,
		GroupName: row.GroupName,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		code := "internal_error"
		if errors.Is(err, apperrors.ErrConflict) {
			code = "conflict"
		} else if errors.Is(err, apperrors.ErrValidation) {
			code = "validation_error"
		}
		return &domain.BatchRowError{Row: row.RowNumber, Email: row.Email, Code: code, Message: err.Error()}
	}

	if err := s.SendRegistrationEmail(ctx, requestID, outboxID, row.Email, rawToken); err != nil {
		// request created; email will be retried
		_ = err
	}

	return nil
}

func (s *registrationService) GetBatchStatus(ctx context.Context, batchID string) (*domain.BatchStatusResult, error) {
	batch, err := s.repo.GetBatch(ctx, batchID)
	if err != nil {
		return nil, err
	}
	requests, err := s.repo.ListRequestsByBatch(ctx, batchID)
	if err != nil {
		return nil, err
	}
	return &domain.BatchStatusResult{Batch: *batch, Requests: requests}, nil
}

func (s *registrationService) PreviewComplete(ctx context.Context, token string) (*domain.CompletePreview, error) {
	req, err := s.lookupByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return &domain.CompletePreview{
		Email:     req.Email,
		Fio:       req.Fio,
		Role:      req.Role,
		GroupName: req.GroupName,
	}, nil
}

func (s *registrationService) CompleteRegistration(ctx context.Context, token, password, phone string) (int32, error) {
	tokenHash := tokenutil.Hash(token)
	return s.repo.CompleteRequestByToken(ctx, tokenHash, password, phone)
}

func (s *registrationService) SendRegistrationEmail(ctx context.Context, requestID, outboxID int32, email, rawToken string) error {
	link := fmt.Sprintf("%s/api/v1/auth/complete-registration?token=%s", s.cfg.RegistrationBaseURL, rawToken)

	if err := s.mailer.SendRegistrationLink(ctx, email, link); err != nil {
		_ = s.repo.MarkOutboxFailed(ctx, outboxID, err.Error())
		_ = s.retryBackend.Enqueue(ctx, domain.EmailRetryMessage{
			RequestID: requestID,
			OutboxID:  outboxID,
			Email:     email,
			TokenRaw:  rawToken,
		})
		return err
	}

	if err := s.repo.MarkOutboxSent(ctx, outboxID); err != nil {
		return err
	}
	return nil
}

func (s *registrationService) lookupByToken(ctx context.Context, token string) (*domain.RegistrationRequest, error) {
	tokenHash := tokenutil.Hash(token)
	req, err := s.repo.GetRequestByToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}

	switch req.Status {
	case "completed":
		return nil, apperrors.ErrConflict
	case "revoked":
		return nil, apperrors.ErrGone
	case "pending":
		if !req.ExpiresAt.Valid || req.ExpiresAt.Time.Before(time.Now()) {
			return nil, apperrors.ErrGone
		}
		return req, nil
	default:
		return nil, apperrors.ErrGone
	}
}

func (s *registrationService) BuildRegistrationLink(rawToken string) string {
	return fmt.Sprintf("%s/api/v1/auth/complete-registration?token=%s", s.cfg.RegistrationBaseURL, rawToken)
}

// ResendEmail is used by retry backends.
func (s *registrationService) ResendEmail(ctx context.Context, msg domain.EmailRetryMessage) error {
	req, err := s.repo.GetRequestByID(ctx, msg.RequestID)
	if err != nil {
		return err
	}
	if req.Status != "pending" {
		return nil
	}
	if !req.ExpiresAt.Valid || req.ExpiresAt.Time.Before(time.Now()) {
		return nil
	}

	rawToken := msg.TokenRaw
	if rawToken == "" {
		rawToken = req.InviteToken
	}

	link := s.BuildRegistrationLink(rawToken)
	if err := s.mailer.SendRegistrationLink(ctx, msg.Email, link); err != nil {
		return err
	}
	return s.repo.MarkOutboxSent(ctx, msg.OutboxID)
}
