package web_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/Meat-Hook/back-template/internal/libs/metrics"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/web"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/web/generated/client"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/web/generated/models"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/web/generated/restapi/operations"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/swag"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	metrics.InitMetrics()

	os.Exit(m.Run())
}

func start(t *testing.T) (string, *Mockapplication, *client.ServiceUser) {
	t.Helper()

	ctrl := gomock.NewController(t)
	mockApp := NewMockapplication(ctrl)

	log := zerolog.New(os.Stdout)
	server, err := web.New(mockApp, log, web.Config{})
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

	return url, mockApp, c
}

// APIError returns model.Error with given msg.
func APIError(msg string) *models.Error {
	return &models.Error{
		Message: swag.String(msg),
	}
}

func errPayload(err interface{}) *models.Error {
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
	default:
		return nil
	}
}
