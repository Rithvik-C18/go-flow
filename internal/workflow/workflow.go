package workflow

import (
	"errors"
	"fmt"
)

type NodeDefinition struct {
	Id     string
	Type   string
	Config map[string]any
}

type Workflow struct {
	Id    string
	Name  string
	Nodes map[string]*NodeDefinition
	Edges map[string][]string
}

func NewWorkflow(id, name string) *Workflow {
	return &Workflow{
		Id:    id,
		Name:  name,
		Nodes: make(map[string]*NodeDefinition),
		Edges: make(map[string][]string),
	}
}

func NewNodeDefinition(id, nodeType string, config map[string]any) (*NodeDefinition, error) {
	if id == "" {
		return nil, fmt.Errorf("node id is required")
	}

	if nodeType == "" {
		return nil, fmt.Errorf("node type is required")
	}

	if config == nil {
		config = make(map[string]any)
	}

	return &NodeDefinition{
		Id:     id,
		Type:   nodeType,
		Config: config,
	}, nil
}

func (d *Workflow) AddNode(node *NodeDefinition) error {
	if node == nil {
		return errors.New("node is nil")
	}

	id := node.Id

	if id == "" {
		return errors.New("node Id is empty")
	}

	if _, ok := d.Nodes[id]; ok {
		return errors.New("node already exists")
	}

	d.Nodes[id] = node
	return nil
}

func (d *Workflow) AddEdge(from, to string) error {
	if from == "" {
		return errors.New("from node is empty")
	}
	if to == "" {
		return errors.New("to node is empty")
	}

	if from == to {
		return errors.New("self loop is not accepted")
	}

	if _, ok := d.Nodes[from]; !ok {
		return errors.New("from node does not exist")
	}

	if _, ok := d.Nodes[to]; !ok {
		return errors.New("to node does not exist")
	}

	for _, existing := range d.Edges[from] {
		if existing == to {
			return errors.New("edge already exists")
		}
	}

	if d.hasPath(to, from) {
		return errors.New("adding this edge creates a cycle")
	}

	d.Edges[from] = append(d.Edges[from], to)

	return nil
}

func (d *Workflow) DeleteEdge(from, to string) error {
	if from == "" {
		return errors.New("from node is empty")
	}
	if to == "" {
		return errors.New("to node is empty")
	}

	if _, ok := d.Nodes[from]; !ok {
		return errors.New("from node does not exist")
	}

	if _, ok := d.Nodes[to]; !ok {
		return errors.New("to node does not exist")
	}

	edges := d.Edges[from]
	for i, existing := range edges {
		if existing == to {
			d.Edges[from] = append(edges[:i], edges[i+1:]...)
			if len(d.Edges[from]) == 0 {
				delete(d.Edges, from)
			}
			return nil
		}
	}

	return errors.New("edge does not exist")
}

func (d *Workflow) DeleteNode(id string) error {
	if id == "" {
		return errors.New("node id is empty")
	}

	if _, ok := d.Nodes[id]; !ok {
		return errors.New("node does not exist")
	}

	delete(d.Nodes, id)
	delete(d.Edges, id)

	for from, neighbors := range d.Edges {
		var remaining []string
		for _, to := range neighbors {
			if to != id {
				remaining = append(remaining, to)
			}
		}
		d.Edges[from] = remaining
	}

	return nil
}
