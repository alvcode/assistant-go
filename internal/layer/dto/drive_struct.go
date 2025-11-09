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
	SHA256                *string
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
	IsChunk   bool      `db:"is_chunk" json:"is_chunk"`
	SHA256    *string   `db:"sha256" json:"sha256"`
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

type DriveChunkPrepare struct {
	Filename string  `json:"filename" validate:"required,min=1,max=300"`
	FullSize int64   `json:"full_size" validate:"required"`
	ParentID *int    `json:"parent_id"`
	SHA256   *string `json:"sha256"`
}

func (dto *DriveChunkPrepare) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}

type DriveChunkPrepareIn struct {
	DriveChunkPrepare
	MaxSizeBytes          int64
	StorageMaxSizePerUser int64
	SHA256                *string
}

type DriveChunkPrepareResponse struct {
	StructID int `json:"struct_id"`
}

type DriveUploadChunk struct {
	File                  multipart.File
	StructID              int
	ChunkNumber           int
	MaxSizeBytes          int64
	StorageMaxSizePerUser int64
	SavePath              string
}

type DriveChunkEnd struct {
	StructID int `json:"struct_id" validate:"required"`
}

func (dto *DriveChunkEnd) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}

type DriveChunksInfo struct {
	StartNumber int `json:"start_number"`
	EndNumber   int `json:"end_number"`
}
