package dbutil

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

func Rollback(ctx context.Context, tx pgx.Tx) {
	if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		slog.Error("transaction rollback failed", "error", err)
	}
}
