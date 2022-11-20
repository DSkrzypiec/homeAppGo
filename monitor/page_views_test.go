package monitor

import "testing"

func TestIsEndpointMatched(t *testing.T) {
	endpoints := map[string]struct{}{
		"/":     {},
		"/docs": {},
	}
	pv := NewPageViews(false, endpoints)

	if !pv.isEndpointMatched("/") {
		t.Errorf("endpoint [/] should be matched")
	}
	if !pv.isEndpointMatched("/docs") {
		t.Errorf("endpoint [/docs] should be matched")
	}
	if !pv.isEndpointMatched("/docs?id=10") {
		t.Errorf("endpoint [/docs?id=10] should be matched")
	}
	if pv.isEndpointMatched("/crap") {
		t.Errorf("endpoint [/crap] should not be matched")
	}
}
