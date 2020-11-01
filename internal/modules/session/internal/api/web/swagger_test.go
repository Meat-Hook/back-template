package web_test

import (
	"net/http"
	"testing"

	"github.com/Meat-Hook/back-template/internal/modules/session/internal/api/web/generated/restapi"
	"github.com/go-openapi/loads"
	"github.com/stretchr/testify/assert"
)

func TestServeSwagger(t *testing.T) {
	t.Parallel()

	url, _, _, _ := start(t)

	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	assert.NoError(t, err)
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
		resp, err := c.Get("http://" + url + tc.path)
		assert.Nil(t, err, tc.path)
		assert.Equal(t, tc.want, resp.StatusCode, tc.path)
	}
}
