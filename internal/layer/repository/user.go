package repository

import (
	"assistant-go/internal/layer/entity"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(in entity.User) (*entity.User, error)
	Find(login string) (*entity.User, error)
	FindById(id int) (*entity.User, error)
	FindUserToken(token string) (*entity.UserToken, error)
	SetUserToken(in entity.UserToken) (*entity.UserToken, error)
	Delete(userID int) error
	DeleteUserTokensByID(userID int) error
}

type userRepository struct {
	ctx context.Context
	db  *pgxpool.Pool
}

func NewUserRepository(ctx context.Context, db *pgxpool.Pool) UserRepository {
	return &userRepository{
		ctx: ctx,
		db:  db,
	}
}

func (ur *userRepository) Create(in entity.User) (*entity.User, error) {
	query := `INSERT INTO users (login, password, created_at, updated_at) VALUES ($1, $2, $3, $4) RETURNING id`

	row := ur.db.QueryRow(ur.ctx, query, in.Login, in.Password, in.CreatedAt, in.UpdatedAt)

	if err := row.Scan(&in.ID); err != nil {
		return nil, err
	}

	return &in, nil
}

func (ur *userRepository) Find(login string) (*entity.User, error) {
	query := `SELECT id, login, password, created_at, updated_at FROM users WHERE login = $1`

	row := ur.db.QueryRow(ur.ctx, query, login)

	var user entity.User
	if err := row.Scan(&user.ID, &user.Login, &user.Password, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *userRepository) FindById(id int) (*entity.User, error) {
	query := `SELECT id, login, password, created_at, updated_at FROM users WHERE id = $1`

	row := ur.db.QueryRow(ur.ctx, query, id)

	var user entity.User
	if err := row.Scan(&user.ID, &user.Login, &user.Password, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *userRepository) FindUserToken(token string) (*entity.UserToken, error) {
	query := `SELECT user_id, token, refresh_token, expired_to FROM user_tokens WHERE token = $1`

	row := ur.db.QueryRow(ur.ctx, query, token)

	var userToken entity.UserToken
	if err := row.Scan(&userToken.UserId, &userToken.Token, &userToken.RefreshToken, &userToken.ExpiredTo); err != nil {
		return nil, err
	}
	return &userToken, nil
}

func (ur *userRepository) SetUserToken(in entity.UserToken) (*entity.UserToken, error) {
	query := `INSERT INTO user_tokens (user_id, token, refresh_token, expired_to) VALUES ($1, $2, $3, $4)`

	_, err := ur.db.Exec(ur.ctx, query, in.UserId, in.Token, in.RefreshToken, in.ExpiredTo)
	return &in, err
}

func (ur *userRepository) Delete(userID int) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := ur.db.Exec(ur.ctx, query, userID)
	if err != nil {
		return err
	}
	return nil
}

func (ur *userRepository) DeleteUserTokensByID(userID int) error {
	query := `DELETE FROM user_tokens WHERE user_id = $1`

	_, err := ur.db.Exec(ur.ctx, query, userID)
	if err != nil {
		return err
	}
	return nil
}
