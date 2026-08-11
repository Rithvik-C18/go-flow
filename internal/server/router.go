package server

import (
	"errors"
	"net/http"

	ex "github.com/Rithvik-C18/go-flow/internal/execution"
	wf "github.com/Rithvik-C18/go-flow/internal/workflow"
	"github.com/gin-gonic/gin"
)

type handler struct {
	store *Store
	exec  *ex.Executor
}

func NewRouter(store *Store) *gin.Engine {
	h := &handler{
		store: store,
		exec:  ex.NewExecutor(),
	}

	r := gin.Default()

	r.POST("/workflows", h.createWorkflow)

	workflows := r.Group("/workflows/:id")
	{
		workflows.POST("/run", h.runWorkflow)
		workflows.DELETE("", h.deleteWorkflow)

		workflows.GET("/nodes", h.listNodes)
		workflows.POST("/nodes", h.addNode)
		workflows.DELETE("/nodes/:nodeId", h.deleteNode)

		workflows.POST("/edges", h.addEdge)
		workflows.DELETE("/edges", h.deleteEdge)
	}

	return r
}

func (h *handler) fail(c *gin.Context, status int, err error) {
	c.JSON(status, gin.H{"error": err.Error()})
}

func (h *handler) createWorkflow(c *gin.Context) {
	var req struct {
		Id   string `json:"id" binding:"required"`
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.fail(c, http.StatusBadRequest, errors.New("id is required"))
		return
	}

	if err := h.store.Create(req.Id, req.Name); err != nil {
		h.fail(c, http.StatusConflict, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": req.Id, "name": req.Name})
}

func (h *handler) deleteWorkflow(c *gin.Context) {
	if err := h.store.Delete(c.Param("id")); err != nil {
		h.fail(c, http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": c.Param("id")})
}

func (h *handler) runWorkflow(c *gin.Context) {
	w, err := h.store.Get(c.Param("id"))
	if err != nil {
		h.fail(c, http.StatusNotFound, err)
		return
	}

	var params map[string]any
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&params); err != nil {
			h.fail(c, http.StatusBadRequest, errors.New("invalid params body"))
			return
		}
	}

	if err := h.exec.Run(w, params); err != nil {
		h.fail(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "completed"})
}

func (h *handler) listNodes(c *gin.Context) {
	w, err := h.store.Get(c.Param("id"))
	if err != nil {
		h.fail(c, http.StatusNotFound, err)
		return
	}

	nodes := make([]gin.H, 0, len(w.Nodes))
	for id, node := range w.Nodes {
		nodes = append(nodes, gin.H{
			"id":     id,
			"type":   node.Type,
			"config": node.Config,
		})
	}

	c.JSON(http.StatusOK, nodes)
}

func (h *handler) addNode(c *gin.Context) {
	w, err := h.store.Get(c.Param("id"))
	if err != nil {
		h.fail(c, http.StatusNotFound, err)
		return
	}

	var req struct {
		Id     string         `json:"id" binding:"required"`
		Type   string         `json:"type" binding:"required"`
		Config map[string]any `json:"config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.fail(c, http.StatusBadRequest, errors.New("id and type are required"))
		return
	}

	node, err := wf.NewNodeDefinition(req.Id, req.Type, req.Config)
	if err != nil {
		h.fail(c, http.StatusBadRequest, err)
		return
	}

	if err := w.AddNode(node); err != nil {
		h.fail(c, http.StatusConflict, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": req.Id, "type": req.Type})
}

func (h *handler) deleteNode(c *gin.Context) {
	w, err := h.store.Get(c.Param("id"))
	if err != nil {
		h.fail(c, http.StatusNotFound, err)
		return
	}

	if err := w.DeleteNode(c.Param("nodeId")); err != nil {
		h.fail(c, http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": c.Param("nodeId")})
}

func (h *handler) addEdge(c *gin.Context) {
	w, err := h.store.Get(c.Param("id"))
	if err != nil {
		h.fail(c, http.StatusNotFound, err)
		return
	}

	var req struct {
		From string `json:"from" binding:"required"`
		To   string `json:"to" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.fail(c, http.StatusBadRequest, errors.New("from and to are required"))
		return
	}

	if err := w.AddEdge(req.From, req.To); err != nil {
		h.fail(c, http.StatusConflict, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"from": req.From, "to": req.To})
}

func (h *handler) deleteEdge(c *gin.Context) {
	w, err := h.store.Get(c.Param("id"))
	if err != nil {
		h.fail(c, http.StatusNotFound, err)
		return
	}

	var req struct {
		From string `json:"from" binding:"required"`
		To   string `json:"to" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.fail(c, http.StatusBadRequest, errors.New("from and to are required"))
		return
	}

	if err := w.DeleteEdge(req.From, req.To); err != nil {
		h.fail(c, http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": req.From + "->" + req.To})
}
