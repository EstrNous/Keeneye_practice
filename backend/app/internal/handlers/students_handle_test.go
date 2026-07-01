package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"keeneye_practice/app/internal/domain"
	"keeneye_practice/app/internal/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStudentHandler_GetAll_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := mocks.NewStudentService(t)
	h := NewStudentHandler(mockSvc)

	expected := []domain.Student{
		{ID: 1, Fio: "Ivan", GroupName: "101"},
	}

	mockSvc.On("GetStudentsList", mock.Anything, "admin", int32(0)).Return(expected, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("role", "admin")
	c.Set("profileID", int32(0))

	c.Request = httptest.NewRequest(http.MethodGet, "/api/base/students", nil)

	h.GetAll(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var actual []domain.Student
	err := json.Unmarshal(w.Body.Bytes(), &actual)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestStudentHandler_Update_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := mocks.NewStudentService(t)
	h := NewStudentHandler(mockSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("role", "admin")
	c.Set("profileID", int32(0))

	c.Request = httptest.NewRequest(http.MethodPut, "/api/base/students", bytes.NewBufferString("{bad json}"))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertNotCalled(t, "ModifyStudent")
}
