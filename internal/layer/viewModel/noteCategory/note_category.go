package vmNoteCategory

import (
	"assistant-go/internal/layer/entity"
)

type NoteCategory struct {
	ID       uint32 `json:"id"`
	UserId   uint32 `json:"user_id" db:"user_id"`
	Name     string `json:"name" db:"name"`
	ParentId string `json:"parent_id" db:"parent_id"`
}

func NoteCategoryVMFromEnity(entity *entity.NoteCategory) *NoteCategory {
	return &NoteCategory{
		ID:       entity.ID,
		UserId:   entity.UserId,
		Name:     entity.Name,
		ParentId: entity.ParentId,
	}
}
