package domain

import (
	"time"

	"github.com/google/uuid"
)

type Tag struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	Name      string    `json:"name"`
	Color     *string   `json:"color"`
	Icon      *string   `json:"icon"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
