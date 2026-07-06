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

func TestTeacherService_ListTeachers_AdminOnly(t *testing.T) {
	ctx := context.Background()
	expected := []domain.Teacher{{ID: 1, Fio: "T"}}

	t.Run("admin ok", func(t *testing.T) {
		repo := mocks.NewTeacherRepository(t)
		svc := NewTeacherService(repo)
		repo.On("ListAll", mock.Anything).Return(expected, nil)

		got, err := svc.ListTeachers(ctx, "admin", 0)
		assert.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("teacher forbidden", func(t *testing.T) {
		svc := NewTeacherService(mocks.NewTeacherRepository(t))
		_, err := svc.ListTeachers(ctx, "teacher", 1)
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
	})
}

func TestTeacherService_GetTeacher(t *testing.T) {
	ctx := context.Background()
	teacher := &domain.Teacher{ID: 2, Fio: "T"}

	t.Run("teacher reads self", func(t *testing.T) {
		repo := mocks.NewTeacherRepository(t)
		svc := NewTeacherService(repo)
		repo.On("GetByID", mock.Anything, int32(2)).Return(teacher, nil)

		got, err := svc.GetTeacher(ctx, "teacher", 2, 2)
		assert.NoError(t, err)
		assert.Equal(t, teacher, got)
	})

	t.Run("teacher cannot read other", func(t *testing.T) {
		svc := NewTeacherService(mocks.NewTeacherRepository(t))
		_, err := svc.GetTeacher(ctx, "teacher", 1, 2)
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
	})

	t.Run("admin reads any", func(t *testing.T) {
		repo := mocks.NewTeacherRepository(t)
		svc := NewTeacherService(repo)
		repo.On("GetByID", mock.Anything, int32(2)).Return(teacher, nil)

		got, err := svc.GetTeacher(ctx, "admin", 0, 2)
		assert.NoError(t, err)
		assert.Equal(t, teacher, got)
	})
}

func TestTeacherService_AssignGroup_AdminOnly(t *testing.T) {
	ctx := context.Background()

	t.Run("admin assigns", func(t *testing.T) {
		repo := mocks.NewTeacherRepository(t)
		svc := NewTeacherService(repo)
		repo.On("AssignGroup", mock.Anything, int32(1), int32(2)).Return(nil)

		assert.NoError(t, svc.AssignGroup(ctx, "admin", 1, 2))
	})

	t.Run("teacher forbidden", func(t *testing.T) {
		svc := NewTeacherService(mocks.NewTeacherRepository(t))
		err := svc.AssignGroup(ctx, "teacher", 1, 2)
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
	})
}

func TestTeacherService_DeleteTeacher(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewTeacherRepository(t)
	svc := NewTeacherService(repo)

	repo.On("GetByID", mock.Anything, int32(3)).Return(&domain.Teacher{ID: 3, UserID: 10}, nil)
	repo.On("DeleteWithUser", mock.Anything, int32(3), int32(10)).Return(nil)

	assert.NoError(t, svc.DeleteTeacher(ctx, "admin", 3))
}
