package service

import (
	"errors"
	"time"

	"github.com/doodledoc/backend/internal/domain"
	"github.com/doodledoc/backend/internal/infrastructure"
	"github.com/doodledoc/backend/internal/readmodel"
	"github.com/google/uuid"
)

// ErrNotFound is returned when a document does not exist in the event store.
var ErrNotFound = errors.New("document not found")

// Broadcaster is satisfied by the WebSocket hub (Phase 5).
// A no-op implementation is used until the hub is wired up.
type Broadcaster interface {
	BroadcastDocumentCreated(id, title string)
	BroadcastDocumentUpdated(id string)
	BroadcastDocumentDeleted(id string)
	BroadcastEventAdded(docID, eventType, description string, ts time.Time)
	BroadcastCommentAdded(docID string)
}

// NoopBroadcaster satisfies Broadcaster without doing anything.
// Replaced by the real hub in Phase 5.
type NoopBroadcaster struct{}

func (NoopBroadcaster) BroadcastDocumentCreated(_, _ string)            {}
func (NoopBroadcaster) BroadcastDocumentUpdated(_ string)               {}
func (NoopBroadcaster) BroadcastDocumentDeleted(_ string)               {}
func (NoopBroadcaster) BroadcastEventAdded(_, _, _ string, _ time.Time) {}
func (NoopBroadcaster) BroadcastCommentAdded(_ string)                  {}

// HistoryEntry is one row in the event timeline shown on the frontend.
type HistoryEntry struct {
	EventID     string    `json:"eventId"`
	DocumentID  string    `json:"documentId"`
	EventType   string    `json:"eventType"`
	Description string    `json:"description"`
	UserID      string    `json:"userId"`
	UserName    string    `json:"userName"`
	OccurredAt  time.Time `json:"occurredAt"`
	Version     int       `json:"version"`
}

// DocumentService orchestrates CQRS commands and queries.
// Commands mutate aggregates → events → event store → projection store.
// Queries read from the projection store (fast) or replay events (version travel).
type DocumentService struct {
	eventStore      infrastructure.EventStore
	projectionStore readmodel.ProjectionStore
	eventHandler    readmodel.EventHandler
	broadcaster     Broadcaster
}

func NewDocumentService(
	es infrastructure.EventStore,
	ps readmodel.ProjectionStore,
	eh readmodel.EventHandler,
	b Broadcaster,
) *DocumentService {
	return &DocumentService{
		eventStore:      es,
		projectionStore: ps,
		eventHandler:    eh,
		broadcaster:     b,
	}
}

// ── Commands ─────────────────────────────────────────────────────────────────

// CreateDocument creates a new document and returns its projection.
func (s *DocumentService) CreateDocument(title, userID, userName string) (*readmodel.DocumentProjection, error) {
	if title == "" {
		title = "Untitled Document"
	}
	documentID := uuid.NewString()
	agg := domain.Create(documentID, title, userID, userName)

	if err := s.persistAndProject(agg); err != nil {
		return nil, err
	}
	s.broadcaster.BroadcastDocumentCreated(documentID, title)
	return s.projectionStore.Get(documentID)
}

// UpdateDocument applies title and/or content changes and returns the updated projection.
func (s *DocumentService) UpdateDocument(id, title, content, userID, userName string) (*readmodel.DocumentProjection, error) {
	agg, err := s.loadAggregate(id)
	if err != nil {
		return nil, err
	}

	if agg.Title != title {
		agg.UpdateTitle(title, userID, userName)
	}
	if agg.Content != content {
		agg.UpdateContent(content, userID, userName, "text")
	}

	if err := s.persistAndProject(agg); err != nil {
		return nil, err
	}
	s.broadcaster.BroadcastDocumentUpdated(id)
	return s.projectionStore.Get(id)
}

// DeleteDocument removes the document from the projection store.
func (s *DocumentService) DeleteDocument(id, userID, userName string) error {
	agg, err := s.loadAggregate(id)
	if err != nil {
		return err
	}
	agg.Delete(userID, userName)

	if err := s.persistAndProject(agg); err != nil {
		return err
	}
	s.broadcaster.BroadcastDocumentDeleted(id)
	return nil
}

// AddComment attaches a comment to a document.
func (s *DocumentService) AddComment(docID, text, author string) (*readmodel.DocumentProjection, error) {
	agg, err := s.loadAggregate(docID)
	if err != nil {
		return nil, err
	}
	commentID := uuid.NewString()
	agg.AddComment(commentID, text, author)

	if err := s.persistAndProject(agg); err != nil {
		return nil, err
	}
	s.broadcaster.BroadcastCommentAdded(docID)
	return s.projectionStore.Get(docID)
}

// DeleteComment removes a comment from a document.
func (s *DocumentService) DeleteComment(docID, commentID string) (*readmodel.DocumentProjection, error) {
	agg, err := s.loadAggregate(docID)
	if err != nil {
		return nil, err
	}
	agg.DeleteComment(commentID)

	if err := s.persistAndProject(agg); err != nil {
		return nil, err
	}
	return s.projectionStore.Get(docID)
}

