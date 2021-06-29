package auth_test

import (
	"testing"

	app2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/app"
	auth2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/auth"
	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
)

func TestAuth_TokenAndSubject(t *testing.T) {
	t.Parallel()

	r := require.New(t)
	a := auth2.New("super-duper-secret-key-qwertyuio")

	subject := app2.Subject{SessionID: xid.New().String()}
	appToken, err := a.Token(subject)
	r.Nil(err)
	r.NotNil(appToken)

	res, err := a.Subject(appToken.Value)
	r.Nil(err)
	r.Equal(&subject, res)
}
