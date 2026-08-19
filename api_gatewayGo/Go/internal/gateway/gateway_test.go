package gateway

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"api-gateway-go/internal/config"
)

func TestGatewayProxiesRawRequest(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/base/upload/file" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.URL.RawQuery != "mode=raw&item=1" {
			t.Errorf("raw query = %s", r.URL.RawQuery)
		}
		if r.Header.Get("Content-Type") != "application/x-custom" {
			t.Errorf("content type = %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("X-Forwarded-Prefix") != "/api/files" {
			t.Errorf("forwarded prefix = %s", r.Header.Get("X-Forwarded-Prefix"))
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		if string(body) != "raw=body" {
			t.Errorf("body = %q", string(body))
		}

		w.Header().Set("X-Upstream", "ok")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("proxied"))
	}))
	defer upstream.Close()

	store := storeFromJSON(t, `{"files": "`+upstream.URL+`/base"}`)
	handler := New(store, Options{})

	req := httptest.NewRequest(http.MethodPatch, "/api/files/upload/file?mode=raw&item=1", strings.NewReader("raw=body"))
	req.Header.Set("Content-Type", "application/x-custom")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("X-Upstream") != "ok" {
		t.Fatalf("upstream header was not preserved")
	}
	if recorder.Body.String() != "proxied" {
		t.Fatalf("body = %q", recorder.Body.String())
	}
}

func TestGatewayHandlesCORSPreflight(t *testing.T) {
	store := storeFromJSON(t, `{"files": "http://example.com"}`)
	handler := New(store, Options{
		CORS: CORSOptions{
			AllowedOrigins:   []string{"*"},
			AllowCredentials: true,
		},
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/files/upload", nil)
	req.Header.Set("Origin", "http://frontend.example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "Authorization, Content-Type")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d", recorder.Code)
	}
	if recorder.Header().Get("Access-Control-Allow-Origin") != "http://frontend.example.com" {
		t.Fatalf("allow origin = %q", recorder.Header().Get("Access-Control-Allow-Origin"))
	}
	if recorder.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatalf("allow credentials missing")
	}
}

func storeFromJSON(t *testing.T, body string) *config.Store {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "services-*.json")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	if _, err := file.WriteString(body); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close config: %v", err)
	}

	store := config.NewStore(file.Name(), time.Nanosecond)
	if err := store.Load(); err != nil {
		t.Fatalf("store.Load() error = %v", err)
	}
	return store
}
