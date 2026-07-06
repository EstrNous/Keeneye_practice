package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"keeneye_practice/app/internal/apperrors"
	"keeneye_practice/app/internal/config"
	"keeneye_practice/app/internal/db"
	"keeneye_practice/app/internal/dbutil"
	"keeneye_practice/app/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	q          *db.Queries
	pool       *pgxpool.Pool
	jwtSecret  string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewAuthService(q *db.Queries, pool *pgxpool.Pool, cfg *config.Config) domain.AuthService {
	return &authService{
		q:          q,
		pool:       pool,
		jwtSecret:  cfg.JWTSecret,
		accessTTL:  cfg.JWTAccessTTL,
		refreshTTL: cfg.JWTRefreshTTL,
	}
}

func (s *authService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", apperrors.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", apperrors.ErrUnauthorized
	}

	profileID, err := s.resolveProfileID(ctx, user)
	if err != nil {
		return "", "", err
	}

	return s.issueTokenPair(ctx, user, profileID)
}

func (s *authService) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	hash := hashToken(refreshToken)
	stored, err := s.q.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		return "", "", apperrors.ErrUnauthorized
	}

	if !stored.ExpiresAt.Valid || stored.ExpiresAt.Time.Before(time.Now()) {
		return "", "", apperrors.ErrUnauthorized
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", "", err
	}
	defer dbutil.Rollback(ctx, tx)

	qTx := s.q.WithTx(tx)

	if err := qTx.RevokeRefreshToken(ctx, stored.ID); err != nil {
		return "", "", err
	}

	user, err := qTx.GetUserByID(ctx, stored.UserID)
	if err != nil {
		return "", "", apperrors.ErrUnauthorized
	}

	profileID, err := s.resolveProfileID(ctx, user)
	if err != nil {
		return "", "", err
	}

	access, refresh, err := s.issueTokenPairTx(ctx, qTx, user, profileID)
	if err != nil {
		return "", "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

func (s *authService) Register(ctx context.Context, email, password, role, phone, profileFIO string, groupID *int32) (int32, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer dbutil.Rollback(ctx, tx)

	qTx := s.q.WithTx(tx)

	user, err := qTx.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: string(hash),
		Role:         db.UserRole(role),
		PhoneNumber:  phone,
	})
	if err != nil {
		return 0, apperrors.MapPG(err)
	}

	switch db.UserRole(role) {
	case db.UserRoleStudent:
		if groupID == nil {
			return 0, apperrors.NewValidation("group_id is required for student")
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
	case db.UserRoleAdmin:
	default:
		return 0, apperrors.NewValidation("invalid role")
	}

	if err != nil {
		return 0, apperrors.MapPG(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return user.ID, nil
}

func (s *authService) resolveProfileID(ctx context.Context, user db.User) (int32, error) {
	switch user.Role {
	case db.UserRoleStudent:
		student, err := s.q.GetStudentByUserID(ctx, pgtype.Int4{Int32: user.ID, Valid: true})
		if err != nil {
			return 0, apperrors.ErrNotFound
		}
		return student.ID, nil
	case db.UserRoleTeacher:
		teacher, err := s.q.GetTeacherByUserID(ctx, pgtype.Int4{Int32: user.ID, Valid: true})
		if err != nil {
			return 0, apperrors.ErrNotFound
		}
		return teacher.ID, nil
	case db.UserRoleAdmin:
		return 0, nil
	default:
		return 0, errors.New("unknown user role")
	}
}

func (s *authService) issueTokenPair(ctx context.Context, user db.User, profileID int32) (string, string, error) {
	access, err := s.signAccessToken(user, profileID)
	if err != nil {
		return "", "", err
	}

	refresh, hash, expiresAt, err := newRefreshToken(s.refreshTTL)
	if err != nil {
		return "", "", err
	}

	_, err = s.q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

func (s *authService) issueTokenPairTx(ctx context.Context, qTx *db.Queries, user db.User, profileID int32) (string, string, error) {
	access, err := s.signAccessToken(user, profileID)
	if err != nil {
		return "", "", err
	}

	refresh, hash, expiresAt, err := newRefreshToken(s.refreshTTL)
	if err != nil {
		return "", "", err
	}

	_, err = qTx.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

func (s *authService) signAccessToken(user db.User, profileID int32) (string, error) {
	claims := domain.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:    user.ID,
		Role:      string(user.Role),
		ProfileID: profileID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func newRefreshToken(ttl time.Duration) (token string, hash string, expiresAt pgtype.Timestamp, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", pgtype.Timestamp{}, err
	}
	token = hex.EncodeToString(b)
	hash = hashToken(token)
	expiresAt = pgtype.Timestamp{Time: time.Now().Add(ttl), Valid: true}
	return token, hash, expiresAt, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
