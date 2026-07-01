package service

import (
	"context"
	"errors"

	"keeneye_practice/app/internal/domain"
)

type StudentServiceImpl struct {
	repo domain.StudentRepository
}

func NewStudentService(repo domain.StudentRepository) domain.StudentService {
	return &StudentServiceImpl{repo: repo}
}

func (s *StudentServiceImpl) GetStudentsList(ctx context.Context, role string, profileID int32) ([]domain.Student, error) {
	switch role {
	case "admin":
		return s.repo.ListAll(ctx)
	case "teacher":
		return s.repo.ListByTeacherGroups(ctx, profileID)
	case "student":
		return s.repo.ListClassmates(ctx, profileID)
	default:
		return nil, errors.New("unknown role")
	}
}

func (s *StudentServiceImpl) GetStudent(ctx context.Context, role string, actorProfileID int32, targetID int32) (*domain.Student, error) {
	if role == "student" && actorProfileID != targetID {
		return nil, errors.New("access denied")
	}

	return s.repo.GetByID(ctx, targetID)
}

func (s *StudentServiceImpl) ModifyStudent(ctx context.Context, role string, actorProfileID int32, targetStudentID int32, sData *domain.Student) error {
	if role == "student" && actorProfileID != targetStudentID {
		return errors.New("access denied: you can only modify your own profile")
	}

	if role == "teacher" {
		targetStudent, err := s.repo.GetByID(ctx, targetStudentID)
		if err != nil {
			return errors.New("student not found")
		}
		hasAccess, err := s.repo.CheckTeacherGroup(ctx, actorProfileID, targetStudent.GroupID)
		if err != nil || !hasAccess {
			return errors.New("access denied: student is not in your group")
		}
	}

	return s.repo.Update(ctx, targetStudentID, sData)
}

func (s *StudentServiceImpl) RemoveStudent(ctx context.Context, id int32) error {
	return s.repo.Delete(ctx, id)
}
