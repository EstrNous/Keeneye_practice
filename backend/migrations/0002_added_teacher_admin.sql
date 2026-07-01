-- +goose Up
CREATE TYPE user_role AS ENUM ('student', 'teacher', 'admin');

CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       email VARCHAR(255) UNIQUE NOT NULL,
                       password_hash VARCHAR(255) NOT NULL,
                       role user_role NOT NULL,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE groups (
                        id SERIAL PRIMARY KEY NOT NULL,
                        name VARCHAR(50) UNIQUE NOT NULL
);

CREATE TABLE teachers (
                          id SERIAL PRIMARY KEY NOT NULL,
                          user_id INT UNIQUE REFERENCES users(id) ON DELETE CASCADE,
                          fio VARCHAR(255) NOT NULL
);

CREATE TABLE teacher_groups (
                                teacher_id INT  REFERENCES teachers(id) ON DELETE CASCADE,
                                group_id INT  REFERENCES groups(id) ON DELETE CASCADE,
                                PRIMARY KEY (teacher_id, group_id)
);

ALTER TABLE students DROP COLUMN group_of_students;

ALTER TABLE students ADD COLUMN user_id INT UNIQUE REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE students ADD COLUMN group_id INT REFERENCES groups(id) ON DELETE SET NULL;
-- +goose Down
DROP TABLE IF EXISTS teacher_groups;
DROP TABLE IF EXISTS teachers;
DROP TABLE IF EXISTS students;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_role;