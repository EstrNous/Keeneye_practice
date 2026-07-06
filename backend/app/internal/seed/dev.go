package seed

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"keeneye_practice/app/internal/db"
	"keeneye_practice/app/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

func DevData(ctx context.Context, q *db.Queries, auth domain.AuthService) error {
	count, err := q.CountUsersByEmail(ctx, "admin@local.dev")
	if err != nil {
		return err
	}
	if count > 0 {
		slog.Info("dev seed skipped: data already exists")
		return nil
	}

	password := os.Getenv("DEV_SEED_PASSWORD")
	if password == "" {
		password = "DevPassword1!"
	}

	vm, err := getOrCreateGroup(ctx, q, "VM")
	if err != nil {
		return err
	}
	av, err := getOrCreateGroup(ctx, q, "AV")
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin, err := q.CreateUser(ctx, db.CreateUserParams{
		Email:        "admin@local.dev",
		PasswordHash: string(hash),
		Role:         db.UserRoleAdmin,
		PhoneNumber:  "+79000000001",
	})
	if err != nil {
		return err
	}
	_ = admin

	teacherUserID, err := auth.Register(ctx, "teacher@local.dev", password, "teacher", "+79000000002", "Teacher One", nil)
	if err != nil {
		return err
	}

	teacher, err := q.GetTeacherByUserID(ctx, pgtype.Int4{Int32: teacherUserID, Valid: true})
	if err != nil {
		return err
	}

	if err := q.AssignTeacherToGroup(ctx, db.AssignTeacherToGroupParams{TeacherID: teacher.ID, GroupID: vm.ID}); err != nil {
		return err
	}
	if err := q.AssignTeacherToGroup(ctx, db.AssignTeacherToGroupParams{TeacherID: teacher.ID, GroupID: av.ID}); err != nil {
		return err
	}

	vmID := vm.ID
	avID := av.ID
	_, err = auth.Register(ctx, "student1@local.dev", password, "student", "+79000000003", "Student One", &vmID)
	if err != nil {
		return err
	}
	_, err = auth.Register(ctx, "student2@local.dev", password, "student", "+79000000004", "Student Two", &vmID)
	if err != nil {
		return err
	}
	_, err = auth.Register(ctx, "student3@local.dev", password, "student", "+79000000005", "Student Three", &avID)
	if err != nil {
		return err
	}

	slog.Info("dev seed data created", "admin", "admin@local.dev", "password_env", "DEV_SEED_PASSWORD")
	return nil
}

func getOrCreateGroup(ctx context.Context, q *db.Queries, name string) (db.Group, error) {
	group, err := q.GetGroupByName(ctx, name)
	if err == nil {
		return group, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return db.Group{}, err
	}
	return q.CreateGroup(ctx, name)
}
