package services

import (
	"context"
	"errors"

	"github.com/GFernandesS/go-bid/internal/store/pgstore"
	"github.com/GFernandesS/go-bid/internal/usecase/users"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	queries *pgstore.Queries
	pool    *pgxpool.Pool
}

func NewUserService(pool *pgxpool.Pool) UserService {
	return UserService{
		pool:    pool,
		queries: pgstore.New(pool),
	}
}

var (
	ErrDuplicatedEmailOrPassword = errors.New("username or email already exists")
	ErrInvalidUserOrPassword     = errors.New("invalid user or password")
)

func (us *UserService) CreateUser(ctx context.Context, request users.CreateUserRequest) (uuid.UUID, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), 12)

	if err != nil {
		return uuid.Nil, err
	}

	args := pgstore.CreateUserParams{
		UserName:     request.UserName,
		Email:        request.Email,
		PasswordHash: hash,
		Bio:          request.Bio,
	}

	id, err := us.queries.CreateUser(ctx, args)

	if err != nil {
		pgErr, _ := errors.AsType[*pgconn.PgError](err)

		if pgErr.Code == "23505" {
			return uuid.Nil, ErrDuplicatedEmailOrPassword
		}

		return uuid.Nil, err
	}

	return id, nil
}

func (us *UserService) AuthenticateUser(ctx context.Context, request users.LoginUserRequest) (uuid.UUID, error) {
	user, err := us.queries.GetUserByEmail(ctx, request.Email)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrInvalidUserOrPassword
		}

		return uuid.Nil, err
	}

	if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(request.Password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return uuid.Nil, ErrInvalidUserOrPassword
		}

		return uuid.Nil, err
	}

	return user.ID, nil
}
