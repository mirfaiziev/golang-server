package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/mirfaiziev/golang-server/internal/app/stats"
)

type AnalyzeRequest []stats.InputItem

type AnalyzeErrResponse struct {
	Message string `json:"message"`
}

type AnalyzeHandler struct {
	s *stats.Service
}

func NewAnalyzeHandler(s *stats.Service) *AnalyzeHandler {
	return &AnalyzeHandler{s: s}
}

func (h *AnalyzeHandler) Analyze(w http.ResponseWriter, r *http.Request) {
	nweeks, err := h.getNweeks(r)
	if err != nil {
		h.validationError(w, err, http.StatusBadRequest)
		return
	}

	var req AnalyzeRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		h.validationError(w, fmt.Errorf("failed to decode analyze request body: %w", err), http.StatusBadRequest)

		return
	}

	result, err := h.s.Analyze(nweeks, time.Now(), req...)
	if err != nil {
		h.validationError(w, fmt.Errorf("failed to analyze: %w", err), http.StatusBadRequest)

		return
	}

	jsonResp, _ := json.Marshal(result)

	w.Write(jsonResp)
	w.WriteHeader(http.StatusOK)
}

func (h *AnalyzeHandler) getNweeks(r *http.Request) (int, error) {
	nweeksParam := r.URL.Query().Get("nweeks")

	if nweeksParam == "" {
		return 0, fmt.Errorf("nweeks param is required")
	}

	nweeks, err := strconv.Atoi(nweeksParam)
	if err != nil || nweeks <= 0 || nweeks > stats.MaxWeeks {
		return 0, fmt.Errorf("nweeks param must be a positive number, less than %d", stats.MaxWeeks)
	}

	return nweeks, nil
}

func (h *AnalyzeHandler) validationError(w http.ResponseWriter, err error, statusCode int) {
	// todo: log error
	if statusCode == 0 {
		statusCode = http.StatusBadRequest
	}
	jsonResp, _ := json.Marshal(AnalyzeErrResponse{Message: err.Error()})

	w.Write(jsonResp)
	w.WriteHeader(http.StatusBadRequest)
}
