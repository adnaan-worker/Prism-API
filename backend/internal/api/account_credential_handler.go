package api

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AccountCredentialHandler struct {
	credentialService *service.AccountCredentialService
}

func NewAccountCredentialHandler(credentialService *service.AccountCredentialService) *AccountCredentialHandler {
	return &AccountCredentialHandler{
		credentialService: credentialService,
	}
}

// GetCredentials godoc
// @Summary Get all credentials
// @Description Get all account credentials with optional filters
// @Tags Account Credentials
// @Accept json
// @Produce json
// @Param pool_id query int false "Filter by pool ID"
// @Param provider query string false "Filter by provider"
// @Param status query string false "Filter by status"
// @Success 200 {array} models.AccountCredential
// @Router /api/admin/account-credentials [get]
func (h *AccountCredentialHandler) GetCredentials(c *gin.Context) {
	var poolID *uint
	if poolIDStr := c.Query("pool_id"); poolIDStr != "" {
		id, err := strconv.ParseUint(poolIDStr, 10, 32)
		if err == nil {
			pid := uint(id)
			poolID = &pid
		}
	}

	provider := c.Query("provider")
	if provider == "" {
		provider = ""
	}

	status := c.Query("status")
	if status == "" {
		status = ""
	}

	credentials, err := h.credentialService.GetCredentials(c.Request.Context(), poolID, provider, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, credentials)
}

// GetCredential 获取指定凭据
func (h *AccountCredentialHandler) GetCredential(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credential ID"})
		return
	}

	credential, err := h.credentialService.GetCredential(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if credential == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Credential not found"})
		return
	}

	c.JSON(http.StatusOK, credential)
}

// CreateCredential godoc
// @Summary Create a new credential
// @Description Create a new account credential
// @Tags Account Credentials
// @Accept json
// @Produce json
// @Param credential body models.AccountCredential true "Credential data"
// @Success 201 {object} models.AccountCredential
// @Router /api/admin/account-credentials [post]
func (h *AccountCredentialHandler) CreateCredential(c *gin.Context) {
	var req models.AccountCredential
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	credential, err := h.credentialService.CreateCredential(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, credential)
}

// UpdateCredential godoc
// @Summary Update a credential
// @Description Update an existing account credential
// @Tags Account Credentials
// @Accept json
// @Produce json
// @Param id path int true "Credential ID"
// @Param credential body models.AccountCredential true "Credential data"
// @Success 200 {object} models.AccountCredential
// @Router /api/admin/account-credentials/:id [put]
func (h *AccountCredentialHandler) UpdateCredential(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credential ID"})
		return
	}

	var req models.AccountCredential
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ID = uint(id)
	credential, err := h.credentialService.UpdateCredential(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, credential)
}

// DeleteCredential godoc
// @Summary Delete a credential
// @Description Delete an account credential
// @Tags Account Credentials
// @Accept json
// @Produce json
// @Param id path int true "Credential ID"
// @Success 204
// @Router /api/admin/account-credentials/:id [delete]
func (h *AccountCredentialHandler) DeleteCredential(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credential ID"})
		return
	}

	if err := h.credentialService.DeleteCredential(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// RefreshCredential godoc
// @Summary Refresh credential token
// @Description Manually trigger token refresh for a credential
// @Tags Account Credentials
// @Accept json
// @Produce json
// @Param id path int true "Credential ID"
// @Success 200 {object} models.AccountCredential
// @Router /api/admin/account-credentials/:id/refresh [post]
func (h *AccountCredentialHandler) RefreshCredential(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credential ID"})
		return
	}

	credential, err := h.credentialService.RefreshCredential(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, credential)
}

// UpdateCredentialStatus godoc
// @Summary Update credential status
// @Description Update the status of a credential
// @Tags Account Credentials
// @Accept json
// @Produce json
// @Param id path int true "Credential ID"
// @Param status body object{is_active=bool} true "Status data"
// @Success 200 {object} models.AccountCredential
// @Router /api/admin/account-credentials/:id/status [put]
func (h *AccountCredentialHandler) UpdateCredentialStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credential ID"})
		return
	}

	var req struct {
		IsActive bool `json:"is_active" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	credential, err := h.credentialService.UpdateCredentialStatus(c.Request.Context(), uint(id), req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, credential)
}
