package web_test

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	web2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/api/web"
	client2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/api/web/generated/client"
	operations2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/api/web/generated/client/operations"
	models2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/api/web/generated/models"
	restapi2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/api/web/generated/restapi"
	app2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/app"
	metrics2 "github.com/Meat-Hook/back-template/internal/libs/metrics"
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

	user = app2.User{
		ID:    uuid.Must(uuid.NewV4()),
		Email: "email",
		Name:  "username",
	}

	session = app2.Session{
		ID:     "id",
		UserID: user.ID,
	}

	token      = "token"
	apiKeyAuth = httptransport.APIKeyAuth("Cookie", "header", "authKey="+token)
)

func TestMain(m *testing.M) {
	metrics2.InitMetrics()

	os.Exit(m.Run())
}

func start(t *testing.T) (string, *Mockapplication, *client2.UserService, *require.Assertions) {
	t.Helper()

	ctrl := gomock.NewController(t)
	mockApp := NewMockapplication(ctrl)

	log := zerolog.New(os.Stdout)
	m := metrics2.HTTP(strings.ReplaceAll(t.Name(), "/", "_"), restapi2.FlatSwaggerJSON)
	server, err := web2.New(mockApp, log, &m, web2.Config{})
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

	url := fmt.Sprintf("%s:%d", client2.DefaultHost, server.Port)

	transport := httptransport.New(url, client2.DefaultBasePath, client2.DefaultSchemes)
	c := client2.New(transport, nil)

	return url, mockApp, c, require.New(t)
}

// APIError returns model.Error with given msg.
func APIError(msg string) *models2.Error {
	return &models2.Error{
		Message: swag.String(msg),
	}
}

func errPayload(err interface{}) *models2.Error {
	if err == nil {
		return nil
	}

	switch err := err.(type) {
	case *operations2.VerificationEmailDefault:
		return err.Payload
	case *operations2.VerificationUsernameDefault:
		return err.Payload
	case *operations2.CreateUserDefault:
		return err.Payload
	case *operations2.GetUserDefault:
		return err.Payload
	case *operations2.DeleteUserDefault:
		return err.Payload
	case *operations2.UpdatePasswordDefault:
		return err.Payload
	case *operations2.UpdateUsernameDefault:
		return err.Payload
	case *operations2.GetUsersDefault:
		return err.Payload
	default:
		return nil
	}
}
