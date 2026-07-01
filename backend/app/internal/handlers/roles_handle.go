package handlers

import (
	"keeneye_practice/app/internal/domain"
	"net/http"
	"strconv"

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

	students, err := h.svc.GetStudentsList(c.Request.Context(), role, profileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, students)
}

func (h *StudentHandler) GetByID(c *gin.Context) {
	role := c.GetString("role")
	actorProfileID := c.GetInt32("profileID")

	targetID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student id"})
		return
	}

	student, err := h.svc.GetStudent(c.Request.Context(), role, actorProfileID, int32(targetID))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, student)
}

func (h *StudentHandler) Update(c *gin.Context) {
	role := c.GetString("role")
	actorProfileID := c.GetInt32("profileID")

	var req domain.Student
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.ModifyStudent(c.Request.Context(), role, actorProfileID, req.ID, &req); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *StudentHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.svc.RemoveStudent(c.Request.Context(), int32(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
