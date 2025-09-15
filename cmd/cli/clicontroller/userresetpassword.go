package clicontroller

import (
	"assistant-go/internal/config"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/layer/ucase"
	"assistant-go/internal/logging"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
)

func UserResetPassword(ctx context.Context, cfg *config.Config, db *pgxpool.Pool, minio *minio.Client, login string, password string) {
	fmt.Println("start reset-password cli command")
	logging.GetLogger(ctx).Println("start clean-db cli command")
	repos := repository.NewRepositories(ctx, cfg, db, minio)

	userUseCase := ucase.NewUserUseCase(ctx, repos)
	err := userUseCase.ChangePasswordWithoutCurrent(login, password)
	if err != nil {
		fmt.Printf("Error change password without current: %v", err)
		logging.GetLogger(ctx).Errorf("Error change password without current: %v", err)
		return
	}

	db.Close()
	fmt.Println("successfully")
	logging.GetLogger(ctx).Println("successfully")
}
