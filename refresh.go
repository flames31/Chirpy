package main

import (
	"context"
	"log"
	"net/http"

	"github.com/flames31/Chirpy/internal/auth"
)

func (cfg *apiConfig) handleRefresh(w http.ResponseWriter, req *http.Request) {
	type respJSON struct {
		Token string `json:"token"`
	}
	refresh_token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Error while retieving refresh token: %s", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
		})
		return
	}

	refresh_token_record, err := cfg.db.GetRefreshToken(context.Background(), refresh_token)
	if err != nil {
		log.Printf("No refersh token record: %s", err)
		writeJSON(w, http.StatusUnauthorized, errorJSON{
			Error: "No user with give refersh token",
		})
		return
	}

	if refresh_token_record.RevokedAt.Valid {
		writeJSON(w, http.StatusUnauthorized, errorJSON{
			Error: "Refersh token revoked!",
		})
		return
	}

	user, err := cfg.db.GetUserFromRefreshToken(req.Context(), refresh_token)
	if err != nil {
		log.Printf("No user with give refersh token: %s", err)
		writeJSON(w, http.StatusUnauthorized, errorJSON{
			Error: "No user with give refersh token",
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

	writeJSON(w, http.StatusOK, respJSON{
		Token: token,
	})
}

func (cfg *apiConfig) handleRevoke(w http.ResponseWriter, req *http.Request) {
	refresh_token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Error while retieving refresh token: %s", err)
		writeJSON(w, http.StatusInternalServerError, errorJSON{
			Error: "Something went wrong",
		})
		return
	}

	err = cfg.db.RevokeRefreshToken(req.Context(), refresh_token)
	if err != nil {
		log.Printf("No record found with given refresh token: %s", err)
		writeJSON(w, http.StatusNotFound, errorJSON{
			Error: "No record found with given refresh token",
		})
		return
	}
	w.WriteHeader(http.StatusNoContent)

}
