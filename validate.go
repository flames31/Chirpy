package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func handlerValidateChirp(w http.ResponseWriter, req *http.Request) {
	type incoming struct {
		Body string `json:"body"`
	}

	type respJSON struct {
		Valid bool `json:"valid"`
	}

	type errorJSON struct {
		Error string `json:"error"`
	}

	incomingJSON := incoming{}
	if err := json.NewDecoder(req.Body).Decode(&incomingJSON); err != nil {
		log.Printf("Error decoding json: %s", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
		})
		return
	}

	if len(incomingJSON.Body) > 140 {
		writeJSON(w, http.StatusBadRequest, errorJSON{
			Error: "Chirp is too long",
		})
		return
	}

	writeJSON(w, http.StatusBadRequest, respJSON{
		Valid: true,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}
