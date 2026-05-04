package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/doodledoc/backend/internal/service"
)

// DocumentHandler handles all /api/document/* endpoints.
type DocumentHandler struct {
	svc *service.DocumentService
}

func NewDocumentHandler(svc *service.DocumentService) *DocumentHandler {
	return &DocumentHandler{svc: svc}
}

// request bodies
type createDocumentRequest struct {
	Title    string `json:"title"`
	UserID   string `json:"userId"`
	UserName string `json:"userName"`
}

type updateDocumentRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	UserID   string `json:"userId"`
	UserName string `json:"userName"`
}

// GET /api/document
//
// @Summary      List all documents
// @Description  Returns all document projections, newest first.
// @Tags         documents
// @Produce      json
// @Success      200  {array}   readmodel.DocumentProjection
// @Router       /api/document [get]
func (h *DocumentHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	docs, err := h.svc.GetAllDocuments()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, docs)
}

// GET /api/document/{id}
//
// @Summary      Get a document
// @Description  Returns the current projection for a document.
// @Tags         documents
// @Produce      json
// @Param        id   path      string  true  "Document ID"
// @Success      200  {object}  readmodel.DocumentProjection
// @Failure      404  {object}  map[string]string
// @Router       /api/document/{id} [get]
func (h *DocumentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	doc, err := h.svc.GetDocument(id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeError(w, http.StatusNotFound, "document not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

// GET /api/document/{id}/history
//
// @Summary      Get event history
// @Description  Returns the full ordered event log for a document.
// @Tags         documents
// @Produce      json
// @Param        id   path      string  true  "Document ID"
// @Success      200  {array}   service.HistoryEntry
// @Failure      404  {object}  map[string]string
// @Router       /api/document/{id}/history [get]
func (h *DocumentHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	entries, err := h.svc.GetHistory(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

// GET /api/document/{id}/version/{n}
//
// @Summary      Get document at version
// @Description  Replays events up to version N — enables time travel / undo.
// @Tags         documents
// @Produce      json
// @Param        id       path      string  true  "Document ID"
// @Param        version  path      int     true  "Version number"
// @Success      200  {object}  readmodel.DocumentProjection
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/document/{id}/version/{version} [get]
func (h *DocumentHandler) GetAtVersion(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	versionStr := r.PathValue("version")
	version, err := strconv.Atoi(versionStr)
	if err != nil || version < 1 {
		writeError(w, http.StatusBadRequest, "version must be a positive integer")
		return
	}
	doc, err := h.svc.GetDocumentAtVersion(id, version)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeError(w, http.StatusNotFound, "document or version not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

// POST /api/document
//
// @Summary      Create a document
// @Description  Creates a new document and returns its projection.
// @Tags         documents
// @Accept       json
// @Produce      json
// @Param        body  body      createDocumentRequest  true  "Create request"
// @Success      201   {object}  readmodel.DocumentProjection
// @Failure      400   {object}  map[string]string
// @Router       /api/document [post]
func (h *DocumentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	doc, err := h.svc.CreateDocument(req.Title, req.UserID, req.UserName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, doc)
}

// PUT /api/document/{id}
//
// @Summary      Update a document
// @Description  Updates title and/or content of an existing document.
// @Tags         documents
// @Accept       json
// @Produce      json
// @Param        id    path      string                 true  "Document ID"
// @Param        body  body      updateDocumentRequest  true  "Update request"
// @Success      200   {object}  readmodel.DocumentProjection
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Router       /api/document/{id} [put]
func (h *DocumentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if strings.TrimSpace(id) == "" {
		writeError(w, http.StatusBadRequest, "document ID is required")
		return
	}
	var req updateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	doc, err := h.svc.UpdateDocument(id, req.Title, req.Content, req.UserID, req.UserName)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeError(w, http.StatusNotFound, "document not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

// DELETE /api/document/{id}
//
// @Summary      Delete a document
// @Description  Deletes a document by ID.
// @Tags         documents
// @Param        id        path   string  true   "Document ID"
// @Param        userId    query  string  false  "User ID"
// @Param        userName  query  string  false  "User name"
// @Success      200
// @Failure      404  {object}  map[string]string
// @Router       /api/document/{id} [delete]
func (h *DocumentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	userID := r.URL.Query().Get("userId")
	userName := r.URL.Query().Get("userName")

	err := h.svc.DeleteDocument(id, userID, userName)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeError(w, http.StatusNotFound, "document not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
