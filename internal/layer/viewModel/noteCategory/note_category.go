package vmNoteCategory

import (
	"assistant-go/internal/layer/entity"
)

type NoteCategory struct {
	ID       int    `json:"id"`
	UserId   int    `json:"user_id" db:"user_id"`
	Name     string `json:"name" db:"name"`
	ParentId int    `json:"parent_id" db:"parent_id"`
}

func NoteCategoryVMFromEnity(entity *entity.NoteCategory) *NoteCategory {
	return &NoteCategory{
		ID:       entity.ID,
		UserId:   entity.UserId,
		Name:     entity.Name,
		ParentId: *entity.ParentId,
	}
}
