package repository

import (
	"context"

	"keeneye_practice/app/internal/apperrors"
	"keeneye_practice/app/internal/db"
	"keeneye_practice/app/internal/domain"
)

type postgresGroupRepository struct {
	q *db.Queries
}

func NewPostgresGroupRepository(q *db.Queries) domain.GroupRepository {
	return &postgresGroupRepository{q: q}
}

func (r *postgresGroupRepository) ListAll(ctx context.Context) ([]domain.Group, error) {
	rows, err := r.q.ListGroups(ctx)
	if err != nil {
		return nil, apperrors.MapPG(err)
	}
	res := make([]domain.Group, len(rows))
	for i, g := range rows {
		res[i] = domain.Group{ID: g.ID, Name: g.Name}
	}
	return res, nil
}

func (r *postgresGroupRepository) GetByID(ctx context.Context, id int32) (*domain.Group, error) {
	row, err := r.q.GetGroupByID(ctx, id)
	if err != nil {
		return nil, apperrors.MapPG(err)
	}
	return &domain.Group{ID: row.ID, Name: row.Name}, nil
}

func (r *postgresGroupRepository) Create(ctx context.Context, name string) (*domain.Group, error) {
	row, err := r.q.CreateGroup(ctx, name)
	if err != nil {
		return nil, apperrors.MapPG(err)
	}
	return &domain.Group{ID: row.ID, Name: row.Name}, nil
}
