package ucase

import (
	"assistant-go/internal/layer/repository"
	"context"
	"time"
)

type BlockEventUseCase interface {
	CleanOld(ctx context.Context) error
}

type blockEventUseCase struct {
	repositories repository.Repositories
}

func NewBlockEventUseCase(repositories *repository.Repositories) BlockEventUseCase {
	return &blockEventUseCase{
		repositories: *repositories,
	}
}

func (uc *blockEventUseCase) CleanOld(ctx context.Context) error {
	deleteTime := time.Now().Add(-30 * time.Minute).UTC()
	err := uc.repositories.BlockEventRepository.RemoveByDateExpired(ctx, deleteTime)
	if err != nil {
		return err
	}
	return nil
}
