package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
)

// Configure CORS and wrap handler with CORS middleware
func ConfigureCORS(handler http.Handler) http.Handler {

	corsConfig := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	corsHandler := corsConfig.Handler(handler)

	return corsHandler
}
