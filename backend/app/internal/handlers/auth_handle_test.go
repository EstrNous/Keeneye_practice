package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"keeneye_practice/app/internal/apperrors"
	"keeneye_practice/app/internal/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupAuthRoutes(r *gin.Engine, h *AuthHandler) {
	r.POST("/api/v1/auth/login", h.Login)
	r.POST("/api/v1/auth/register", h.Register)
	r.POST("/api/v1/auth/refresh", h.Refresh)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockSvc := mocks.NewAuthService(t)
	h := NewAuthHandler(mockSvc)

	mockSvc.On("Login", mock.Anything, "admin@local.dev", "secret12").Return("access", "refresh", nil)

	w := httptest.NewRecorder()
	r := newTestRouter(t, "", 0)
	setupAuthRoutes(r, h)

	body, _ := json.Marshal(map[string]string{"email": "admin@local.dev", "password": "secret12"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "access_token")
	assert.Contains(t, w.Body.String(), "refresh_token")
}

func TestAuthHandler_Login_Unauthorized(t *testing.T) {
	mockSvc := mocks.NewAuthService(t)
	h := NewAuthHandler(mockSvc)

	mockSvc.On("Login", mock.Anything, "admin@local.dev", "wrongpass").Return("", "", apperrors.ErrUnauthorized)

	w := httptest.NewRecorder()
	r := newTestRouter(t, "", 0)
	setupAuthRoutes(r, h)

	body, _ := json.Marshal(map[string]string{"email": "admin@local.dev", "password": "wrongpass"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Login_InvalidEmail(t *testing.T) {
	h := NewAuthHandler(mocks.NewAuthService(t))

	w := httptest.NewRecorder()
	r := newTestRouter(t, "", 0)
	setupAuthRoutes(r, h)

	body := []byte(`{"email":"not-an-email","password":"secret12"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Register_Success(t *testing.T) {
	mockSvc := mocks.NewAuthService(t)
	h := NewAuthHandler(mockSvc)

	mockSvc.On("Register", mock.Anything, "user@example.com", "secret12", "student", "+79001112233", "User", mock.Anything).Return(int32(1), nil)

	w := httptest.NewRecorder()
	r := newTestRouter(t, "", 0)
	setupAuthRoutes(r, h)

	body := `{"email":"user@example.com","password":"secret12","role":"student","phone_number":"+79001112233","profile_fio":"User","group_id":1}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAuthHandler_Refresh_Success(t *testing.T) {
	mockSvc := mocks.NewAuthService(t)
	h := NewAuthHandler(mockSvc)

	mockSvc.On("RefreshTokens", mock.Anything, "old-refresh").Return("new-access", "new-refresh", nil)

	w := httptest.NewRecorder()
	r := newTestRouter(t, "", 0)
	setupAuthRoutes(r, h)

	body := `{"refresh_token":"old-refresh"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "new-access")
}
