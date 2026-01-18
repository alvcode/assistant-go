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

func CleanDB(ctx context.Context, cfg *config.Config, db *pgxpool.Pool, minio *minio.Client) {
	fmt.Println("start clean-db cli command")
	logging.GetLogger(ctx).Println("start clean-db cli command")
	repos := repository.NewRepositories(cfg, db, minio)

	blockIpUseCase := ucase.NewBlockIpUseCase(repos)
	err := blockIpUseCase.CleanOld(ctx)
	if err != nil {
		fmt.Printf("Error clean block ip: %v", err)
		logging.GetLogger(ctx).Errorf("Error clean block ip: %v", err)
		return
	}

	userUseCase := ucase.NewUserUseCase(repos)
	err = userUseCase.CleanOldTokens(ctx)
	if err != nil {
		fmt.Printf("Error clean user tokens: %v", err)
		logging.GetLogger(ctx).Errorf("Error clean user tokens: %v", err)
		return
	}

	blockEventUseCase := ucase.NewBlockEventUseCase(repos)
	err = blockEventUseCase.CleanOld(ctx)
	if err != nil {
		fmt.Printf("Error clean block events: %v", err)
		logging.GetLogger(ctx).Errorf("Error clean block events: %v", err)
		return
	}

	fileUseCase := ucase.NewFileUseCase(repos)
	err = fileUseCase.CleanUnused(ctx, cfg.File.SavePath)
	if err != nil {
		fmt.Printf("Error clean unused files: %v", err)
		logging.GetLogger(ctx).Errorf("Error clean unused files: %v", err)
		return
	}

	rateLimiterUseCase := ucase.NewRateLimiterUseCase(repos)
	err = rateLimiterUseCase.Clean(ctx)
	if err != nil {
		fmt.Printf("Error clean rate limiters: %v", err)
		logging.GetLogger(ctx).Errorf("Error clean rate limiters: %v", err)
		return
	}

	db.Close()
	fmt.Println("successfully")
	logging.GetLogger(ctx).Println("successfully")
}
