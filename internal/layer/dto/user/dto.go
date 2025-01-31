package dtoUser

import "github.com/go-playground/validator/v10"

type CreateDto struct {
	Login    string `json:"login" validate:"required,max=100"`
	Password string `json:"password" validate:"required,max=255"`
}

func (dto *CreateDto) Validate() error {
	validate := validator.New()
	return validate.Struct(dto)
}
