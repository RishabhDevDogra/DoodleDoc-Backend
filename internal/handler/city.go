package handler

import (
	"encoding/json"
	"net/http"

	"github.com/doodledoc/backend/internal/service"
)

// CitiesResponse is the response shape for GET /cities.
type CitiesResponse struct {
	Cities []string `json:"cities"`
}

// CityHandler is a controller-like HTTP adapter for city endpoints.
type CityHandler struct {
	cityService service.CityService
}

// NewCityHandler creates a new CityHandler.
func NewCityHandler(cityService service.CityService) *CityHandler {
	return &CityHandler{cityService: cityService}
}

// ListCities returns available city names.
//
// @Summary List cities
// @Description Returns the list of city names.
// @Tags cities
// @Produce json
// @Success 200 {object} handler.CitiesResponse
// @Router /cities [get]
func (h *CityHandler) ListCities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CitiesResponse{Cities: h.cityService.ListCityNames()})
}
