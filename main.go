package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/flames31/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	db             *database.Queries
	fileServerHits atomic.Int32
	jwtToken       string
	polkaAPIKey    string
}

func main() {
	const filePathRoot = "."
	const port = "8080"

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Issue with opening DB")
	}
	dbQueries := database.New(db)

	mux := http.NewServeMux()
	cfg := apiConfig{
		fileServerHits: atomic.Int32{},
		db:             dbQueries,
		jwtToken:       os.Getenv("JWT_TOKEN"),
		polkaAPIKey:    os.Getenv("POLKA_KEY"),
	}
	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", cfg.handleMetrics)
	mux.HandleFunc("GET /api/chirps", cfg.handleGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handleGetChirp)
	mux.HandleFunc("POST /admin/reset", cfg.handleReset)
	mux.HandleFunc("POST /api/chirps", cfg.handleCreateChirp)
	mux.HandleFunc("POST /api/users", cfg.handlerCreateUser)
	mux.HandleFunc("POST /api/login", cfg.handleLogin)
	mux.HandleFunc("POST /api/refresh", cfg.handleRefresh)
	mux.HandleFunc("POST /api/revoke", cfg.handleRevoke)
	mux.HandleFunc("PUT /api/users", cfg.handleUpdateCredentials)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.handleDeleteChirp)
	mux.HandleFunc("POST /api/polka/webhooks", cfg.handlerUpdateChirpyRed)
	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Error while listening : %v", err)
	}

}
