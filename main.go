package main

import (
	"database/sql"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/trbute-boot/chirpy/internal/database"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	jwtSecret      string
	polkaKey       string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env")
	}

	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	jwtSecret := os.Getenv("JWT_SECRET")
	polkaKey := os.Getenv("POLKA_KEY")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Unable to connect to db")
	}

	port := "8080"
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	apiCfg := apiConfig{
		db:        database.New(db),
		platform:  platform,
		jwtSecret: jwtSecret,
		polkaKey:  polkaKey,
	}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /admin/metrics", apiCfg.handleGetMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handleReset)
	mux.HandleFunc("GET /api/chirps", apiCfg.handleGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handleGetChirpById)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handleDeleteChirpById)
	mux.HandleFunc("POST /api/chirps", apiCfg.handleCreateChirp)
	mux.HandleFunc("POST /api/users", apiCfg.handleCreateUser)
	mux.HandleFunc("PUT /api/users", apiCfg.handleUpdateUser)
	mux.HandleFunc("POST /api/login", apiCfg.handleLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.handleRefresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.handleRevoke)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlePolkaWebhook)
	mux.HandleFunc("GET /api/healthz", handleHealth)

	server.ListenAndServe()
}
