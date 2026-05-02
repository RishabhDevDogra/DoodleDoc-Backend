package router

import (
	"net/http"

	"github.com/doodledoc/backend/internal/handler"
	"github.com/doodledoc/backend/internal/repository"
	"github.com/doodledoc/backend/internal/service"
	httpSwagger "github.com/swaggo/http-swagger"
)

// New returns a configured HTTP mux.
func New() http.Handler {
	mux := http.NewServeMux()
	cityRepo := repository.NewInMemoryCityRepository()
	cityService := service.NewCityService(cityRepo)
	cityHandler := handler.NewCityHandler(cityService)

	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("GET /cities", cityHandler.ListCities)
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	return mux
}
