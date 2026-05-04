package domain

import "time"

// DocumentAggregate is the aggregate root for a document.
// State is reconstructed entirely by replaying domain events.
type DocumentAggregate struct {
	ID        string
	Title     string
	Content   string
	IsDeleted bool
	CreatedAt time.Time
	UpdatedAt time.Time
	Version   int // number of events applied so far

	changes []DomainEvent // uncommitted events
}

// Create is the factory that raises a DocumentCreated event and returns
// a fully-initialised aggregate (equivalent to C# static Create()).
func Create(documentID, title, userID, userName string) *DocumentAggregate {
	agg := &DocumentAggregate{}
	event := NewDocumentCreated(documentID, title, userID, userName)
	agg.apply(event)
	agg.changes = append(agg.changes, event)
	return agg
}

// UpdateContent raises a ContentUpdated event.
func (a *DocumentAggregate) UpdateContent(content, userID, userName, contentType string) {
	event := NewContentUpdated(a.ID, content, contentType, userID, userName)
	a.apply(event)
	a.changes = append(a.changes, event)
}

// UpdateTitle raises a TitleUpdated event.
func (a *DocumentAggregate) UpdateTitle(newTitle, userID, userName string) {
	event := NewTitleUpdated(a.ID, newTitle, userID, userName)
	a.apply(event)
	a.changes = append(a.changes, event)
}

// Delete raises a DocumentDeleted event.
func (a *DocumentAggregate) Delete(userID, userName string) {
	event := NewDocumentDeleted(a.ID, userID, userName)
	a.apply(event)
	a.changes = append(a.changes, event)
}

// AddComment raises a CommentAdded event.
func (a *DocumentAggregate) AddComment(commentID, text, author string) {
	event := NewCommentAdded(a.ID, commentID, text, author)
	a.apply(event)
	a.changes = append(a.changes, event)
}

// DeleteComment raises a CommentDeleted event.
func (a *DocumentAggregate) DeleteComment(commentID string) {
	event := NewCommentDeleted(a.ID, commentID)
	a.apply(event)
	a.changes = append(a.changes, event)
}

// FromEvents reconstructs an aggregate from a persisted event stream.
// Equivalent to the C# static FromEvents().
func FromEvents(events []DomainEvent) *DocumentAggregate {
	agg := &DocumentAggregate{}
	for _, e := range events {
		agg.apply(e)
		agg.Version++
	}
	return agg
}

// GetUncommittedChanges returns events raised since the last save.
func (a *DocumentAggregate) GetUncommittedChanges() []DomainEvent {
	out := make([]DomainEvent, len(a.changes))
	copy(out, a.changes)
	return out
}

// MarkChangesAsCommitted clears the uncommitted event list after a save.
func (a *DocumentAggregate) MarkChangesAsCommitted() {
	a.changes = nil
}

// apply updates aggregate state based on the concrete event type.
// This is the only place where business state mutation happens.
func (a *DocumentAggregate) apply(event DomainEvent) {
	switch e := event.(type) {
	case *DocumentCreated:
		a.ID = e.DocumentID
		a.Title = e.Title
		a.Content = ""
		a.CreatedAt = e.OccurredAt
		a.UpdatedAt = e.OccurredAt

	case *ContentUpdated:
		a.Content = e.Content
		a.UpdatedAt = e.OccurredAt

	case *TitleUpdated:
		a.Title = e.NewTitle
		a.UpdatedAt = e.OccurredAt

	case *DocumentDeleted:
		a.IsDeleted = true
		a.UpdatedAt = e.OccurredAt

	case *CommentAdded, *CommentDeleted:
		// Comments are tracked in the projection, not aggregate state.
		a.UpdatedAt = event.GetOccurredAt()
	}
}
