package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/trbute-boot/chirpy/internal/auth"
	"net/http"
)

func (cfg *apiConfig) handlePolkaWebhook(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		UserID uuid.UUID `json:"user_id"`
	}

	type Parameters struct {
		Event string `json:"event"`
		Data  Data   `json:"data"`
	}

	token, err := auth.GetAPIKey(r.Header)
	if err != nil || token != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Unable to retrieve token", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to decode parameters", err)
		return
	}

	if params.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}

	_, err = cfg.db.GetUserById(r.Context(), params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed getting user", err)
		return
	}

	err = cfg.db.SetChirpyRed(r.Context(), params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed updating user", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
