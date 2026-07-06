package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"keeneye_practice/app/internal/apperrors"
	"keeneye_practice/app/internal/db"
	"keeneye_practice/app/internal/dbutil"
	"keeneye_practice/app/internal/domain"
)

type postgresStudentRepository struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewPostgresStudentRepository(q *db.Queries, pool *pgxpool.Pool) domain.StudentRepository {
	return &postgresStudentRepository{q: q, pool: pool}
}

func (r *postgresStudentRepository) GetByID(ctx context.Context, id int32) (*domain.Student, error) {
	row, err := r.q.GetStudentByID(ctx, id)
	if err != nil {
		return nil, apperrors.MapPG(err)
	}
	return mapStudentRow(row.ID, row.Fio, row.UserID, row.GroupID, row.GroupName), nil
}

func (r *postgresStudentRepository) Create(ctx context.Context, userID int32, groupID int32, fio string) (*domain.Student, error) {
	row, err := r.q.CreateStudent(ctx, db.CreateStudentParams{
		UserID:  pgtype.Int4{Int32: userID, Valid: true},
		GroupID: pgtype.Int4{Int32: groupID, Valid: true},
		Fio:     fio,
	})
	if err != nil {
		return nil, apperrors.MapPG(err)
	}
	return &domain.Student{
		ID: row.ID, UserID: row.UserID.Int32, GroupID: row.GroupID.Int32, Fio: row.Fio,
	}, nil
}

func (r *postgresStudentRepository) Update(ctx context.Context, id int32, fio string, groupID int32) error {
	err := r.q.UpdateStudent(ctx, db.UpdateStudentParams{
		ID:      id,
		GroupID: pgtype.Int4{Int32: groupID, Valid: groupID != 0},
		Fio:     fio,
	})
	return apperrors.MapPG(err)
}

func (r *postgresStudentRepository) DeleteWithUser(ctx context.Context, studentID int32, userID int32) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer dbutil.Rollback(ctx, tx)

	qTx := r.q.WithTx(tx)
	if err := qTx.DeleteStudent(ctx, studentID); err != nil {
		return apperrors.MapPG(err)
	}
	if err := qTx.DeleteUser(ctx, userID); err != nil {
		return apperrors.MapPG(err)
	}
	return tx.Commit(ctx)
}

func (r *postgresStudentRepository) ListAll(ctx context.Context) ([]domain.Student, error) {
	rows, err := r.q.ListAllStudents(ctx)
	if err != nil {
		return nil, apperrors.MapPG(err)
	}
	return mapListRows(rows), nil
}

func (r *postgresStudentRepository) ListByGroupID(ctx context.Context, groupID int32) ([]domain.Student, error) {
	rows, err := r.q.GetStudentsByGroupID(ctx, groupID)
	if err != nil {
		return nil, apperrors.MapPG(err)
	}
	res := make([]domain.Student, len(rows))
	for i, row := range rows {
		res[i] = *mapStudentRow(row.ID, row.Fio, row.UserID, row.GroupID, row.GroupName)
	}
	return res, nil
}

func (r *postgresStudentRepository) ListClassmates(ctx context.Context, studentID int32) ([]domain.Student, error) {
	rows, err := r.q.GetClassmates(ctx, studentID)
	if err != nil {
		return nil, apperrors.MapPG(err)
	}
	res := make([]domain.Student, len(rows))
	for i, row := range rows {
		res[i] = *mapStudentRow(row.ID, row.Fio, row.UserID, row.GroupID, row.GroupName)
	}
	return res, nil
}

func (r *postgresStudentRepository) ListByTeacherGroups(ctx context.Context, teacherID int32) ([]domain.Student, error) {
	rows, err := r.q.GetStudentsByTeacherGroups(ctx, teacherID)
	if err != nil {
		return nil, apperrors.MapPG(err)
	}
	res := make([]domain.Student, len(rows))
	for i, row := range rows {
		res[i] = domain.Student{
			ID: row.ID, UserID: row.UserID.Int32, GroupID: row.GroupID.Int32,
			GroupName: row.GroupName, Fio: row.Fio,
		}
	}
	return res, nil
}

func (r *postgresStudentRepository) GetByUserID(ctx context.Context, userID int32) (*domain.Student, error) {
	row, err := r.q.GetStudentByUserID(ctx, pgtype.Int4{Int32: userID, Valid: true})
	if err != nil {
		return nil, apperrors.MapPG(err)
	}
	student, err := r.GetByID(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	return student, nil
}

func (r *postgresStudentRepository) CheckTeacherGroup(ctx context.Context, teacherID, groupID int32) (bool, error) {
	return r.q.CheckTeacherHasGroup(ctx, db.CheckTeacherHasGroupParams{
		TeacherID: teacherID,
		GroupID:   groupID,
	})
}

func mapListRows(rows []db.ListAllStudentsRow) []domain.Student {
	res := make([]domain.Student, len(rows))
	for i, row := range rows {
		res[i] = *mapStudentRow(row.ID, row.Fio, row.UserID, row.GroupID, row.GroupName)
	}
	return res
}

func mapStudentRow(id int32, fio string, userID, groupID pgtype.Int4, groupName pgtype.Text) *domain.Student {
	return &domain.Student{
		ID: id, Fio: fio,
		UserID: userID.Int32, GroupID: groupID.Int32,
		GroupName: groupName.String,
	}
}
