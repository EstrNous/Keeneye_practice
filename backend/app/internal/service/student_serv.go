package service

import (
	"context"
	"errors"
	"keenye_practice/app/internal/domain"
)

type StudentServiceImpl struct {
	repo domain.StudentRepository
}

func NewStudentService(repo domain.StudentRepository) domain.StudentService {
	return &StudentServiceImpl{repo: repo}
}

func (s *StudentServiceImpl) GetStudent(ctx context.Context, id int64) (*domain.Student, error) {
	if id <= 0 {
		return nil, errors.New("bad request: invalid student ID")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *StudentServiceImpl) TemplateCheck() string {
	return "Привет из Go-версии проекта!"
}

func (s *StudentServiceImpl) GetAllStudents(ctx context.Context) ([]domain.Student, error) {
	return s.repo.List(ctx)
}

func (s *StudentServiceImpl) GetStudentsByGroup(ctx context.Context, group string) ([]domain.Student, error) {
	if group == "" {
		return nil, errors.New("group is required")
	}
	return s.repo.ListByGroup(ctx, group)
}

func (s *StudentServiceImpl) RegisterStudent(ctx context.Context, student *domain.Student) (*domain.Student, error) {
	if student.Fio == "" || student.Group == "" {
		return nil, errors.New("validation failed: empty fields")
	}
	return s.repo.Create(ctx, student)
}

func (s *StudentServiceImpl) ModifyStudent(ctx context.Context, student *domain.Student) (*domain.Student, error) {
	if student.ID <= 0 {
		return nil, errors.New("validation failed: missing student ID for update")
	}
	return s.repo.Update(ctx, student)
}

func (s *StudentServiceImpl) RemoveStudent(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
