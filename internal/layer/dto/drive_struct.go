package dto

import (
	"assistant-go/pkg/vld"
)

type DriveCreateDirectory struct {
	Name     string `json:"name" validate:"required,min=1,max=350"`
	ParentId *int   `json:"parent_id"`
}

func (dto *DriveCreateDirectory) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}
