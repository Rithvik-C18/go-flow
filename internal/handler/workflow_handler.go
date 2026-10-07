package handler

import (
	"net/http"

	"github.com/Rithvik-C18/go-flow/internal/service"
	"github.com/gin-gonic/gin"
)

type WorkflowHandler struct {
	service *service.WorkflowService
}

func NewWorkflowHandler(service *service.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{service: service}
}

func (h *WorkflowHandler) ListWorkflows(c *gin.Context) {
	workflows, err := h.service.ForUser(c.GetUint("userId")).List()
	if err != nil {
		respondError(c, err)
		return
	}
	respondJSON(c, http.StatusOK, workflows)
}

func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
	var req struct {
		Id   string `json:"id" binding:"required"`
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, "id is required")
		return
	}

	if err := h.service.ForUser(c.GetUint("userId")).Create(req.Id, req.Name); err != nil {
		respondError(c, err)
		return
	}

	respondJSON(c, http.StatusCreated, gin.H{"id": req.Id, "name": req.Name})
}

func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
	workflow, err := h.service.ForUser(c.GetUint("userId")).Get(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	respondJSON(c, http.StatusOK, workflow)
}

func (h *WorkflowHandler) DeleteWorkflow(c *gin.Context) {
	if err := h.service.ForUser(c.GetUint("userId")).Delete(c.Param("id")); err != nil {
		respondError(c, err)
		return
	}
	respondJSON(c, http.StatusOK, gin.H{"deleted": c.Param("id")})
}

func (h *WorkflowHandler) RunWorkflow(c *gin.Context) {
	var params map[string]any
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&params); err != nil {
			respondValidationError(c, "invalid params body")
			return
		}
	}

	if err := h.service.ForUser(c.GetUint("userId")).Run(c.Param("id"), params); err != nil {
		respondError(c, err)
		return
	}

	respondJSON(c, http.StatusOK, gin.H{"status": "completed"})
}

func (h *WorkflowHandler) ListNodes(c *gin.Context) {
	workflow, err := h.service.ForUser(c.GetUint("userId")).Get(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	respondJSON(c, http.StatusOK, workflow.Nodes)
}

func (h *WorkflowHandler) AddNode(c *gin.Context) {
	var req struct {
		Id     string         `json:"id" binding:"required"`
		Type   string         `json:"type" binding:"required"`
		Config map[string]any `json:"config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, "id and type are required")
		return
	}

	if err := h.service.ForUser(c.GetUint("userId")).AddNode(c.Param("id"), req.Id, req.Type, req.Config); err != nil {
		respondError(c, err)
		return
	}

	respondJSON(c, http.StatusCreated, gin.H{"id": req.Id, "type": req.Type})
}

func (h *WorkflowHandler) DeleteNode(c *gin.Context) {
	if err := h.service.ForUser(c.GetUint("userId")).DeleteNode(c.Param("id"), c.Param("nodeId")); err != nil {
		respondError(c, err)
		return
	}
	respondJSON(c, http.StatusOK, gin.H{"deleted": c.Param("nodeId")})
}

func (h *WorkflowHandler) AddEdge(c *gin.Context) {
	var req struct {
		From string `json:"from" binding:"required"`
		To   string `json:"to" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, "from and to are required")
		return
	}

	if err := h.service.ForUser(c.GetUint("userId")).AddEdge(c.Param("id"), req.From, req.To); err != nil {
		respondError(c, err)
		return
	}

	respondJSON(c, http.StatusCreated, gin.H{"from": req.From, "to": req.To})
}

func (h *WorkflowHandler) DeleteEdge(c *gin.Context) {
	var req struct {
		From string `json:"from" binding:"required"`
		To   string `json:"to" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, "from and to are required")
		return
	}

	if err := h.service.ForUser(c.GetUint("userId")).DeleteEdge(c.Param("id"), req.From, req.To); err != nil {
		respondError(c, err)
		return
	}

	respondJSON(c, http.StatusOK, gin.H{"deleted": req.From + "->" + req.To})
}
