package service

import (
	"context"
	"testing"

	"keeneye_practice/app/internal/apperrors"
	"keeneye_practice/app/internal/domain"
	"keeneye_practice/app/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGroupService_ListGroups(t *testing.T) {
	ctx := context.Background()
	expected := []domain.Group{{ID: 1, Name: "VM"}}

	t.Run("admin ok", func(t *testing.T) {
		repo := mocks.NewGroupRepository(t)
		svc := NewGroupService(repo)
		repo.On("ListAll", mock.Anything).Return(expected, nil)

		got, err := svc.ListGroups(ctx, "admin")
		assert.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("teacher ok", func(t *testing.T) {
		repo := mocks.NewGroupRepository(t)
		svc := NewGroupService(repo)
		repo.On("ListAll", mock.Anything).Return(expected, nil)

		got, err := svc.ListGroups(ctx, "teacher")
		assert.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("student forbidden", func(t *testing.T) {
		svc := NewGroupService(mocks.NewGroupRepository(t))
		_, err := svc.ListGroups(ctx, "student")
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
	})
}

func TestGroupService_CreateGroup_AdminOnly(t *testing.T) {
	ctx := context.Background()

	t.Run("admin creates", func(t *testing.T) {
		repo := mocks.NewGroupRepository(t)
		svc := NewGroupService(repo)
		repo.On("Create", mock.Anything, "NEW").Return(&domain.Group{ID: 5, Name: "NEW"}, nil)

		got, err := svc.CreateGroup(ctx, "admin", "NEW")
		assert.NoError(t, err)
		assert.Equal(t, "NEW", got.Name)
	})

	t.Run("teacher forbidden", func(t *testing.T) {
		svc := NewGroupService(mocks.NewGroupRepository(t))
		_, err := svc.CreateGroup(ctx, "teacher", "X")
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
	})
}
