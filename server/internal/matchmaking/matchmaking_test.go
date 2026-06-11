package matchmaking

import (
	"testing"
)

func TestMatchUpVsDown3m(t *testing.T) {
	mm := NewMatchmaker()

	mm.Join("player1", SideUp, Duration3m, 100)
	mm.Join("player2", SideDown, Duration3m, 100)

	duel := mm.TryMatch()
	if duel == nil {
		t.Fatal("expected duel, got nil")
	}

	if duel.Player1Side != SideUp || duel.Player2Side != SideDown {
		t.Fatalf("wrong sides: %s vs %s", duel.Player1Side, duel.Player2Side)
	}

	if duel.Duration != Duration3m {
		t.Fatalf("wrong duration: %d", duel.Duration)
	}

	t.Logf("Duel: %s(%s) vs %s(%s) — %dm", duel.Player1, duel.Player1Side, duel.Player2, duel.Player2Side, duel.Duration)
}

func TestMatchUpVsDown1m(t *testing.T) {
	mm := NewMatchmaker()

	mm.Join("player1", SideUp, Duration1m, 50)
	mm.Join("player2", SideDown, Duration1m, 50)

	duel := mm.TryMatch()
	if duel == nil {
		t.Fatal("expected duel, got nil")
	}

	if duel.Duration != Duration1m {
		t.Fatalf("wrong duration: %d", duel.Duration)
	}

	t.Logf("Duel: %s(%s) vs %s(%s) — %dm", duel.Player1, duel.Player1Side, duel.Player2, duel.Player2Side, duel.Duration)
}

func TestMatchUpVsFlat5m(t *testing.T) {
	mm := NewMatchmaker()

	mm.Join("player1", SideUp, Duration5m, 200)
	mm.Join("player2", SideFlat, Duration5m, 200)

	duel := mm.TryMatch()
	if duel == nil {
		t.Fatal("expected duel, got nil")
	}

	if duel.Player1Side != SideUp || duel.Player2Side != SideFlat {
		t.Fatalf("wrong sides: %s vs %s", duel.Player1Side, duel.Player2Side)
	}

	t.Logf("Duel: %s(%s) vs %s(%s) — %dm", duel.Player1, duel.Player1Side, duel.Player2, duel.Player2Side, duel.Duration)
}

func TestMatchDownVsFlat3m(t *testing.T) {
	mm := NewMatchmaker()

	mm.Join("player1", SideDown, Duration3m, 150)
	mm.Join("player2", SideFlat, Duration3m, 300)

	duel := mm.TryMatch()
	if duel == nil {
		t.Fatal("expected duel, got nil")
	}

	if duel.Credits != 150 {
		t.Fatalf("expected min credits 150, got %d", duel.Credits)
	}

	t.Logf("Duel: %s(%s) vs %s(%s) — credits: %d", duel.Player1, duel.Player1Side, duel.Player2, duel.Player2Side, duel.Credits)
}

func TestNoMatchDifferentDuration(t *testing.T) {
	mm := NewMatchmaker()

	mm.Join("player1", SideUp, Duration1m, 100)
	mm.Join("player2", SideDown, Duration3m, 100)

	duel := mm.TryMatch()
	if duel != nil {
		t.Fatal("expected nil — different durations should not match")
	}

	t.Log("Correctly returned nil for different durations")
}

func TestNoMatchSameSide(t *testing.T) {
	mm := NewMatchmaker()

	mm.Join("player1", SideUp, Duration3m, 100)
	mm.Join("player2", SideUp, Duration3m, 100)

	duel := mm.TryMatch()
	if duel != nil {
		t.Fatal("expected nil — same side should not match")
	}

	t.Log("Correctly returned nil for same side")
}

func TestPriorityOneMInute(t *testing.T) {
	mm := NewMatchmaker()

	mm.Join("player1", SideUp, Duration1m, 100)
	mm.Join("player2", SideDown, Duration1m, 100)
	mm.Join("player3", SideUp, Duration3m, 100)
	mm.Join("player4", SideDown, Duration3m, 100)

	duel := mm.TryMatch()
	if duel == nil {
		t.Fatal("expected duel, got nil")
	}

	if duel.Duration != Duration1m {
		t.Fatalf("expected 1m duel first, got %dm", duel.Duration)
	}

	t.Logf("Correctly matched 1m duel first: %s vs %s", duel.Player1, duel.Player2)
}
