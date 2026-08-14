package ws

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// TokenValidator resolves a bearer token to a user id. It is the same
// credential the REST API takes — a JWT, an API key or an OAuth token.
type TokenValidator interface {
	UserIDForToken(ctx context.Context, token string) string
}

// HandleWebSocket upgrades to a WebSocket and registers the connection.
//
// With auth enabled the connection must carry a token, and it arrives as a
// query parameter rather than a header: browsers cannot set headers on a
// WebSocket handshake. That is the same accommodation the SSE transport makes.
func HandleWebSocket(hub *Hub, validator TokenValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := ""
		if hub.AuthEnabled() {
			token := bearerToken(r)
			if token == "" || validator == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if userID = validator.UserIDForToken(r.Context(), token); userID == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		client := NewClient(hub, conn, userID)
		hub.Register(client)

		go client.WritePump()
		go client.ReadPump()
	}
}

func bearerToken(r *http.Request) string {
	if h := r.Header.Get("Authorization"); len(h) > 7 && h[:7] == "Bearer " {
		return h[7:]
	}
	return r.URL.Query().Get("token")
}
