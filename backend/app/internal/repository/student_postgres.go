package repository

import (
	"context"
	"keenye_practice/app/internal/db"
	"keenye_practice/app/internal/domain"
)

type PostgresStudentRepository struct {
	queries *db.Queries
}

func NewPostgresStudentRepository(q *db.Queries) domain.StudentRepository {
	return &PostgresStudentRepository{queries: q}
}

func (r *PostgresStudentRepository) GetByID(ctx context.Context, id int64) (*domain.Student, error) {
	res, err := r.queries.GetStudent(ctx, id)
	if err != nil {
		return nil, err
	}
	return &domain.Student{ID: res.ID, Fio: res.Fio, Group: res.GroupOfStudents, PhoneNumber: res.PhoneNumber}, nil
}

func (r *PostgresStudentRepository) List(ctx context.Context) ([]domain.Student, error) {
	rows, err := r.queries.ListStudents(ctx)
	if err != nil {
		return nil, err
	}
	students := make([]domain.Student, len(rows))
	for i, row := range rows {
		students[i] = domain.Student{ID: row.ID, Fio: row.Fio, Group: row.GroupOfStudents, PhoneNumber: row.PhoneNumber}
	}
	return students, nil
}

func (r *PostgresStudentRepository) ListByGroup(ctx context.Context, group string) ([]domain.Student, error) {
	rows, err := r.queries.ListStudentsByGroup(ctx, group)
	if err != nil {
		return nil, err
	}
	students := make([]domain.Student, len(rows))
	for i, row := range rows {
		students[i] = domain.Student{ID: row.ID, Fio: row.Fio, Group: row.GroupOfStudents, PhoneNumber: row.PhoneNumber}
	}
	return students, nil
}

func (r *PostgresStudentRepository) Create(ctx context.Context, s *domain.Student) (*domain.Student, error) {
	res, err := r.queries.CreateStudent(ctx, db.CreateStudentParams{
		Fio:             s.Fio,
		GroupOfStudents: s.Group,
		PhoneNumber:     s.PhoneNumber,
	})
	if err != nil {
		return nil, err
	}
	s.ID = res.ID
	return s, nil
}

func (r *PostgresStudentRepository) Update(ctx context.Context, s *domain.Student) (*domain.Student, error) {
	res, err := r.queries.UpdateStudent(ctx, db.UpdateStudentParams{
		ID:              s.ID,
		Fio:             s.Fio,
		GroupOfStudents: s.Group,
		PhoneNumber:     s.PhoneNumber,
	})
	if err != nil {
		return nil, err
	}
	return &domain.Student{ID: res.ID, Fio: res.Fio, Group: res.GroupOfStudents, PhoneNumber: res.PhoneNumber}, nil
}

func (r *PostgresStudentRepository) Delete(ctx context.Context, id int64) error {
	return r.queries.DeleteStudent(ctx, id)
}
