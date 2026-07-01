package repository

import (
	"context"
	"keeneye_practice/app/internal/db"
	"keeneye_practice/app/internal/domain"

	"github.com/jackc/pgx/v5/pgtype"
)

type postgresStudentRepository struct {
	q *db.Queries
}

func NewPostgresStudentRepository(q *db.Queries) domain.StudentRepository {
	return &postgresStudentRepository{q: q}
}

func (r *postgresStudentRepository) GetByID(ctx context.Context, id int32) (*domain.Student, error) {
	row, err := r.q.GetStudentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &domain.Student{
		ID:          row.ID,
		UserID:      row.UserID.Int32,
		GroupID:     row.GroupID.Int32,
		GroupName:   row.GroupName.String,
		Fio:         row.Fio,
		PhoneNumber: row.PhoneNumber,
	}, nil
}

func (r *postgresStudentRepository) Create(ctx context.Context, s *domain.Student) (*domain.Student, error) {
	row, err := r.q.CreateStudent(ctx, db.CreateStudentParams{
		UserID:      pgtype.Int4{Int32: s.UserID, Valid: true},
		GroupID:     pgtype.Int4{Int32: s.GroupID, Valid: s.GroupID != 0},
		Fio:         s.Fio,
		PhoneNumber: s.PhoneNumber,
	})
	if err != nil {
		return nil, err
	}

	return &domain.Student{
		ID:          row.ID,
		UserID:      row.UserID.Int32,
		GroupID:     row.GroupID.Int32,
		GroupName:   "",
		Fio:         row.Fio,
		PhoneNumber: row.PhoneNumber,
	}, nil
}

func (r *postgresStudentRepository) ListAll(ctx context.Context) ([]domain.Student, error) {
	rows, err := r.q.ListAllStudents(ctx)
	if err != nil {
		return nil, err
	}
	return mapRowsToDomain(rows), nil
}

func (r *postgresStudentRepository) ListClassmates(ctx context.Context, studentID int32) ([]domain.Student, error) {
	rows, err := r.q.GetClassmates(ctx, studentID)
	if err != nil {
		return nil, err
	}
	return mapClassmatesToDomain(rows), nil
}

func (r *postgresStudentRepository) ListByTeacherGroups(ctx context.Context, teacherID int32) ([]domain.Student, error) {
	rows, err := r.q.GetStudentsByTeacherGroups(ctx, teacherID)
	if err != nil {
		return nil, err
	}
	return mapTeacherRowsToDomain(rows), nil
}

func (r *postgresStudentRepository) Update(ctx context.Context, id int32, s *domain.Student) error {
	return r.q.UpdateStudent(ctx, db.UpdateStudentParams{
		ID:          id,
		GroupID:     pgtype.Int4{Int32: s.GroupID, Valid: s.GroupID != 0},
		Fio:         s.Fio,
		PhoneNumber: s.PhoneNumber,
	})
}

func (r *postgresStudentRepository) Delete(ctx context.Context, id int32) error {
	return r.q.DeleteStudent(ctx, id)
}

func (r *postgresStudentRepository) CheckTeacherGroup(ctx context.Context, teacherID, groupID int32) (bool, error) {
	return r.q.CheckTeacherHasGroup(ctx, db.CheckTeacherHasGroupParams{
		TeacherID: teacherID,
		GroupID:   groupID,
	})
}

func mapRowsToDomain(rows []db.ListAllStudentsRow) []domain.Student {
	res := make([]domain.Student, len(rows))
	for i, r := range rows {
		res[i] = domain.Student{
			ID: r.ID, UserID: r.UserID.Int32, GroupID: r.GroupID.Int32,
			GroupName: r.GroupName.String, Fio: r.Fio, PhoneNumber: r.PhoneNumber,
		}
	}
	return res
}

func mapClassmatesToDomain(rows []db.GetClassmatesRow) []domain.Student {
	res := make([]domain.Student, len(rows))
	for i, r := range rows {
		res[i] = domain.Student{
			ID: r.ID, UserID: r.UserID.Int32, GroupID: r.GroupID.Int32,
			GroupName: r.GroupName.String, Fio: r.Fio, PhoneNumber: r.PhoneNumber,
		}
	}
	return res
}

func mapTeacherRowsToDomain(rows []db.GetStudentsByTeacherGroupsRow) []domain.Student {
	res := make([]domain.Student, len(rows))
	for i, r := range rows {
		res[i] = domain.Student{
			ID: r.ID, UserID: r.UserID.Int32, GroupID: r.GroupID.Int32,
			GroupName: r.GroupName,
			Fio:       r.Fio, PhoneNumber: r.PhoneNumber,
		}
	}
	return res
}
