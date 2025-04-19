package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/flames31/Chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, req *http.Request) {
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

	user, err := cfg.db.GetUserByEmail(req.Context(), incomingJSON.Email)
	if err != nil {
		log.Printf("Incorrect email or password: %s", err)
		writeJSON(w, http.StatusUnauthorized, errorJSON{
			Error: "Incorrect email or password",
		})
		return
	}
	err = auth.CheckPasswordHash(user.HashedPassword, incomingJSON.Password)
	if err != nil {
		log.Printf("Incorrect email or password: %s", err)
		writeJSON(w, http.StatusUnauthorized, errorJSON{
			Error: "Incorrect email or password",
		})
		return
	}

	writeJSON(w, http.StatusOK, respJSON{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}
