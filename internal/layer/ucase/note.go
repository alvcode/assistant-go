package ucase

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/logging"
	"assistant-go/internal/storage/postgres"
	"assistant-go/pkg/utils"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/tidwall/gjson"
	"time"
)

var (
	ErrNoteNotFound = errors.New("note not found")
)

type NoteUseCase interface {
	Create(in dto.NoteCreate, userEntity *entity.User) (*entity.Note, error)
	GetAll(catIdStruct dto.RequiredID, userEntity *entity.User) ([]*entity.Note, error)
	Update(in dto.NoteUpdate, userEntity *entity.User) (*entity.Note, error)
	GetOne(noteIdStruct dto.RequiredID, userEntity *entity.User) (*entity.Note, error)
	DeleteOne(noteIdStruct dto.RequiredID, userEntity *entity.User) error
	Pin(noteIdStruct dto.RequiredID, userEntity *entity.User) error
	UnPin(noteIdStruct dto.RequiredID, userEntity *entity.User) error
}

type noteUseCase struct {
	ctx          context.Context
	repositories repository.Repositories
}

func NewNoteUseCase(ctx context.Context, repositories *repository.Repositories) NoteUseCase {
	return &noteUseCase{
		ctx:          ctx,
		repositories: *repositories,
	}
}

func (uc *noteUseCase) Create(in dto.NoteCreate, userEntity *entity.User) (*entity.Note, error) {
	_, err := uc.repositories.NoteCategoryRepository.FindByIDAndUser(uc.ctx, userEntity.ID, in.CategoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	timeNow := time.Now().UTC()

	var pinned bool
	if in.Pinned == nil {
		pinned = false
	} else {
		pinned = *in.Pinned
	}

	noteEntity := entity.Note{
		CategoryID: in.CategoryID,
		NoteBlocks: in.NoteBlocks,
		CreatedAt:  timeNow,
		UpdatedAt:  timeNow,
		Title:      uc.getNoteTitle(in.Title, string(in.NoteBlocks)),
		Pinned:     pinned,
	}

	data, err := uc.repositories.NoteRepository.Create(uc.ctx, noteEntity)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	fileIDs, _ := getFileIDsByBlocks(string(in.NoteBlocks))
	err = uc.repositories.FileNoteLinkRepository.Upsert(uc.ctx, data.ID, fileIDs)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (uc *noteUseCase) GetAll(catIdStruct dto.RequiredID, userEntity *entity.User) ([]*entity.Note, error) {
	categories, err := uc.repositories.NoteCategoryRepository.FindByIDAndUserWithChildren(uc.ctx, userEntity.ID, catIdStruct.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	catIds := make([]int, 0)
	for _, cat := range categories {
		catIds = append(catIds, cat.ID)
	}
	if len(catIds) == 0 {
		return nil, ErrCategoryNotFound
	}

	notes, err := uc.repositories.NoteRepository.GetMinimalByCategoryIds(uc.ctx, catIds)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			logging.GetLogger(uc.ctx).Error(err)
			return nil, postgres.ErrUnexpectedDBError
		}
	}
	return notes, nil
}

func (uc *noteUseCase) Update(in dto.NoteUpdate, userEntity *entity.User) (*entity.Note, error) {
	currentNote, err := uc.repositories.NoteRepository.GetById(uc.ctx, in.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoteNotFound
		}
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	// проверим, что текущая категория заметки принадлежит пользователю
	_, err = uc.repositories.NoteCategoryRepository.FindByIDAndUser(uc.ctx, userEntity.ID, currentNote.CategoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoteNotFound
		}
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	// проверим, что новая категория заметки принадлежит пользователю, если она изменяется
	if currentNote.CategoryID != in.CategoryID {
		_, err := uc.repositories.NoteCategoryRepository.FindByIDAndUser(uc.ctx, userEntity.ID, in.CategoryID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrCategoryNotFound
			}
			logging.GetLogger(uc.ctx).Error(err)
			return nil, postgres.ErrUnexpectedDBError
		}
	}

	var pinned bool
	if in.Pinned == nil {
		pinned = currentNote.Pinned
	} else {
		pinned = *in.Pinned
	}

	currentNote.NoteBlocks = in.NoteBlocks
	currentNote.CategoryID = in.CategoryID
	currentNote.Title = uc.getNoteTitle(in.Title, string(in.NoteBlocks))
	currentNote.UpdatedAt = time.Now().UTC()
	currentNote.Pinned = pinned

	err = uc.repositories.NoteRepository.Update(uc.ctx, currentNote)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	fileIDs, _ := getFileIDsByBlocks(string(in.NoteBlocks))
	err = uc.repositories.FileNoteLinkRepository.Upsert(uc.ctx, currentNote.ID, fileIDs)
	if err != nil {
		return nil, err
	}

	return currentNote, nil
}

