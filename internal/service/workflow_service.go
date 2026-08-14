package service

import (
	"encoding/json"
	"fmt"

	ex "github.com/Rithvik-C18/go-flow/internal/execution"
	"github.com/Rithvik-C18/go-flow/internal/repository"
	wf "github.com/Rithvik-C18/go-flow/internal/workflow"
)

type NodeDTO struct {
	ID     string         `json:"id"`
	Type   string         `json:"type"`
	Config map[string]any `json:"config"`
}

type EdgeDTO struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type WorkflowSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type WorkflowDetail struct {
	ID    string    `json:"id"`
	Name  string    `json:"name"`
	Nodes []NodeDTO `json:"nodes"`
	Edges []EdgeDTO `json:"edges"`
}

type WorkflowService struct {
	repo     *repository.WorkflowRepository
	executor *ex.Executor
}

func NewWorkflowService(repo *repository.WorkflowRepository, executor *ex.Executor) *WorkflowService {
	return &WorkflowService{
		repo:     repo,
		executor: executor,
	}
}

func (s *WorkflowService) List() ([]WorkflowSummary, error) {
	records, err := s.repo.List()
	if err != nil {
		return nil, err
	}

	workflows := make([]WorkflowSummary, 0, len(records))
	for _, record := range records {
		workflows = append(workflows, WorkflowSummary{ID: record.ID, Name: record.Name})
	}

	return workflows, nil
}

func (s *WorkflowService) Create(id, name string) error {
	return s.repo.Create(id, name)
}

func (s *WorkflowService) Get(id string) (*WorkflowDetail, error) {
	workflow, err := s.repo.Load(id)
	if err != nil {
		return nil, err
	}
	return toDetail(workflow), nil
}

func (s *WorkflowService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *WorkflowService) AddNode(workflowID, id, nodeType string, config map[string]any) error {
	if !ex.IsValidNodeType(nodeType) {
		return fmt.Errorf("unknown node type %q: %w", nodeType, ErrValidation)
	}

	workflow, err := s.repo.Load(workflowID)
	if err != nil {
		return err
	}

	if err := workflow.AddNode(&wf.NodeDefinition{Id: id, Type: nodeType, Config: config}); err != nil {
		return err
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to encode node config: %w", err)
	}

	return s.repo.AddNode(workflowID, id, nodeType, string(configJSON))
}

func (s *WorkflowService) DeleteNode(workflowID, nodeID string) error {
	workflow, err := s.repo.Load(workflowID)
	if err != nil {
		return err
	}

	if err := workflow.DeleteNode(nodeID); err != nil {
		return err
	}

	return s.repo.DeleteNode(workflowID, nodeID)
}

func (s *WorkflowService) AddEdge(workflowID, from, to string) error {
	workflow, err := s.repo.Load(workflowID)
	if err != nil {
		return err
	}

	if err := workflow.AddEdge(from, to); err != nil {
		return err
	}

	return s.repo.AddEdge(workflowID, from, to)
}

func (s *WorkflowService) DeleteEdge(workflowID, from, to string) error {
	workflow, err := s.repo.Load(workflowID)
	if err != nil {
		return err
	}

	if err := workflow.DeleteEdge(from, to); err != nil {
		return err
	}

	return s.repo.DeleteEdge(workflowID, from, to)
}

func (s *WorkflowService) Run(workflowID string, params map[string]any) error {
	workflow, err := s.repo.Load(workflowID)
	if err != nil {
		return err
	}

	return s.executor.Run(workflow, params)
}

func toDetail(workflow *wf.Workflow) *WorkflowDetail {
	detail := &WorkflowDetail{
		ID:    workflow.Id,
		Name:  workflow.Name,
		Nodes: make([]NodeDTO, 0, len(workflow.Nodes)),
		Edges: make([]EdgeDTO, 0),
	}

	for id, node := range workflow.Nodes {
		detail.Nodes = append(detail.Nodes, NodeDTO{
			ID:     id,
			Type:   node.Type,
			Config: node.Config,
		})
	}

	for from, neighbors := range workflow.Edges {
		for _, to := range neighbors {
			detail.Edges = append(detail.Edges, EdgeDTO{From: from, To: to})
		}
	}

	return detail
}
