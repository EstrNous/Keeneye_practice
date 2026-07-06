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

func setupTeacherRoutes(r *gin.Engine, h *TeacherHandler) {
	api := r.Group("/api/v1")
	api.GET("/teachers", h.List)
	api.GET("/teachers/:id", h.GetByID)
	api.PUT("/teachers/:id", h.Update)
	api.POST("/teachers/:id/groups/:groupId", h.AssignGroup)
}

func TestTeacherHandler_List_Success(t *testing.T) {
	mockSvc := mocks.NewTeacherService(t)
	h := NewTeacherHandler(mockSvc)

	expected := []domain.Teacher{{ID: 1, Fio: "T"}}
	mockSvc.On("ListTeachers", mock.Anything, "admin", int32(0)).Return(expected, nil)

	w := httptest.NewRecorder()
	r := newTestRouter(t, "admin", 0)
	setupTeacherRoutes(r, h)
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/teachers", nil))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTeacherHandler_GetByID_Forbidden(t *testing.T) {
	mockSvc := mocks.NewTeacherService(t)
	h := NewTeacherHandler(mockSvc)

	mockSvc.On("GetTeacher", mock.Anything, "teacher", int32(1), int32(2)).Return(nil, apperrors.ErrForbidden)

	w := httptest.NewRecorder()
	r := newTestRouter(t, "teacher", 1)
	setupTeacherRoutes(r, h)
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/teachers/2", nil))

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestTeacherHandler_Update_Success(t *testing.T) {
	mockSvc := mocks.NewTeacherService(t)
	h := NewTeacherHandler(mockSvc)

	mockSvc.On("UpdateTeacher", mock.Anything, "teacher", int32(2), int32(2), "New Name").Return(nil)

	w := httptest.NewRecorder()
	r := newTestRouter(t, "teacher", 2)
	setupTeacherRoutes(r, h)

	body := `{"fio":"New Name"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teachers/2", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTeacherHandler_AssignGroup_Success(t *testing.T) {
	mockSvc := mocks.NewTeacherService(t)
	h := NewTeacherHandler(mockSvc)

	mockSvc.On("AssignGroup", mock.Anything, "admin", int32(1), int32(3)).Return(nil)

	w := httptest.NewRecorder()
	r := newTestRouter(t, "admin", 0)
	setupTeacherRoutes(r, h)
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/teachers/1/groups/3", nil))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTeacherHandler_GetByID_InvalidID(t *testing.T) {
	h := NewTeacherHandler(mocks.NewTeacherService(t))

	w := httptest.NewRecorder()
	r := newTestRouter(t, "admin", 0)
	setupTeacherRoutes(r, h)
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/teachers/abc", nil))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
