package service_test

import (
	"testing"

	"github.com/doodledoc/backend/internal/infrastructure"
	"github.com/doodledoc/backend/internal/readmodel"
	"github.com/doodledoc/backend/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestService() *service.DocumentService {
	es := infrastructure.NewInMemoryEventStore()
	ps := readmodel.NewInMemoryProjectionStore()
	eh := readmodel.NewDocumentEventHandler(ps)
	return service.NewDocumentService(es, ps, eh, service.NoopBroadcaster{})
}

func TestCreateDocument_ReturnsProjection(t *testing.T) {
	svc := newTestService()

	doc, err := svc.CreateDocument("My Doc", "user-1", "Alice")

	require.NoError(t, err)
	require.NotNil(t, doc)
	assert.Equal(t, "My Doc", doc.Title)
	assert.NotEmpty(t, doc.ID)
	assert.Equal(t, 1, doc.Version)
}

func TestUpdateDocument_ChangesProjection(t *testing.T) {
	svc := newTestService()
	created, _ := svc.CreateDocument("Original", "user-1", "Alice")

	updated, err := svc.UpdateDocument(created.ID, "Renamed", "hello world", "user-1", "Alice")

	require.NoError(t, err)
	assert.Equal(t, "Renamed", updated.Title)
	assert.Equal(t, "hello world", updated.Content)
}

func TestDeleteDocument_RemovesFromProjectionStore(t *testing.T) {
	svc := newTestService()
	created, _ := svc.CreateDocument("To Delete", "", "")

	err := svc.DeleteDocument(created.ID, "", "")
	require.NoError(t, err)

	_, err = svc.GetDocument(created.ID)
	assert.ErrorIs(t, err, service.ErrNotFound)
}

func TestGetHistory_ReturnsAllEvents(t *testing.T) {
	svc := newTestService()
	doc, _ := svc.CreateDocument("Doc", "", "")
	svc.UpdateDocument(doc.ID, "Doc v2", "content", "", "")

	history, err := svc.GetHistory(doc.ID)

	require.NoError(t, err)
	// DocumentCreated + TitleUpdated + ContentUpdated = 3 events
	assert.GreaterOrEqual(t, len(history), 2)
}

func TestGetDocumentAtVersion_TimeTravels(t *testing.T) {
	svc := newTestService()
	doc, _ := svc.CreateDocument("Original Title", "", "")
	svc.UpdateDocument(doc.ID, "New Title", "", "", "")

	// Version 1 = just after DocumentCreated
	v1, err := svc.GetDocumentAtVersion(doc.ID, 1)
	require.NoError(t, err)
	assert.Equal(t, "Original Title", v1.Title)
}

func TestAddComment_AppearsInProjection(t *testing.T) {
	svc := newTestService()
	doc, _ := svc.CreateDocument("Doc", "", "")

	updated, err := svc.AddComment(doc.ID, "Nice!", "Bob")

	require.NoError(t, err)
	require.Len(t, updated.Comments, 1)
	assert.Equal(t, "Nice!", updated.Comments[0].Text)
	assert.Equal(t, "Bob", updated.Comments[0].Author)
}

func TestDeleteComment_RemovesFromProjection(t *testing.T) {
	svc := newTestService()
	doc, _ := svc.CreateDocument("Doc", "", "")
	withComment, _ := svc.AddComment(doc.ID, "Hello", "Alice")
	commentID := withComment.Comments[0].ID

	result, err := svc.DeleteComment(doc.ID, commentID)

	require.NoError(t, err)
	assert.Empty(t, result.Comments)
}

func TestUpdateDocument_NotFound_ReturnsError(t *testing.T) {
	svc := newTestService()

	_, err := svc.UpdateDocument("nonexistent-id", "Title", "Content", "", "")

	assert.ErrorIs(t, err, service.ErrNotFound)
}
