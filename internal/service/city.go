package service

import "github.com/doodledoc/backend/internal/repository"

// CityService defines city-related business operations.
type CityService interface {
	ListCityNames() []string
}

// DefaultCityService contains business logic for city use cases.
type DefaultCityService struct {
	cityRepo repository.CityRepository
}

// NewCityService creates a CityService implementation.
func NewCityService(cityRepo repository.CityRepository) *DefaultCityService {
	return &DefaultCityService{cityRepo: cityRepo}
}

// ListCityNames transforms city entities to API-friendly names.
func (s *DefaultCityService) ListCityNames() []string {
	cities := s.cityRepo.ListCities()
	names := make([]string, 0, len(cities))
	for _, city := range cities {
		names = append(names, city.Name)
	}

	return names
}
