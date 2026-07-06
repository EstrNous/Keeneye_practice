package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"keeneye_practice/app/internal/apperrors"
	"keeneye_practice/app/internal/domain"
	"keeneye_practice/app/internal/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupGroupRoutes(r *gin.Engine, h *GroupHandler) {
	api := r.Group("/api/v1")
	api.GET("/groups", h.List)
	api.POST("/groups", h.Create)
}

func TestGroupHandler_List_Success(t *testing.T) {
	mockSvc := mocks.NewGroupService(t)
	h := NewGroupHandler(mockSvc)

	expected := []domain.Group{{ID: 1, Name: "VM"}}
	mockSvc.On("ListGroups", mock.Anything, "teacher").Return(expected, nil)

	w := httptest.NewRecorder()
	r := newTestRouter(t, "teacher", 1)
	setupGroupRoutes(r, h)
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/groups", nil))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGroupHandler_Create_Success(t *testing.T) {
	mockSvc := mocks.NewGroupService(t)
	h := NewGroupHandler(mockSvc)

	mockSvc.On("CreateGroup", mock.Anything, "admin", "NEW").Return(&domain.Group{ID: 5, Name: "NEW"}, nil)

	w := httptest.NewRecorder()
	r := newTestRouter(t, "admin", 0)
	setupGroupRoutes(r, h)

	body := `{"name":"NEW"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestGroupHandler_Create_Forbidden(t *testing.T) {
	mockSvc := mocks.NewGroupService(t)
	h := NewGroupHandler(mockSvc)

	mockSvc.On("CreateGroup", mock.Anything, "teacher", "X").Return(nil, apperrors.ErrForbidden)

	w := httptest.NewRecorder()
	r := newTestRouter(t, "teacher", 1)
	setupGroupRoutes(r, h)

	body := `{"name":"X"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
