package database

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username     string `gorm:"uniqueIndex;not null"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
}

type Workflow struct {
	ID        string `gorm:"primaryKey;size:255"`
	Name      string `gorm:"not null"`
	Nodes     []Node `gorm:"constraint:OnDelete:CASCADE"`
	Edges     []Edge `gorm:"constraint:OnDelete:CASCADE"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Node struct {
	WorkflowID string `gorm:"primaryKey;size:255"`
	ID         string `gorm:"primaryKey;size:255"`
	Name       string
	Type       string
	Config     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Edge struct {
	ID           uint   `gorm:"primaryKey"`
	WorkflowID   string `gorm:"uniqueIndex:idx_workflow_edge;size:255"`
	SourceNodeID string `gorm:"uniqueIndex:idx_workflow_edge;size:255"`
	TargetNodeID string `gorm:"uniqueIndex:idx_workflow_edge;size:255"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
