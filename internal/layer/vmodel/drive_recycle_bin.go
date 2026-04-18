package vmodel

import (
	"assistant-go/internal/layer/entity"
	"time"
)

type DriveRecycleBinStruct struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Type          int8      `json:"type"`
	DriveStructID int       `json:"drive_struct_id"`
	CreatedAt     time.Time `json:"created_at"`
	OriginalPath  string    `json:"original_path"`
}

func DriveRecycleBinStructFromEntity(entity *entity.DriveRecycleBinStruct) *DriveRecycleBinStruct {
	return &DriveRecycleBinStruct{
		ID:            entity.ID,
		Name:          entity.Name,
		Type:          entity.Type,
		DriveStructID: entity.DriveStructID,
		CreatedAt:     entity.CreatedAt,
		OriginalPath:  entity.OriginalPath,
	}
}
