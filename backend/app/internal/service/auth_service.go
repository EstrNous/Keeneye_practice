package service

import (
	"context"
	"errors"
	"time"

	"keeneye_practice/app/internal/db"
	"keeneye_practice/app/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	q         *db.Queries
	db        *pgxpool.Pool
	jwtSecret string
}

func NewAuthService(q *db.Queries, dbPool *pgxpool.Pool, jwtSecret string) domain.AuthService {
	return &authService{
		q:         q,
		db:        dbPool,
		jwtSecret: jwtSecret,
	}
}

func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	var profileID int32

	switch user.Role {
	case db.UserRoleStudent:
		student, err := s.q.GetStudentByUserID(ctx, pgtype.Int4{Int32: user.ID, Valid: true})
		if err != nil {
			return "", errors.New("student profile not found")
		}
		profileID = student.ID
	case db.UserRoleTeacher:
		teacher, err := s.q.GetTeacherByUserID(ctx, pgtype.Int4{Int32: user.ID, Valid: true})
		if err != nil {
			return "", errors.New("teacher profile not found")
		}
		profileID = teacher.ID
	case db.UserRoleAdmin:
		profileID = 0
	default:
		return "", errors.New("unknown user role")
	}

	claims := domain.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:    user.ID,
		Role:      string(user.Role),
		ProfileID: profileID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *authService) Register(ctx context.Context, email, password, role string, profileFIO string, groupID *int32) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qTx := s.q.WithTx(tx)

	user, err := qTx.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: string(hash),
		Role:         db.UserRole(role),
	})
	if err != nil {
		return err
	}

	switch db.UserRole(role) {
	case db.UserRoleStudent:
		if groupID == nil {
			return errors.New("group_id is required for student")
		}
		_, err = qTx.CreateStudent(ctx, db.CreateStudentParams{
			UserID:  pgtype.Int4{Int32: user.ID, Valid: true},
			GroupID: pgtype.Int4{Int32: *groupID, Valid: true},
			Fio:     profileFIO,
		})
	case db.UserRoleTeacher:
		_, err = qTx.CreateTeacher(ctx, db.CreateTeacherParams{
			UserID: pgtype.Int4{Int32: user.ID, Valid: true},
			Fio:    profileFIO,
		})
	}

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
