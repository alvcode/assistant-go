package ucase

import (
	"assistant-go/internal/layer/repository"
	"context"
)

type RateLimiterUseCase interface {
	Clean() error
}

type rateLimiterUseCase struct {
	ctx          context.Context
	repositories *repository.Repositories
}

func NewRateLimiterUseCase(ctx context.Context, repositories *repository.Repositories) RateLimiterUseCase {
	return &rateLimiterUseCase{
		ctx:          ctx,
		repositories: repositories,
	}
}

func (uc *rateLimiterUseCase) Clean() error {
	err := uc.repositories.RateLimiterRepository.Clean(uc.ctx)
	if err != nil {
		return err
	}
	return nil
}
