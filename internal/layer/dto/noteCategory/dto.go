package dtoNoteCategory

import "assistant-go/pkg/vld"

type Create struct {
	Name     string `json:"name" validate:"required,max=255"`
	ParentId uint32 `json:"parent_id"`
}

func (dto *Create) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}
