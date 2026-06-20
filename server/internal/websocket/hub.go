package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/Sqwerzyyy/pvparena/internal/matchmaking"
)

type Hub struct {
	clients    map[string]*Client
	mu         sync.RWMutex
	matchmaker *matchmaking.Matchmaker
}

type OutgoingMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func NewHub(matchmaker *matchmaking.Matchmaker) *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		matchmaker: matchmaker,
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client.ID] = client
	log.Printf("Hub: registered client %s (total: %d)", client.ID, len(h.clients))
}

func (h *Hub) Unregister(clientID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, clientID)
	log.Printf("Hub: unregistered client %s (total: %d)", clientID, len(h.clients))
}

func (h *Hub) Get(clientID string) (*Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	c, ok := h.clients[clientID]
	return c, ok
}

func (h *Hub) SendTo(clientID string, msgType string, data interface{}) {
	h.mu.RLock()
	client, ok := h.clients[clientID]
	h.mu.RUnlock()

	if !ok {
		log.Printf("Hub: cannot send to %s — not connected", clientID)
		return
	}

	payload, err := json.Marshal(OutgoingMessage{Type: msgType, Data: data})
	if err != nil {
		log.Printf("Hub: failed to marshal message for %s: %v", clientID, err)
		return
	}

	select {
	case client.Send <- payload:
	default:
		log.Printf("Hub: send buffer full for %s, dropping message", clientID)
	}
}

func (h *Hub) BroadcastDuel(duel *matchmaking.Duel) {
	h.SendTo(duel.Player1, "duel_found", duel)
	h.SendTo(duel.Player2, "duel_found", duel)
}

func (h *Hub) TryMatchAndNotify(ctx context.Context) {
	duel, err := h.matchmaker.TryMatch(ctx)
	if err != nil {
		log.Printf("Hub: matchmaking error: %v", err)
		return
	}
	if duel == nil {
		return
	}
	h.BroadcastDuel(duel)
}

func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
