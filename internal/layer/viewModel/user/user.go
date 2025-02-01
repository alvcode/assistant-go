package vmUser

import (
	"assistant-go/internal/layer/entity"
	"time"
)

type User struct {
	ID        uint32    `json:"id"`
	Login     string    `json:"login"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func UserVMFromEnity(entity *entity.User) *User {
	return &User{
		ID:        entity.ID,
		Login:     entity.Login,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}
