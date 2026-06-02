package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"nhooyr.io/websocket"
)

type WSClient struct {
	baseURL string
	conn    *websocket.Conn
	cancel  context.CancelFunc
}

func NewWSClient(baseURL string) *WSClient {
	return &WSClient{baseURL: strings.TrimRight(baseURL, "/")}
}

func (w *WSClient) SetBaseURL(baseURL string) {
	w.baseURL = strings.TrimRight(baseURL, "/")
}

func (w *WSClient) Subscribe(event string, sink chan<- string) error {
	_ = w.Close()

	u, err := url.Parse(w.baseURL)
	if err != nil {
		return err
	}
	q := u.Query()
	q.Set("event", event)
	u.RawQuery = q.Encode()

	ctx, cancel := context.WithCancel(context.Background())
	conn, _, err := websocket.Dial(ctx, u.String(), nil)
	if err != nil {
		cancel()
		return err
	}
	w.conn = conn
	w.cancel = cancel

	go func() {
		for {
			rCtx, rCancel := context.WithTimeout(ctx, 90*time.Second)
			_, message, readErr := conn.Read(rCtx)
			rCancel()
			if readErr != nil {
				select {
				case sink <- fmt.Sprintf("websocket closed: %v", readErr):
				default:
				}
				return
			}
			select {
			case sink <- string(message):
			default:
			}
		}
	}()

	return nil
}

func (w *WSClient) Close() error {
	if w.cancel != nil {
		w.cancel()
		w.cancel = nil
	}
	if w.conn != nil {
		err := w.conn.Close(websocket.StatusNormalClosure, "bye")
		w.conn = nil
		return err
	}
	return nil
}
