package services

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/struckchure/idp/internals"
)

func TestBuildAuthorizeLinkIncludesRedirectAndState(t *testing.T) {
	svc := &GithubService{
		env: internals.Env{
			GH_APP_CLIENT_ID:    "client-id",
			GH_APP_REDIRECT_URL: "http://localhost:3000/api/v1/callback/github/",
		},
	}

	link := svc.buildAuthorizeLink("state-token")
	parsed, err := url.Parse(link)
	require.NoError(t, err)

	q := parsed.Query()
	assert.Equal(t, "client-id", q.Get("client_id"))
	assert.Equal(t, "state-token", q.Get("state"))
	assert.Equal(t, "http://localhost:3000/api/v1/callback/github/", q.Get("redirect_uri"))
}

func TestOAuthConfigUsesRedirectURL(t *testing.T) {
	svc := &GithubService{
		env: internals.Env{
			GH_APP_CLIENT_ID:     "client-id",
			GH_APP_CLIENT_SECRET: "client-secret",
			GH_APP_REDIRECT_URL:  "http://localhost:3000/api/v1/callback/github/",
		},
	}

	cfg := svc.oauthConfig()
	assert.Equal(t, "http://localhost:3000/api/v1/callback/github/", cfg.RedirectURL)
	assert.Equal(t, "client-id", cfg.ClientID)
	assert.Equal(t, "client-secret", cfg.ClientSecret)
}

func TestMapOAuthExchangeErrorBadVerificationCode(t *testing.T) {
	err := mapOAuthExchangeError(assert.AnError)
	httpErr, ok := err.(internals.HttpError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.StatusCode)

	err = mapOAuthExchangeError(&dummyError{msg: `oauth2: "bad_verification_code"`})
	httpErr, ok = err.(internals.HttpError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.StatusCode)
	assert.True(t, strings.Contains(httpErr.Error(), "invalid or expired"))
}

func TestGenerateOAuthStateToken(t *testing.T) {
	token, err := generateOAuthStateToken()
	require.NoError(t, err)
	assert.Len(t, token, 32)
}

type dummyError struct {
	msg string
}

func (d *dummyError) Error() string {
	return d.msg
}
