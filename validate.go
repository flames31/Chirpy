package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func handlerValidateChirp(w http.ResponseWriter, req *http.Request) {
	type incoming struct {
		Body string `json:"body"`
	}

	type respJSON struct {
		Cleaned_body string `json:"cleaned_body"`
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

	writeJSON(w, http.StatusOK, respJSON{
		Cleaned_body: getCleanBody(incomingJSON.Body),
	})
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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}
