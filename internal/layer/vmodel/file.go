package vmodel

import (
	"assistant-go/internal/layer/entity"
	"time"
)

type File struct {
	ID               int       `json:"id"`
	OriginalFilename string    `json:"original_filename"`
	Ext              string    `json:"ext"`
	SizeBytes        int       `json:"size_bytes"`
	Url              string    `json:"url"`
	CreatedAt        time.Time `json:"created_at"`
}

func FileFromEntity(entity *entity.File, url string) *File {
	return &File{
		ID:               entity.ID,
		OriginalFilename: entity.OriginalFilename,
		Ext:              entity.Ext,
		SizeBytes:        entity.Size,
		Url:              url,
		CreatedAt:        entity.CreatedAt,
	}
}
