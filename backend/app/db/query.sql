-- name: GetStudent :one
SELECT * FROM students
WHERE id = $1 LIMIT 1;

-- name: ListStudents :many
SELECT * FROM students
ORDER BY id;

-- name: ListStudentsByGroup :many
SELECT * FROM students
WHERE group_of_students = $1
ORDER BY fio;

-- name: CreateStudent :one
INSERT INTO students (fio, group_of_students, phone_number)
VALUES ($1, $2, $3)
    RETURNING *;

-- name: UpdateStudent :one
UPDATE students
SET fio = $2, group_of_students = $3, phone_number = $4
WHERE id = $1
    RETURNING *;

-- name: DeleteStudent :exec
DELETE FROM students
WHERE id = $1;