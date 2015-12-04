package proxy

import (
	"encoding/json"
	"fmt"
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

// Tests if the Proxy handler will use the Simple Upstream correctly when the
// Simple Upstream url has been passed with correct user and password in its
// net.URL argument.
func TestBasichAuthenticateInSimpleUpstream(t *testing.T) {
	t.Parallel()

	correctUser, correctPassword := "parolata", "zausera"

	challengeServer := BasicAuthHandler{username: correctUser, password: correctPassword}

	ts := httptest.NewServer(challengeServer)
	defer ts.Close()

	upstreamURL, err := url.Parse(ts.URL)

	if err != nil {
		t.Fatal(err)
	}

	// we want a copy of the URL object here
	upstreamURLWithoutAuth := *upstreamURL

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

	if upstreamURLWithoutAuth.User != nil {
		t.Fatal("Wrong test: the test request shouldn't have any Basic Auth credentials")
	}

	req, err := http.NewRequest("GET", upstreamURLWithoutAuth.String(), nil)

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

// Tests the balanced upstream from end-to-end. From parsing the config to using it
// in the proxy handler.
func TestBasichAuthenticateParsedFromConfig(t *testing.T) {
	t.Parallel()

	correctUser, correctPassword := "parolata", "zausera"

	challengeServer := BasicAuthHandler{username: correctUser, password: correctPassword}

	ts := httptest.NewServer(challengeServer)
	defer ts.Close()

	upstreamURL, err := url.Parse(ts.URL)

	if err != nil {
		t.Fatal(err)
	}

	configUpstreamServer := fmt.Sprintf("%s://%s:%s@%s",
		upstreamURL.Scheme, correctUser, correctPassword, upstreamURL.Host)

	upstreamConfigString := fmt.Sprintf(`
	{
		"balancing": "random",
		"addresses": [
			"%s"
		]
	}`, configUpstreamServer)

	var cfgUp config.Upstream

	json.Unmarshal([]byte(upstreamConfigString), &cfgUp)

	up, err := upstream.New(&cfgUp, mock.NewLogger())

	if err != nil {
		t.Fatalf("Failed to create upstream: %s", err)
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
		t.Errorf("Expected %d for request with BasicAuth upstream. Got %d. "+
			"Upstream config was:\n%s",
			http.StatusNoContent, resp.Code, upstreamConfigString)
	}

}
