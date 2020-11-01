package auth_test

import (
	"testing"

	"github.com/Meat-Hook/back-template/internal/modules/session/internal/app"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/auth"
	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
)

func TestAuth_TokenAndSubject(t *testing.T) {
	t.Parallel()

	r := require.New(t)
	a := auth.New("super-duper-secret-key-qwertyuio")

	subject := app.Subject{SessionID: xid.New().String()}
	appToken, err := a.Token(subject)
	r.Nil(err)
	r.NotNil(appToken)

	res, err := a.Subject(appToken.Value)
	r.Nil(err)
	r.Equal(subject.SessionID, res.SessionID)
}
