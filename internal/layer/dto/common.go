package dto

import "assistant-go/pkg/vld"

type RequiredID struct {
	ID int `json:"id" validate:"required"`
}

func (dto *RequiredID) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}
