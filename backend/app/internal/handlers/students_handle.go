package handlers

import (
	"net/http"
	"strconv"

	"keeneye_practice/app/internal/apperrors"
	"keeneye_practice/app/internal/domain"
	"keeneye_practice/app/internal/dto"

	"github.com/gin-gonic/gin"
)

type StudentHandler struct {
	svc domain.StudentService
}

func NewStudentHandler(svc domain.StudentService) *StudentHandler {
	return &StudentHandler{svc: svc}
}

func (h *StudentHandler) GetAll(c *gin.Context) {
	role := c.GetString("role")
	profileID := c.GetInt32("profileID")

	var groupID *int32
	if raw := c.Query("group_id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			_ = c.Error(apperrors.NewValidation("invalid group_id"))
			return
		}
		gid := int32(id)
		groupID = &gid
	}

	students, err := h.svc.GetStudentsList(c.Request.Context(), role, profileID, groupID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, students)
}

func (h *StudentHandler) GetByID(c *gin.Context) {
	role := c.GetString("role")
	actorProfileID := c.GetInt32("profileID")

	targetID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apperrors.NewValidation("invalid student id"))
		return
	}

	student, err := h.svc.GetStudent(c.Request.Context(), role, actorProfileID, int32(targetID))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, student)
}

func (h *StudentHandler) Create(c *gin.Context) {
	role := c.GetString("role")

	var req dto.CreateStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewValidation(err.Error()))
		return
	}

	student, err := h.svc.CreateStudent(c.Request.Context(), role, domain.CreateStudentInput{
		Email: req.Email, Password: req.Password,
		PhoneNumber: req.PhoneNumber, GroupID: req.GroupID, Fio: req.Fio,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, student)
}

func (h *StudentHandler) Update(c *gin.Context) {
	role := c.GetString("role")
	actorProfileID := c.GetInt32("profileID")

	targetID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apperrors.NewValidation("invalid student id"))
		return
	}

	var req dto.UpdateStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewValidation(err.Error()))
		return
	}

	if err := h.svc.ModifyStudent(c.Request.Context(), role, actorProfileID, int32(targetID), req.Fio, req.GroupID); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *StudentHandler) Delete(c *gin.Context) {
	role := c.GetString("role")

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apperrors.NewValidation("invalid id"))
		return
	}

	if err := h.svc.RemoveStudent(c.Request.Context(), role, int32(id)); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
