package ucase

import (
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/logging"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrDriveRecycleBinNotFound = errors.New("drive recycle bin not found")
)

type DriveRecycleBinUseCase interface {
	GetAll(ctx context.Context, user *entity.User) ([]*entity.DriveRecycleBinStruct, error)
	RestoreOne(ctx context.Context, user *entity.User, recycleBinID int) error
}

type driveRecycleBinUseCase struct {
	repositories *repository.Repositories
}

func NewDriveRecycleBinUseCase(repositories *repository.Repositories) DriveRecycleBinUseCase {
	return &driveRecycleBinUseCase{
		repositories: repositories,
	}
}

func (uc *driveRecycleBinUseCase) GetAll(ctx context.Context, user *entity.User) ([]*entity.DriveRecycleBinStruct, error) {
	result, err := uc.repositories.DriveRecycleBinRepository.GetAll(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (uc *driveRecycleBinUseCase) RestoreOne(ctx context.Context, user *entity.User, recycleBinID int) error {
	recycleBin, err := uc.repositories.DriveRecycleBinRepository.GetByID(ctx, recycleBinID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrDriveRecycleBinNotFound
		}
		logging.GetLogger(ctx).Error(err)
		return err
	}

	driveStruct, err := uc.repositories.DriveStructRepository.GetByID(ctx, recycleBin.DriveStructID, true)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrDriveRecycleBinNotFound
		}
		logging.GetLogger(ctx).Error(err)
		return err
	}

	if driveStruct.UserID != user.ID {
		return ErrDriveRecycleBinNotFound
	}

	parts := make([]string, 0)
	cleanPath := strings.Trim(recycleBin.OriginalPath, "/")
	if cleanPath != "" {
		parts = strings.Split(cleanPath, "/")
	}

	var lastStructID *int = nil
	for _, part := range parts {
		foundStruct, err := uc.repositories.DriveStructRepository.FindRow(
			ctx, user.ID, part, typeDirectory, lastStructID, false,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				createEntity := &entity.DriveStruct{
					UserID:    user.ID,
					Name:      part,
					Type:      typeDirectory,
					ParentID:  lastStructID,
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
				}
				createdStruct, err := uc.repositories.DriveStructRepository.Create(ctx, createEntity)
				if err != nil {
					logging.GetLogger(ctx).Error(err)
					return err
				}

				lastStructID = &createdStruct.ID
				continue
			} else {
				logging.GetLogger(ctx).Error(err)
				return err
			}
		}

		if lastStructID != nil && foundStruct.ParentID != lastStructID {
			foundStruct.ParentID = lastStructID
			err = uc.repositories.DriveStructRepository.Update(ctx, foundStruct)
			if err != nil {
				logging.GetLogger(ctx).Error(err)
				return err
			}
		}

		lastStructID = &foundStruct.ID
	}

	if driveStruct.ParentID != lastStructID {
		driveStruct.ParentID = lastStructID
		err = uc.repositories.DriveStructRepository.Update(ctx, driveStruct)
		if err != nil {
			logging.GetLogger(ctx).Error(err)
			return err
		}
	}

	err = uc.repositories.DriveRecycleBinRepository.DeleteByID(ctx, recycleBinID)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return err
	}

	return nil
}
