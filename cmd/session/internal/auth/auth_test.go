package auth_test

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Meat-Hook/back-template/cmd/session/internal/app"
	"github.com/Meat-Hook/back-template/cmd/session/internal/auth"
)

func TestAuth_TokenAndSubject(t *testing.T) {
	t.Parallel()

	assert := require.New(t)
	a := auth.New("super-duper-secret-key-qwertyuio")

	subject := app.Subject{SessionID: uuid.Must(uuid.NewV4())}
	appToken, err := a.Token(subject)
	assert.NoError(err)
	assert.NotNil(appToken)

	res, err := a.Subject(appToken.Value)
	assert.NoError(err)
	assert.Equal(&subject, res)
}
