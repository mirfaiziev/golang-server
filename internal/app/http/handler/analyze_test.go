package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestHandler(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(HandlerSuite))
}

type HandlerSuite struct {
	suite.Suite
}

func (s *HandlerSuite) TestAnalyzeHandler_nweeks_absent() {
	req := httptest.NewRequest(http.MethodGet, "/analyze", nil)
	w := httptest.NewRecorder()

	h := NewAnalyzeHandler(nil, nil)
	h.Analyze(w, req)

	s.Equal(http.StatusBadRequest, http.StatusBadRequest)
	s.Contains(w.Body.String(), "nweeks param is required")
}

func (s *HandlerSuite) TestAnalyzeHandler_nweeks_wrong() {
	req := httptest.NewRequest(http.MethodGet, "/analyze?nweeks=some-string", nil)
	w := httptest.NewRecorder()

	h := NewAnalyzeHandler(nil, nil)
	h.Analyze(w, req)

	s.Equal(http.StatusBadRequest, http.StatusBadRequest)
	s.Contains(w.Body.String(), "nweeks param must be a positive number, less than 6000")
}

func (s *HandlerSuite) TestAnalyzeHandler_nweeks_too_big() {
	req := httptest.NewRequest(http.MethodGet, "/analyze?nweeks=7000", nil)
	w := httptest.NewRecorder()

	h := NewAnalyzeHandler(nil, nil)
	h.Analyze(w, req)

	s.Equal(http.StatusBadRequest, http.StatusBadRequest)
	s.Contains(w.Body.String(), "nweeks param must be a positive number, less than 6000")
}

func (s *HandlerSuite) TestAnalyzeHandler_wrong_body() {
	req := httptest.NewRequest(http.MethodGet, "/analyze?nweeks=1", strings.NewReader("some string"))
	w := httptest.NewRecorder()

	h := NewAnalyzeHandler(nil, nil)
	h.Analyze(w, req)

	s.Equal(http.StatusBadRequest, http.StatusBadRequest)
	s.Contains(w.Body.String(), "failed to decode analyze request body")
}
