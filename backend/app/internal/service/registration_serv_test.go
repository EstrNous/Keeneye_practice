package service_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"keeneye_practice/app/internal/apperrors"
	"keeneye_practice/app/internal/config"
	"keeneye_practice/app/internal/domain"
	"keeneye_practice/app/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockRegRepo struct {
	mock.Mock
}

func (m *mockRegRepo) CreateBatch(ctx context.Context, batchID string, createdBy int32, totalRows int32) (*domain.RegistrationBatch, error) {
	args := m.Called(ctx, batchID, createdBy, totalRows)
	return args.Get(0).(*domain.RegistrationBatch), args.Error(1)
}
func (m *mockRegRepo) UpdateBatchCounts(ctx context.Context, batchID string, success, failed int32) error {
	return m.Called(ctx, batchID, success, failed).Error(0)
}
func (m *mockRegRepo) GetBatch(ctx context.Context, batchID string) (*domain.RegistrationBatch, error) {
	return nil, nil
}
func (m *mockRegRepo) ListRequestsByBatch(ctx context.Context, batchID string) ([]domain.RegistrationRequest, error) {
	return nil, nil
}
func (m *mockRegRepo) CreateRequestWithOutbox(ctx context.Context, input domain.CreateRegistrationInput) (int32, int32, string, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(int32), args.Get(1).(int32), args.Get(2).(string), args.Error(3)
}
func (m *mockRegRepo) CompleteRequestByToken(ctx context.Context, tokenHash, password, phone string) (int32, error) {
	return 0, nil
}
func (m *mockRegRepo) RevokeExpired(ctx context.Context, limit int32) ([]int32, error) {
	return nil, nil
}
func (m *mockRegRepo) MarkOutboxSent(ctx context.Context, outboxID int32) error {
	return m.Called(ctx, outboxID).Error(0)
}
func (m *mockRegRepo) MarkOutboxFailed(ctx context.Context, outboxID int32, errMsg string) error {
	return nil
}
func (m *mockRegRepo) MarkOutboxDead(ctx context.Context, outboxID int32, errMsg string) error {
	return nil
}
func (m *mockRegRepo) ListOutboxForRetry(ctx context.Context, maxAttempts, limit int32) ([]domain.EmailOutboxItem, error) {
	return nil, nil
}
func (m *mockRegRepo) GetRequestByID(ctx context.Context, requestID int32) (*domain.RegistrationRequest, error) {
	return nil, nil
}
func (m *mockRegRepo) CountUsersByEmail(ctx context.Context, email string) (int64, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockRegRepo) HasPendingRequest(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}
func (m *mockRegRepo) GetRequestByToken(ctx context.Context, tokenHash string) (*domain.RegistrationRequest, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RegistrationRequest), args.Error(1)
}
func (m *mockRegRepo) GetGroupByName(ctx context.Context, name string) (*domain.Group, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Group), args.Error(1)
}

type mockMailer struct{ mock.Mock }

func (m *mockMailer) SendRegistrationLink(ctx context.Context, email, link string) error {
	return m.Called(ctx, email, link).Error(0)
}

type mockRetry struct{ mock.Mock }

func (m *mockRetry) Enqueue(ctx context.Context, msg domain.EmailRetryMessage) error {
	return m.Called(ctx, msg).Error(0)
}
func (m *mockRetry) Start(ctx context.Context) error { return nil }

func TestProcessBatchCSVPartialSuccess(t *testing.T) {
	repo := new(mockRegRepo)
	mailer := new(mockMailer)
	retry := new(mockRetry)
	cfg := &config.Config{
		RegistrationLinkTTL: 24 * time.Hour,
		RegistrationBaseURL: "http://localhost",
	}

	svc := service.NewRegistrationService(repo, mailer, retry, cfg)

	csvData := `fio,role,email,group_name
Good,student,good@x.com,G1
Bad,student,bad@x.com,MissingGroup
`
	repo.On("CreateBatch", mock.Anything, mock.Anything, int32(1), int32(2)).Return(&domain.RegistrationBatch{ID: "batch"}, nil)
	repo.On("CountUsersByEmail", mock.Anything, "good@x.com").Return(int64(0), nil)
	repo.On("HasPendingRequest", mock.Anything, "good@x.com").Return(false, nil)
	repo.On("GetGroupByName", mock.Anything, "G1").Return(&domain.Group{ID: 1, Name: "G1"}, nil)
	repo.On("CreateRequestWithOutbox", mock.Anything, mock.Anything).Return(int32(1), int32(1), "token", nil)
	mailer.On("SendRegistrationLink", mock.Anything, "good@x.com", mock.Anything).Return(nil)
	repo.On("MarkOutboxSent", mock.Anything, int32(1)).Return(nil)

	repo.On("CountUsersByEmail", mock.Anything, "bad@x.com").Return(int64(0), nil)
	repo.On("HasPendingRequest", mock.Anything, "bad@x.com").Return(false, nil)
	repo.On("GetGroupByName", mock.Anything, "MissingGroup").Return(nil, apperrors.ErrNotFound)

	repo.On("UpdateBatchCounts", mock.Anything, mock.Anything, int32(1), int32(1)).Return(nil)

	result, err := svc.ProcessBatchCSV(context.Background(), 1, strings.NewReader(csvData))
	require.NoError(t, err)
	assert.Equal(t, 1, result.Created)
	assert.Equal(t, 1, result.Failed)
}

func TestPreviewCompleteRevoked(t *testing.T) {
	repo := new(mockRegRepo)
	cfg := &config.Config{RegistrationBaseURL: "http://localhost"}
	svc := service.NewRegistrationService(repo, new(mockMailer), new(mockRetry), cfg)

	repo.On("GetRequestByToken", mock.Anything, mock.Anything).Return(&domain.RegistrationRequest{
		Status: "revoked",
	}, nil)

	_, err := svc.PreviewComplete(context.Background(), "sometoken")
	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrGone)
}
