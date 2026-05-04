package router

import (
	"net/http"

	"github.com/doodledoc/backend/internal/handler"
	"github.com/doodledoc/backend/internal/hub"
	"github.com/doodledoc/backend/internal/infrastructure"
	"github.com/doodledoc/backend/internal/readmodel"
	"github.com/doodledoc/backend/internal/service"
	httpSwagger "github.com/swaggo/http-swagger"
)

// New returns a configured HTTP mux.
func New() http.Handler {
	mux := http.NewServeMux()

	// ── DoodleDoc CQRS wiring ─────────────────────────────────────────────────
	eventStore := infrastructure.NewInMemoryEventStore()
	projStore := readmodel.NewInMemoryProjectionStore()
	eventHandler := readmodel.NewDocumentEventHandler(projStore)
	wsHub := hub.New()
	docService := service.NewDocumentService(
		eventStore,
		projStore,
		eventHandler,
		wsHub,
	)

	docHandler := handler.NewDocumentHandler(docService)
	commentHandler := handler.NewCommentHandler(docService)

	// documents
	mux.HandleFunc("GET /api/document", docHandler.GetAll)
	mux.HandleFunc("GET /api/document/{id}", docHandler.GetByID)
	mux.HandleFunc("GET /api/document/{id}/history", docHandler.GetHistory)
	mux.HandleFunc("GET /api/document/{id}/version/{version}", docHandler.GetAtVersion)
	mux.HandleFunc("POST /api/document", docHandler.Create)
	mux.HandleFunc("PUT /api/document/{id}", docHandler.Update)
	mux.HandleFunc("DELETE /api/document/{id}", docHandler.Delete)

	// comments
	mux.HandleFunc("GET /api/document/{id}/comments", commentHandler.List)
	mux.HandleFunc("POST /api/document/{id}/comments", commentHandler.Add)
	mux.HandleFunc("DELETE /api/document/{id}/comments/{commentId}", commentHandler.Delete)

	// ── infra ─────────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("GET /hubs/document", wsHub.ServeWS)
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	return mux
}
