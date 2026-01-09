package clicontroller

import (
	"assistant-go/internal/config"
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/layer/ucase"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
)

func UserRegister(ctx context.Context, cfg *config.Config, db *pgxpool.Pool, minio *minio.Client, login string, password string) {
	repos := repository.NewRepositories(ctx, cfg, db, minio)

	userUseCase := ucase.NewUserUseCase(ctx, repos)

	createDTO := dto.UserLoginAndPassword{
		Login:    login,
		Password: password,
	}

	_, err := userUseCase.Create(createDTO)
	if err != nil {
		fmt.Printf("Error create a new user: %v", err)
		return
	}

	db.Close()
	fmt.Println("successfully")
}