func getFileIDsByBlocks(blocks string) ([]int, error) {
	var result []int
	attaches := gjson.Get(blocks, `#(type="attaches")#.data.file.id`)
	images := gjson.Get(blocks, `#(type="image")#.data.file.id`)

	collect := func(values gjson.Result) {
		if values.IsArray() {
			for _, id := range values.Array() {
				if id.Type == gjson.Number {
					result = append(result, int(id.Int()))
				}
			}
		} else if values.Type == gjson.Number {
			result = append(result, int(values.Int()))
		}
	}

	collect(attaches)
	collect(images)

	return result, nil
}

func (uc *noteUseCase) GetOne(noteIdStruct dto.RequiredID, userEntity *entity.User) (*entity.Note, error) {
	currentNote, err := uc.repositories.NoteRepository.GetById(uc.ctx, noteIdStruct.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoteNotFound
		}
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	// проверим, что текущая категория заметки принадлежит пользователю
	_, err = uc.repositories.NoteCategoryRepository.FindByIDAndUser(uc.ctx, userEntity.ID, currentNote.CategoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoteNotFound
		}
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	return currentNote, nil
}

func (uc *noteUseCase) DeleteOne(noteIdStruct dto.RequiredID, userEntity *entity.User) error {
	currentNote, err := uc.repositories.NoteRepository.GetById(uc.ctx, noteIdStruct.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoteNotFound
		}
		logging.GetLogger(uc.ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}

	// проверим, что текущая категория заметки принадлежит пользователю
	_, err = uc.repositories.NoteCategoryRepository.FindByIDAndUser(uc.ctx, userEntity.ID, currentNote.CategoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoteNotFound
		}
		logging.GetLogger(uc.ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}

	err = uc.repositories.NoteRepository.DeleteOne(uc.ctx, currentNote.ID)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}

	err = uc.repositories.FileNoteLinkRepository.DeleteByNoteID(uc.ctx, currentNote.ID)
	if err != nil {
		return err
	}

	return nil
}

func (uc *noteUseCase) getNoteTitle(title string, blocks string) *string {
	stringUtils := utils.NewStringUtils()

	var titleTruncate string
	if title != "" {
		titleTruncate = stringUtils.Trim(stringUtils.TruncateString(title, 150))
	} else {
		firstBlockText := gjson.Get(blocks, `0.data.text`)
		titleWithoutHtml := stringUtils.RemoveHTMLTags(firstBlockText.Str)
		titleTruncate = stringUtils.Trim(stringUtils.TruncateString(titleWithoutHtml, 150))
	}

	if titleTruncate == "" {
		return nil
	} else {
		return &titleTruncate
	}
}

func (uc *noteUseCase) Pin(noteIdStruct dto.RequiredID, userEntity *entity.User) error {
	currentNote, err := uc.repositories.NoteRepository.GetById(uc.ctx, noteIdStruct.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoteNotFound
		}
		logging.GetLogger(uc.ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}

	// проверим, что текущая категория заметки принадлежит пользователю
	_, err = uc.repositories.NoteCategoryRepository.FindByIDAndUser(uc.ctx, userEntity.ID, currentNote.CategoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoteNotFound
		}
		logging.GetLogger(uc.ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}

	err = uc.repositories.NoteRepository.Pin(uc.ctx, currentNote.ID)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}
	return nil
}

func (uc *noteUseCase) UnPin(noteIdStruct dto.RequiredID, userEntity *entity.User) error {
	currentNote, err := uc.repositories.NoteRepository.GetById(uc.ctx, noteIdStruct.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoteNotFound
		}
		logging.GetLogger(uc.ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}

	_, err = uc.repositories.NoteCategoryRepository.FindByIDAndUser(uc.ctx, userEntity.ID, currentNote.CategoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoteNotFound
		}
		logging.GetLogger(uc.ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}

	err = uc.repositories.NoteRepository.UnPin(uc.ctx, currentNote.ID)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}
	return nil
}
