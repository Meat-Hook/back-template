package web_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Meat-Hook/back-template/internal/microservices/session/internal/api/web/generated/restapi"
	"github.com/go-openapi/loads"
)

func TestServeSwagger(t *testing.T) {
	t.Parallel()

	url, _, _, assert := start(t)

	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	assert.NoError(err)
	basePath := swaggerSpec.BasePath()

	testCases := []struct {
		path string
		want int
	}{
		{"", 404},
		{"/swagger.yml", 404},
		{"/swagger.yaml", 404},
		{"/swagger.json", 200},
		{basePath + "/", 404},
		{basePath + "/docs", 200},
		{basePath + "/swagger.json", 200},
	}

	c := &http.Client{}

	for _, tc := range testCases {
		req, err := http.NewRequestWithContext(context.Background(), "GET", "http://"+url+tc.path, nil)
		assert.Nil(err)
		resp, err := c.Do(req)
		assert.Nil(err, tc.path)
		assert.Equal(tc.want, resp.StatusCode, tc.path)
	}
}
