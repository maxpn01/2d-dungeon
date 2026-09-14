package server

import (
	"net/http"
)

type rootResponse struct {
	Message string `json:"message"`
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	res := rootResponse{
		Message: "nothing to see here yet, go to /health",
	}

	writeJSON(w, res)
}
