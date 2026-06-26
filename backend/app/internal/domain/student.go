package domain

import "context"

// Student — доменная модель
type Student struct {
	ID          int64  `json:"id"`
	Fio         string `json:"fio"`
	Group       string `json:"group"`
	PhoneNumber string `json:"phoneNumber"`
}

// StudentRepository — интерфейс для инфраструктуры хранения данных.
type StudentRepository interface {
	GetByID(ctx context.Context, id int64) (*Student, error)
	List(ctx context.Context) ([]Student, error)
	ListByGroup(ctx context.Context, group string) ([]Student, error)
	Create(ctx context.Context, s *Student) (*Student, error)
	Update(ctx context.Context, s *Student) (*Student, error)
	Delete(ctx context.Context, id int64) error
}

// StudentService — Порт для применения бизнес-логики.
type StudentService interface {
	GetStudent(ctx context.Context, id int64) (*Student, error)
	GetAllStudents(ctx context.Context) ([]Student, error)
	GetStudentsByGroup(ctx context.Context, group string) ([]Student, error)
	RegisterStudent(ctx context.Context, s *Student) (*Student, error)
	ModifyStudent(ctx context.Context, s *Student) (*Student, error)
	RemoveStudent(ctx context.Context, id int64) error
}
