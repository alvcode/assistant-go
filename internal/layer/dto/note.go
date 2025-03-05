package dto

import (
	"assistant-go/pkg/vld"
	"encoding/json"
)

type NoteCreate struct {
	CategoryID int             `json:"category_id" validate:"required"`
	NoteBlocks json.RawMessage `json:"note_blocks" validate:"json"`
}

func (dto *NoteCreate) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}

type NoteUpdate struct {
	ID         int             `json:"id" validate:"required"`
	CategoryID int             `json:"category_id" validate:"required"`
	NoteBlocks json.RawMessage `json:"note_blocks" validate:"json"`
}

func (dto *NoteUpdate) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}
