package web_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/swag"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/Meat-Hook/back-template/cmd/user/internal/api/web"
	"github.com/Meat-Hook/back-template/cmd/user/internal/api/web/generated/client"
	"github.com/Meat-Hook/back-template/cmd/user/internal/api/web/generated/client/operations"
	"github.com/Meat-Hook/back-template/cmd/user/internal/api/web/generated/models"
	"github.com/Meat-Hook/back-template/cmd/user/internal/api/web/generated/restapi"
	"github.com/Meat-Hook/back-template/cmd/user/internal/app"
	"github.com/Meat-Hook/back-template/libs/metrics"
	libweb "github.com/Meat-Hook/back-template/libs/web"
)

var (
	errAny = errors.New("any error")

	user = app.User{
		ID:    uuid.Must(uuid.NewV4()),
		Email: "email@email.test",
		Name:  "username",
	}

	session = app.Session{
		ID:     uuid.Must(uuid.NewV4()),
		UserID: user.ID,
	}

	token      = "token"
	apiKeyAuth = httptransport.APIKeyAuth("Cookie", "header", "authKey="+token)

	reg = prometheus.NewPedanticRegistry()
)

func TestMain(m *testing.M) {
	metrics.InitMetrics(reg)

	os.Exit(m.Run())
}

func start(t *testing.T) (string, *Mockapplication, *client.UserService, *require.Assertions) {
	t.Helper()

	ctrl := gomock.NewController(t)
	mockApp := NewMockapplication(ctrl)
	assert := require.New(t)

	logger := zerolog.New(os.Stdout)
	webMetric := libweb.NewMetric(reg, strings.Replace(t.Name(), "/", "_", -1), restapi.FlatSwaggerJSON)
	server, err := web.New(logger.WithContext(context.Background()), mockApp, &webMetric, web.Config{})
	assert.NoError(err, "web.New")
	assert.NoError(server.Listen(), "server.Listen")

	errc := make(chan error, 1)
	go func() { errc <- server.Serve() }()
	t.Cleanup(func() {
		t.Helper()

		assert.NoError(server.Shutdown(), "server.Shutdown")
		assert.NoError(<-errc, "server.Serve")
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
	if err == nil {
		return nil
	}

	switch err := err.(type) {
	case *operations.VerificationEmailDefault:
		return err.Payload
	case *operations.VerificationUsernameDefault:
		return err.Payload
	case *operations.CreateUserDefault:
		return err.Payload
	case *operations.GetUserDefault:
		return err.Payload
	case *operations.DeleteUserDefault:
		return err.Payload
	case *operations.UpdatePasswordDefault:
		return err.Payload
	case *operations.UpdateUsernameDefault:
		return err.Payload
	case *operations.GetUsersDefault:
		return err.Payload
	case *operations.LoginDefault:
		return err.Payload
	case *operations.LogoutDefault:
		return err.Payload
	case *operations.NewAvatarDefault:
		return err.Payload
	case *operations.DeleteAvatarDefault:
		return err.Payload
	default:
		return nil
	}
}
