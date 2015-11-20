package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

func newLocationWithHandler(name string) *types.Location {
	return &types.Location{
		Name: name,
		Handler: types.RequestHandlerFunc(func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusOK)
			if _, err := rw.Write([]byte(name)); err != nil {
				panic(err)
			}
		}),
	}
}

const (
	notLocation          = `virtualHost`
	exactStat            = `= /stat`
	status               = `/status`
	picturesWithoutRegex = `^~ /pictures`
	specialJpG           = `~ jpG$`
	jpgs                 = `~* jpg$`
	notFound             = "404 page not found\n"
)

func TestLocationMatching(t *testing.T) {
	t.Parallel()
	muxer, err := NewLocationMuxer(
		[]*types.Location{
			newLocationWithHandler(exactStat),
			newLocationWithHandler(status),
			newLocationWithHandler(picturesWithoutRegex),
			newLocationWithHandler(specialJpG),
			newLocationWithHandler(jpgs),
		},
	)
	if err != nil {
		t.Fatal("Error while creating test LocationMuxer", err)
	}
	app := &Application{
		cfg:                  new(config.Config),
		notConfiguredHandler: newNotConfiguredHandler(),
		virtualHosts: map[string]*VirtualHost{
			"localhost": {
				Location: types.Location{
					Name: "localhost",
					Handler: types.RequestHandlerFunc(func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
						if req.Host != "localhost" {
							t.Fatalf("VirtualHost handler got requst for %s.", req.Host)
						}
						rw.WriteHeader(200)
						if _, err := rw.Write([]byte(notLocation)); err != nil {
							t.Fatalf("Unexpected Write error: %s", err)
						}
					}),
				},
				Muxer: muxer,
			},

			"localhost2": {
				Location: types.Location{
					Name: "localhost2",
				},
				Muxer: muxer,
			},
		},
		stats: new(applicationStats),
	}

	var mat = map[string]string{
		"http://localhost/notinlocaitons":       notLocation,
		"http://localhost/statu":                notLocation,
		"http://localhost/stat/":                notLocation,
		"http://localhost/stat":                 exactStat,
		"http://localhost/status/":              status,
		"http://localhost/status/somewhereElse": status,
		"http://localhost/test.jpg":             jpgs,
		"http://localhost/test.jpG":             specialJpG,
		"http://localhost/pictures/test.jpg":    picturesWithoutRegex,
		"http://localhost/pictures/test.jpG":    picturesWithoutRegex,
		"http://localhost/Pictures/Terst.jpG":   picturesWithoutRegex,
		// not in virtualhosts
		"http://localhost.com/pictures/test.jpG": notFound,
		// localhost2
		"http://localhost2/notinlocaitons":       notFound,
		"http://localhost2/statu":                notFound,
		"http://localhost2/stat/":                notFound,
		"http://localhost2/stat":                 exactStat,
		"http://localhost2/status/":              status,
		"http://localhost2/status/somewhereElse": status,
		"http://localhost2/test.jpg":             jpgs,
		"http://localhost2/test.jpG":             specialJpG,
		"http://localhost2/pictures/test.jpg":    picturesWithoutRegex,
		"http://localhost2/pictures/test.jpG":    picturesWithoutRegex,
		"http://localhost2/Pictures/Terst.jpG":   picturesWithoutRegex,
	}
	recorder := httptest.NewRecorder()
	for url, expected := range mat {
		recorder.Body.Reset()
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			t.Fatalf("Error while creating request - %s", err)
		}
		app.ServeHTTP(recorder, req)
		got := recorder.Body.String()
		if got != expected {
			t.Errorf("Expected %s got %s in the body for url %s", expected, got, url)

		}
	}

	var stats = app.Stats()
	var expectedRequests, expectedResponded, expectedNotConfigured uint64 = 23, 19, 4
	if stats.Requests != expectedRequests {
		t.Errorf("expected requests are %d but got %d", expectedRequests, stats.Requests)
	}
	if stats.Responded != expectedResponded {
		t.Errorf("expected responded requsts are %d but got %d", expectedResponded, stats.Responded)
	}
	if stats.NotConfigured != expectedNotConfigured {
		t.Errorf("expected non configured requests are %d but got %d", expectedNotConfigured, stats.NotConfigured)
	}
}
