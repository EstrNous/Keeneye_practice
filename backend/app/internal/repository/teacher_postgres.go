package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"keeneye_practice/app/internal/apperrors"
	"keeneye_practice/app/internal/db"
	"keeneye_practice/app/internal/dbutil"
	"keeneye_practice/app/internal/domain"
)

type postgresTeacherRepository struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewPostgresTeacherRepository(q *db.Queries, pool *pgxpool.Pool) domain.TeacherRepository {
	return &postgresTeacherRepository{q: q, pool: pool}
}

func (r *postgresTeacherRepository) ListAll(ctx context.Context) ([]domain.Teacher, error) {
	rows, err := r.q.ListTeachers(ctx)
	if err != nil {
		return nil, apperrors.MapPG(err)
	}
	res := make([]domain.Teacher, len(rows))
	for i, row := range rows {
		res[i] = domain.Teacher{
			ID: row.ID, UserID: row.UserID.Int32, Fio: row.Fio,
			Email: row.Email, PhoneNumber: row.PhoneNumber,
		}
	}
	return res, nil
}

func (r *postgresTeacherRepository) GetByID(ctx context.Context, id int32) (*domain.Teacher, error) {
	row, err := r.q.GetTeacherByID(ctx, id)
	if err != nil {
		return nil, apperrors.MapPG(err)
	}
	return &domain.Teacher{
		ID: row.ID, UserID: row.UserID.Int32, Fio: row.Fio,
		Email: row.Email, PhoneNumber: row.PhoneNumber,
	}, nil
}

func (r *postgresTeacherRepository) Update(ctx context.Context, id int32, fio string) error {
	return apperrors.MapPG(r.q.UpdateTeacher(ctx, db.UpdateTeacherParams{ID: id, Fio: fio}))
}

func (r *postgresTeacherRepository) DeleteWithUser(ctx context.Context, teacherID int32, userID int32) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer dbutil.Rollback(ctx, tx)

	qTx := r.q.WithTx(tx)
	if err := qTx.DeleteTeacher(ctx, teacherID); err != nil {
		return apperrors.MapPG(err)
	}
	if err := qTx.DeleteUser(ctx, userID); err != nil {
		return apperrors.MapPG(err)
	}
	return tx.Commit(ctx)
}

func (r *postgresTeacherRepository) ListGroups(ctx context.Context, teacherID int32) ([]domain.Group, error) {
	rows, err := r.q.ListTeacherGroups(ctx, teacherID)
	if err != nil {
		return nil, apperrors.MapPG(err)
	}
	return mapGroups(rows), nil
}

func (r *postgresTeacherRepository) AssignGroup(ctx context.Context, teacherID int32, groupID int32) error {
	return apperrors.MapPG(r.q.AssignTeacherToGroup(ctx, db.AssignTeacherToGroupParams{
		TeacherID: teacherID,
		GroupID:   groupID,
	}))
}

func (r *postgresTeacherRepository) RemoveGroup(ctx context.Context, teacherID int32, groupID int32) error {
	return apperrors.MapPG(r.q.RemoveTeacherFromGroup(ctx, db.RemoveTeacherFromGroupParams{
		TeacherID: teacherID,
		GroupID:   groupID,
	}))
}

func mapGroups(rows []db.Group) []domain.Group {
	res := make([]domain.Group, len(rows))
	for i, g := range rows {
		res[i] = domain.Group{ID: g.ID, Name: g.Name}
	}
	return res
}
