-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone_number VARCHAR(20);

UPDATE users SET phone_number = '+79000000000' WHERE phone_number IS NULL;

ALTER TABLE users ALTER COLUMN phone_number SET NOT NULL;

ALTER TABLE students DROP COLUMN IF EXISTS phone_number;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_check;
ALTER TABLE users ADD CONSTRAINT users_email_check
    CHECK (email ~* '^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$');

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_phone_check;
ALTER TABLE users ADD CONSTRAINT users_phone_check
    CHECK (phone_number ~ '^\+7[0-9]{10}$');

-- +goose Down
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_phone_check;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_check;
ALTER TABLE users DROP COLUMN IF EXISTS phone_number;
