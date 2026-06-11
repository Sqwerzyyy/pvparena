package duel

import (
	"testing"
	"time"
)

func TestDuelUpWins(t *testing.T) {
	d := NewDuel("test1", "player1", "UP", "player2", "DOWN", 100, 3*time.Minute, 100000.0)

	result := d.Finish(101000.0)
	if result.Winner != "player1" {
		t.Fatalf("expected player1 to win, got %s", result.Winner)
	}

	t.Logf("Winner: %s — %s", result.Winner, result.Reason)
}

func TestDuelDownWins(t *testing.T) {
	d := NewDuel("test2", "player1", "UP", "player2", "DOWN", 100, 3*time.Minute, 100000.0)

	result := d.Finish(99000.0)
	if result.Winner != "player2" {
		t.Fatalf("expected player2 to win, got %s", result.Winner)
	}

	t.Logf("Winner: %s — %s", result.Winner, result.Reason)
}

func TestDuelFlatWins(t *testing.T) {
	d := NewDuel("test3", "player1", "UP", "player2", "FLAT", 100, 3*time.Minute, 100000.0)

	result := d.Finish(100100.0)
	if result.Winner != "player2" {
		t.Fatalf("expected player2 (FLAT) to win, got %s", result.Winner)
	}

	t.Logf("Winner: %s — %s", result.Winner, result.Reason)
}

func TestDuelFold(t *testing.T) {
	d := NewDuel("test4", "player1", "UP", "player2", "DOWN", 100, 3*time.Minute, 100000.0)

	d.MakeRaise("player1", 50)
	d.RespondRaise("player2", false)

	if d.Winner != "player1" {
		t.Fatalf("expected player1 to win after fold, got %s", d.Winner)
	}

	t.Logf("Winner after fold: %s", d.Winner)
}

func TestDuelRaiseAccepted(t *testing.T) {
	d := NewDuel("test5", "player1", "UP", "player2", "DOWN", 100, 3*time.Minute, 100000.0)

	d.MakeRaise("player1", 50)
	d.RespondRaise("player2", true)

	if d.Credits != 150 {
		t.Fatalf("expected 150 credits after raise, got %d", d.Credits)
	}

	t.Logf("Credits after raise: %d", d.Credits)
}
