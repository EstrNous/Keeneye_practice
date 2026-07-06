-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    EXECUTE format('ALTER DATABASE %I REFRESH COLLATION VERSION', current_database());
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'collation refresh skipped: %', SQLERRM;
END $$;
-- +goose StatementEnd

-- +goose Down
