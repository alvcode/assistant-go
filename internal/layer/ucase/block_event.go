package ucase

import (
	"assistant-go/internal/layer/repository"
	"context"
	"time"
)

type BlockEventUseCase interface {
	CleanOld() error
}

type blockEventUseCase struct {
	ctx          context.Context
	repositories repository.Repositories
}

func NewBlockEventUseCase(ctx context.Context, repositories *repository.Repositories) BlockEventUseCase {
	return &blockEventUseCase{
		ctx:          ctx,
		repositories: *repositories,
	}
}

func (uc *blockEventUseCase) CleanOld() error {
	deleteTime := time.Now().Add(-30 * time.Minute).UTC()
	err := uc.repositories.BlockEventRepository.RemoveByDateExpired(uc.ctx, deleteTime)
	if err != nil {
		return err
	}
	return nil
}
