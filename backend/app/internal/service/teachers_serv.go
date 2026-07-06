package service

import (
	"context"

	"keeneye_practice/app/internal/apperrors"
	"keeneye_practice/app/internal/domain"
)

type TeacherServiceImpl struct {
	repo domain.TeacherRepository
}

func NewTeacherService(repo domain.TeacherRepository) domain.TeacherService {
	return &TeacherServiceImpl{repo: repo}
}

func (s *TeacherServiceImpl) ListTeachers(ctx context.Context, role string, actorProfileID int32) ([]domain.Teacher, error) {
	if role != "admin" {
		return nil, apperrors.ErrForbidden
	}
	return s.repo.ListAll(ctx)
}

func (s *TeacherServiceImpl) GetTeacher(ctx context.Context, role string, actorProfileID int32, targetID int32) (*domain.Teacher, error) {
	if role == "teacher" && actorProfileID != targetID {
		return nil, apperrors.ErrForbidden
	}
	if role != "admin" && role != "teacher" {
		return nil, apperrors.ErrForbidden
	}
	return s.repo.GetByID(ctx, targetID)
}

func (s *TeacherServiceImpl) UpdateTeacher(ctx context.Context, role string, actorProfileID int32, targetID int32, fio string) error {
	if role == "teacher" && actorProfileID != targetID {
		return apperrors.ErrForbidden
	}
	if role != "admin" && role != "teacher" {
		return apperrors.ErrForbidden
	}
	return s.repo.Update(ctx, targetID, fio)
}

func (s *TeacherServiceImpl) DeleteTeacher(ctx context.Context, role string, targetID int32) error {
	if role != "admin" {
		return apperrors.ErrForbidden
	}
	teacher, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		return err
	}
	return s.repo.DeleteWithUser(ctx, targetID, teacher.UserID)
}

func (s *TeacherServiceImpl) ListTeacherGroups(ctx context.Context, role string, actorProfileID int32, teacherID int32) ([]domain.Group, error) {
	if role == "teacher" && actorProfileID != teacherID {
		return nil, apperrors.ErrForbidden
	}
	if role != "admin" && role != "teacher" {
		return nil, apperrors.ErrForbidden
	}
	return s.repo.ListGroups(ctx, teacherID)
}

func (s *TeacherServiceImpl) AssignGroup(ctx context.Context, role string, teacherID int32, groupID int32) error {
	if role != "admin" {
		return apperrors.ErrForbidden
	}
	return s.repo.AssignGroup(ctx, teacherID, groupID)
}

func (s *TeacherServiceImpl) RemoveGroup(ctx context.Context, role string, teacherID int32, groupID int32) error {
	if role != "admin" {
		return apperrors.ErrForbidden
	}
	return s.repo.RemoveGroup(ctx, teacherID, groupID)
}
