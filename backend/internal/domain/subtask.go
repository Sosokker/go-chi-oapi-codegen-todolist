package domain

import (
	"time"

	"github.com/google/uuid"
)

type Subtask struct {
	ID          uuid.UUID `json:"id"`
	TodoID      uuid.UUID `json:"todoId"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
