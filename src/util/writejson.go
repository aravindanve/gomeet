package util

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func WriteJSONResponse(w http.ResponseWriter, statusCode int, body interface{}) {
	b, err := json.Marshal(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"error":true,"message":"%s"}`, err)))
	}

	w.WriteHeader(statusCode)
	w.Header().Add("content-type", "application/json")
	w.Write(b)
}

func WriteJSONError(w http.ResponseWriter, statusCode int, message string) {
	WriteJSONResponse(w, statusCode, map[string]any{
		"error":   true,
		"message": message,
	})
}
