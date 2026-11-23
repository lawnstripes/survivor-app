package sse

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Broker struct {
	clients map[chan string]bool
	mutex   sync.Mutex
}

func NewBroker() *Broker {
	return &Broker{
		clients: make(map[chan string]bool),
	}
}

// Stream sends the SSE headers and keeps the connection open
func (b *Broker) Stream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	log.Printf("client connected")

	// Create a new buffered channel for this specific browser tab
	messageChan := make(chan string, 8)

	b.mutex.Lock()
	b.clients[messageChan] = true
	b.mutex.Unlock()

	// Clean up when the user closes the tab
	defer func() {
		b.mutex.Lock()
		delete(b.clients, messageChan)
		b.mutex.Unlock()
		close(messageChan)
	}()

	// Flush initial SSE comment immediately so fetch()-based SSE clients (htmx 4 hx-sse)
	// receive the 200 OK + text/event-stream headers right away.
	fmt.Fprintf(w, ": connected\n\n")
	w.(http.Flusher).Flush()

	for {
		select {
		case msg := <-messageChan:
			log.Printf("sent %s", msg)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", msg, msg)
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// Notify sends a message to all connected clients
func (b *Broker) Notify(msg string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	for client := range b.clients {
		select {
		case client <- msg:
		default:
		}
	}
}
