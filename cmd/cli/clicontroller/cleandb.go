package clicontroller

import (
	"assistant-go/internal/config"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/layer/ucase"
	"assistant-go/internal/logging"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CleanDBInit(ctx context.Context, cfg *config.Config, db *pgxpool.Pool) {
	fmt.Println("start clean-db cli command")
	logging.GetLogger(ctx).Println("start clean-db cli command")
	repos := repository.NewRepositories(ctx, db)

	blockIpUseCase := ucase.NewBlockIpUseCase(ctx, repos)
	err := blockIpUseCase.CleanOld()
	if err != nil {
		fmt.Printf("Error clean block ip: %v", err)
		logging.GetLogger(ctx).Errorf("Error clean block ip: %v", err)
		return
	}

	userUseCase := ucase.NewUserUseCase(ctx, repos)
	err = userUseCase.CleanOldTokens()
	if err != nil {
		fmt.Printf("Error clean user tokens: %v", err)
		logging.GetLogger(ctx).Errorf("Error clean user tokens: %v", err)
		return
	}

	blockEventUseCase := ucase.NewBlockEventUseCase(ctx, repos)
	err = blockEventUseCase.CleanOld()
	if err != nil {
		fmt.Printf("Error clean block events: %v", err)
		logging.GetLogger(ctx).Errorf("Error clean block events: %v", err)
		return
	}

	db.Close()
	fmt.Println("successfully")
	logging.GetLogger(ctx).Println("successfully")
}
