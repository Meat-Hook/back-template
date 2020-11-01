// Package notification needed for provide message to nats.
package notification

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Meat-Hook/back-template/internal/modules/user/internal/app"
	"github.com/nats-io/nats.go"
)

var _ app.Notification = &Client{}

// Client to nats connection.
type Client struct {
	conn *nats.Conn
}

// New build and returns new Client.
func New(conn *nats.Conn) *Client {
	return &Client{conn: conn}
}

type message struct {
	Contact string `json:"contact"`
	Kind    string `json:"kind"`
	Content string `json:"content"`
}

const subj = `notification`

// Send for implements app.Notification.
func (c *Client) Send(_ context.Context, contact string, msg app.Message) error {
	m := message{
		Contact: contact,
		Kind:    msg.Kind.String(),
		Content: msg.Content,
	}

	buf, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	err = c.conn.Publish(subj, buf)
	if err != nil {
		return fmt.Errorf("publish message: %w", err)
	}

	return nil
}
