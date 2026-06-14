package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DuelRecord struct {
	ID          string
	Player1ID   string
	Player2ID   *string
	Asset       string
	Timeframe   int
	CreditsBet  int
	Status      string
	DirectionP1 *string
	DirectionP2 *string
	PriceOpen   *float64
	PriceClose  *float64
	WinnerID    *string
	StartsAt    *time.Time
	EndsAt      *time.Time
	CreatedAt   time.Time
}

type DuelRepository struct {
	pool *pgxpool.Pool
}

func NewDuelRepository(pool *pgxpool.Pool) *DuelRepository {
	return &DuelRepository{pool: pool}
}

func (r *DuelRepository) Create(ctx context.Context, player1ID, asset, directionP1 string, timeframe, credits int) (*DuelRecord, error) {
	query := `
		INSERT INTO duels (player1_id, asset, timeframe, credits_bet, status, direction_p1)
		VALUES ($1, $2, $3, $4, 'waiting', $5)
		RETURNING id, player1_id, player2_id, asset, timeframe, credits_bet, status,
		          direction_p1, direction_p2, price_open, price_close, winner_id,
		          starts_at, ends_at, created_at`

	d := &DuelRecord{}
	err := r.pool.QueryRow(ctx, query, player1ID, asset, timeframe, credits, directionP1).Scan(
		&d.ID, &d.Player1ID, &d.Player2ID, &d.Asset, &d.Timeframe, &d.CreditsBet,
		&d.Status, &d.DirectionP1, &d.DirectionP2, &d.PriceOpen, &d.PriceClose,
		&d.WinnerID, &d.StartsAt, &d.EndsAt, &d.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (r *DuelRepository) Join(ctx context.Context, duelID, player2ID, directionP2 string, priceOpen float64) error {
	now := time.Now()
	query := `
		UPDATE duels
		SET player2_id   = $1,
		    direction_p2 = $2,
		    status       = 'active',
		    price_open   = $3,
		    starts_at    = $4,
		    ends_at      = $5
		WHERE id = $6 AND status = 'waiting'`

	tag, err := r.pool.Exec(ctx, query, player2ID, directionP2, priceOpen, now, now.Add(time.Duration(1)*time.Minute), duelID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrDuelNotAvailable
	}
	return nil
}

func (r *DuelRepository) Finish(ctx context.Context, duelID, winnerID string, priceClose float64) error {
	query := `
		UPDATE duels
		SET status      = 'finished',
		    winner_id   = $1,
		    price_close = $2
		WHERE id = $3 AND status = 'active'`

	_, err := r.pool.Exec(ctx, query, winnerID, priceClose, duelID)
	return err
}

func (r *DuelRepository) GetByID(ctx context.Context, duelID string) (*DuelRecord, error) {
	query := `
		SELECT id, player1_id, player2_id, asset, timeframe, credits_bet, status,
		       direction_p1, direction_p2, price_open, price_close, winner_id,
		       starts_at, ends_at, created_at
		FROM duels
		WHERE id = $1`

	d := &DuelRecord{}
	err := r.pool.QueryRow(ctx, query, duelID).Scan(
		&d.ID, &d.Player1ID, &d.Player2ID, &d.Asset, &d.Timeframe, &d.CreditsBet,
		&d.Status, &d.DirectionP1, &d.DirectionP2, &d.PriceOpen, &d.PriceClose,
		&d.WinnerID, &d.StartsAt, &d.EndsAt, &d.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (r *DuelRepository) Cancel(ctx context.Context, duelID string) error {
	query := `UPDATE duels SET status = 'cancelled' WHERE id = $1 AND status = 'waiting'`
	_, err := r.pool.Exec(ctx, query, duelID)
	return err
}
