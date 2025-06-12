package vmodel

import (
	"assistant-go/internal/layer/entity"
	"time"
)

type DriveStruct struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	Type      int8      `json:"type"`
	ParentID  *int      `json:"parent_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func DriveStructFromEntity(entity *entity.DriveStruct) *DriveStruct {
	return &DriveStruct{
		ID:        entity.ID,
		UserID:    entity.UserID,
		Name:      entity.Name,
		Type:      entity.Type,
		ParentID:  entity.ParentID,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}

func DriveStructsFromEntities(entities []*entity.DriveStruct) []*DriveStruct {
	result := make([]*DriveStruct, 0, len(entities))
	for _, one := range entities {
		result = append(result, DriveStructFromEntity(one))
	}
	return result
}
