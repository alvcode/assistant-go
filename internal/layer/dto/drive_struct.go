package dto

import (
	"assistant-go/pkg/vld"
	"mime/multipart"
	"time"
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

type DriveRenameStruct struct {
	Name string `json:"name" validate:"required,min=1,max=350"`
}

func (dto *DriveRenameStruct) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}

type DriveSpace struct {
	Total int64 `json:"total"`
	Used  int64 `json:"used"`
}

type DriveTree struct {
	ID        int       `db:"id" json:"id"`
	UserID    int       `db:"user_id" json:"user_id"`
	Name      string    `db:"name" json:"name"`
	Type      int8      `db:"type" json:"type"`
	Size      int64     `db:"size" json:"size"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type DriveRenMov struct {
	StructIDs []int `json:"struct_ids" validate:"required"`
	ParentID  *int  `json:"parent_id"`
}

func (dto *DriveRenMov) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}
