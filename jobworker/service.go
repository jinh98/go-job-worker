package jobworker

import (
	"sync"
)

// Service is an in memory storage interface for workers.
type Service struct {
	workersMap map[string]*Worker
	mu         sync.RWMutex
}

// NewService makes a new empty map of workers, the worker id is the key.
func NewService() (*Service, error) {

	return &Service{
		workersMap: make(map[string]*Worker),
	}, nil
}

// AddWorker adds a worker to the map.
func (s *Service) AddWorker(w *Worker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workersMap[w.ID] = w
}

// GetWorker returns a worker.
func (s *Service) GetWorker(id string) (w *Worker) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w = s.workersMap[id]
	return w
}
