package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"alpoGame/app/model"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type appState struct {
	mu   sync.Mutex
	game *model.Game
}

const defaultServerAddr = ":8081"

type joinRequest struct {
	Name string `json:"name"`
}

type joinResponse struct {
	PlayerID     string `json:"playerId"`
	PlayerNumber int    `json:"playerNumber"`
}

type placeRequest struct {
	PlayerID string                `json:"playerId"`
	Ships    []model.ShipPlacement `json:"ships"`
}

type shotRequest struct {
	PlayerID string `json:"playerId"`
	Row      int    `json:"row"`
	Col      int    `json:"col"`
}

type resetRequest struct {
	PlayerID string `json:"playerId"`
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *loggingResponseWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	written, err := w.ResponseWriter.Write(data)
	w.bytes += written
	return written, err
}

func newAppState() *appState {
	return &appState{game: model.NewGame()}
}

func (a *appState) joinHandler(w http.ResponseWriter, r *http.Request) {
	var req joinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && r.ContentLength > 0 {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	playerID, err := randomID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create player id")
		return
	}

	a.mu.Lock()
	playerNumber, err := a.game.Join(playerID, req.Name)
	a.mu.Unlock()
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, joinResponse{PlayerID: playerID, PlayerNumber: playerNumber + 1})
}

func (a *appState) stateHandler(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("playerId")
	if playerID == "" {
		writeError(w, http.StatusBadRequest, "playerId is required")
		return
	}

	a.mu.Lock()
	view, err := a.game.View(playerID)
	a.mu.Unlock()
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, view)
}

func (a *appState) placeHandler(w http.ResponseWriter, r *http.Request) {
	var req placeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	a.mu.Lock()
	err := a.game.PlaceFleet(req.PlayerID, req.Ships)
	view, viewErr := a.game.View(req.PlayerID)
	a.mu.Unlock()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if viewErr != nil {
		writeError(w, http.StatusBadRequest, viewErr.Error())
		return
	}

	writeJSON(w, http.StatusOK, view)
}

func (a *appState) shootHandler(w http.ResponseWriter, r *http.Request) {
	var req shotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	a.mu.Lock()
	result, err := a.game.Shoot(req.PlayerID, model.Coord{Row: req.Row, Col: req.Col})
	view, viewErr := a.game.View(req.PlayerID)
	a.mu.Unlock()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if viewErr != nil {
		writeError(w, http.StatusBadRequest, viewErr.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"result": result,
		"view":   view,
	})
}

func (a *appState) resetHandler(w http.ResponseWriter, r *http.Request) {
	var req resetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.PlayerID == "" {
		writeError(w, http.StatusBadRequest, "playerId is required")
		return
	}

	a.mu.Lock()
	if _, err := a.game.View(req.PlayerID); err != nil {
		a.mu.Unlock()
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	a.game = model.NewGame()
	a.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]string{"status": "reset"})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &loggingResponseWriter{ResponseWriter: w}
		next.ServeHTTP(recorder, r)
		if recorder.status == 0 {
			recorder.status = http.StatusOK
		}

		entry := log.WithFields(log.Fields{
			"method":  r.Method,
			"path":    r.URL.Path,
			"remote":  r.RemoteAddr,
			"agent":   r.UserAgent(),
			"status":  recorder.status,
			"bytes":   recorder.bytes,
			"latency": time.Since(start).String(),
		})

		message := fmt.Sprintf("%s %s %d", r.Method, r.URL.Path, recorder.status)
		switch {
		case recorder.status >= http.StatusInternalServerError:
			entry.Error(message)
		case recorder.status >= http.StatusBadRequest:
			entry.Warn(message)
		case r.URL.Path == "/api/state":
			entry.Debug(message)
		default:
			entry.Info(message)
		}
	})
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.WithError(err).Warn("write response")
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func randomID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
	log.Info("Initialize...")

	state := newAppState()
	router := mux.NewRouter()
	router.Use(loggingMiddleware)
	router.HandleFunc("/api/join", state.joinHandler).Methods("POST")
	router.HandleFunc("/api/state", state.stateHandler).Methods("GET")
	router.HandleFunc("/api/place", state.placeHandler).Methods("POST")
	router.HandleFunc("/api/shoot", state.shootHandler).Methods("POST")
	router.HandleFunc("/api/reset", state.resetHandler).Methods("POST")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("web")))

	log.Infof("Starting server on %s", defaultServerAddr)
	if err := newHTTPServer(defaultServerAddr, router).ListenAndServe(); err != nil {
		log.WithError(err).Fatal("server stopped")
	}
}
