package repository

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Rithvik-C18/go-flow/internal/database"
	wf "github.com/Rithvik-C18/go-flow/internal/workflow"
	"gorm.io/gorm"
)

type WorkflowRepository struct {
	db     *gorm.DB
	userID uint
}

func NewWorkflowRepository(db *gorm.DB) *WorkflowRepository {
	return &WorkflowRepository{db: db}
}

func (r *WorkflowRepository) ForUser(userID uint) *WorkflowRepository {
	return &WorkflowRepository{db: r.db, userID: userID}
}

func (r *WorkflowRepository) Create(id, name string) error {
	record := &database.Workflow{ID: id, Name: name, UserID: r.userID}
	if err := r.db.Create(record).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("workflow %q: %w", id, ErrConflict)
		}
		return err
	}
	return nil
}

func (r *WorkflowRepository) List() ([]database.Workflow, error) {
	var records []database.Workflow
	if err := r.db.Where("user_id = ?", r.userID).Order("created_at").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *WorkflowRepository) Load(id string) (*wf.Workflow, error) {
	var record database.Workflow
	err := r.db.Preload("Nodes").Preload("Edges").First(&record, "id = ? AND user_id = ?", id, r.userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("workflow %q: %w", id, ErrNotFound)
	}
	if err != nil {
		return nil, err
	}

	workflow := wf.NewWorkflow(record.ID, record.Name)

	for _, node := range record.Nodes {
		config := map[string]any{}
		if node.Config != "" {
			if err := json.Unmarshal([]byte(node.Config), &config); err != nil {
				return nil, fmt.Errorf("workflow %q: node %q has invalid config: %w", id, node.ID, err)
			}
		}
		workflow.Nodes[node.ID] = &wf.NodeDefinition{
			Id:     node.ID,
			Type:   node.Type,
			Config: config,
		}
	}

	for _, edge := range record.Edges {
		workflow.Edges[edge.SourceNodeID] = append(workflow.Edges[edge.SourceNodeID], edge.TargetNodeID)
	}

	return workflow, nil
}

func (r *WorkflowRepository) Delete(id string) error {
	if _, err := r.Load(id); err != nil {
		return err
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("workflow_id = ?", id).Delete(&database.Node{}).Error; err != nil {
			return err
		}
		if err := tx.Where("workflow_id = ?", id).Delete(&database.Edge{}).Error; err != nil {
			return err
		}
		return tx.Delete(&database.Workflow{}, "id = ? AND user_id = ?", id, r.userID).Error
	})
}

func (r *WorkflowRepository) AddNode(workflowID, id, nodeType, config string) error {
	record := &database.Node{
		WorkflowID: workflowID,
		ID:         id,
		Type:       nodeType,
		Config:     config,
	}
	if err := r.db.Create(record).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("node %q: %w", id, ErrConflict)
		}
		return err
	}
	return nil
}

func (r *WorkflowRepository) DeleteNode(workflowID, nodeID string) error {
	tx := r.db.Where(&database.Node{WorkflowID: workflowID, ID: nodeID}).Delete(&database.Node{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return fmt.Errorf("node %q: %w", nodeID, ErrNotFound)
	}

	r.db.Where(
		"workflow_id = ? AND (source_node_id = ? OR target_node_id = ?)",
		workflowID, nodeID, nodeID,
	).Delete(&database.Edge{})

	return nil
}

func (r *WorkflowRepository) AddEdge(workflowID, from, to string) error {
	record := &database.Edge{
		WorkflowID:   workflowID,
		SourceNodeID: from,
		TargetNodeID: to,
	}
	if err := r.db.Create(record).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("edge %s->%s: %w", from, to, ErrConflict)
		}
		return err
	}
	return nil
}

func (r *WorkflowRepository) DeleteEdge(workflowID, from, to string) error {
	tx := r.db.Where(&database.Edge{
		WorkflowID:   workflowID,
		SourceNodeID: from,
		TargetNodeID: to,
	}).Delete(&database.Edge{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return fmt.Errorf("edge %s->%s: %w", from, to, ErrNotFound)
	}
	return nil
}
