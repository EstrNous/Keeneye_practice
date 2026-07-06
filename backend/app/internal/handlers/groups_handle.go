package handlers

import (
	"net/http"

	"keeneye_practice/app/internal/apperrors"
	"keeneye_practice/app/internal/domain"
	"keeneye_practice/app/internal/dto"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	svc domain.GroupService
}

func NewGroupHandler(svc domain.GroupService) *GroupHandler {
	return &GroupHandler{svc: svc}
}

func (h *GroupHandler) List(c *gin.Context) {
	role := c.GetString("role")

	groups, err := h.svc.ListGroups(c.Request.Context(), role)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, groups)
}

func (h *GroupHandler) Create(c *gin.Context) {
	role := c.GetString("role")

	var req dto.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewValidation(err.Error()))
		return
	}

	group, err := h.svc.CreateGroup(c.Request.Context(), role, req.Name)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, group)
}
