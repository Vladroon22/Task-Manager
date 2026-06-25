package models

import (
	"slices"
	"time"
)

var DefaultTags = []string{"отчетность", "операции", "звонок"}

type Tag struct {
	ID        int       `json:"id"`
	Names     []string  `json:"names"`
	CreatedAt time.Time `json:"created_at"`
	UpdateAt  time.Time `json:"updated_at"`
}

type CreateTagRequest struct {
	Name []string `json:"names" validate:"required"`
}

func IsDefaultTag(names []string) bool {
	for _, tag := range names {
		if slices.Contains(DefaultTags, tag) {
			return true
		}
	}
	return false
}
