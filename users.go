package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/flames31/Chirpy/internal/auth"
	"github.com/flames31/Chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, req *http.Request) {
	type incoming struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type respJSON struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	incomingJSON := incoming{}
	if err := json.NewDecoder(req.Body).Decode(&incomingJSON); err != nil {
		log.Printf("Error decoding json: %s", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
		})
		return
	}

	hashed_password, err := auth.HashPassword(incomingJSON.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
		})
		return
	}

	user, err := cfg.db.CreateUser(req.Context(), database.CreateUserParams{
		Email:          incomingJSON.Email,
		HashedPassword: hashed_password,
	})
	if err != nil {
		log.Printf("Error creating user: %s", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
		})
		return
	}

	writeJSON(w, http.StatusCreated, respJSON{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}

func (cfg *apiConfig) handleReset(w http.ResponseWriter, req *http.Request) {
	platform := os.Getenv("PLATFORM")
	if platform != "dev" {
		writeJSON(w, http.StatusForbidden, errorJSON{
			Error: "Endpoint forbidden!",
		})
		return
	}

	err := cfg.db.DeleteAllUsers(req.Context())

	if err != nil {
		log.Printf("Error deleting all users: %s", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
		})
		return
	}

	writeJSON(w, http.StatusOK, struct{}{})
}
