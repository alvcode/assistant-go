package dto

import "assistant-go/pkg/vld"

type NoteCategoryCreate struct {
	Name     string `json:"name" validate:"required,max=255"`
	ParentId *int   `json:"parent_id"`
}

func (dto *NoteCategoryCreate) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}

type NoteCategoryUpdate struct {
	ID       int    `json:"id" validate:"required"`
	Name     string `json:"name" validate:"required,max=255"`
	ParentID *int   `json:"parent_id"`
}

func (dto *NoteCategoryUpdate) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}
