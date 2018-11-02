package render

import (
	"encoding/json"
	"net/http"
)

func JSON(res http.ResponseWriter, code int, body interface{}) error {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.WriteHeader(code)
	return json.NewEncoder(res).Encode(body)
}
