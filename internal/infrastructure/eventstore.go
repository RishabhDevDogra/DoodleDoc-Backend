package infrastructure

import (
	"sync"

	"github.com/doodledoc/backend/internal/domain"
)

// EventStore is the write-side storage interface.
// Every action a user takes produces events; this is where they live.
// Swap this out for Postgres/SQL Server later without touching anything else.
type EventStore interface {
	SaveEvents(documentID string, events []domain.DomainEvent) error
	GetEvents(documentID string) ([]domain.DomainEvent, error)
	GetAllDocumentIDs() ([]string, error)
}

// InMemoryEventStore stores events in RAM — lost on restart, perfect for now.
// Protected by a RWMutex so concurrent WebSocket writes don't race.
type InMemoryEventStore struct {
	mu     sync.RWMutex
	events map[string][]domain.DomainEvent // documentID → ordered event list
}

func NewInMemoryEventStore() *InMemoryEventStore {
	return &InMemoryEventStore{
		events: make(map[string][]domain.DomainEvent),
	}
}

// SaveEvents appends events for a document and stamps each with its version number.
func (s *InMemoryEventStore) SaveEvents(documentID string, events []domain.DomainEvent) error {
	if len(events) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	current := len(s.events[documentID])
	for _, e := range events {
		current++
		e.SetVersion(current)
		s.events[documentID] = append(s.events[documentID], e)
	}
	return nil
}

// GetEvents returns the full ordered event history for a document.
func (s *InMemoryEventStore) GetEvents(documentID string) ([]domain.DomainEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	src := s.events[documentID]
	out := make([]domain.DomainEvent, len(src))
	copy(out, src)
	return out, nil
}

// GetAllDocumentIDs returns every document ID that has at least one event.
func (s *InMemoryEventStore) GetAllDocumentIDs() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.events))
	for id := range s.events {
		ids = append(ids, id)
	}
	return ids, nil
}
