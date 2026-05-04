package domain

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent is the interface all domain events must satisfy.
type DomainEvent interface {
	GetEventID() string
	GetDocumentID() string
	GetOccurredAt() time.Time
	GetVersion() int
	SetVersion(v int)
}

// BaseEvent holds the common fields shared by every domain event.
type BaseEvent struct {
	EventID    string
	DocumentID string
	OccurredAt time.Time
	Version    int // sequence number within the document timeline
}

func newBase(documentID string) BaseEvent {
	return BaseEvent{
		EventID:    uuid.NewString(),
		DocumentID: documentID,
		OccurredAt: time.Now().UTC(),
	}
}

func (b *BaseEvent) GetEventID() string       { return b.EventID }
func (b *BaseEvent) GetDocumentID() string    { return b.DocumentID }
func (b *BaseEvent) GetOccurredAt() time.Time { return b.OccurredAt }
func (b *BaseEvent) GetVersion() int          { return b.Version }
func (b *BaseEvent) SetVersion(v int)         { b.Version = v }
