package domain_test

import (
	"testing"

	"github.com/doodledoc/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate_RaisesDocumentCreatedEvent(t *testing.T) {
	agg := domain.Create("doc-1", "Test Document", "", "")

	changes := agg.GetUncommittedChanges()
	require.Len(t, changes, 1)

	e, ok := changes[0].(*domain.DocumentCreated)
	require.True(t, ok, "expected *DocumentCreated")
	assert.Equal(t, "doc-1", e.DocumentID)
	assert.Equal(t, "Test Document", e.Title)
}

func TestUpdateTitle_RaisesTitleUpdatedEvent(t *testing.T) {
	agg := domain.Create("doc-2", "Original Title", "", "")
	agg.MarkChangesAsCommitted()

	agg.UpdateTitle("New Title", "", "")

	changes := agg.GetUncommittedChanges()
	require.Len(t, changes, 1)

	e, ok := changes[0].(*domain.TitleUpdated)
	require.True(t, ok, "expected *TitleUpdated")
	assert.Equal(t, "New Title", e.NewTitle)
}

func TestUpdateContent_RaisesContentUpdatedEvent(t *testing.T) {
	agg := domain.Create("doc-3", "Test", "", "")
	agg.MarkChangesAsCommitted()

	agg.UpdateContent("New content here", "", "", "text")

	changes := agg.GetUncommittedChanges()
	require.Len(t, changes, 1)

	e, ok := changes[0].(*domain.ContentUpdated)
	require.True(t, ok, "expected *ContentUpdated")
	assert.Equal(t, "New content here", e.Content)
}

func TestDelete_RaisesDocumentDeletedEvent(t *testing.T) {
	agg := domain.Create("doc-4", "To Delete", "", "")
	agg.MarkChangesAsCommitted()

	agg.Delete("", "")

	changes := agg.GetUncommittedChanges()
	require.Len(t, changes, 1)

	_, ok := changes[0].(*domain.DocumentDeleted)
	require.True(t, ok, "expected *DocumentDeleted")
}

func TestMarkChangesAsCommitted_ClearsUncommittedEvents(t *testing.T) {
	agg := domain.Create("doc-5", "Test", "", "")
	require.Len(t, agg.GetUncommittedChanges(), 1)

	agg.MarkChangesAsCommitted()

	assert.Empty(t, agg.GetUncommittedChanges())
}

func TestFromEvents_ReconstructsState(t *testing.T) {
	// Build a history: create → update title → update content
	original := domain.Create("doc-6", "Initial Title", "", "")
	original.UpdateTitle("Updated Title", "", "")
	original.UpdateContent("Some content", "", "", "text")

	events := original.GetUncommittedChanges()

	// Reconstruct from those events
	rebuilt := domain.FromEvents(events)

	assert.Equal(t, "doc-6", rebuilt.ID)
	assert.Equal(t, "Updated Title", rebuilt.Title)
	assert.Equal(t, "Some content", rebuilt.Content)
	assert.Equal(t, 3, rebuilt.Version)
}

func TestFromEvents_DeletedState(t *testing.T) {
	original := domain.Create("doc-7", "Will Be Deleted", "", "")
	original.Delete("", "")

	rebuilt := domain.FromEvents(original.GetUncommittedChanges())

	assert.True(t, rebuilt.IsDeleted)
}

func TestAddComment_RaisesCommentAddedEvent(t *testing.T) {
	agg := domain.Create("doc-8", "Doc with Comments", "", "")
	agg.MarkChangesAsCommitted()

	agg.AddComment("comment-1", "Great doc!", "Rishabh")

	changes := agg.GetUncommittedChanges()
	require.Len(t, changes, 1)

	e, ok := changes[0].(*domain.CommentAdded)
	require.True(t, ok, "expected *CommentAdded")
	assert.Equal(t, "comment-1", e.CommentID)
	assert.Equal(t, "Great doc!", e.Text)
	assert.Equal(t, "Rishabh", e.Author)
}

// TestDeleteComment_RaisesCommentDeletedEvent
func TestDeleteComment_RaisesCommentDeletedEvent(t *testing.T) {
	agg := domain.Create("doc-9", "Doc", "", "")
	agg.AddComment("comment-1", "Hello", "Alice")
	agg.MarkChangesAsCommitted()

	agg.DeleteComment("comment-1")

	changes := agg.GetUncommittedChanges()
	require.Len(t, changes, 1)

	e, ok := changes[0].(*domain.CommentDeleted)
	require.True(t, ok, "expected *CommentDeleted")
	assert.Equal(t, "comment-1", e.CommentID)
}
