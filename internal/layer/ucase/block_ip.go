package ucase

import (
	"assistant-go/internal/layer/repository"
	"context"
	"time"
)

type BlockIpUseCase interface {
	CleanOld(ctx context.Context) error
}

type blockIpUseCase struct {
	repositories repository.Repositories
}

func NewBlockIpUseCase(repositories *repository.Repositories) BlockIpUseCase {
	return &blockIpUseCase{
		repositories: *repositories,
	}
}

func (uc *blockIpUseCase) CleanOld(ctx context.Context) error {
	err := uc.repositories.BlockIPRepository.RemoveByDateExpired(ctx, time.Now().UTC())
	if err != nil {
		return err
	}
	return nil
}
