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
		ID          uuid.UUID `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red`
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
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
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

func (cfg *apiConfig) handleUpdateCredentials(w http.ResponseWriter, req *http.Request) {
	type incoming struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type respJSON struct {
		Email string `json:"email"`
	}

	incomingJSON := incoming{}
	if err := json.NewDecoder(req.Body).Decode(&incomingJSON); err != nil {
		log.Printf("Error decoding json: %s", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
		})
		return
	}

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorJSON{
			Error: "User not authorized",
		})
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtToken)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorJSON{
			Error: "User not authorized",
		})
		return
	}

	hashedPassword, err := auth.HashPassword(incomingJSON.Password)
	if err != nil {
		log.Printf("Error hashing password :%v", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
		})
		return
	}

	err = cfg.db.UpdateUserCredential(req.Context(), database.UpdateUserCredentialParams{
		ID:             userID,
		Email:          incomingJSON.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		log.Printf("Error saving to DB :%v", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
		})
		return
	}

	writeJSON(w, http.StatusOK, respJSON{
		Email: incomingJSON.Email,
	})
}

func (cfg *apiConfig) handlerUpdateChirpyRed(w http.ResponseWriter, req *http.Request) {
	type incoming struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	apiKey, err := auth.GetAPIKey(req.Header)
	if err != nil || apiKey != cfg.polkaAPIKey {
		log.Printf("Incorrect / No apiKey present: %s", err)
		writeJSON(w, http.StatusUnauthorized, errorJSON{
			Error: "Missing/Incorrect apiKey",
		})
		return
	}

	incomingJSON := incoming{}
	if err := json.NewDecoder(req.Body).Decode(&incomingJSON); err != nil {
		log.Printf("Error decoding json: %s", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
		})
		return
	}

	if incomingJSON.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userIDInUUID, err := uuid.Parse(incomingJSON.Data.UserID)
	if err != nil {
		log.Printf("Error parsing to UUID: %s", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
		})
		return
	}

	err = cfg.db.UpdateChirpyRed(req.Context(), database.UpdateChirpyRedParams{
		ID:          userIDInUUID,
		IsChirpyRed: true,
	})
	if err != nil {
		log.Printf("User not found: %s", err)
		writeJSON(w, http.StatusNotFound, errorJSON{
			Error: "User not present",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
