package vmodel

import (
	"assistant-go/internal/layer/entity"
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Login     string    `json:"login"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func UserFromEnity(entity *entity.User) *User {
	return &User{
		ID:        entity.ID,
		Login:     entity.Login,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}

type UserToken struct {
	UserId       int    `json:"user_id"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiredTo    int    `json:"expired_to"`
}

func UserTokenFromEnity(entity *entity.UserToken) *UserToken {
	return &UserToken{
		UserId:       entity.UserId,
		Token:        entity.Token,
		RefreshToken: entity.RefreshToken,
		ExpiredTo:    entity.ExpiredTo,
	}
}
