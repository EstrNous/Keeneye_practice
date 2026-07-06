package handlers

import (
	"testing"

	"keeneye_practice/app/internal/middleware"
	"keeneye_practice/app/internal/validators"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
)

func newTestRouter(t *testing.T, role string, profileID int32) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		require.NoError(t, validators.RegisterCustomValidations(v))
	}
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	if role != "" {
		r.Use(func(c *gin.Context) {
			c.Set("role", role)
			c.Set("profileID", profileID)
			c.Next()
		})
	}
	return r
}
