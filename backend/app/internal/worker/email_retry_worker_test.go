package worker_test

import (
	"context"
	"testing"
	"time"

	"keeneye_practice/app/internal/config"
	"keeneye_practice/app/internal/domain"
	"keeneye_practice/app/internal/worker"

	"github.com/stretchr/testify/mock"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) CreateBatch(ctx context.Context, batchID string, createdBy int32, totalRows int32) (*domain.RegistrationBatch, error) {
	return nil, nil
}
func (m *mockRepo) UpdateBatchCounts(ctx context.Context, batchID string, success, failed int32) error {
	return nil
}
func (m *mockRepo) GetBatch(ctx context.Context, batchID string) (*domain.RegistrationBatch, error) {
	return nil, nil
}
func (m *mockRepo) ListRequestsByBatch(ctx context.Context, batchID string) ([]domain.RegistrationRequest, error) {
	return nil, nil
}
func (m *mockRepo) CreateRequestWithOutbox(ctx context.Context, input domain.CreateRegistrationInput) (int32, int32, string, error) {
	return 0, 0, "", nil
}
func (m *mockRepo) CompleteRequestByToken(ctx context.Context, tokenHash, password, phone string) (int32, error) {
	return 0, nil
}
func (m *mockRepo) RevokeExpired(ctx context.Context, limit int32) ([]int32, error) {
	return nil, nil
}
func (m *mockRepo) MarkOutboxSent(ctx context.Context, outboxID int32) error {
	args := m.Called(ctx, outboxID)
	return args.Error(0)
}
func (m *mockRepo) MarkOutboxFailed(ctx context.Context, outboxID int32, errMsg string) error {
	args := m.Called(ctx, outboxID, errMsg)
	return args.Error(0)
}
func (m *mockRepo) MarkOutboxDead(ctx context.Context, outboxID int32, errMsg string) error {
	args := m.Called(ctx, outboxID, errMsg)
	return args.Error(0)
}
func (m *mockRepo) ListOutboxForRetry(ctx context.Context, maxAttempts, limit int32) ([]domain.EmailOutboxItem, error) {
	args := m.Called(ctx, maxAttempts, limit)
	return args.Get(0).([]domain.EmailOutboxItem), args.Error(1)
}
func (m *mockRepo) GetRequestByID(ctx context.Context, requestID int32) (*domain.RegistrationRequest, error) {
	args := m.Called(ctx, requestID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RegistrationRequest), args.Error(1)
}
func (m *mockRepo) CountUsersByEmail(ctx context.Context, email string) (int64, error) {
	return 0, nil
}
func (m *mockRepo) HasPendingRequest(ctx context.Context, email string) (bool, error) {
	return false, nil
}
func (m *mockRepo) GetRequestByToken(ctx context.Context, tokenHash string) (*domain.RegistrationRequest, error) {
	return nil, nil
}
func (m *mockRepo) GetGroupByName(ctx context.Context, name string) (*domain.Group, error) {
	return nil, nil
}

type mockSender struct {
	mock.Mock
}

func (m *mockSender) ResendEmail(ctx context.Context, msg domain.EmailRetryMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func TestEmailRetryWorkerSimulatedFailureDoesNotMarkSent(t *testing.T) {
	repo := new(mockRepo)
	sender := new(mockSender)

	cfg := &config.Config{
		EmailRetrySimulateFailRate: 1.0,
		EmailRetryMaxAttempts:      5,
		EmailRetryWorkerInterval:   time.Second,
	}

	w := worker.NewEmailRetryWorker(repo, sender, cfg)

	repo.On("ListOutboxForRetry", mock.Anything, int32(5), int32(20)).Return([]domain.EmailOutboxItem{
		{OutboxID: 1, RequestID: 10, Attempts: 0},
	}, nil)
	repo.On("GetRequestByID", mock.Anything, int32(10)).Return(&domain.RegistrationRequest{
		ID: 10, Email: "a@x.com", Status: "pending", InviteToken: "tok",
	}, nil)
	repo.On("MarkOutboxFailed", mock.Anything, int32(1), mock.Anything).Return(nil)

	w.TickForTest(context.Background())

	sender.AssertNotCalled(t, "ResendEmail", mock.Anything, mock.Anything)
	repo.AssertCalled(t, "MarkOutboxFailed", mock.Anything, int32(1), mock.Anything)
}

func TestEmailRetryWorkerSuccessMarksSent(t *testing.T) {
	repo := new(mockRepo)
	sender := new(mockSender)

	cfg := &config.Config{
		EmailRetrySimulateFailRate: 0,
		EmailRetryMaxAttempts:      5,
	}

	w := worker.NewEmailRetryWorker(repo, sender, cfg)

	repo.On("ListOutboxForRetry", mock.Anything, int32(5), int32(20)).Return([]domain.EmailOutboxItem{
		{OutboxID: 2, RequestID: 20, Attempts: 1},
	}, nil)
	repo.On("GetRequestByID", mock.Anything, int32(20)).Return(&domain.RegistrationRequest{
		ID: 20, Email: "b@x.com", Status: "pending", InviteToken: "tok2",
	}, nil)
	sender.On("ResendEmail", mock.Anything, mock.Anything).Return(nil)

	w.TickForTest(context.Background())

	sender.AssertCalled(t, "ResendEmail", mock.Anything, mock.Anything)
}
