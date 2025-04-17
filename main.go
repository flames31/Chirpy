package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func main() {
	const filePathRoot = "."
	const port = "8080"

	mux := http.NewServeMux()
	cfg := apiConfig{
		fileServerHits: atomic.Int32{},
	}
	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", cfg.handleMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.handlerMetricsReset)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Error while listening : %v", err)
	}

}
