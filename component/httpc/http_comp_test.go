package httpc

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type dummyResponse struct {
	Message string `json:"message"`
}

func TestHTTPComponentLifecycle(t *testing.T) {
	comp := NewHTTPComponent("http")
	if comp.ID() != "http" {
		t.Fatalf("expected ID 'http', got %s", comp.ID())
	}

	comp.InitFlags()

	if err := comp.Activate(nil); err != nil {
		t.Fatalf("Activate error: %v", err)
	}

	if comp.GetHttpClient() == nil {
		t.Fatal("expected non-nil http client")
	}

	if err := comp.Stop(); err != nil {
		t.Fatalf("Stop error: %v", err)
	}
}

func TestHTTPComponentMakeRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom-Header") != "TestValue" {
			http.Error(w, "missing custom header", http.StatusBadRequest)
			return
		}

		if r.Method == http.MethodPost {
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(dummyResponse{Message: "received: " + body["name"]})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(dummyResponse{Message: "hello world"})
	}))
	defer server.Close()

	comp := NewHTTPComponent("http2")
	comp.InitFlags()
	_ = comp.Activate(nil)

	t.Run("GET Request without output unmarshal", func(t *testing.T) {
		header := make(http.Header)
		header.Set("X-Custom-Header", "TestValue")

		resp, err := comp.MakeRequest(context.Background(), http.MethodGet, server.URL, &ReqOption{
			Header:       header,
			TraceReqBody: true,
		}, nil)

		if err != nil {
			t.Fatalf("MakeRequest error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("POST Request with JSON body and output unmarshal", func(t *testing.T) {
		header := make(http.Header)
		header.Set("X-Custom-Header", "TestValue")
		header.Set("Content-Type", "application/json")

		bodyBytes, _ := json.Marshal(map[string]string{"name": "tester"})
		var result dummyResponse

		resp, err := comp.MakeRequest(context.Background(), http.MethodPost, server.URL, &ReqOption{
			Body:         bytes.NewReader(bodyBytes),
			Header:       header,
			Timeout:      2 * time.Second,
			TraceReqBody: true,
		}, &result)

		if err != nil {
			t.Fatalf("MakeRequest error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}
		if result.Message != "received: tester" {
			t.Fatalf("expected message 'received: tester', got '%s'", result.Message)
		}
	})

	t.Run("Request with nil option", func(t *testing.T) {
		_, _ = comp.MakeRequest(context.Background(), http.MethodGet, server.URL, nil, nil)
	})

	t.Run("Request with invalid URL", func(t *testing.T) {
		_, err := comp.MakeRequest(context.Background(), http.MethodGet, "::invalid-url", nil, nil)
		if err == nil {
			t.Fatal("expected error for invalid URL")
		}
	})
}
