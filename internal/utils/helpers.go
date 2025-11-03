package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func MustGetenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func WriteJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		log.Printf("write json failed: %v", err)
	}
}

func HTTPError(w http.ResponseWriter, err error, code int) {
	w.WriteHeader(code)
	WriteJSON(w, map[string]string{"error": err.Error()})
}
