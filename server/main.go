package main

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Vonucka, Hello!!"))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)

		log.WithFields(log.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"agent":  r.UserAgent(),
			"time":   time.Since(start),
		}).Debug("Handle request")
	})
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
	log.Info("Initialize...")

	router := mux.NewRouter()
	router.Use(loggingMiddleware)
	router.HandleFunc("/", homeHandler).Methods("GET")

	log.Info("Starting server on :8080")
	http.ListenAndServe(":8080", router)
}
