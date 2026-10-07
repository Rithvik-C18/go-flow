package service

import (
	"encoding/json"
	"fmt"

	"github.com/Rithvik-C18/go-flow/internal/execution"
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
	executor *execution.Executor
}

func NewWorkflowService(repo *repository.WorkflowRepository, executor *execution.Executor) *WorkflowService {
	return &WorkflowService{
		repo:     repo,
		executor: executor,
	}
}

func (s *WorkflowService) ForUser(userID uint) *WorkflowService {
	return &WorkflowService{repo: s.repo.ForUser(userID), executor: s.executor}
}

func (s *WorkflowService) SetExecutor(executor *execution.Executor) {
	s.executor = executor
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
	if !execution.IsValidNodeType(nodeType) {
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

func (s *WorkflowService) UpdateNode(workflowID, id, nodeType string, config map[string]any) error {
	if !execution.IsValidNodeType(nodeType) {
		return fmt.Errorf("unknown node type %q: %w", nodeType, ErrValidation)
	}
	workflow, err := s.repo.Load(workflowID)
	if err != nil {
		return err
	}
	if _, ok := workflow.Nodes[id]; !ok {
		return fmt.Errorf("node %q: %w", id, repository.ErrNotFound)
	}
	encoded, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to encode node config: %w", err)
	}
	return s.repo.UpdateNode(workflowID, id, nodeType, string(encoded))
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
	_, err := s.RunDetailed(workflowID, params)
	return err
}

func (s *WorkflowService) RunDetailed(workflowID string, params map[string]any) (map[string]map[string]any, error) {
	workflow, err := s.repo.Load(workflowID)
	if err != nil {
		return nil, err
	}

	return s.executor.RunDetailed(workflow, params)
}

func (s *WorkflowService) RunNode(workflowID, nodeID string, params map[string]any) (map[string]map[string]any, error) {
	workflow, err := s.repo.Load(workflowID)
	if err != nil {
		return nil, err
	}
	if _, ok := workflow.Nodes[nodeID]; !ok {
		return nil, fmt.Errorf("node %q: %w", nodeID, repository.ErrNotFound)
	}
	needed := map[string]bool{nodeID: true}
	var visit func(string)
	visit = func(target string) {
		for from, destinations := range workflow.Edges {
			for _, to := range destinations {
				if to == target && !needed[from] {
					needed[from] = true
					visit(from)
				}
			}
		}
	}
	visit(nodeID)
	partial := wf.NewWorkflow(workflow.Id, workflow.Name)
	for id := range needed {
		partial.Nodes[id] = workflow.Nodes[id]
	}
	for from, destinations := range workflow.Edges {
		if !needed[from] {
			continue
		}
		for _, to := range destinations {
			if needed[to] {
				partial.Edges[from] = append(partial.Edges[from], to)
			}
		}
	}
	return s.executor.RunDetailed(partial, params)
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
