package web_test

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Meat-Hook/back-template/cmd/session/internal/api/web"
	"github.com/Meat-Hook/back-template/cmd/session/internal/api/web/generated/client"
	"github.com/Meat-Hook/back-template/cmd/session/internal/api/web/generated/client/operations"
	"github.com/Meat-Hook/back-template/cmd/session/internal/api/web/generated/models"
	"github.com/Meat-Hook/back-template/cmd/session/internal/api/web/generated/restapi"
	"github.com/Meat-Hook/back-template/cmd/session/internal/app"
	"github.com/Meat-Hook/back-template/libs/metrics"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/swag"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errAny = errors.New("any error")

	user = app.User{
		ID:    uuid.Must(uuid.NewV4()),
		Email: "email",
		Name:  "username",
	}

	token      = "token"
	apiKeyAuth = httptransport.APIKeyAuth("Cookie", "header", "authKey="+token)
)

func TestMain(m *testing.M) {
	metrics.InitMetrics()

	os.Exit(m.Run())
}

func start(t *testing.T) (string, *Mockapplication, *client.SessionService, *require.Assertions) {
	t.Helper()

	ctrl := gomock.NewController(t)
	mockApp := NewMockapplication(ctrl)

	log := zerolog.New(os.Stdout)
	m := metrics.HTTP(strings.ReplaceAll(t.Name(), "/", "_"), restapi.FlatSwaggerJSON)
	server, err := web.New(mockApp, log, &m, web.Config{})
	assert.NoError(t, err, "web.New")
	assert.NoError(t, server.Listen(), "server.Listen")

	errc := make(chan error, 1)
	go func() { errc <- server.Serve() }()
	t.Cleanup(func() {
		t.Helper()

		assert.Nil(t, server.Shutdown(), "server.Shutdown")
		assert.Nil(t, <-errc, "server.Serve")
		ctrl.Finish()
	})

	url := fmt.Sprintf("%s:%d", client.DefaultHost, server.Port)

	transport := httptransport.New(url, client.DefaultBasePath, client.DefaultSchemes)
	c := client.New(transport, nil)

	return url, mockApp, c, require.New(t)
}

// APIError returns model.Error with given msg.
func APIError(msg string) *models.Error {
	return &models.Error{
		Message: swag.String(msg),
	}
}

func errPayload(err interface{}) *models.Error {
	switch err := err.(type) {
	case *operations.LoginDefault:
		return err.Payload
	case *operations.LogoutDefault:
		return err.Payload
	default:
		return nil
	}
}
