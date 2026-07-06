package service

import (
	"context"

	"keeneye_practice/app/internal/apperrors"
	"keeneye_practice/app/internal/domain"
)

type StudentServiceImpl struct {
	repo      domain.StudentRepository
	authSvc   domain.AuthService
	groupRepo domain.GroupRepository
}

func NewStudentService(repo domain.StudentRepository, authSvc domain.AuthService, groupRepo domain.GroupRepository) domain.StudentService {
	return &StudentServiceImpl{repo: repo, authSvc: authSvc, groupRepo: groupRepo}
}

func (s *StudentServiceImpl) GetStudentsList(ctx context.Context, role string, profileID int32, groupID *int32) ([]domain.Student, error) {
	if groupID != nil {
		return s.listByGroupFilter(ctx, role, profileID, *groupID)
	}

	switch role {
	case "admin":
		return s.repo.ListAll(ctx)
	case "teacher":
		return s.repo.ListByTeacherGroups(ctx, profileID)
	case "student":
		return s.repo.ListClassmates(ctx, profileID)
	default:
		return nil, apperrors.ErrForbidden
	}
}

func (s *StudentServiceImpl) listByGroupFilter(ctx context.Context, role string, profileID int32, groupID int32) ([]domain.Student, error) {
	switch role {
	case "admin":
		return s.repo.ListByGroupID(ctx, groupID)
	case "teacher":
		hasAccess, err := s.repo.CheckTeacherGroup(ctx, profileID, groupID)
		if err != nil || !hasAccess {
			return nil, apperrors.ErrForbidden
		}
		return s.repo.ListByGroupID(ctx, groupID)
	default:
		return nil, apperrors.ErrForbidden
	}
}

func (s *StudentServiceImpl) GetStudent(ctx context.Context, role string, actorProfileID int32, targetID int32) (*domain.Student, error) {
	student, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		return nil, err
	}

	switch role {
	case "admin":
		return student, nil
	case "student":
		if actorProfileID != targetID {
			return nil, apperrors.ErrForbidden
		}
		return student, nil
	case "teacher":
		hasAccess, err := s.repo.CheckTeacherGroup(ctx, actorProfileID, student.GroupID)
		if err != nil || !hasAccess {
			return nil, apperrors.ErrForbidden
		}
		return student, nil
	default:
		return nil, apperrors.ErrForbidden
	}
}

func (s *StudentServiceImpl) CreateStudent(ctx context.Context, role string, req domain.CreateStudentInput) (*domain.Student, error) {
	if role != "admin" {
		return nil, apperrors.ErrForbidden
	}

	if _, err := s.groupRepo.GetByID(ctx, req.GroupID); err != nil {
		return nil, err
	}

	groupID := req.GroupID
	userID, err := s.authSvc.Register(ctx, req.Email, req.Password, "student", req.PhoneNumber, req.Fio, &groupID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetByUserID(ctx, userID)
}

func (s *StudentServiceImpl) ModifyStudent(ctx context.Context, role string, actorProfileID int32, targetStudentID int32, fio string, groupID int32) error {
	targetStudent, err := s.repo.GetByID(ctx, targetStudentID)
	if err != nil {
		return err
	}

	if role == "student" {
		if actorProfileID != targetStudentID {
			return apperrors.ErrForbidden
		}
		groupID = targetStudent.GroupID
	}

	if role == "teacher" {
		hasAccess, err := s.repo.CheckTeacherGroup(ctx, actorProfileID, targetStudent.GroupID)
		if err != nil || !hasAccess {
			return apperrors.ErrForbidden
		}
	}

	return s.repo.Update(ctx, targetStudentID, fio, groupID)
}

func (s *StudentServiceImpl) RemoveStudent(ctx context.Context, role string, id int32) error {
	if role != "admin" {
		return apperrors.ErrForbidden
	}

	student, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return s.repo.DeleteWithUser(ctx, id, student.UserID)
}
