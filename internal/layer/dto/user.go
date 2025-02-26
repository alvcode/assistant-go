package dto

import (
	"assistant-go/pkg/vld"
)

type UserLoginAndPassword struct {
	Login    string `json:"login" validate:"required,max=100"`
	Password string `json:"password" validate:"required,max=200"`
}

func (dto *UserLoginAndPassword) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}

type UserRefreshToken struct {
	Token        string `json:"token" validate:"required,min=50,max=100"`
	RefreshToken string `json:"refresh_token" validate:"required,min=50,max=100"`
}

func (dto *UserRefreshToken) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}

type UserToken struct {
	Token string `json:"token" validate:"required,min=50,max=100"`
}

func (dto *UserToken) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}

type UserChangePassword struct {
	CurrentPassword string `json:"current_password" validate:"required,max=200"`
	NewPassword     string `json:"new_password" validate:"required,max=200"`
}

func (dto *UserChangePassword) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}
