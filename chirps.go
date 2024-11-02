package main

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/trbute-boot/chirpy/internal/auth"
	"github.com/trbute-boot/chirpy/internal/database"
	"net/http"
	"slices"
	"strings"
	"time"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to decode parameters", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to retrieve token", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token invalid", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	params.Body = replaceProfanity(params.Body)
	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: userID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) handleGetAllChirps(w http.ResponseWriter, r *http.Request) {
	authorQuery := r.URL.Query().Get("author_id")
	sortQuery := r.URL.Query().Get("sort")
	if sortQuery != "" && sortQuery != "asc" && sortQuery != "desc" {
		respondWithError(w, http.StatusBadRequest, "Invalid sort argument", errors.New("Invalid sort argument"))
		return
	}

	var chirps []database.Chirp
	var err error

	if authorQuery == "" {
		if sortQuery == "" || sortQuery == "asc" {
			chirps, err = cfg.db.GetAllChirps(r.Context())
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Unable to retrieve chirps", err)
				return
			}
		} else {
			chirps, err = cfg.db.GetAllChirpsDesc(r.Context())
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Unable to retrieve chirps", err)
				return
			}
		}
	} else {
		authorId, err := uuid.Parse(authorQuery)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author ID", err)
			return
		}

		if sortQuery == "" || sortQuery == "asc" {
			chirps, err = cfg.db.GetChirpsByAuthor(r.Context(), authorId)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Unable to retrieve chirps", err)
				return
			}
		} else {
			chirps, err = cfg.db.GetChirpsByAuthorDesc(r.Context(), authorId)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Unable to retrieve chirps", err)
				return
			}
		}
	}

	chirpRes := []Chirp{}
	for _, chirp := range chirps {
		chirpRes = append(chirpRes, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, chirpRes)

}

func (cfg *apiConfig) handleGetChirpById(w http.ResponseWriter, r *http.Request) {
	chirpId, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}

	chirp, err := cfg.db.GetChirpById(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) handleDeleteChirpById(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to retrieve token", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token invalid", err)
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}

	chirp, err := cfg.db.GetChirpById(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Requester and chirp userid do not match", err)
		return
	}

	err = cfg.db.DeleteChirpById(r.Context(), database.DeleteChirpByIdParams{
		ID:     chirpID,
		UserID: userID,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed deleting chirp", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

func replaceProfanity(s string) string {
	words := strings.Split(s, " ")
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	for index, word := range words {
		if slices.Contains(badWords, strings.ToLower(word)) {
			words[index] = "****"
		}
	}

	return strings.Join(words, " ")
}
