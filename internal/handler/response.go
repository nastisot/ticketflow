package handler

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, value any) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err = w.Write(data); err != nil {
		return true, err
	}
	return true, nil
}
