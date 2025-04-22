package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/flames31/Chirpy/internal/auth"
	"github.com/flames31/Chirpy/internal/database"
	"github.com/google/uuid"
)

type chirpJSON struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, req *http.Request) {
	type incoming struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
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
		log.Printf("Error validating token: %s", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
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

	cleanBody, ok := validateChirp(incomingJSON.Body)
	if !ok {
		writeJSON(w, http.StatusBadRequest, errorJSON{
			Error: "Chirp is too long",
		})
		return
	}

	chirp, err := cfg.db.CreateChirp(req.Context(), database.CreateChirpParams{
		Body:   cleanBody,
		UserID: userID,
	})
	if err != nil {
		log.Printf("Error creating chirp: %s", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
		})
		return
	}

	writeJSON(w, http.StatusCreated, chirpJSON{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) handleGetAllChirps(w http.ResponseWriter, req *http.Request) {
	chirps, err := cfg.db.GetAllChirps(req.Context())
	if err != nil {
		log.Printf("Error fetching all chirps: %s", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
		})
		return
	}
	chirpsJSON := []chirpJSON{}

	for _, chirp := range chirps {
		chirpsJSON = append(chirpsJSON, chirpJSON{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	writeJSON(w, http.StatusOK, chirpsJSON)
}

func (cfg *apiConfig) handleGetChirp(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("chirpID")

	chirpID, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("Error while converting to UUID!")
		writeJSON(w, http.StatusInternalServerError, errorJSON{Error: "Something went wrong"})
		return
	}

	chirp, err := cfg.db.GetChirpByID(req.Context(), chirpID)
	if err != nil {
		log.Printf("Error while retrieveing chirp / Given ChirpID does not exist!")
		writeJSON(w, http.StatusNotFound, errorJSON{Error: "Something went wrong"})
		return
	}

	writeJSON(w, http.StatusOK, chirpJSON{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) handleDeleteChirp(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("chirpID")

	chirpID, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("Error while converting to UUID!")
		writeJSON(w, http.StatusInternalServerError, errorJSON{Error: "Something went wrong"})
		return
	}

	chirp, err := cfg.db.GetChirpByID(req.Context(), chirpID)
	if err != nil {
		log.Printf("Error while retrieveing chirp / Given ChirpID does not exist!")
		writeJSON(w, http.StatusNotFound, errorJSON{Error: "Something went wrong"})
		return
	}

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Error validating token: %s", err)
		writeJSON(w, http.StatusUnauthorized, errorJSON{
			Error: "Incorrect toke / No token provided",
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

	if chirp.UserID != userID {
		writeJSON(w, http.StatusForbidden, errorJSON{
			Error: "Forbidden",
		})
		return
	}

	err = cfg.db.DeleteChirpByID(req.Context(), chirpID)
	if err != nil {
		log.Printf("Error deleting chirp: %s", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func validateChirp(body string) (string, bool) {
	if len(body) > 140 {
		return "", false
	}
	return getCleanBody(body), true
}

func getCleanBody(str string) string {
	const asteriks = "****"
	badWords := map[string]bool{"kerfuffle": true, "sharbert": true, "fornax": true}

	strSlice := strings.Split(str, " ")
	for i, s := range strSlice {
		if _, ok := badWords[strings.ToLower(s)]; ok {
			strSlice[i] = asteriks
		}
	}

	return strings.Join(strSlice, " ")
}
