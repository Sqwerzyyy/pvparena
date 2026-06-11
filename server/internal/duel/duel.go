package duel

import (
	"fmt"
	"sync"
	"time"
)

type Status string
type RaiseStatus string

const (
	StatusWaiting  Status = "WAITING"
	StatusActive   Status = "ACTIVE"
	StatusFinished Status = "FINISHED"

	RaisePending  RaiseStatus = "PENDING"
	RaiseAccepted RaiseStatus = "ACCEPTED"
	RaiseFolded   RaiseStatus = "FOLDED"
	RaiseExpired  RaiseStatus = "EXPIRED"

	RaiseTimeout = 15 * time.Second
)

type Raise struct {
	FromPlayer string
	Amount     int
	Status     RaiseStatus
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

type Duel struct {
	ID          string
	Player1     string
	Player1Side string
	Player2     string
	Player2Side string
	Credits     int
	Duration    time.Duration
	Threshold   float64
	Status      Status
	StartPrice  float64
	EndPrice    float64
	Winner      string
	ActiveRaise *Raise
	StartsAt    time.Time
	EndsAt      time.Time
	mu          sync.Mutex
}

type Result struct {
	Winner  string
	Loser   string
	Credits int
	Reason  string
}

func NewDuel(id, p1, p1side, p2, p2side string, credits int, duration time.Duration, startPrice float64) *Duel {
	now := time.Now()
	return &Duel{
		ID:          id,
		Player1:     p1,
		Player1Side: p1side,
		Player2:     p2,
		Player2Side: p2side,
		Credits:     credits,
		Duration:    duration,
		Threshold:   0.5,
		Status:      StatusActive,
		StartPrice:  startPrice,
		StartsAt:    now,
		EndsAt:      now.Add(duration),
	}
}

func (d *Duel) MakeRaise(fromPlayer string, amount int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.Status != StatusActive {
		return fmt.Errorf("duel is not active")
	}

	if d.ActiveRaise != nil && d.ActiveRaise.Status == RaisePending {
		return fmt.Errorf("raise already pending")
	}

	if fromPlayer != d.Player1 && fromPlayer != d.Player2 {
		return fmt.Errorf("player not in this duel")
	}

	now := time.Now()
	d.ActiveRaise = &Raise{
		FromPlayer: fromPlayer,
		Amount:     amount,
		Status:     RaisePending,
		CreatedAt:  now,
		ExpiresAt:  now.Add(RaiseTimeout),
	}

	fmt.Printf("Raise: %s raised by %d — opponent has 15 seconds\n", fromPlayer, amount)

	go d.watchRaiseTimeout()

	return nil
}

func (d *Duel) RespondRaise(fromPlayer string, accept bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.ActiveRaise == nil || d.ActiveRaise.Status != RaisePending {
		return fmt.Errorf("no pending raise")
	}

	if fromPlayer == d.ActiveRaise.FromPlayer {
		return fmt.Errorf("cannot respond to your own raise")
	}

	if accept {
		d.ActiveRaise.Status = RaiseAccepted
		d.Credits += d.ActiveRaise.Amount
		fmt.Printf("Raise accepted — new pot: %d credits\n", d.Credits)
	} else {
		d.ActiveRaise.Status = RaiseFolded
		d.Status = StatusFinished
		d.Winner = d.ActiveRaise.FromPlayer
		fmt.Printf("Raise folded — %s wins %d credits\n", d.Winner, d.Credits)
	}

	return nil
}

func (d *Duel) watchRaiseTimeout() {
	time.Sleep(RaiseTimeout)

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.ActiveRaise == nil || d.ActiveRaise.Status != RaisePending {
		return
	}

	d.ActiveRaise.Status = RaiseExpired
	d.Status = StatusFinished
	d.Winner = d.ActiveRaise.FromPlayer
	fmt.Printf("Raise expired — %s wins by timeout %d credits\n", d.Winner, d.Credits)
}

func (d *Duel) Finish(endPrice float64) *Result {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.Status == StatusFinished {
		return &Result{Winner: d.Winner, Credits: d.Credits, Reason: "fold or raise timeout"}
	}

	d.Status = StatusFinished
	d.EndPrice = endPrice

	priceDiff := ((endPrice - d.StartPrice) / d.StartPrice) * 100

	winner := d.determineWinner(priceDiff)
	d.Winner = winner

	loser := d.Player2
	if winner == d.Player2 {
		loser = d.Player1
	}

	fmt.Printf("Duel finished — price diff: %.2f%% — winner: %s\n", priceDiff, winner)

	return &Result{
		Winner:  winner,
		Loser:   loser,
		Credits: d.Credits,
		Reason:  fmt.Sprintf("price diff %.2f%%", priceDiff),
	}
}

func (d *Duel) determineWinner(priceDiff float64) string {
	p1wins := false

	switch d.Player1Side {
	case "UP":
		if d.Player2Side == "DOWN" {
			p1wins = priceDiff > 0
		} else if d.Player2Side == "FLAT" {
			p1wins = priceDiff > d.Threshold
		}
	case "DOWN":
		if d.Player2Side == "UP" {
			p1wins = priceDiff < 0
		} else if d.Player2Side == "FLAT" {
			p1wins = priceDiff < -d.Threshold
		}
	case "FLAT":
		if d.Player2Side == "UP" {
			p1wins = priceDiff <= d.Threshold
		} else if d.Player2Side == "DOWN" {
			p1wins = priceDiff >= -d.Threshold
		}
	}

	if p1wins {
		return d.Player1
	}
	return d.Player2
}
