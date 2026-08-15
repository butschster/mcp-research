package ws

import (
	"time"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMsgSize = 512
	sendBufLen = 64

	// CloseUnauthorized closes a connection whose credential stopped being
	// valid. It is in the application range (4000-4999) precisely so a client
	// can tell it apart from a network drop: one calls for signing in again,
	// the other for reconnecting, and retrying the first forever is what the
	// browser would otherwise do.
	CloseUnauthorized = 4401
)

// Client is a single WebSocket connection, and the person behind it.
//
// userID is fixed for the life of the connection — it is who authenticated —
// but what that person may see is re-checked on every event, so a membership
// taken away lands on the open socket rather than waiting for a reconnect.
//
// token is kept for the same reason in the other direction. Authenticating only
// at the handshake meant a socket outlived the credential that opened it: an
// API key deleted an hour ago still fed events for as long as the tab stayed
// open, because the per-event check asks whether the *user* may read, never
// whether the credential is still good.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID string
	token  string

	// share is set instead of userID for a connection opened from a share link.
	// The two are mutually exclusive by construction: a share visitor has no
	// account, and a signed-in reader is scoped by membership rather than by a
	// token in a URL.
	//
	// It is re-resolved on the same timer as a user's credential, so revoking a
	// link closes the sockets it opened rather than waiting for the tab to be
	// closed.
	share      *auth.Share
	shareToken string
	// shareUnlock is what the visitor got back from the password exchange. It
	// has to be kept for the re-check: without it a protected link would fail
	// its first re-validation and close a connection that is perfectly valid.
	shareUnlock string
}

func NewClient(hub *Hub, conn *websocket.Conn, userID, token string) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, sendBufLen),
		userID: userID,
		token:  token,
	}
}

// NewShareClient is the connection behind a shared page. It carries no user,
// which is what keeps it out of every rule written in terms of one.
func NewShareClient(hub *Hub, conn *websocket.Conn, share *auth.Share, token, unlock string) *Client {
	return &Client{
		hub:         hub,
		conn:        conn,
		send:        make(chan []byte, sendBufLen),
		share:       share,
		shareToken:  token,
		shareUnlock: unlock,
	}
}

// ReadPump reads messages from the client (only pong/close).
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMsgSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

// WritePump writes messages from hub to the WebSocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			// The keepalive is also when the credential is re-examined. Doing
			// it here costs one lookup a minute per connection and means a
			// revoked key stops delivering within that minute, rather than
			// whenever the person happens to close the tab.
			if !c.hub.credentialStillValid(c) {
				c.conn.SetWriteDeadline(time.Now().Add(writeWait))
				c.conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(CloseUnauthorized, "credential no longer valid"))
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
