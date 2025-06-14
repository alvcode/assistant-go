package dto

import (
	"assistant-go/pkg/vld"
	"mime/multipart"
)

type DriveCreateDirectory struct {
	Name     string `json:"name" validate:"required,min=1,max=350"`
	ParentID *int   `json:"parent_id"`
}

func (dto *DriveCreateDirectory) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}

type DriveUploadFile struct {
	File                  multipart.File
	OriginalFilename      string `validate:"min=1,max=300"`
	MaxSizeBytes          int64
	StorageMaxSizePerUser int64
	SavePath              string
	ParentID              *int
}

func (dto *DriveUploadFile) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}
