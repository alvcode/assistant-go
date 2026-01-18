package ucase

import (
	"assistant-go/internal/layer/repository"
	"context"
)

type RateLimiterUseCase interface {
	Clean(ctx context.Context) error
}

type rateLimiterUseCase struct {
	repositories *repository.Repositories
}

func NewRateLimiterUseCase(repositories *repository.Repositories) RateLimiterUseCase {
	return &rateLimiterUseCase{
		repositories: repositories,
	}
}

func (uc *rateLimiterUseCase) Clean(ctx context.Context) error {
	err := uc.repositories.RateLimiterRepository.Clean(ctx)
	if err != nil {
		return err
	}
	return nil
}
