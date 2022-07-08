package util

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestWriteJSONResponse(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()

	WriteJSONResponse(w, http.StatusFound, struct {
		Key string `json:"key"`
	}{
		Key: "value",
	})

	s := w.Result().StatusCode
	if s != http.StatusFound {
		t.Errorf("expected status to be %#v got %#v", http.StatusFound, s)
		return
	}

	var a = map[string]string{"key": "value"}
	var b map[string]string
	err := json.NewDecoder(w.Result().Body).Decode(&b)
	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}
	if !reflect.DeepEqual(a, b) {
		t.Errorf("expected body to be %#v got %#v", a, b)
		return
	}
}

func TestWriteJSONError(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()

	message := "this is an error message"
	WriteJSONError(w, http.StatusUnauthorized, message)

	s := w.Result().StatusCode
	if s != http.StatusUnauthorized {
		t.Errorf("expected status to be %#v got %#v", http.StatusUnauthorized, s)
		return
	}

	var a = map[string]any{"error": true, "message": message}
	var b map[string]any
	err := json.NewDecoder(w.Result().Body).Decode(&b)
	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}
	if !reflect.DeepEqual(a, b) {
		t.Errorf("expected body to be %#v got %#v", a, b)
		return
	}
}
