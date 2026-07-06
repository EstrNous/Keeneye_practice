package handlers

import (
	"net/http"
	"strconv"

	"keeneye_practice/app/internal/apperrors"
	"keeneye_practice/app/internal/domain"
	"keeneye_practice/app/internal/dto"

	"github.com/gin-gonic/gin"
)

type TeacherHandler struct {
	svc domain.TeacherService
}

func NewTeacherHandler(svc domain.TeacherService) *TeacherHandler {
	return &TeacherHandler{svc: svc}
}

func (h *TeacherHandler) List(c *gin.Context) {
	role := c.GetString("role")
	profileID := c.GetInt32("profileID")

	teachers, err := h.svc.ListTeachers(c.Request.Context(), role, profileID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, teachers)
}

func (h *TeacherHandler) GetByID(c *gin.Context) {
	role := c.GetString("role")
	profileID := c.GetInt32("profileID")

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apperrors.NewValidation("invalid teacher id"))
		return
	}

	teacher, err := h.svc.GetTeacher(c.Request.Context(), role, profileID, int32(id))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, teacher)
}

func (h *TeacherHandler) Update(c *gin.Context) {
	role := c.GetString("role")
	profileID := c.GetInt32("profileID")

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apperrors.NewValidation("invalid teacher id"))
		return
	}

	var req dto.UpdateTeacherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewValidation(err.Error()))
		return
	}

	if err := h.svc.UpdateTeacher(c.Request.Context(), role, profileID, int32(id), req.Fio); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *TeacherHandler) Delete(c *gin.Context) {
	role := c.GetString("role")

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apperrors.NewValidation("invalid teacher id"))
		return
	}

	if err := h.svc.DeleteTeacher(c.Request.Context(), role, int32(id)); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *TeacherHandler) ListGroups(c *gin.Context) {
	role := c.GetString("role")
	profileID := c.GetInt32("profileID")

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apperrors.NewValidation("invalid teacher id"))
		return
	}

	groups, err := h.svc.ListTeacherGroups(c.Request.Context(), role, profileID, int32(id))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, groups)
}

func (h *TeacherHandler) AssignGroup(c *gin.Context) {
	role := c.GetString("role")

	teacherID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apperrors.NewValidation("invalid teacher id"))
		return
	}

	groupID, err := strconv.ParseInt(c.Param("groupId"), 10, 32)
	if err != nil {
		_ = c.Error(apperrors.NewValidation("invalid group id"))
		return
	}

	if err := h.svc.AssignGroup(c.Request.Context(), role, int32(teacherID), int32(groupID)); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "assigned"})
}

func (h *TeacherHandler) RemoveGroup(c *gin.Context) {
	role := c.GetString("role")

	teacherID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apperrors.NewValidation("invalid teacher id"))
		return
	}

	groupID, err := strconv.ParseInt(c.Param("groupId"), 10, 32)
	if err != nil {
		_ = c.Error(apperrors.NewValidation("invalid group id"))
		return
	}

	if err := h.svc.RemoveGroup(c.Request.Context(), role, int32(teacherID), int32(groupID)); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}
