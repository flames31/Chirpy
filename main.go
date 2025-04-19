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
	}
	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", cfg.handleMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.handleReset)
	mux.HandleFunc("POST /api/chirps", cfg.handleCreateChirp)
	mux.HandleFunc("POST /api/users", cfg.handlerCreateUser)
	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Error while listening : %v", err)
	}

}
