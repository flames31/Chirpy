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
		Email     string `json:"email"`
		Password  string `json:"password"`
		ExpiresIn int    `json:"expires_in_seconds"`
	}

	type respJSON struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
	}

	incomingJSON := incoming{}
	if err := json.NewDecoder(req.Body).Decode(&incomingJSON); err != nil {
		log.Printf("Error decoding json: %s", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
		})
		return
	}

	expiresIn := time.Duration(incomingJSON.ExpiresIn) * time.Second

	if expiresIn == 0 || expiresIn > time.Hour {
		expiresIn = time.Hour
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

	token, err := auth.MakeJWT(user.ID, cfg.jwtToken, expiresIn)
	if err != nil {
		log.Printf("Error while creating token: %s", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
		})
		return
	}

	writeJSON(w, http.StatusOK, respJSON{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	})
}
