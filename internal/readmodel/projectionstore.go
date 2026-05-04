package readmodel

import (
	"sort"
	"sync"
)

// ProjectionStore is the read-side storage interface.
// Completely separate from the event store — optimised for fast reads.
type ProjectionStore interface {
	Save(p *DocumentProjection) error
	Get(documentID string) (*DocumentProjection, error)
	GetAll() ([]*DocumentProjection, error)
	Delete(documentID string) error
}

// InMemoryProjectionStore keeps projections in RAM, sorted by UpdatedAt on reads.
type InMemoryProjectionStore struct {
	mu          sync.RWMutex
	projections map[string]*DocumentProjection
}

func NewInMemoryProjectionStore() *InMemoryProjectionStore {
	return &InMemoryProjectionStore{
		projections: make(map[string]*DocumentProjection),
	}
}

func (s *InMemoryProjectionStore) Save(p *DocumentProjection) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.projections[p.ID] = p
	return nil
}

func (s *InMemoryProjectionStore) Get(documentID string) (*DocumentProjection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.projections[documentID]
	if !ok {
		return nil, nil
	}
	return p, nil
}

// GetAll returns all projections ordered newest-first (mirrors C# OrderByDescending).
func (s *InMemoryProjectionStore) GetAll() ([]*DocumentProjection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*DocumentProjection, 0, len(s.projections))
	for _, p := range s.projections {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out, nil
}

func (s *InMemoryProjectionStore) Delete(documentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.projections, documentID)
	return nil
}
