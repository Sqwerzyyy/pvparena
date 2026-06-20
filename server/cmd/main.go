package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Sqwerzyyy/pvparena/internal/matchmaking"
	"github.com/Sqwerzyyy/pvparena/internal/repository"
	ws "github.com/Sqwerzyyy/pvparena/internal/websocket"
)

func main() {
	fmt.Println("DARX server starting...")

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://darx:darx_dev_pass@localhost:5433/darx?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("postgres ping failed: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	userRepo := repository.NewUserRepository(pool)
	duelRepo := repository.NewDuelRepository(pool)

	matchmaker := matchmaking.NewMatchmaker(duelRepo)
	hub := ws.NewHub(matchmaker)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		client := ws.HandleWS(w, r)
		if client == nil {
			return
		}

		user, err := userRepo.GetOrCreate(context.Background(), client.ID)
		if err != nil {
			log.Printf("Failed to get/create user %s: %v", client.ID, err)
			return
		}
		log.Printf("Player connected: %s | credits: %d | rating: %d", user.WalletAddress, user.Credits, user.Rating)

		hub.Register(client)
		defer hub.Unregister(client.ID)

		go client.WritePump()
		client.ReadPump(func(clientID string, msg []byte) {
			log.Printf("Message from %s: %s", clientID, string(msg))
			client.Send <- []byte("got your message: " + string(msg))
		})
	})

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
