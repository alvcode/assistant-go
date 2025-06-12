package ucase

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/logging"
	"assistant-go/internal/storage/postgres"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"time"
)

const (
	typeDirectory = 0
	typeFile      = 1
)

var (
	ErrDriveDirectoryExists  = errors.New("directory exists")
	ErrDriveParentIdNotFound = errors.New("drive parent id does not exist")
)

type DriveUseCase interface {
	CreateDirectory(dto *dto.DriveCreateDirectory, user *entity.User) ([]*entity.DriveStruct, error)
	GetTree(parentID *int, user *entity.User) ([]*entity.DriveStruct, error)
}

type driveUseCase struct {
	ctx          context.Context
	repositories *repository.Repositories
}

func NewDriveUseCase(ctx context.Context, repositories *repository.Repositories) DriveUseCase {
	return &driveUseCase{
		ctx:          ctx,
		repositories: repositories,
	}
}

func (uc *driveUseCase) GetTree(parentID *int, user *entity.User) ([]*entity.DriveStruct, error) {
	list, err := uc.repositories.DriveStructRepository.ListByUserID(user.ID, parentID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			logging.GetLogger(uc.ctx).Error(err)
			return nil, postgres.ErrUnexpectedDBError
		}
	}
	return list, nil
}

func (uc *driveUseCase) CreateDirectory(dto *dto.DriveCreateDirectory, user *entity.User) ([]*entity.DriveStruct, error) {
	if dto.ParentID != nil {
		parentStruct, err := uc.repositories.DriveStructRepository.GetByID(*dto.ParentID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrDriveParentIdNotFound
			}
		}
		if parentStruct.UserID != user.ID {
			return nil, ErrDriveDirectoryExists
		}
	}
	_, err := uc.repositories.DriveStructRepository.FindRow(user.ID, dto.Name, typeDirectory, dto.ParentID)

	if err == nil {
		return nil, ErrDriveDirectoryExists
	}
	createEntity := &entity.DriveStruct{
		UserID:    user.ID,
		Name:      dto.Name,
		Type:      typeDirectory,
		ParentID:  dto.ParentID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	err = uc.repositories.DriveStructRepository.CreateDirectory(createEntity)
	if err != nil {
		return nil, err
	}

	treeList, err := uc.GetTree(dto.ParentID, user)
	if err != nil {
		return nil, err
	}
	return treeList, nil
}
