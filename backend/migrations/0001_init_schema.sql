-- +goose Up
CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    fio VARCHAR(255) NOT NULL,
    group_of_students VARCHAR(50) NOT NULL,
    phone_number VARCHAR(20) NOT NULL
    );

-- +goose Down
DROP TABLE IF EXISTS students;