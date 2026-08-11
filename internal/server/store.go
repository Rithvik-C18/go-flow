package server

import (
	"errors"
	"sync"

	wf "github.com/Rithvik-C18/go-flow/internal/workflow"
)

type Store struct {
	mu        sync.Mutex
	workflows map[string]*wf.Workflow
}

func NewStore() *Store {
	return &Store{
		workflows: make(map[string]*wf.Workflow),
	}
}

func (s *Store) Create(id, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.workflows[id]; ok {
		return errors.New("workflow already exists")
	}

	s.workflows[id] = wf.NewWorkflow(id, name)
	return nil
}

func (s *Store) Get(id string) (*wf.Workflow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	w, ok := s.workflows[id]
	if !ok {
		return nil, errors.New("workflow not found")
	}

	return w, nil
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.workflows[id]; !ok {
		return errors.New("workflow not found")
	}

	delete(s.workflows, id)
	return nil
}
