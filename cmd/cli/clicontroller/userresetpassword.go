package clicontroller

import (
	"assistant-go/internal/config"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/layer/ucase"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
)

func UserResetPassword(ctx context.Context, cfg *config.Config, db *pgxpool.Pool, minio *minio.Client, login string, password string) {
	repos := repository.NewRepositories(cfg, db, minio)

	userUseCase := ucase.NewUserUseCase(repos)
	err := userUseCase.ChangePasswordWithoutCurrent(ctx, login, password)
	if err != nil {
		fmt.Printf("Error change password without current: %v", err)
		return
	}

	db.Close()
	fmt.Println("successfully")
}
