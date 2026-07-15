package repository

import (
	"context"
	"errors"
	"time"

	"keeneye_practice/app/internal/apperrors"
	"keeneye_practice/app/internal/db"
	"keeneye_practice/app/internal/dbutil"
	"keeneye_practice/app/internal/domain"
	"keeneye_practice/app/internal/tokenutil"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type postgresRegistrationRepository struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewPostgresRegistrationRepository(q *db.Queries, pool *pgxpool.Pool) domain.RegistrationRepository {
	return &postgresRegistrationRepository{q: q, pool: pool}
}

func parseUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return u, nil
}

func uuidToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return u.String()
}

func (r *postgresRegistrationRepository) CreateBatch(ctx context.Context, batchID string, createdBy int32, totalRows int32) (*domain.RegistrationBatch, error) {
	id, err := parseUUID(batchID)
	if err != nil {
		return nil, apperrors.NewValidation("invalid batch id")
	}
	row, err := r.q.CreateRegistrationBatch(ctx, db.CreateRegistrationBatchParams{
		ID:           id,
		CreatedBy:    createdBy,
		TotalRows:    totalRows,
		SuccessCount: 0,
		ErrorCount:   0,
	})
	if err != nil {
		return nil, apperrors.MapPG(err)
	}
	return mapBatch(row), nil
}

func (r *postgresRegistrationRepository) UpdateBatchCounts(ctx context.Context, batchID string, success, failed int32) error {
	id, err := parseUUID(batchID)
	if err != nil {
		return apperrors.NewValidation("invalid batch id")
	}
	return apperrors.MapPG(r.q.UpdateRegistrationBatchCounts(ctx, db.UpdateRegistrationBatchCountsParams{
		ID:           id,
		SuccessCount: success,
		ErrorCount:   failed,
	}))
}

func (r *postgresRegistrationRepository) GetBatch(ctx context.Context, batchID string) (*domain.RegistrationBatch, error) {
	id, err := parseUUID(batchID)
	if err != nil {
		return nil, apperrors.NewValidation("invalid batch id")
	}
	row, err := r.q.GetRegistrationBatch(ctx, id)
	if err != nil {
		return nil, apperrors.MapPG(err)
	}
	return mapBatch(row), nil
}

func (r *postgresRegistrationRepository) ListRequestsByBatch(ctx context.Context, batchID string) ([]domain.RegistrationRequest, error) {
	id, err := parseUUID(batchID)
	if err != nil {
		return nil, apperrors.NewValidation("invalid batch id")
	}
	rows, err := r.q.ListRegistrationRequestsByBatch(ctx, id)
	if err != nil {
		return nil, apperrors.MapPG(err)
	}
	res := make([]domain.RegistrationRequest, len(rows))
	for i, row := range rows {
		res[i] = *mapRequest(row)
	}
	return res, nil
}

