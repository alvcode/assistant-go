package ucase

import (
	"assistant-go/internal/layer/repository"
	"context"
	"time"
)

type BlockIpUseCase interface {
	CleanOld() error
}

type blockIpUseCase struct {
	ctx          context.Context
	repositories repository.Repositories
}

func NewBlockIpUseCase(ctx context.Context, repositories *repository.Repositories) BlockIpUseCase {
	return &blockIpUseCase{
		ctx:          ctx,
		repositories: *repositories,
	}
}

func (uc *blockIpUseCase) CleanOld() error {
	err := uc.repositories.BlockIPRepository.RemoveByDateExpired(time.Now().UTC())
	if err != nil {
		return err
	}
	return nil
}
