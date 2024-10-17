package main

import (
	"encoding/json"
	"github.com/trbute-boot/chirpy/internal/auth"
	"net/http"
)

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.:Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
ExpiresInSeconds string `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to decode parameters", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	err = auth.CheckPasswordHash(user.HashedPassword, params.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	expires, ok := params.time_in_seconds
	if !ok || expires > 3600 {
		expires = 3600
	}

	expireDuration := Duration(expires * time.Second)


	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, 

	respondWithJSON(w, http.StatusOK, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:	   token,
	})
}
