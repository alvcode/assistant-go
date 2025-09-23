package dto

import (
	"assistant-go/pkg/vld"
	"io"
	"mime/multipart"
)

type UploadFile struct {
	File             multipart.File
	OriginalFilename string `validate:"min=1,max=200"`
	MaxSizeBytes     int64
	StorageMaxSize   int64
	SavePath         string
}

func (dto *UploadFile) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}

type SaveFile struct {
	File      io.Reader
	SavePath  string
	SizeBytes int64
}

type GetFileByHash struct {
	Hash     string `validate:"required,min=80,max=80"`
	SavePath string
}

func (dto *GetFileByHash) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}

type FileResponse struct {
	File             io.Reader
	OriginalFilename string
	SizeBytes        int64
}
