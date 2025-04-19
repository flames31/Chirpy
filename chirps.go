package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/flames31/Chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, req *http.Request) {
	type incoming struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	type respJSON struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}
	incomingJSON := incoming{}
	if err := json.NewDecoder(req.Body).Decode(&incomingJSON); err != nil {
		log.Printf("Error decoding json: %s", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
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
		UserID: incomingJSON.UserID,
	})
	if err != nil {
		log.Printf("Error creating chirp: %s", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
		})
		return
	}

	writeJSON(w, http.StatusCreated, respJSON{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
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
