package server

import (
	"net/http"
)

type healthResponse struct {
	Status string `json:"status"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	res := healthResponse{
		Status: "ok",
	}

	writeJSON(w, res)
}
