package service

import (
	"context"

	"keeneye_practice/app/internal/apperrors"
	"keeneye_practice/app/internal/domain"
)

type GroupServiceImpl struct {
	repo domain.GroupRepository
}

func NewGroupService(repo domain.GroupRepository) domain.GroupService {
	return &GroupServiceImpl{repo: repo}
}

func (s *GroupServiceImpl) ListGroups(ctx context.Context, role string) ([]domain.Group, error) {
	if role != "admin" && role != "teacher" {
		return nil, apperrors.ErrForbidden
	}
	return s.repo.ListAll(ctx)
}

func (s *GroupServiceImpl) CreateGroup(ctx context.Context, role string, name string) (*domain.Group, error) {
	if role != "admin" {
		return nil, apperrors.ErrForbidden
	}
	return s.repo.Create(ctx, name)
}
