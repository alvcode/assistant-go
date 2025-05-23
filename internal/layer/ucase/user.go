package ucase

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/logging"
	"assistant-go/internal/storage/postgres"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"time"
)

const userTokenLifeHours = 4

var (
	ErrUserIncorrectUsernameOrPassword = errors.New("incorrect username or password")
	ErrUserAlreadyExists               = errors.New("user already exists")
	ErrRefreshTokenNotFound            = errors.New("refresh token not found")
	ErrUserNotFound                    = errors.New("user not found")
	ErrUserPasswordsAreNotIdentical    = errors.New("passwords are not identical")
)

type UserUseCase interface {
	Create(in dto.UserLoginAndPassword) (*entity.User, error)
	Login(in dto.UserLoginAndPassword) (*entity.UserToken, error)
	RefreshToken(in dto.UserRefreshToken) (*entity.UserToken, error)
	Delete(userID int) error
	ChangePassword(userID int, in dto.UserChangePassword) error
	CleanOldTokens() error
}

type userUseCase struct {
	ctx          context.Context
	repositories *repository.Repositories
}

func NewUserUseCase(ctx context.Context, repositories *repository.Repositories) UserUseCase {
	return &userUseCase{
		ctx:          ctx,
		repositories: repositories,
	}
}

func (uc *userUseCase) Create(in dto.UserLoginAndPassword) (*entity.User, error) {
	existingUser, err := uc.repositories.UserRepository.Find(in.Login)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), 11)
	if err != nil {
		return nil, err
	}

	userEntity := entity.User{
		Login:     in.Login,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, err := uc.repositories.UserRepository.Create(userEntity)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}
	return data, nil
}

func (uc *userUseCase) Login(in dto.UserLoginAndPassword) (*entity.UserToken, error) {
	existingUser, err := uc.repositories.UserRepository.Find(in.Login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserIncorrectUsernameOrPassword
		}
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(in.Password))
	if err != nil {
		return nil, ErrUserIncorrectUsernameOrPassword
	}

	userTokenEntity, err := uc.generateTokenPair(existingUser.ID)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, ErrUnexpectedError
	}

	data, err := uc.repositories.UserRepository.SetUserToken(*userTokenEntity)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}
	return data, nil
}

func (uc *userUseCase) RefreshToken(in dto.UserRefreshToken) (*entity.UserToken, error) {
	existingToken, err := uc.repositories.UserRepository.FindUserToken(in.Token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, postgres.ErrUnexpectedDBError
	}
	if existingToken.RefreshToken != in.RefreshToken {
		return nil, ErrRefreshTokenNotFound
	}

	userTokenEntity, err := uc.generateTokenPair(int(existingToken.UserId))
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, ErrUnexpectedError
	}
	data, err := uc.repositories.UserRepository.SetUserToken(*userTokenEntity)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}
	return data, nil
}

func (uc *userUseCase) generateTokenPair(userId int) (*entity.UserToken, error) {
	var userTokenEntity *entity.UserToken
	for {
		token, err := generateAPIToken()
		if err != nil {
			return nil, err
		}
		refreshToken, err := generateAPIToken()
		if err != nil {
			return nil, err
		}

		_, err = uc.repositories.UserRepository.FindUserToken(token)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				userTokenEntity = &entity.UserToken{
					UserId:       userId,
					Token:        token,
					RefreshToken: refreshToken,
					ExpiredTo:    int(time.Now().Add(userTokenLifeHours * time.Hour).Unix()),
				}
				break
			}
		}
	}
	return userTokenEntity, nil
}

func generateAPIToken() (string, error) {
	bytes := make([]byte, 48)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (uc *userUseCase) Delete(userID int) error {

	err := uc.repositories.UserRepository.Delete(userID)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}

	err = uc.repositories.UserRepository.DeleteUserTokensByID(userID)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}

	err = uc.repositories.NoteCategoryRepository.DeleteByUserId(userID)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}

	return nil
}

func (uc *userUseCase) ChangePassword(userID int, in dto.UserChangePassword) error {
	user, err := uc.repositories.UserRepository.FindById(userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		logging.GetLogger(uc.ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.CurrentPassword))
	if err != nil {
		return ErrUserPasswordsAreNotIdentical
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), 11)
	if err != nil {
		return err
	}

	err = uc.repositories.UserRepository.ChangePassword(userID, string(hashedPassword))
	if err != nil {
		return err
	}

	err = uc.repositories.UserRepository.DeleteUserTokensByID(userID)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}

	return nil
}

func (uc *userUseCase) CleanOldTokens() error {
	err := uc.repositories.UserRepository.RemoveTokensByDateExpired(
		int(time.Now().AddDate(0, 0, -30).Unix()),
	)

	if err != nil {
		return err
	}
	return nil
}
