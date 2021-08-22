package web_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/swag"
	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/Meat-Hook/back-template/cmd/file/internal/api/web"
	"github.com/Meat-Hook/back-template/cmd/file/internal/api/web/generated/client"
	"github.com/Meat-Hook/back-template/cmd/file/internal/api/web/generated/client/operations"
	"github.com/Meat-Hook/back-template/cmd/file/internal/api/web/generated/models"
	"github.com/Meat-Hook/back-template/cmd/file/internal/api/web/generated/restapi"
	"github.com/Meat-Hook/back-template/libs/metrics"
	libweb "github.com/Meat-Hook/back-template/libs/web"
)

var (
	reg = prometheus.NewPedanticRegistry()
)

const testFile = `test.jpg`

func TestMain(m *testing.M) {
	metrics.InitMetrics(reg)

	os.Exit(m.Run())
}

func start(t *testing.T) (string, *Mockapplication, *client.FileService, *require.Assertions) {
	t.Helper()

	ctrl := gomock.NewController(t)
	mockApp := NewMockapplication(ctrl)
	assert := require.New(t)

	log := zerolog.New(os.Stdout).With().Caller().Timestamp().Logger()
	webMetric := libweb.NewMetric(reg, strings.Replace(t.Name(), "/", "_", -1), restapi.FlatSwaggerJSON)
	server, err := web.New(log.WithContext(context.Background()), mockApp, &webMetric, web.Config{})
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
	transport.Consumers["image/jpeg"] = runtime.ByteStreamConsumer()
	transport.Consumers["image/png"] = runtime.ByteStreamConsumer()
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
	case *operations.GetFileDefault:
		return err.Payload
	default:
		return nil
	}
}
