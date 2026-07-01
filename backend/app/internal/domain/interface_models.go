package domain

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID    int32  `json:"user_id"`
	Role      string `json:"role"`
	ProfileID int32  `json:"profile_id,omitempty"`
}

type User struct {
	ID           int32
	Email        string
	PasswordHash string
	Role         string
}

type Student struct {
	ID          int32  `json:"id"`
	UserID      int32  `json:"user_id"`
	GroupID     int32  `json:"group_id"`
	GroupName   string `json:"group_name"`
	Fio         string `json:"fio"`
	PhoneNumber string `json:"phone_number"`
}

type Teacher struct {
	ID     int32  `json:"id"`
	UserID int32  `json:"user_id"`
	Fio    string `json:"fio"`
}

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, error)
	Register(ctx context.Context, email, password, role string, profileFIO string, groupID *int32) error
}

type StudentService interface {
	GetStudentsList(ctx context.Context, role string, profileID int32) ([]Student, error)
	GetStudent(ctx context.Context, role string, actorProfileID int32, targetID int32) (*Student, error)
	ModifyStudent(ctx context.Context, role string, actorProfileID int32, targetStudentID int32, s *Student) error
	RemoveStudent(ctx context.Context, id int32) error
}

type StudentRepository interface {
	GetByID(ctx context.Context, id int32) (*Student, error)
	Create(ctx context.Context, s *Student) (*Student, error)
	Update(ctx context.Context, id int32, s *Student) error
	Delete(ctx context.Context, id int32) error

	ListAll(ctx context.Context) ([]Student, error)
	ListClassmates(ctx context.Context, studentID int32) ([]Student, error)
	ListByTeacherGroups(ctx context.Context, teacherID int32) ([]Student, error)

	CheckTeacherGroup(ctx context.Context, teacherID, groupID int32) (bool, error)
}
