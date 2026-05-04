package domain

import "time"

// DocumentCreated is fired when a new document is created.
type DocumentCreated struct {
	BaseEvent
	Title    string
	UserID   string
	UserName string
}

func NewDocumentCreated(documentID, title, userID, userName string) *DocumentCreated {
	return &DocumentCreated{
		BaseEvent: newBase(documentID),
		Title:     title,
		UserID:    userID,
		UserName:  userName,
	}
}

// ContentUpdated is fired when document content (text or drawing) changes.
type ContentUpdated struct {
	BaseEvent
	Content     string
	ContentType string // "text" or "drawing"
	UserID      string
	UserName    string
}

func NewContentUpdated(documentID, content, contentType, userID, userName string) *ContentUpdated {
	if contentType == "" {
		contentType = "text"
	}
	return &ContentUpdated{
		BaseEvent:   newBase(documentID),
		Content:     content,
		ContentType: contentType,
		UserID:      userID,
		UserName:    userName,
	}
}

// TitleUpdated is fired when the document title changes.
type TitleUpdated struct {
	BaseEvent
	NewTitle string
	UserID   string
	UserName string
}

func NewTitleUpdated(documentID, newTitle, userID, userName string) *TitleUpdated {
	return &TitleUpdated{
		BaseEvent: newBase(documentID),
		NewTitle:  newTitle,
		UserID:    userID,
		UserName:  userName,
	}
}

// DocumentDeleted is fired when a document is deleted.
type DocumentDeleted struct {
	BaseEvent
	UserID   string
	UserName string
}

func NewDocumentDeleted(documentID, userID, userName string) *DocumentDeleted {
	return &DocumentDeleted{
		BaseEvent: newBase(documentID),
		UserID:    userID,
		UserName:  userName,
	}
}

// CommentAdded is fired when a comment is added to a document.
type CommentAdded struct {
	BaseEvent
	CommentID string
	Text      string
	Author    string
	Timestamp time.Time
}

func NewCommentAdded(documentID, commentID, text, author string) *CommentAdded {
	now := time.Now().UTC()
	return &CommentAdded{
		BaseEvent: newBase(documentID),
		CommentID: commentID,
		Text:      text,
		Author:    author,
		Timestamp: now,
	}
}

// CommentDeleted is fired when a comment is removed from a document.
type CommentDeleted struct {
	BaseEvent
	CommentID string
	Timestamp time.Time
}

func NewCommentDeleted(documentID, commentID string) *CommentDeleted {
	now := time.Now().UTC()
	return &CommentDeleted{
		BaseEvent: newBase(documentID),
		CommentID: commentID,
		Timestamp: now,
	}
}
