package handler

import "net/http"

func Hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello word"))
	w.WriteHeader(http.StatusOK)
}
