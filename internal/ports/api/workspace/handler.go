package workspace

import (
	"booking/internal/entity"
	"booking/internal/usecase/workspace"
	"booking/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
)

type WorkspaceHandler struct {
	service workspace.Service
}

func NewWorkspaceHandler(service workspace.Service) *WorkspaceHandler {
	return &WorkspaceHandler{service: service}
}

func (h *WorkspaceHandler) ListWorkspaces(c *gin.Context) {
	workspaces, err := h.service.ListWorkspaces(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, workspaces)
}

type CreateWorkspaceRequest struct {
	Name        string  `json:"name" binding:"required"`
	Type        string  `json:"type" binding:"required"`
	HourlyRate  float64 `json:"hourly_rate" binding:"required"`
	Description string  `json:"description"`
	Capacity    int     `json:"capacity" binding:"required"`
}

type CreateWorkspaceResponse struct {
	ID int64 `json:"id"`
}

func (h *WorkspaceHandler) CreateWorkspace(c *gin.Context) {
	var req CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	workspace := &entity.Workspace{
		Name:        req.Name,
		Type:        req.Type,
		HourlyRate:  req.HourlyRate,
		Description: req.Description,
		Capacity:    req.Capacity,
	}
	id, err := h.service.CreateWorkspace(c.Request.Context(), workspace)
	if err != nil {
		logger.Error("Failed to create workspace: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, CreateWorkspaceResponse{ID: id})
}
