package dto

import (
	"assistant-go/pkg/vld"
	"mime/multipart"
)

type UploadFile struct {
	File             multipart.File
	OriginalFilename string `validate:"min=1,max=150"`
	MaxSizeBytes     int64
}

func (dto *UploadFile) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}
