package handlers

import (
	"bytes"
	"encoding/json"
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

func setupStudentRoutes(r *gin.Engine, h *StudentHandler) {
	api := r.Group("/api/v1")
	api.GET("/students", h.GetAll)
	api.GET("/students/:id", h.GetByID)
	api.POST("/students", h.Create)
	api.PUT("/students/:id", h.Update)
	api.DELETE("/students/:id", h.Delete)
}

func TestStudentHandler_GetAll_Success(t *testing.T) {
	mockSvc := mocks.NewStudentService(t)
	h := NewStudentHandler(mockSvc)

	expected := []domain.Student{{ID: 1, Fio: "Ivan", GroupName: "101"}}
	mockSvc.On("GetStudentsList", mock.Anything, "admin", int32(0), (*int32)(nil)).Return(expected, nil)

	w := httptest.NewRecorder()
	r := newTestRouter(t, "admin", 0)
	setupStudentRoutes(r, h)
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/students", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	var actual []domain.Student
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &actual))
	assert.Equal(t, expected, actual)
}

func TestStudentHandler_GetAll_WithGroupFilter(t *testing.T) {
	mockSvc := mocks.NewStudentService(t)
	h := NewStudentHandler(mockSvc)

	gid := int32(2)
	mockSvc.On("GetStudentsList", mock.Anything, "admin", int32(0), &gid).Return([]domain.Student{}, nil)

	w := httptest.NewRecorder()
	r := newTestRouter(t, "admin", 0)
	setupStudentRoutes(r, h)
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/students?group_id=2", nil))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStudentHandler_GetAll_InvalidGroupID(t *testing.T) {
	h := NewStudentHandler(mocks.NewStudentService(t))

	w := httptest.NewRecorder()
	r := newTestRouter(t, "admin", 0)
	setupStudentRoutes(r, h)
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/students?group_id=abc", nil))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStudentHandler_GetByID_Success(t *testing.T) {
	mockSvc := mocks.NewStudentService(t)
	h := NewStudentHandler(mockSvc)

	expected := &domain.Student{ID: 3, Fio: "S"}
	mockSvc.On("GetStudent", mock.Anything, "student", int32(3), int32(3)).Return(expected, nil)

	w := httptest.NewRecorder()
	r := newTestRouter(t, "student", 3)
	setupStudentRoutes(r, h)
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/students/3", nil))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStudentHandler_GetByID_Forbidden(t *testing.T) {
	mockSvc := mocks.NewStudentService(t)
	h := NewStudentHandler(mockSvc)

	mockSvc.On("GetStudent", mock.Anything, "student", int32(1), int32(3)).Return(nil, apperrors.ErrForbidden)

	w := httptest.NewRecorder()
	r := newTestRouter(t, "student", 1)
	setupStudentRoutes(r, h)
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/students/3", nil))

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestStudentHandler_Create_Success(t *testing.T) {
	mockSvc := mocks.NewStudentService(t)
	h := NewStudentHandler(mockSvc)

	body := `{"email":"new@example.com","password":"secret12","phone_number":"+79001112233","group_id":1,"fio":"New"}`
	created := &domain.Student{ID: 9, Fio: "New"}
	mockSvc.On("CreateStudent", mock.Anything, "admin", mock.AnythingOfType("domain.CreateStudentInput")).Return(created, nil)

	w := httptest.NewRecorder()
	r := newTestRouter(t, "admin", 0)
	setupStudentRoutes(r, h)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/students", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestStudentHandler_Update_BadJSON(t *testing.T) {
	mockSvc := mocks.NewStudentService(t)
	h := NewStudentHandler(mockSvc)

	w := httptest.NewRecorder()
	r := newTestRouter(t, "admin", 0)
	setupStudentRoutes(r, h)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/students/1", bytes.NewBufferString("{bad json}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertNotCalled(t, "ModifyStudent")
}

func TestStudentHandler_Update_Success(t *testing.T) {
	mockSvc := mocks.NewStudentService(t)
	h := NewStudentHandler(mockSvc)

	body := `{"fio":"Updated","group_id":1}`
	mockSvc.On("ModifyStudent", mock.Anything, "admin", int32(0), int32(1), "Updated", int32(1)).Return(nil)

	w := httptest.NewRecorder()
	r := newTestRouter(t, "admin", 0)
	setupStudentRoutes(r, h)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/students/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStudentHandler_Delete_Success(t *testing.T) {
	mockSvc := mocks.NewStudentService(t)
	h := NewStudentHandler(mockSvc)

	mockSvc.On("RemoveStudent", mock.Anything, "admin", int32(2)).Return(nil)

	w := httptest.NewRecorder()
	r := newTestRouter(t, "admin", 0)
	setupStudentRoutes(r, h)
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/students/2", nil))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStudentHandler_Delete_InvalidID(t *testing.T) {
	h := NewStudentHandler(mocks.NewStudentService(t))

	w := httptest.NewRecorder()
	r := newTestRouter(t, "admin", 0)
	setupStudentRoutes(r, h)
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/students/x", nil))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
