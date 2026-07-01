-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (email, password_hash, role)
VALUES ($1, $2, $3) RETURNING *;

-- name: ListAllStudents :many
SELECT s.*, g.name as group_name FROM students s
                                          LEFT JOIN groups g ON s.group_id = g.id;

-- name: GetClassmates :many
SELECT s.*, g.name as group_name FROM students s
                                          LEFT JOIN groups g ON s.group_id = g.id
WHERE s.group_id = (SELECT group_id FROM students as s WHERE s.id = $1);

-- name: GetStudentsByTeacherGroups :many
SELECT s.*, g.name as group_name FROM students s
                                          JOIN groups g ON s.group_id = g.id
                                          JOIN teacher_groups tg ON tg.group_id = s.group_id
WHERE tg.teacher_id = $1;

-- name: CreateStudent :one
INSERT INTO students (user_id, group_id, fio, phone_number)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdateStudent :exec
UPDATE students SET group_id = $2, fio = $3, phone_number = $4 WHERE id = $1;

-- name: CreateTeacher :one
INSERT INTO teachers (user_id, fio) VALUES ($1, $2) RETURNING *;

-- name: AssignTeacherToGroup :exec
INSERT INTO teacher_groups (teacher_id, group_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: RemoveTeacherFromGroup :exec
DELETE FROM teacher_groups WHERE teacher_id = $1 AND group_id = $2;

-- name: GetStudentByID :one
SELECT s.*, g.name as group_name FROM students s
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