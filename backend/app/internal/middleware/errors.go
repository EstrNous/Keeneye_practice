package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"keeneye_practice/app/internal/apperrors"
	"keeneye_practice/app/internal/dto"

	"github.com/gin-gonic/gin"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			b := make([]byte, 16)
			if _, err := randRead(b); err == nil {
				id = hexEncode(b)
			}
		}
		c.Set("requestID", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		if c.Writer.Written() {
			return
		}
		err := c.Errors.Last().Err
		status, body := mapError(err)
		if status >= http.StatusInternalServerError {
			slog.Error("request failed",
				"request_id", c.GetString("requestID"),
				"path", c.Request.URL.Path,
				"error", err.Error(),
			)
		}
		c.JSON(status, dto.ErrorResponse{Error: body})
	}
}

func mapError(err error) (int, dto.ErrorBody) {
	var ve *apperrors.ValidationError
	if errors.As(err, &ve) {
		return http.StatusBadRequest, dto.ErrorBody{Code: "validation_error", Message: ve.Error()}
	}

	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		return http.StatusNotFound, dto.ErrorBody{Code: "not_found", Message: "resource not found"}
	case errors.Is(err, apperrors.ErrForbidden):
		return http.StatusForbidden, dto.ErrorBody{Code: "forbidden", Message: "access denied"}
	case errors.Is(err, apperrors.ErrUnauthorized):
		return http.StatusUnauthorized, dto.ErrorBody{Code: "unauthorized", Message: "unauthorized"}
	case errors.Is(err, apperrors.ErrConflict):
		return http.StatusConflict, dto.ErrorBody{Code: "conflict", Message: "resource conflict"}
	case errors.Is(err, apperrors.ErrValidation):
		return http.StatusBadRequest, dto.ErrorBody{Code: "validation_error", Message: "validation failed"}
	default:
		return http.StatusInternalServerError, dto.ErrorBody{Code: "internal_error", Message: "internal server error"}
	}
}
