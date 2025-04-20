package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/flames31/Chirpy/internal/auth"
	"github.com/flames31/Chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, req *http.Request) {
	type incoming struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type respJSON struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
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

	token, err := auth.MakeJWT(user.ID, cfg.jwtToken)
	if err != nil {
		log.Printf("Error while creating token: %s", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
		})
		return
	}

	refresh_token, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Error while creating refresh token: %s", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
		})
		return
	}

	_, err = cfg.db.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{
		Token:     refresh_token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		log.Printf("Error while creating refresh token record in DB: %s", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
		})
		return
	}

	writeJSON(w, http.StatusOK, respJSON{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refresh_token,
	})
}
