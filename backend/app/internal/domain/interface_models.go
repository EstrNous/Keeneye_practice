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
	ID           int32  `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
	PhoneNumber  string `json:"phone_number"`
}

type Student struct {
	ID        int32  `json:"id"`
	UserID    int32  `json:"user_id"`
	GroupID   int32  `json:"group_id"`
	GroupName string `json:"group_name"`
	Fio       string `json:"fio"`
}

type Teacher struct {
	ID          int32  `json:"id"`
	UserID      int32  `json:"user_id"`
	Fio         string `json:"fio"`
	Email       string `json:"email,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

type Group struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type AuthService interface {
	Login(ctx context.Context, email, password string) (accessToken string, refreshToken string, err error)
	RefreshTokens(ctx context.Context, refreshToken string) (newAccessToken string, newRefreshToken string, err error)
	Register(ctx context.Context, email, password, role, phone, profileFIO string, groupID *int32) (int32, error)
}

type StudentService interface {
	GetStudentsList(ctx context.Context, role string, profileID int32, groupID *int32) ([]Student, error)
	GetStudent(ctx context.Context, role string, actorProfileID int32, targetID int32) (*Student, error)
	CreateStudent(ctx context.Context, role string, req CreateStudentInput) (*Student, error)
	ModifyStudent(ctx context.Context, role string, actorProfileID int32, targetStudentID int32, fio string, groupID int32) error
	RemoveStudent(ctx context.Context, role string, id int32) error
}

type CreateStudentInput struct {
	Email       string
	Password    string
	PhoneNumber string
	GroupID     int32
	Fio         string
}

type StudentRepository interface {
	GetByID(ctx context.Context, id int32) (*Student, error)
	Create(ctx context.Context, userID int32, groupID int32, fio string) (*Student, error)
	Update(ctx context.Context, id int32, fio string, groupID int32) error
	DeleteWithUser(ctx context.Context, studentID int32, userID int32) error

	ListAll(ctx context.Context) ([]Student, error)
	ListByGroupID(ctx context.Context, groupID int32) ([]Student, error)
	ListClassmates(ctx context.Context, studentID int32) ([]Student, error)
	ListByTeacherGroups(ctx context.Context, teacherID int32) ([]Student, error)

	GetByUserID(ctx context.Context, userID int32) (*Student, error)
	CheckTeacherGroup(ctx context.Context, teacherID, groupID int32) (bool, error)
}

type TeacherService interface {
	ListTeachers(ctx context.Context, role string, actorProfileID int32) ([]Teacher, error)
	GetTeacher(ctx context.Context, role string, actorProfileID int32, targetID int32) (*Teacher, error)
	UpdateTeacher(ctx context.Context, role string, actorProfileID int32, targetID int32, fio string) error
	DeleteTeacher(ctx context.Context, role string, targetID int32) error
	ListTeacherGroups(ctx context.Context, role string, actorProfileID int32, teacherID int32) ([]Group, error)
	AssignGroup(ctx context.Context, role string, teacherID int32, groupID int32) error
	RemoveGroup(ctx context.Context, role string, teacherID int32, groupID int32) error
}

type TeacherRepository interface {
	ListAll(ctx context.Context) ([]Teacher, error)
	GetByID(ctx context.Context, id int32) (*Teacher, error)
	Update(ctx context.Context, id int32, fio string) error
	DeleteWithUser(ctx context.Context, teacherID int32, userID int32) error
	ListGroups(ctx context.Context, teacherID int32) ([]Group, error)
	AssignGroup(ctx context.Context, teacherID int32, groupID int32) error
	RemoveGroup(ctx context.Context, teacherID int32, groupID int32) error
}

type GroupService interface {
	ListGroups(ctx context.Context, role string) ([]Group, error)
	CreateGroup(ctx context.Context, role string, name string) (*Group, error)
}

type GroupRepository interface {
	ListAll(ctx context.Context) ([]Group, error)
	GetByID(ctx context.Context, id int32) (*Group, error)
	Create(ctx context.Context, name string) (*Group, error)
}