func (r *postgresRegistrationRepository) CreateRequestWithOutbox(ctx context.Context, input domain.CreateRegistrationInput) (int32, int32, string, error) {
	batchID, err := parseUUID(input.BatchID)
	if err != nil {
		return 0, 0, "", apperrors.NewValidation("invalid batch id")
	}

	rawToken, tokenHash, err := tokenutil.NewToken()
	if err != nil {
		return 0, 0, "", err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, 0, "", err
	}
	defer dbutil.Rollback(ctx, tx)

	qTx := r.q.WithTx(tx)

	var groupID pgtype.Int4
	var groupName pgtype.Text
	if input.GroupID != nil {
		groupID = pgtype.Int4{Int32: *input.GroupID, Valid: true}
	}
	if input.GroupName != "" {
		groupName = pgtype.Text{String: input.GroupName, Valid: true}
	}

	req, err := qTx.CreateRegistrationRequest(ctx, db.CreateRegistrationRequestParams{
		BatchID:     batchID,
		Email:       input.Email,
		Fio:         input.Fio,
		Role:        db.UserRole(input.Role),
		GroupID:     groupID,
		GroupName:   groupName,
		InviteToken: rawToken,
		TokenHash:   tokenHash,
		ExpiresAt:   input.ExpiresAt,
	})
	if err != nil {
		return 0, 0, "", apperrors.MapPG(err)
	}

	outbox, err := qTx.CreateRegistrationEmailOutbox(ctx, req.ID)
	if err != nil {
		return 0, 0, "", apperrors.MapPG(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, 0, "", err
	}

	return req.ID, outbox.ID, rawToken, nil
}

func (r *postgresRegistrationRepository) CompleteRequestByToken(ctx context.Context, tokenHash, password, phone string) (int32, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer dbutil.Rollback(ctx, tx)

	qTx := r.q.WithTx(tx)

	req, err := qTx.GetRegistrationRequestByTokenHashForUpdate(ctx, tokenHash)
	if err != nil {
		return 0, apperrors.MapPG(err)
	}

	if req.Status != db.RegistrationRequestStatusPending {
		if req.Status == db.RegistrationRequestStatusCompleted {
			return 0, apperrors.ErrConflict
		}
		return 0, apperrors.ErrGone
	}
	if !req.ExpiresAt.Valid || req.ExpiresAt.Time.Before(time.Now()) {
		return 0, apperrors.ErrGone
	}

	user, err := qTx.CreateUser(ctx, db.CreateUserParams{
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         req.Role,
		PhoneNumber:  phone,
	})
	if err != nil {
		return 0, apperrors.MapPG(err)
	}

	switch req.Role {
	case db.UserRoleStudent:
		if !req.GroupID.Valid {
			return 0, apperrors.NewValidation("group_id is required for student")
		}
		_, err = qTx.CreateStudent(ctx, db.CreateStudentParams{
			UserID:  pgtype.Int4{Int32: user.ID, Valid: true},
			GroupID: req.GroupID,
			Fio:     req.Fio,
		})
	case db.UserRoleTeacher:
		_, err = qTx.CreateTeacher(ctx, db.CreateTeacherParams{
			UserID: pgtype.Int4{Int32: user.ID, Valid: true},
			Fio:    req.Fio,
		})
	default:
		return 0, apperrors.NewValidation("invalid role")
	}
	if err != nil {
		return 0, apperrors.MapPG(err)
	}

	if err := qTx.CompleteRegistrationRequest(ctx, req.ID); err != nil {
		return 0, apperrors.MapPG(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (r *postgresRegistrationRepository) RevokeExpired(ctx context.Context, limit int32) ([]int32, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer dbutil.Rollback(ctx, tx)

	qTx := r.q.WithTx(tx)
	ids, err := qTx.RevokeExpiredRegistrationRequests(ctx, limit)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *postgresRegistrationRepository) MarkOutboxSent(ctx context.Context, outboxID int32) error {
	return apperrors.MapPG(r.q.MarkRegistrationEmailOutboxSent(ctx, outboxID))
}

func (r *postgresRegistrationRepository) MarkOutboxFailed(ctx context.Context, outboxID int32, errMsg string) error {
	return apperrors.MapPG(r.q.MarkRegistrationEmailOutboxFailed(ctx, db.MarkRegistrationEmailOutboxFailedParams{
		ID:        outboxID,
		LastError: pgtype.Text{String: errMsg, Valid: true},
	}))
}

func (r *postgresRegistrationRepository) MarkOutboxDead(ctx context.Context, outboxID int32, errMsg string) error {
	return apperrors.MapPG(r.q.MarkRegistrationEmailOutboxDead(ctx, db.MarkRegistrationEmailOutboxDeadParams{
		ID:        outboxID,
		LastError: pgtype.Text{String: errMsg, Valid: true},
	}))
}

func (r *postgresRegistrationRepository) ListOutboxForRetry(ctx context.Context, maxAttempts, limit int32) ([]domain.EmailOutboxItem, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer dbutil.Rollback(ctx, tx)

	qTx := r.q.WithTx(tx)
	rows, err := qTx.ListRegistrationEmailOutboxForRetry(ctx, db.ListRegistrationEmailOutboxForRetryParams{
		Attempts: maxAttempts,
		Limit:    limit,
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	items := make([]domain.EmailOutboxItem, len(rows))
	for i, row := range rows {
		items[i] = domain.EmailOutboxItem{
			OutboxID:  row.ID,
			RequestID: row.RequestID,
			Attempts:  row.Attempts,
		}
	}
	return items, nil
}

func (r *postgresRegistrationRepository) GetRequestByID(ctx context.Context, requestID int32) (*domain.RegistrationRequest, error) {
	row, err := r.q.GetRegistrationRequestByID(ctx, requestID)
	if err != nil {
		return nil, apperrors.MapPG(err)
	}
	return mapRequest(row), nil
}

func (r *postgresRegistrationRepository) CountUsersByEmail(ctx context.Context, email string) (int64, error) {
	return r.q.CountUsersByEmail(ctx, email)
}

func (r *postgresRegistrationRepository) HasPendingRequest(ctx context.Context, email string) (bool, error) {
	_, err := r.q.GetPendingRegistrationRequestByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, apperrors.MapPG(err)
	}
	return true, nil
}

func (r *postgresRegistrationRepository) GetRequestByToken(ctx context.Context, tokenHash string) (*domain.RegistrationRequest, error) {
	row, err := r.q.GetRegistrationRequestByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, apperrors.MapPG(err)
	}
	return mapRequest(row), nil
}

func (r *postgresRegistrationRepository) GetGroupByName(ctx context.Context, name string) (*domain.Group, error) {
	row, err := r.q.GetGroupByName(ctx, name)
	if err != nil {
		return nil, apperrors.MapPG(err)
	}
	return &domain.Group{ID: row.ID, Name: row.Name}, nil
}

func mapBatch(row db.RegistrationBatch) *domain.RegistrationBatch {
	return &domain.RegistrationBatch{
		ID:           uuidToString(row.ID),
		CreatedBy:    row.CreatedBy,
		TotalRows:    row.TotalRows,
		SuccessCount: row.SuccessCount,
		ErrorCount:   row.ErrorCount,
	}
}

func mapRequest(row db.RegistrationRequest) *domain.RegistrationRequest {
	req := &domain.RegistrationRequest{
		ID:          row.ID,
		BatchID:     uuidToString(row.BatchID),
		Email:       row.Email,
		Fio:         row.Fio,
		Role:        string(row.Role),
		Status:      string(row.Status),
		ExpiresAt:   row.ExpiresAt,
		CompletedAt: row.CompletedAt,
	}
	if row.GroupID.Valid {
		gid := row.GroupID.Int32
		req.GroupID = &gid
	}
	if row.GroupName.Valid {
		req.GroupName = row.GroupName.String
	}
	req.InviteToken = row.InviteToken
	return req
}
