-- +goose Up
CREATE TYPE registration_request_status AS ENUM ('pending', 'completed', 'revoked', 'failed');

CREATE TYPE registration_email_outbox_status AS ENUM ('pending', 'sent', 'failed');

CREATE TABLE registration_batches (
    id UUID PRIMARY KEY,
    created_by INT NOT NULL REFERENCES users(id),
    total_rows INT NOT NULL DEFAULT 0,
    success_count INT NOT NULL DEFAULT 0,
    error_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE registration_requests (
    id SERIAL PRIMARY KEY,
    batch_id UUID NOT NULL REFERENCES registration_batches(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    fio VARCHAR(255) NOT NULL,
    role user_role NOT NULL,
    group_id INT REFERENCES groups(id),
    group_name VARCHAR(50),
    invite_token VARCHAR(64) NOT NULL,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    status registration_request_status NOT NULL DEFAULT 'pending',
    expires_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT registration_requests_role_check CHECK (role IN ('student', 'teacher')),
    CONSTRAINT registration_requests_student_group_check CHECK (
        (role = 'student' AND group_id IS NOT NULL) OR
        (role = 'teacher' AND group_id IS NULL)
    )
);

CREATE UNIQUE INDEX registration_requests_pending_email_idx
    ON registration_requests (LOWER(email))
    WHERE status = 'pending';

CREATE TABLE registration_email_outbox (
    id SERIAL PRIMARY KEY,
    request_id INT NOT NULL REFERENCES registration_requests(id) ON DELETE CASCADE,
    status registration_email_outbox_status NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX registration_email_outbox_retry_idx
    ON registration_email_outbox (status, attempts)
    WHERE status IN ('pending', 'failed');

CREATE INDEX registration_requests_expires_idx
    ON registration_requests (expires_at)
    WHERE status = 'pending';

-- +goose Down
DROP TABLE IF EXISTS registration_email_outbox;
DROP TABLE IF EXISTS registration_requests;
DROP TABLE IF EXISTS registration_batches;
DROP TYPE IF EXISTS registration_email_outbox_status;
DROP TYPE IF EXISTS registration_request_status;
