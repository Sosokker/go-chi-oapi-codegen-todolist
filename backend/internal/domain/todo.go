package domain

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type TodoStatus string

const (
	StatusPending    TodoStatus = "pending"
	StatusInProgress TodoStatus = "in-progress"
	StatusCompleted  TodoStatus = "completed"
)

type Todo struct {
	ID            uuid.UUID   `json:"id"`
	UserID        uuid.UUID   `json:"userId"`
	Title         string      `json:"title"`
	Description   *string     `json:"description"` // Nullable
	Status        TodoStatus  `json:"status"`
	Deadline      *time.Time  `json:"deadline"`      // Nullable
	TagIDs        []uuid.UUID `json:"tagIds"`        // Populated after fetching
	Tags          []Tag       `json:"-"`             // Loaded separately
	AttachmentUrl *string     `json:"attachmentUrl"` // Renamed and changed type
	Subtasks      []Subtask   `json:"subtasks"`      // Populated after fetching
	CreatedAt     time.Time   `json:"createdAt"`
	UpdatedAt     time.Time   `json:"updatedAt"`
}

// Keep AttachmentInfo for upload responses
type AttachmentInfo struct {
	FileID      string `json:"fileId"`
	FileName    string `json:"fileName"`
	FileURL     string `json:"fileUrl"`
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
}

// Helper functions remain the same
func NullStringToStringPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func NullTimeToTimePtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}
