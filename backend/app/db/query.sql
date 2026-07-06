-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (email, password_hash, role, phone_number)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: ListAllStudents :many
SELECT s.id, s.fio, s.user_id, s.group_id, g.name as group_name FROM students s
LEFT JOIN groups g ON s.group_id = g.id;

-- name: GetStudentsByGroupID :many
SELECT s.id, s.fio, s.user_id, s.group_id, g.name as group_name FROM students s
LEFT JOIN groups g ON s.group_id = g.id
WHERE s.group_id = $1;

-- name: GetClassmates :many
SELECT s.id, s.fio, s.user_id, s.group_id, g.name as group_name FROM students s
LEFT JOIN groups g ON s.group_id = g.id
WHERE s.group_id = (SELECT group_id FROM students WHERE id = $1);

-- name: GetStudentsByTeacherGroups :many
SELECT s.id, s.fio, s.user_id, s.group_id, g.name as group_name FROM students s
JOIN groups g ON s.group_id = g.id
JOIN teacher_groups tg ON tg.group_id = s.group_id
WHERE tg.teacher_id = $1;

-- name: CreateStudent :one
INSERT INTO students (user_id, group_id, fio)
VALUES ($1, $2, $3) RETURNING *;

-- name: UpdateStudent :exec
UPDATE students SET group_id = $2, fio = $3 WHERE id = $1;

-- name: CreateTeacher :one
INSERT INTO teachers (user_id, fio) VALUES ($1, $2) RETURNING *;

-- name: ListTeachers :many
SELECT t.id, t.user_id, t.fio, u.email, u.phone_number FROM teachers t
JOIN users u ON t.user_id = u.id;

-- name: GetTeacherByID :one
SELECT t.id, t.user_id, t.fio, u.email, u.phone_number FROM teachers t
JOIN users u ON t.user_id = u.id
WHERE t.id = $1 LIMIT 1;

-- name: UpdateTeacher :exec
UPDATE teachers SET fio = $2 WHERE id = $1;

-- name: DeleteTeacher :exec
DELETE FROM teachers WHERE id = $1;

-- name: AssignTeacherToGroup :exec
INSERT INTO teacher_groups (teacher_id, group_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: RemoveTeacherFromGroup :exec
DELETE FROM teacher_groups WHERE teacher_id = $1 AND group_id = $2;

-- name: ListTeacherGroups :many
SELECT g.id, g.name FROM groups g
JOIN teacher_groups tg ON tg.group_id = g.id
WHERE tg.teacher_id = $1;

-- name: GetStudentByID :one
SELECT s.id, s.fio, s.user_id, s.group_id, g.name as group_name FROM students s
LEFT JOIN groups g ON s.group_id = g.id
WHERE s.id = $1 LIMIT 1;

-- name: DeleteStudent :exec
DELETE FROM students WHERE id = $1;

-- name: CheckTeacherHasGroup :one
SELECT EXISTS(
    SELECT 1 FROM teacher_groups
    WHERE teacher_id = $1 AND group_id = $2
);

-- name: GetStudentByUserID :one
SELECT * FROM students WHERE user_id = $1 LIMIT 1;

-- name: GetTeacherByUserID :one
SELECT * FROM teachers WHERE user_id = $1 LIMIT 1;

-- name: ListGroups :many
SELECT * FROM groups ORDER BY name;

-- name: GetGroupByID :one
SELECT * FROM groups WHERE id = $1 LIMIT 1;

-- name: GetGroupByName :one
SELECT * FROM groups WHERE name = $1 LIMIT 1;

-- name: CreateGroup :one
INSERT INTO groups (name) VALUES ($1) RETURNING *;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3) RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT * FROM refresh_tokens
WHERE token_hash = $1 AND revoked_at IS NULL
LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: RevokeAllUserRefreshTokens :exec
UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP
WHERE user_id = $1 AND revoked_at IS NULL;

-- name: CountUsersByEmail :one
SELECT COUNT(*)::bigint FROM users WHERE email = $1;
