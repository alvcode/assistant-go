package dtoUser

import (
	"assistant-go/pkg/vld"
)

type LoginAndPassword struct {
	Login    string `json:"login" validate:"required,max=100"`
	Password string `json:"password" validate:"required,max=200"`
}

func (dto *LoginAndPassword) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}

type RefreshToken struct {
	Token        string `json:"token" validate:"required,min=50,max=100"`
	RefreshToken string `json:"refresh_token" validate:"required,min=50,max=100"`
}

func (dto *RefreshToken) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}

type Token struct {
	Token string `json:"token" validate:"required,min=50,max=100"`
}

func (dto *Token) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}
