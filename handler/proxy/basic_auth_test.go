package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/mock"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/upstream"
)

type BasicAuthHandler struct {
	username string // Username to be used for basic authenticate
	password string // Password to be used for basic authenticate
}

// Implements the http.Handler interface and does the actual basic authenticate
// check for every request
func (bah BasicAuthHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	if !bah.authenticate(req) {
		w.Header().Set("WWW-Authenticate", `Basic realm="HTTPMS"`)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Compares the authentication header with the stored user and passwords
// and returns true if they pass.
func (bah BasicAuthHandler) authenticate(req *http.Request) bool {

	user, pass, ok := req.BasicAuth()

	if !ok {
		return false
	}

	return user == bah.username && pass == bah.password
}

// First, we have to make sure the HTTP Handler which does challenge authenticate
// does work as expected. For that end the Go's stdlib is used for making requests
// with or without BasicAuth.
func TestBasicAuthHandler(t *testing.T) {
	t.Parallel()

	correctUser, correctPassword := "parolata", "zausera"

	challengeServer := BasicAuthHandler{username: correctUser, password: correctPassword}

	ts := httptest.NewServer(challengeServer)
	defer ts.Close()

	upstreamURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", upstreamURL.String(), nil)

	if err != nil {
		t.Fatal(err)
	}

	client := http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		t.Fatal(err)
	}

	if http.StatusUnauthorized != resp.StatusCode {
		t.Errorf("Expected %d for request without authentication. Got %d",
			http.StatusUnauthorized, resp.StatusCode)
	}

	req.SetBasicAuth("wronguser", "wrongpassword")

	resp, err = client.Do(req)

	if err != nil {
		t.Fatal(err)
	}

	if http.StatusUnauthorized != resp.StatusCode {
		t.Errorf("Expected %d for request with wrong authentication. Got %d",
			http.StatusUnauthorized, resp.StatusCode)
	}

	req.SetBasicAuth(correctUser, correctPassword)

	resp, err = client.Do(req)

	if err != nil {
		t.Fatal(err)
	}

	if http.StatusNoContent != resp.StatusCode {
		t.Errorf("Expected %d for request with correct authentication. Got %d",
			http.StatusNoContent, resp.StatusCode)
	}
}

func TestBasichAuthenticateInProxySettings(t *testing.T) {
	t.Parallel()

	correctUser, correctPassword := "parolata", "zausera"

	challengeServer := BasicAuthHandler{username: correctUser, password: correctPassword}

	ts := httptest.NewServer(challengeServer)
	defer ts.Close()

	upstreamURL, err := url.Parse(ts.URL)

	if err != nil {
		t.Fatal(err)
	}

	upstreamURL.User = url.UserPassword(correctUser, correctPassword)

	up, err := upstream.NewSimple(upstreamURL)
	if err != nil {
		t.Fatal(err)
	}

	proxy, err := New(
		config.NewHandler("proxy", json.RawMessage(`{}`)),
		&types.Location{
			Name:     "test",
			Logger:   mock.NewLogger(),
			Upstream: up,
		}, nil)

	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", upstreamURL.String(), nil)

	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()

	ctx := contexts.NewAppContext(context.Background(), &mockApp{})

	proxy.ServeHTTP(ctx, resp, req)

	if resp.Code != http.StatusNoContent {
		t.Errorf("Expected %d for request with BasicAuth upstream. Got %d",
			http.StatusNoContent, resp.Code)
	}
}
