package vmodel

import (
	"assistant-go/internal/layer/entity"
	"encoding/json"
	"time"
)

type NoteMinimal struct {
	ID         int       `json:"id"`
	Title      *string   `json:"title"`
	CategoryID int       `json:"category_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func NoteMinimalFromEnity(entity *entity.Note) *NoteMinimal {
	return &NoteMinimal{
		ID:         entity.ID,
		Title:      entity.Title,
		CategoryID: entity.CategoryID,
		CreatedAt:  entity.CreatedAt,
		UpdatedAt:  entity.UpdatedAt,
	}
}

func NotesMinimalFromEntities(entities []*entity.Note) []*NoteMinimal {
	result := make([]*NoteMinimal, 0, len(entities))
	for _, one := range entities {
		result = append(result, NoteMinimalFromEnity(one))
	}
	return result
}

type Note struct {
	ID         int             `json:"id"`
	Title      *string         `json:"title"`
	CategoryID int             `json:"category_id"`
	NoteBlocks json.RawMessage `json:"note_blocks"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

func NoteFromEnity(entity *entity.Note) *Note {
	return &Note{
		ID:         entity.ID,
		Title:      entity.Title,
		CategoryID: entity.CategoryID,
		NoteBlocks: entity.NoteBlocks,
		CreatedAt:  entity.CreatedAt,
		UpdatedAt:  entity.UpdatedAt,
	}
}
