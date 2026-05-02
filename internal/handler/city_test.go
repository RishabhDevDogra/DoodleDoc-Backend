package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type fakeCityService struct {
	names []string
}

func (f fakeCityService) ListCityNames() []string {
	return f.names
}

func TestCityHandlerListCities(t *testing.T) {
	h := NewCityHandler(fakeCityService{names: []string{"Pune", "Jaipur"}})
	req := httptest.NewRequest(http.MethodGet, "/cities", nil)
	rr := httptest.NewRecorder()

	h.ListCities(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}

	var body CitiesResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	want := []string{"Pune", "Jaipur"}
	if !reflect.DeepEqual(body.Cities, want) {
		t.Fatalf("expected cities %v, got %v", want, body.Cities)
	}
}