// ── Queries ───────────────────────────────────────────────────────────────────

// GetDocument returns the current projection for a document.
func (s *DocumentService) GetDocument(id string) (*readmodel.DocumentProjection, error) {
	p, err := s.projectionStore.Get(id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrNotFound
	}
	return p, nil
}

// GetAllDocuments returns all projections, newest first.
func (s *DocumentService) GetAllDocuments() ([]*readmodel.DocumentProjection, error) {
	return s.projectionStore.GetAll()
}

// GetHistory returns the full ordered event log for a document.
func (s *DocumentService) GetHistory(id string) ([]HistoryEntry, error) {
	events, err := s.eventStore.GetEvents(id)
	if err != nil {
		return nil, err
	}
	entries := make([]HistoryEntry, 0, len(events))
	for _, e := range events {
		entries = append(entries, HistoryEntry{
			EventID:     e.GetEventID(),
			DocumentID:  e.GetDocumentID(),
			EventType:   eventTypeName(e),
			Description: eventDescription(e),
			UserID:      eventUserID(e),
			UserName:    eventUserName(e),
			OccurredAt:  e.GetOccurredAt(),
			Version:     e.GetVersion(),
		})
	}
	return entries, nil
}

// GetDocumentAtVersion replays events up to version N — enables time travel.
func (s *DocumentService) GetDocumentAtVersion(id string, version int) (*readmodel.DocumentProjection, error) {
	events, err := s.eventStore.GetEvents(id)
	if err != nil {
		return nil, err
	}

	var slice []domain.DomainEvent
	for _, e := range events {
		if e.GetVersion() <= version {
			slice = append(slice, e)
		}
	}
	if len(slice) == 0 {
		return nil, ErrNotFound
	}

	agg := domain.FromEvents(slice)
	return &readmodel.DocumentProjection{
		ID:        agg.ID,
		Title:     agg.Title,
		Content:   agg.Content,
		CreatedAt: agg.CreatedAt,
		UpdatedAt: agg.UpdatedAt,
		Version:   version,
	}, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (s *DocumentService) loadAggregate(id string) (*domain.DocumentAggregate, error) {
	events, err := s.eventStore.GetEvents(id)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, ErrNotFound
	}
	return domain.FromEvents(events), nil
}

func (s *DocumentService) persistAndProject(agg *domain.DocumentAggregate) error {
	newEvents := agg.GetUncommittedChanges()
	if len(newEvents) == 0 {
		return nil
	}
	if err := s.eventStore.SaveEvents(agg.ID, newEvents); err != nil {
		return err
	}
	for _, e := range newEvents {
		if err := s.eventHandler.Handle(e); err != nil {
			return err
		}
		s.broadcaster.BroadcastEventAdded(
			e.GetDocumentID(),
			eventTypeName(e),
			eventDescription(e),
			e.GetOccurredAt(),
		)
	}
	agg.MarkChangesAsCommitted()
	return nil
}

func eventTypeName(e domain.DomainEvent) string {
	switch e.(type) {
	case *domain.DocumentCreated:
		return "DocumentCreated"
	case *domain.ContentUpdated:
		return "ContentUpdated"
	case *domain.TitleUpdated:
		return "TitleUpdated"
	case *domain.DocumentDeleted:
		return "DocumentDeleted"
	case *domain.CommentAdded:
		return "CommentAdded"
	case *domain.CommentDeleted:
		return "CommentDeleted"
	default:
		return "Unknown"
	}
}

func eventDescription(e domain.DomainEvent) string {
	switch ev := e.(type) {
	case *domain.DocumentCreated:
		return `Document created: "` + ev.Title + `"`
	case *domain.ContentUpdated:
		if len(ev.Content) > 0 {
			return "Content updated"
		}
		return "Content cleared"
	case *domain.TitleUpdated:
		return `Title changed to "` + ev.NewTitle + `"`
	case *domain.DocumentDeleted:
		return "Document deleted"
	case *domain.CommentAdded:
		return "Comment added by " + ev.Author
	case *domain.CommentDeleted:
		return "Comment deleted"
	default:
		return "Unknown event"
	}
}

func eventUserID(e domain.DomainEvent) string {
	switch ev := e.(type) {
	case *domain.DocumentCreated:
		return ev.UserID
	case *domain.ContentUpdated:
		return ev.UserID
	case *domain.TitleUpdated:
		return ev.UserID
	case *domain.DocumentDeleted:
		return ev.UserID
	default:
		return ""
	}
}

func eventUserName(e domain.DomainEvent) string {
	switch ev := e.(type) {
	case *domain.DocumentCreated:
		return ev.UserName
	case *domain.ContentUpdated:
		return ev.UserName
	case *domain.TitleUpdated:
		return ev.UserName
	case *domain.DocumentDeleted:
		return ev.UserName
	default:
		return ""
	}
}
