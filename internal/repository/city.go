package repository

import "github.com/doodledoc/backend/internal/model"

// CityRepository defines city data access operations.
type CityRepository interface {
	ListCities() []model.City
}

// InMemoryCityRepository is a simple in-memory repository implementation.
type InMemoryCityRepository struct {
	cities []model.City
}

// NewInMemoryCityRepository creates a seeded city repository.
func NewInMemoryCityRepository() *InMemoryCityRepository {
	return &InMemoryCityRepository{
		cities: []model.City{
			{Name: "Bengaluru"},
			{Name: "Mumbai"},
			{Name: "Delhi"},
			{Name: "Chennai"},
			{Name: "Hyderabad"},
		},
	}
}

// ListCities returns a copy to avoid external mutation.
func (r *InMemoryCityRepository) ListCities() []model.City {
	cities := make([]model.City, len(r.cities))
	copy(cities, r.cities)
	return cities
}
