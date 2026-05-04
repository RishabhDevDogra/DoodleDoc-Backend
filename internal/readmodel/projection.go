package readmodel

import "time"

// DocumentProjection is the read model — what REST endpoints return to clients.
// Built by replaying events; optimised for queries, not writes.
type DocumentProjection struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Version   int       `json:"version"`
	Comments  []Comment `json:"comments"`
}

// Comment is a denormalised comment stored directly on the projection.
type Comment struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Author    string    `json:"author"`
	Timestamp time.Time `json:"timestamp"`
}
