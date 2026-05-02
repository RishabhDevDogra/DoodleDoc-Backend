package main

import (
	"log"
	"net/http"

	_ "github.com/doodledoc/backend/docs"
	"github.com/doodledoc/backend/internal/router"
)

// @title DoodleDoc Backend API
// @version 1.0
// @description API documentation for the DoodleDoc backend service.
// @host localhost:8080
// @BasePath /

func main() {
	r := router.New()

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
