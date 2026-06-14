package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID             string
	WalletAddress  string
	Username       *string
	AvatarURL      *string
	Credits        int
	TotalWins      int
	TotalLosses    int
	Rating         int
	CreatedAt      time.Time
}

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetOrCreate(ctx context.Context, walletAddress string) (*User, error) {
	user, err := r.GetByWallet(ctx, walletAddress)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	return r.Create(ctx, walletAddress)
}

func (r *UserRepository) Create(ctx context.Context, walletAddress string) (*User, error) {
	query := `
		INSERT INTO users (wallet_address)
		VALUES ($1)
		RETURNING id, wallet_address, username, avatar_url, credits, total_wins, total_losses, rating, created_at`

	user := &User{}
	err := r.pool.QueryRow(ctx, query, walletAddress).Scan(
		&user.ID,
		&user.WalletAddress,
		&user.Username,
		&user.AvatarURL,
		&user.Credits,
		&user.TotalWins,
		&user.TotalLosses,
		&user.Rating,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByWallet(ctx context.Context, walletAddress string) (*User, error) {
	query := `
		SELECT id, wallet_address, username, avatar_url, credits, total_wins, total_losses, rating, created_at
		FROM users
		WHERE wallet_address = $1`

	user := &User{}
	err := r.pool.QueryRow(ctx, query, walletAddress).Scan(
		&user.ID,
		&user.WalletAddress,
		&user.Username,
		&user.AvatarURL,
		&user.Credits,
		&user.TotalWins,
		&user.TotalLosses,
		&user.Rating,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, wallet_address, username, avatar_url, credits, total_wins, total_losses, rating, created_at
		FROM users
		WHERE id = $1`

	user := &User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.WalletAddress,
		&user.Username,
		&user.AvatarURL,
		&user.Credits,
		&user.TotalWins,
		&user.TotalLosses,
		&user.Rating,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) AddCredits(ctx context.Context, userID string, amount int) error {
	query := `UPDATE users SET credits = credits + $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, amount, userID)
	return err
}

func (r *UserRepository) DeductCredits(ctx context.Context, userID string, amount int) error {
	query := `UPDATE users SET credits = credits - $1 WHERE id = $2 AND credits >= $1`
	tag, err := r.pool.Exec(ctx, query, amount, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("insufficient credits")
	}
	return nil
}

func (r *UserRepository) RecordWin(ctx context.Context, userID string) error {
	query := `UPDATE users SET total_wins = total_wins + 1, rating = rating + 25 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *UserRepository) RecordLoss(ctx context.Context, userID string) error {
	query := `UPDATE users SET total_losses = total_losses + 1, rating = GREATEST(0, rating - 20) WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}
