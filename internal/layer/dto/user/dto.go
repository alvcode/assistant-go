package dtoUser

import (
	"assistant-go/pkg/vld"
)

type CreateDto struct {
	Login    string `json:"login" validate:"required,max=100"`
	Password string `json:"password" validate:"required,max=255"`
}

func (dto *CreateDto) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}
