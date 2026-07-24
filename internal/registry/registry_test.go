package registry

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNpmVersions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/left-pad" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write([]byte(`{"name":"left-pad","versions":{"1.2.0":{},"1.3.0":{}}}`))
	}))
	defer srv.Close()

	old := npmRegistry
	npmRegistry = srv.URL
	defer func() { npmRegistry = old }()

	got, err := NpmVersions("left-pad")
	if err != nil {
		t.Fatalf("NpmVersions: %v", err)
	}
	if !got["1.3.0"] || !got["1.2.0"] || got["9.9.9"] {
		t.Errorf("versions = %v, want 1.2.0 and 1.3.0 only", got)
	}
}

func TestNpmVersions_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	old := npmRegistry
	npmRegistry = srv.URL
	defer func() { npmRegistry = old }()

	if _, err := NpmVersions("nope"); err == nil {
		t.Error("expected error for a 404 package")
	}
}
