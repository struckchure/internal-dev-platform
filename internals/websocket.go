package internals

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

type WebSocketHub struct {
	mu          sync.RWMutex
	subscribers map[string]map[*websocket.Conn]struct{}
	server      *http.Server
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		subscribers: map[string]map[*websocket.Conn]struct{}{},
	}
}

func (h *WebSocketHub) Start(env Env) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", h.handleWebSocket)

	h.server = &http.Server{
		Addr:              fmt.Sprintf(":%s", env.SOCKET_PORT),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("websocket server:", err)
		}
	}()
}

func (h *WebSocketHub) Stop() {
	if h.server == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h.server.Shutdown(ctx); err != nil {
		log.Println("websocket shutdown:", err)
	}
}

func (h *WebSocketHub) Publish(event string, content any) {
	h.mu.RLock()
	conns := h.subscribers[event]
	h.mu.RUnlock()
	if len(conns) == 0 {
		return
	}

	payload, err := json.Marshal(map[string]any{
		"event":   event,
		"content": content,
	})
	if err != nil {
		log.Println("websocket publish marshal:", err)
		return
	}

	for conn := range conns {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err = conn.Write(ctx, websocket.MessageText, payload)
		cancel()
		if err != nil {
			h.remove(event, conn)
			_ = conn.Close(websocket.StatusInternalError, "write failed")
		}
	}
}

func (h *WebSocketHub) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	event := r.URL.Query().Get("event")
	if event == "" {
		http.Error(w, "missing event query param", http.StatusBadRequest)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		log.Println("websocket accept:", err)
		return
	}

	h.add(event, conn)
	defer func() {
		h.remove(event, conn)
		_ = conn.Close(websocket.StatusNormalClosure, "closed")
	}()

	for {
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		_, _, err := conn.Read(ctx)
		cancel()
		if err != nil {
			return
		}
	}
}

func (h *WebSocketHub) add(event string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.subscribers[event] == nil {
		h.subscribers[event] = map[*websocket.Conn]struct{}{}
	}
	h.subscribers[event][conn] = struct{}{}
}

func (h *WebSocketHub) remove(event string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.subscribers[event] == nil {
		return
	}

	delete(h.subscribers[event], conn)
	if len(h.subscribers[event]) == 0 {
		delete(h.subscribers, event)
	}
}
