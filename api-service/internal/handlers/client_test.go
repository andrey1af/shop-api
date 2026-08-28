package handlers

import (
	"net/http/httptest"
	"testing"
)

func TestParseClientFilter(t *testing.T) {
	request := httptest.NewRequest("GET", "/api/v1/clients?client_name=Ivan&client_surname=Ivanov&limit=10&offset=2", nil)
	recorder := httptest.NewRecorder()

	filter, ok := parseClientFilter(recorder, request)
	if !ok {
		t.Fatalf("parseClientFilter() rejected valid query with status %d", recorder.Code)
	}
	if filter.ClientName == nil || *filter.ClientName != "Ivan" ||
		filter.ClientSurname == nil || *filter.ClientSurname != "Ivanov" ||
		filter.Limit == nil || *filter.Limit != 10 ||
		filter.Offset == nil || *filter.Offset != 2 {
		t.Fatalf("parseClientFilter() = %#v", filter)
	}
}

func TestParseClientFilterRejectsUnpairedName(t *testing.T) {
	request := httptest.NewRequest("GET", "/api/v1/clients?client_name=Ivan", nil)
	recorder := httptest.NewRecorder()

	if _, ok := parseClientFilter(recorder, request); ok {
		t.Fatal("parseClientFilter() accepted client_name without client_surname")
	}
	if recorder.Code != 400 {
		t.Fatalf("parseClientFilter() status = %d, want 400", recorder.Code)
	}
}

func TestParseClientFilterRejectsInvalidPagination(t *testing.T) {
	for _, query := range []string{"limit=0", "offset=-1", "limit=invalid"} {
		t.Run(query, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/api/v1/clients?"+query, nil)
			recorder := httptest.NewRecorder()

			if _, ok := parseClientFilter(recorder, request); ok {
				t.Fatalf("parseClientFilter() accepted %q", query)
			}
			if recorder.Code != 400 {
				t.Fatalf("parseClientFilter() status = %d, want 400", recorder.Code)
			}
		})
	}
}
