package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/doodledoc/backend/internal/service"
)

// CommentHandler handles /api/document/{id}/comments endpoints.
type CommentHandler struct {
	svc *service.DocumentService
}

func NewCommentHandler(svc *service.DocumentService) *CommentHandler {
	return &CommentHandler{svc: svc}
}

type addCommentRequest struct {
	Text   string `json:"text"`
	Author string `json:"author"`
}

// GET /api/document/{id}/comments
//
// @Summary      List comments
// @Description  Returns all active comments on a document.
// @Tags         comments
// @Produce      json
// @Param        id  path      string  true  "Document ID"
// @Success      200  {array}   readmodel.Comment
// @Failure      404  {object}  map[string]string
// @Router       /api/document/{id}/comments [get]
func (h *CommentHandler) List(w http.ResponseWriter, r *http.Request) {
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
	writeJSON(w, http.StatusOK, doc.Comments)
}

// POST /api/document/{id}/comments
//
// @Summary      Add a comment
// @Description  Adds a comment to a document.
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        id    path      string            true  "Document ID"
// @Param        body  body      addCommentRequest  true  "Comment body"
// @Success      200   {object}  readmodel.DocumentProjection
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Router       /api/document/{id}/comments [post]
func (h *CommentHandler) Add(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req addCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(req.Text) == "" {
		writeError(w, http.StatusBadRequest, "text is required")
		return
	}

	doc, err := h.svc.AddComment(id, req.Text, req.Author)
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

// DELETE /api/document/{id}/comments/{commentId}
//
// @Summary      Delete a comment
// @Description  Removes a comment from a document.
// @Tags         comments
// @Param        id         path  string  true  "Document ID"
// @Param        commentId  path  string  true  "Comment ID"
// @Success      200
// @Failure      404  {object}  map[string]string
// @Router       /api/document/{id}/comments/{commentId} [delete]
func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	commentID := r.PathValue("commentId")

	_, err := h.svc.DeleteComment(id, commentID)
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
