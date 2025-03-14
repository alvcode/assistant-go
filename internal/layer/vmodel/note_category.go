package vmodel

import (
	"assistant-go/internal/layer/entity"
)

type NoteCategory struct {
	ID       int    `json:"id"`
	UserId   int    `json:"user_id"`
	Name     string `json:"name"`
	ParentId *int   `json:"parent_id"`
	Position int    `json:"position"`
}

func NoteCategoryFromEnity(entity *entity.NoteCategory) *NoteCategory {
	return &NoteCategory{
		ID:       entity.ID,
		UserId:   entity.UserId,
		Name:     entity.Name,
		ParentId: entity.ParentId,
	}
}

func NoteCategoriesFromEntities(entities []*entity.NoteCategory) []*NoteCategory {
	vmodelEntities := make([]*NoteCategory, len(entities))
	for i, one := range entities {
		vmodelEntities[i] = NoteCategoryFromEnity(one)
	}
	return vmodelEntities
}
