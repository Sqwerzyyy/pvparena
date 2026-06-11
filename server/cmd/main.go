package main

import (
	"fmt"
	"log"
	"net/http"

	ws "github.com/Sqwerzyyy/pvparena/internal/websocket"
)

func main() {
	fmt.Println("DARX server starting...")

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		client := ws.HandleWS(w, r)
		if client == nil {
			return
		}
		log.Printf("Player connected: %s", client.ID)

		go client.WritePump()
		go client.ReadPump(func(clientID string, msg []byte) {
			log.Printf("Message from %s: %s", clientID, string(msg))
			client.Send <- []byte("got your message: " + string(msg))
		})
	})

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
