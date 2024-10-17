package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleGetMetrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("<html>\n"+
		"\t<body>\n"+
		"\t\t<h1>Welcome, Chirpy Admin</h1>\n"+
		"\t\t<p>Chirpy has been visited %d times!</p>\n"+
		"\t</body>\n"+
		"</html>", cfg.fileserverHits.Load())))
}
