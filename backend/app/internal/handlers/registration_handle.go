package handlers

import (
	"net/http"

	"keeneye_practice/app/internal/apperrors"
	"keeneye_practice/app/internal/domain"
	"keeneye_practice/app/internal/dto"

	"github.com/gin-gonic/gin"
)

type RegistrationHandler struct {
	svc domain.RegistrationService
}

func NewRegistrationHandler(svc domain.RegistrationService) *RegistrationHandler {
	return &RegistrationHandler{svc: svc}
}

func (h *RegistrationHandler) UploadBatch(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		_ = c.Error(apperrors.NewValidation("file is required"))
		return
	}

	f, err := file.Open()
	if err != nil {
		_ = c.Error(err)
		return
	}
	defer f.Close()

	createdBy, _ := c.Get("userID")
	adminID, ok := createdBy.(int32)
	if !ok {
		_ = c.Error(apperrors.ErrUnauthorized)
		return
	}

	result, err := h.svc.ProcessBatchCSV(c.Request.Context(), adminID, f)
	if err != nil {
		_ = c.Error(err)
		return
	}

	resp := dto.BatchUploadResponse{
		BatchID: result.BatchID,
		Total:   result.Total,
		Created: result.Created,
		Failed:  result.Failed,
		Partial: result.Failed > 0 && result.Created > 0,
	}
	for _, e := range result.Errors {
		resp.Errors = append(resp.Errors, dto.BatchRowError{
			Row: e.Row, Email: e.Email, Code: e.Code, Message: e.Message,
		})
	}

	status := http.StatusCreated
	if result.Failed > 0 && result.Created > 0 {
		status = http.StatusMultiStatus
	} else if result.Created == 0 {
		status = http.StatusBadRequest
	}

	c.JSON(status, resp)
}

func (h *RegistrationHandler) GetBatch(c *gin.Context) {
	batchID := c.Param("id")
	result, err := h.svc.GetBatchStatus(c.Request.Context(), batchID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	resp := dto.BatchStatusResponse{
		BatchID:      result.Batch.ID,
		TotalRows:    result.Batch.TotalRows,
		SuccessCount: result.Batch.SuccessCount,
		ErrorCount:   result.Batch.ErrorCount,
	}
	for _, r := range result.Requests {
		resp.Requests = append(resp.Requests, dto.RegistrationRequestItem{
			ID: r.ID, Email: r.Email, Fio: r.Fio, Role: r.Role,
			GroupName: r.GroupName, Status: r.Status,
		})
	}
	c.JSON(http.StatusOK, resp)
}

func (h *RegistrationHandler) PreviewComplete(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		_ = c.Error(apperrors.NewValidation("token is required"))
		return
	}

	preview, err := h.svc.PreviewComplete(c.Request.Context(), token)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.CompleteRegistrationPreview{
		Email: preview.Email, Fio: preview.Fio, Role: preview.Role, GroupName: preview.GroupName,
	})
}

func (h *RegistrationHandler) CompleteRegistration(c *gin.Context) {
	var req dto.CompleteRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewValidation(err.Error()))
		return
	}

	userID, err := h.svc.CompleteRegistration(c.Request.Context(), req.Token, req.Password, req.PhoneNumber)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.CompleteRegistrationResponse{UserID: userID})
}
