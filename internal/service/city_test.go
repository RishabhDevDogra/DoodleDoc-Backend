package service

import (
	"reflect"
	"testing"

	"github.com/doodledoc/backend/internal/model"
)

type fakeCityRepository struct {
	cities []model.City
}

func (f fakeCityRepository) ListCities() []model.City {
	return f.cities
}

func TestDefaultCityServiceListCityNames(t *testing.T) {
	svc := NewCityService(fakeCityRepository{
		cities: []model.City{{Name: "Pune"}, {Name: "Jaipur"}},
	})

	got := svc.ListCityNames()
	want := []string{"Pune", "Jaipur"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}
