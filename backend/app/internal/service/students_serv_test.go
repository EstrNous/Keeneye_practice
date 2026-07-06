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

func TestStudentService_GetStudentsList_ByRole(t *testing.T) {
	ctx := context.Background()
	expected := []domain.Student{{ID: 1, Fio: "A"}}

	t.Run("admin lists all", func(t *testing.T) {
		repo := mocks.NewStudentRepository(t)
		svc := NewStudentService(repo, nil, nil)
		repo.On("ListAll", mock.Anything).Return(expected, nil)

		got, err := svc.GetStudentsList(ctx, "admin", 0, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("teacher lists by groups", func(t *testing.T) {
		repo := mocks.NewStudentRepository(t)
		svc := NewStudentService(repo, nil, nil)
		repo.On("ListByTeacherGroups", mock.Anything, int32(7)).Return(expected, nil)

		got, err := svc.GetStudentsList(ctx, "teacher", 7, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("student lists classmates", func(t *testing.T) {
		repo := mocks.NewStudentRepository(t)
		svc := NewStudentService(repo, nil, nil)
		repo.On("ListClassmates", mock.Anything, int32(3)).Return(expected, nil)

		got, err := svc.GetStudentsList(ctx, "student", 3, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("unknown role forbidden", func(t *testing.T) {
		svc := NewStudentService(mocks.NewStudentRepository(t), nil, nil)
		_, err := svc.GetStudentsList(ctx, "guest", 0, nil)
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
	})
}

func TestStudentService_GetStudentsList_GroupFilter(t *testing.T) {
	ctx := context.Background()
	groupID := int32(2)
	expected := []domain.Student{{ID: 10, GroupID: 2}}

	t.Run("admin by group", func(t *testing.T) {
		repo := mocks.NewStudentRepository(t)
		svc := NewStudentService(repo, nil, nil)
		repo.On("ListByGroupID", mock.Anything, groupID).Return(expected, nil)

		got, err := svc.GetStudentsList(ctx, "admin", 0, &groupID)
		assert.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("teacher allowed in group", func(t *testing.T) {
		repo := mocks.NewStudentRepository(t)
		svc := NewStudentService(repo, nil, nil)
		repo.On("CheckTeacherGroup", mock.Anything, int32(1), groupID).Return(true, nil)
		repo.On("ListByGroupID", mock.Anything, groupID).Return(expected, nil)

		got, err := svc.GetStudentsList(ctx, "teacher", 1, &groupID)
		assert.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("teacher denied outside group", func(t *testing.T) {
		repo := mocks.NewStudentRepository(t)
		svc := NewStudentService(repo, nil, nil)
		repo.On("CheckTeacherGroup", mock.Anything, int32(1), groupID).Return(false, nil)

		_, err := svc.GetStudentsList(ctx, "teacher", 1, &groupID)
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
	})

	t.Run("student cannot filter by group", func(t *testing.T) {
		svc := NewStudentService(mocks.NewStudentRepository(t), nil, nil)
		_, err := svc.GetStudentsList(ctx, "student", 3, &groupID)
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
	})
}

func TestStudentService_GetStudent_AccessMatrix(t *testing.T) {
	ctx := context.Background()
	student := &domain.Student{ID: 5, GroupID: 2, Fio: "S"}

	t.Run("admin can read any", func(t *testing.T) {
		repo := mocks.NewStudentRepository(t)
		svc := NewStudentService(repo, nil, nil)
		repo.On("GetByID", mock.Anything, int32(5)).Return(student, nil)

		got, err := svc.GetStudent(ctx, "admin", 0, 5)
		assert.NoError(t, err)
		assert.Equal(t, student, got)
	})

	t.Run("student can read self", func(t *testing.T) {
		repo := mocks.NewStudentRepository(t)
		svc := NewStudentService(repo, nil, nil)
		repo.On("GetByID", mock.Anything, int32(5)).Return(student, nil)

		got, err := svc.GetStudent(ctx, "student", 5, 5)
		assert.NoError(t, err)
		assert.Equal(t, student, got)
	})

	t.Run("teacher in group", func(t *testing.T) {
		repo := mocks.NewStudentRepository(t)
		svc := NewStudentService(repo, nil, nil)
		repo.On("GetByID", mock.Anything, int32(5)).Return(student, nil)
		repo.On("CheckTeacherGroup", mock.Anything, int32(1), int32(2)).Return(true, nil)

		got, err := svc.GetStudent(ctx, "teacher", 1, 5)
		assert.NoError(t, err)
		assert.Equal(t, student, got)
	})
}

func TestStudentService_GetStudent_TeacherDeniedOutsideGroup(t *testing.T) {
	repo := mocks.NewStudentRepository(t)
	svc := NewStudentService(repo, nil, nil)

	repo.On("GetByID", mock.Anything, int32(5)).Return(&domain.Student{ID: 5, GroupID: 2}, nil)
	repo.On("CheckTeacherGroup", mock.Anything, int32(1), int32(2)).Return(false, nil)

	_, err := svc.GetStudent(context.Background(), "teacher", 1, 5)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestStudentService_GetStudent_StudentSelfOnly(t *testing.T) {
	repo := mocks.NewStudentRepository(t)
	svc := NewStudentService(repo, nil, nil)

	repo.On("GetByID", mock.Anything, int32(3)).Return(&domain.Student{ID: 3}, nil)

	_, err := svc.GetStudent(context.Background(), "student", 1, 3)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestStudentService_CreateStudent(t *testing.T) {
	ctx := context.Background()
	input := domain.CreateStudentInput{
		Email: "s@example.com", Password: "secret12",
		PhoneNumber: "+79001112233", GroupID: 1, Fio: "New",
	}

	t.Run("admin creates student", func(t *testing.T) {
		repo := mocks.NewStudentRepository(t)
		auth := mocks.NewAuthService(t)
		groups := mocks.NewGroupRepository(t)
		svc := NewStudentService(repo, auth, groups)

		groups.On("GetByID", mock.Anything, int32(1)).Return(&domain.Group{ID: 1, Name: "VM"}, nil)
		auth.On("Register", mock.Anything, input.Email, input.Password, "student", input.PhoneNumber, input.Fio, mock.Anything).Return(int32(9), nil)
		repo.On("GetByUserID", mock.Anything, int32(9)).Return(&domain.Student{ID: 4, UserID: 9, Fio: input.Fio}, nil)

		got, err := svc.CreateStudent(ctx, "admin", input)
		assert.NoError(t, err)
		assert.Equal(t, int32(4), got.ID)
	})

	t.Run("non-admin forbidden", func(t *testing.T) {
		svc := NewStudentService(mocks.NewStudentRepository(t), nil, nil)
		_, err := svc.CreateStudent(ctx, "teacher", input)
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
	})
}

func TestStudentService_ModifyStudent(t *testing.T) {
	ctx := context.Background()
	target := &domain.Student{ID: 3, GroupID: 5, Fio: "Old"}

	t.Run("student updates self and keeps group", func(t *testing.T) {
		repo := mocks.NewStudentRepository(t)
		svc := NewStudentService(repo, nil, nil)
		repo.On("GetByID", mock.Anything, int32(3)).Return(target, nil)
		repo.On("Update", mock.Anything, int32(3), "New", int32(5)).Return(nil)

		err := svc.ModifyStudent(ctx, "student", 3, 3, "New", 99)
		assert.NoError(t, err)
	})

	t.Run("teacher denied outside group", func(t *testing.T) {
		repo := mocks.NewStudentRepository(t)
		svc := NewStudentService(repo, nil, nil)
		repo.On("GetByID", mock.Anything, int32(3)).Return(target, nil)
		repo.On("CheckTeacherGroup", mock.Anything, int32(1), int32(5)).Return(false, nil)

		err := svc.ModifyStudent(ctx, "teacher", 1, 3, "New", 5)
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
	})
}

func TestStudentService_RemoveStudent(t *testing.T) {
	ctx := context.Background()

	t.Run("admin deletes with user", func(t *testing.T) {
		repo := mocks.NewStudentRepository(t)
		svc := NewStudentService(repo, nil, nil)
		repo.On("GetByID", mock.Anything, int32(2)).Return(&domain.Student{ID: 2, UserID: 8}, nil)
		repo.On("DeleteWithUser", mock.Anything, int32(2), int32(8)).Return(nil)

		assert.NoError(t, svc.RemoveStudent(ctx, "admin", 2))
	})

	t.Run("non-admin forbidden", func(t *testing.T) {
		svc := NewStudentService(mocks.NewStudentRepository(t), nil, nil)
		err := svc.RemoveStudent(ctx, "teacher", 1)
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
	})
}
