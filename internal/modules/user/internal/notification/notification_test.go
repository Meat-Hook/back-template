package notification_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Meat-Hook/back-template/internal/modules/user/internal/app"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/notification"
	"github.com/nats-io/nats.go"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
)

func TestClient_Send(t *testing.T) {
	t.Parallel()

	conn, assert := start(t)
	defer conn.Close()

	client := notification.New(conn)

	var (
		contact = "email@mail.com"
		msg     = app.Message{
			Kind:    app.Welcome,
			Content: "Hello World!",
		}
	)

	ch := make(chan *nats.Msg, 1)
	sub, err := conn.ChanSubscribe("notification", ch)
	assert.Nil(err)

	err = client.Send(context.Background(), contact, msg)
	assert.Nil(err)
	res := <-ch

	m := make(map[string]string)
	err = json.Unmarshal(res.Data, &m)
	assert.Nil(err)

	assert.Equal(map[string]string{
		"contact": contact,
		"kind":    msg.Kind.String(),
		"content": msg.Content,
	}, m)

	assert.Nil(sub.Unsubscribe())
	assert.Nil(sub.Drain())
}

func start(t *testing.T) (*nats.Conn, *require.Assertions) {
	r := require.New(t)
	pool, err := dockertest.NewPool("")
	r.Nil(err)

	resource, err := pool.Run("nats", "2.1.4", nil)
	r.Nil(err)

	var conn *nats.Conn
	if err := pool.Retry(func() error {
		url := fmt.Sprintf("nats://127.0.0.1:%s", resource.GetPort("4222/tcp"))
		conn, err = nats.Connect(url)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		r.Nil(err)
	}

	t.Cleanup(func() {
		err = pool.Purge(resource)
		r.Nil(err)
	})

	return conn, r
}
