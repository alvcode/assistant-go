package repository

import (
	"assistant-go/internal/layer/entity"
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(in entity.User) (*entity.User, error)
	Find(login string) (*entity.User, error)
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
		return nil, errors.New("failed to insert user: " + err.Error())
	}

	return &in, nil
}

func (ur *userRepository) Find(login string) (*entity.User, error) {
	query := `SELECT id, login, password, created_at, updated_at FROM users WHERE login = $1`

	row := ur.db.QueryRow(ur.ctx, query, login)

	var user entity.User
	if err := row.Scan(&user.ID, &user.Login, &user.Password, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, errors.New("user not found")
	}

	return &user, nil
}
