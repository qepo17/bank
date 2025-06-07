package response

import (
	"encoding/json"
	"net/http"
)

func StatusOnly(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
}

func Json(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Use Marshal instead of Encoder to avoid trailing newline
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Write(jsonData)
}

func JsonError(w http.ResponseWriter, status int, message string) {
	Json(w, status, map[string]string{"error": message})
}
