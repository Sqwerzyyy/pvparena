package matchmaking

import (
	"fmt"
	"sync"
	"time"
)

type Side string
type Duration int

const (
	SideUp   Side = "UP"
	SideDown Side = "DOWN"
	SideFlat Side = "FLAT"
)

const (
	Duration1m Duration = 1
	Duration3m Duration = 3
	Duration5m Duration = 5
)

type QueueEntry struct {
	PlayerID string
	Side     Side
	Duration Duration
	Credits  int
	JoinedAt time.Time
}

type Duel struct {
	ID          string
	Player1     string
	Player1Side Side
	Player2     string
	Player2Side Side
	Credits     int
	Duration    Duration
	Threshold   float64
	StartsAt    time.Time
}

type QueueKey struct {
	Side     Side
	Duration Duration
}

type Matchmaker struct {
	queues map[QueueKey][]*QueueEntry
	mu     sync.Mutex
}

func NewMatchmaker() *Matchmaker {
	return &Matchmaker{
		queues: make(map[QueueKey][]*QueueEntry),
	}
}

func (m *Matchmaker) Join(playerID string, side Side, duration Duration, credits int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry := &QueueEntry{
		PlayerID: playerID,
		Side:     side,
		Duration: duration,
		Credits:  credits,
		JoinedAt: time.Now(),
	}

	key := QueueKey{Side: side, Duration: duration}
	m.queues[key] = append(m.queues[key], entry)

	fmt.Printf("Player %s joined queue: %s %dm for %d credits\n",
		playerID, side, duration, credits)
}

func (m *Matchmaker) TryMatch() *Duel {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, duration := range []Duration{Duration1m, Duration3m, Duration5m} {
		pairs := [][]Side{
			{SideUp, SideDown},
			{SideUp, SideFlat},
			{SideDown, SideFlat},
		}

		for _, pair := range pairs {
			k1 := QueueKey{Side: pair[0], Duration: duration}
			k2 := QueueKey{Side: pair[1], Duration: duration}

			if len(m.queues[k1]) > 0 && len(m.queues[k2]) > 0 {
				p1 := m.pop(k1)
				p2 := m.pop(k2)
				return m.createDuel(p1, p2)
			}
		}
	}

	return nil
}

func (m *Matchmaker) createDuel(p1, p2 *QueueEntry) *Duel {
	duel := &Duel{
		ID:          fmt.Sprintf("duel_%d", time.Now().UnixNano()),
		Player1:     p1.PlayerID,
		Player1Side: p1.Side,
		Player2:     p2.PlayerID,
		Player2Side: p2.Side,
		Credits:     minInt(p1.Credits, p2.Credits),
		Duration:    p1.Duration,
		Threshold:   0.5,
		StartsAt:    time.Now(),
	}

	fmt.Printf("Duel created: %s(%s) vs %s(%s) — %dm for %d credits\n",
		p1.PlayerID, p1.Side, p2.PlayerID, p2.Side, duel.Duration, duel.Credits)

	return duel
}

func (m *Matchmaker) pop(key QueueKey) *QueueEntry {
	e := m.queues[key][0]
	m.queues[key] = m.queues[key][1:]
	return e
}

func (m *Matchmaker) QueueSizes() map[QueueKey]int {
	m.mu.Lock()
	defer m.mu.Unlock()

	sizes := make(map[QueueKey]int)
	for k, v := range m.queues {
		sizes[k] = len(v)
	}
	return sizes
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
