package ucase

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	service "assistant-go/internal/layer/service/note_category"
	"assistant-go/internal/logging"
	"assistant-go/internal/storage/postgres"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrFileTooLarge              = errors.New("file too large")
	ErrFileReading               = errors.New("error reading file")
	ErrFileInvalidType           = errors.New("invalid file type")
	ErrFileResettingPointer      = errors.New("error resetting pointer")
	ErrFileUnableToSeek          = errors.New("error unable to seek")
	ErrFileExtensionDoesNotMatch = errors.New("file extension does not match")
	ErrFileNotSafeFilename       = errors.New("file not safe filename")
)

type FileUseCase interface {
	Upload(in dto.UploadFile, userEntity *entity.User) (*entity.File, error)
}

type fileUseCase struct {
	ctx          context.Context
	repositories *repository.Repositories
}

func NewFileUseCase(ctx context.Context, repositories *repository.Repositories) FileUseCase {
	return &fileUseCase{
		ctx:          ctx,
		repositories: repositories,
	}
}

func (uc *fileUseCase) Upload(in dto.UploadFile, userEntity *entity.User) (*entity.File, error) {
	var allowedMimeTypes = map[string][]string{
		"image/jpeg":      {".jpeg", ".jpg"},
		"image/png":       {".png"},
		"image/gif":       {".gif"},
		"application/pdf": {".pdf"},
		"application/zip": {".zip"},
	}

	limitedReader := io.LimitReader(in.File, in.MaxSizeBytes+1)

	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, err
	}

	if int64(len(data)) > in.MaxSizeBytes {
		return nil, ErrFileTooLarge
	}

	buffer := make([]byte, 512)
	_, err = in.File.Read(buffer)
	if err != nil {
		return nil, ErrFileReading
	}

	mimeType := http.DetectContentType(buffer)
	extAllowed, allowed := allowedMimeTypes[mimeType]
	if !allowed {
		return nil, ErrFileInvalidType
	}

	if seeker, ok := in.File.(io.Seeker); ok {
		_, err := seeker.Seek(0, io.SeekStart)
		if err != nil {
			return nil, ErrFileResettingPointer
		}
	} else {
		return nil, ErrFileUnableToSeek
	}

	fileExt := strings.ToLower(filepath.Ext(in.OriginalFilename))
	var extExists bool
	for _, extName := range extAllowed {
		if extName == fileExt {
			extExists = true
		}
	}
	if !extExists {
		return nil, ErrFileExtensionDoesNotMatch
	}

	safeName := filepath.Base(in.OriginalFilename)
	if strings.Contains(safeName, "..") {
		return nil, ErrFileNotSafeFilename
	}

	newFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), fileExt)
	filePath := filepath.Join(uploadPath, newFilename)

	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			http.Error(w, "failed to close output file", http.StatusInternalServerError)
			return
		}
	}(out)

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	//noteCategoryEntity := entity.NoteCategory{
	//	UserId:   userEntity.ID,
	//	Name:     in.Name,
	//	ParentId: in.ParentId,
	//}
	//
	//if in.ParentId != nil {
	//	_, err := uc.repositories.NoteCategoryRepository.FindByIDAndUser(userEntity.ID, *in.ParentId)
	//	if err != nil {
	//		if errors.Is(err, pgx.ErrNoRows) {
	//			return nil, ErrCategoryParentIdNotFound
	//		}
	//		logging.GetLogger(uc.ctx).Error(err)
	//		return nil, postgres.ErrUnexpectedDBError
	//	}
	//}
	//
	//positionService := service.NewNoteCategory().PositionService(uc.ctx, uc.repositories)
	//newPosition, err := positionService.CalculateForNew(userEntity.ID, in.ParentId)
	//if err != nil {
	//	logging.GetLogger(uc.ctx).Error(err)
	//	return nil, postgres.ErrUnexpectedDBError
	//}
	//
	//noteCategoryEntity.Position = newPosition
	//
	//data, err := uc.repositories.NoteCategoryRepository.Create(noteCategoryEntity)
	//if err != nil {
	//	logging.GetLogger(uc.ctx).Error(err)
	//	return nil, postgres.ErrUnexpectedDBError
	//}
	//return data, nil
}
