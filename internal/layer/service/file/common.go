package service

import (
	"assistant-go/internal/layer/repository"
	"context"
)

type UploadFile interface {
	UploadFileService(ctx context.Context, repositories *repository.Repositories) PositionService
}

type uploadFile struct{}

func NewUploadFile() UploadFile {
	return &uploadFile{}
}

func (ps *uploadFile) UploadFileService(ctx context.Context, repositories *repository.Repositories) PositionService {
	return &positionService{
		ctx:          ctx,
		repositories: repositories,
	}
}
