package helper

import (
	"encoding/json"
	"net/http"
)

func ResponseGenerator(w http.ResponseWriter, Value interface{}, status int) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Value)
}
