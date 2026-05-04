package readmodel

import (
	"fmt"

	"github.com/doodledoc/backend/internal/domain"
)

// EventHandler updates the read model whenever an event is saved.
// This is how the CQRS write side keeps the read side in sync.
type EventHandler interface {
	Handle(event domain.DomainEvent) error
}

// DocumentEventHandler listens to domain events and projects them into
// DocumentProjection records inside the ProjectionStore.
type DocumentEventHandler struct {
	store ProjectionStore
}

func NewDocumentEventHandler(store ProjectionStore) *DocumentEventHandler {
	return &DocumentEventHandler{store: store}
}

func (h *DocumentEventHandler) Handle(event domain.DomainEvent) error {
	switch e := event.(type) {
	case *domain.DocumentCreated:
		return h.store.Save(&DocumentProjection{
			ID:        e.DocumentID,
			Title:     e.Title,
			Content:   "",
			CreatedAt: e.OccurredAt,
			UpdatedAt: e.OccurredAt,
			Version:   e.GetVersion(),
			Comments:  []Comment{},
		})

	case *domain.ContentUpdated:
		p, err := h.store.Get(e.DocumentID)
		if err != nil || p == nil {
			return err
		}
		p.Content = e.Content
		p.UpdatedAt = e.OccurredAt
		p.Version = e.GetVersion()
		return h.store.Save(p)

	case *domain.TitleUpdated:
		p, err := h.store.Get(e.DocumentID)
		if err != nil || p == nil {
			return err
		}
		p.Title = e.NewTitle
		p.UpdatedAt = e.OccurredAt
		p.Version = e.GetVersion()
		return h.store.Save(p)

	case *domain.DocumentDeleted:
		return h.store.Delete(e.DocumentID)

	case *domain.CommentAdded:
		p, err := h.store.Get(e.DocumentID)
		if err != nil || p == nil {
			return err
		}
		p.Comments = append(p.Comments, Comment{
			ID:        e.CommentID,
			Text:      e.Text,
			Author:    e.Author,
			Timestamp: e.Timestamp,
		})
		p.UpdatedAt = e.OccurredAt
		return h.store.Save(p)

	case *domain.CommentDeleted:
		p, err := h.store.Get(e.DocumentID)
		if err != nil || p == nil {
			return err
		}
		filtered := p.Comments[:0]
		for _, c := range p.Comments {
			if c.ID != e.CommentID {
				filtered = append(filtered, c)
			}
		}
		p.Comments = filtered
		p.UpdatedAt = e.OccurredAt
		return h.store.Save(p)

	default:
		return fmt.Errorf("readmodel: unhandled event type %T", event)
	}
}
