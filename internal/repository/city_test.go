package repository

import "testing"

func TestInMemoryCityRepositoryListCities(t *testing.T) {
	repo := NewInMemoryCityRepository()
	cities := repo.ListCities()

	if len(cities) == 0 {
		t.Fatal("expected seeded cities, got empty list")
	}

	if cities[0].Name == "" {
		t.Fatal("expected city name to be populated")
	}
}
