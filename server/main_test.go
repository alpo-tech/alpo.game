package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

func TestNewHTTPServerConfiguresTimeouts(t *testing.T) {
	server := newHTTPServer(":0", http.NewServeMux())

	if server.ReadHeaderTimeout < time.Second {
		t.Fatalf("expected ReadHeaderTimeout to be configured")
	}
	if server.ReadTimeout < time.Second {
		t.Fatalf("expected ReadTimeout to be configured")
	}
	if server.WriteTimeout < time.Second {
		t.Fatalf("expected WriteTimeout to be configured")
	}
	if server.IdleTimeout < time.Second {
		t.Fatalf("expected IdleTimeout to be configured")
	}
}

func TestResetRequiresKnownPlayer(t *testing.T) {
	state := newAppState()
	router := testRouter(state)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/reset", bytes.NewBufferString(`{"playerId":"missing"}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestResetAllowsKnownPlayer(t *testing.T) {
	state := newAppState()
	router := testRouter(state)
	playerID := joinTestPlayer(t, router)

	response := httptest.NewRecorder()
	request := jsonRequest(t, http.MethodPost, "/api/reset", map[string]string{"playerId": playerID})
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
}

func joinTestPlayer(t *testing.T, router http.Handler) string {
	t.Helper()

	response := httptest.NewRecorder()
	request := jsonRequest(t, http.MethodPost, "/api/join", map[string]string{"name": "one"})
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected join status %d, got %d", http.StatusOK, response.Code)
	}

	var body joinResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	return body.PlayerID
}

func jsonRequest(t *testing.T, method, path string, body interface{}) *http.Request {
	t.Helper()

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(method, path, bytes.NewReader(data))
	request.Header.Set("Content-Type", "application/json")
	return request
}

func testRouter(state *appState) http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/api/join", state.joinHandler).Methods(http.MethodPost)
	router.HandleFunc("/api/reset", state.resetHandler).Methods(http.MethodPost)
	return router
}
