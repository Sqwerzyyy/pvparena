package matchmaking

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/Sqwerzyyy/pvparena/internal/repository"
)

// fakeDuelStore is an in-memory implementation of DuelStore for tests.
type fakeDuelStore struct {
	counter atomic.Int64
}

func (f *fakeDuelStore) Create(_ context.Context, _, _, _ string, _ int, _ int) (*repository.DuelRecord, error) {
	id := f.counter.Add(1)
	return &repository.DuelRecord{ID: fmt.Sprintf("duel-%d", id)}, nil
}

func (f *fakeDuelStore) Join(_ context.Context, _, _, _ string, _ float64) error {
	return nil
}

func newFake() DuelStore { return &fakeDuelStore{} }

func TestMatchUpVsDown3m(t *testing.T) {
	mm := NewMatchmaker(newFake())

	mm.Join("player1", "BTC", SideUp, Duration3m, 100)
	mm.Join("player2", "BTC", SideDown, Duration3m, 100)

	duel, err := mm.TryMatch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if duel == nil {
		t.Fatal("expected duel, got nil")
	}

	if duel.Player1Side != SideUp || duel.Player2Side != SideDown {
		t.Fatalf("wrong sides: %s vs %s", duel.Player1Side, duel.Player2Side)
	}

	if duel.Duration != Duration3m {
		t.Fatalf("wrong duration: %d", duel.Duration)
	}

	if duel.Asset != "BTC" {
		t.Fatalf("wrong asset: %s", duel.Asset)
	}

	t.Logf("Duel: %s(%s) vs %s(%s) — %s %dm", duel.Player1, duel.Player1Side, duel.Player2, duel.Player2Side, duel.Asset, duel.Duration)
}

func TestMatchUpVsDown1m(t *testing.T) {
	mm := NewMatchmaker(newFake())

	mm.Join("player1", "ETH", SideUp, Duration1m, 50)
	mm.Join("player2", "ETH", SideDown, Duration1m, 50)

	duel, err := mm.TryMatch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if duel == nil {
		t.Fatal("expected duel, got nil")
	}

	if duel.Duration != Duration1m {
		t.Fatalf("wrong duration: %d", duel.Duration)
	}

	t.Logf("Duel: %s(%s) vs %s(%s) — %dm", duel.Player1, duel.Player1Side, duel.Player2, duel.Player2Side, duel.Duration)
}

func TestMatchUpVsFlat5m(t *testing.T) {
	mm := NewMatchmaker(newFake())

	mm.Join("player1", "SOL", SideUp, Duration5m, 200)
	mm.Join("player2", "SOL", SideFlat, Duration5m, 200)

	duel, err := mm.TryMatch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if duel == nil {
		t.Fatal("expected duel, got nil")
	}

	if duel.Player1Side != SideUp || duel.Player2Side != SideFlat {
		t.Fatalf("wrong sides: %s vs %s", duel.Player1Side, duel.Player2Side)
	}

	t.Logf("Duel: %s(%s) vs %s(%s) — %dm", duel.Player1, duel.Player1Side, duel.Player2, duel.Player2Side, duel.Duration)
}

func TestMatchDownVsFlat3m(t *testing.T) {
	mm := NewMatchmaker(newFake())

	mm.Join("player1", "BTC", SideDown, Duration3m, 150)
	mm.Join("player2", "BTC", SideFlat, Duration3m, 300)

	duel, err := mm.TryMatch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if duel == nil {
		t.Fatal("expected duel, got nil")
	}

	if duel.Credits != 150 {
		t.Fatalf("expected min credits 150, got %d", duel.Credits)
	}

	t.Logf("Duel: %s(%s) vs %s(%s) — credits: %d", duel.Player1, duel.Player1Side, duel.Player2, duel.Player2Side, duel.Credits)
}

func TestNoMatchDifferentDuration(t *testing.T) {
	mm := NewMatchmaker(newFake())

	mm.Join("player1", "BTC", SideUp, Duration1m, 100)
	mm.Join("player2", "BTC", SideDown, Duration3m, 100)

	duel, err := mm.TryMatch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if duel != nil {
		t.Fatal("expected nil — different durations should not match")
	}

	t.Log("Correctly returned nil for different durations")
}

func TestNoMatchSameSide(t *testing.T) {
	mm := NewMatchmaker(newFake())

	mm.Join("player1", "BTC", SideUp, Duration3m, 100)
	mm.Join("player2", "BTC", SideUp, Duration3m, 100)

	duel, err := mm.TryMatch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if duel != nil {
		t.Fatal("expected nil — same side should not match")
	}

	t.Log("Correctly returned nil for same side")
}

func TestNoMatchDifferentAsset(t *testing.T) {
	mm := NewMatchmaker(newFake())

	mm.Join("player1", "BTC", SideUp, Duration3m, 100)
	mm.Join("player2", "ETH", SideDown, Duration3m, 100)

	duel, err := mm.TryMatch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if duel != nil {
		t.Fatal("expected nil — different assets should not match")
	}

	t.Log("Correctly returned nil for different assets")
}

func TestPriorityOneMinute(t *testing.T) {
	mm := NewMatchmaker(newFake())

	mm.Join("player1", "BTC", SideUp, Duration1m, 100)
	mm.Join("player2", "BTC", SideDown, Duration1m, 100)
	mm.Join("player3", "BTC", SideUp, Duration3m, 100)
	mm.Join("player4", "BTC", SideDown, Duration3m, 100)

	duel, err := mm.TryMatch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if duel == nil {
		t.Fatal("expected duel, got nil")
	}

	if duel.Duration != Duration1m {
		t.Fatalf("expected 1m duel first, got %dm", duel.Duration)
	}

	t.Logf("Correctly matched 1m duel first: %s vs %s", duel.Player1, duel.Player2)
}
