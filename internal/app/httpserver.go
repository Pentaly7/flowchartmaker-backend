package app

import (
	"errors"
	"github.com/Pentaly7/flowchartmaker-backend/internal/models"
	"net/http"
)

type App interface {
	Start()
}

type httpServer struct {
	cfg *models.Config
}

func New(cfg *models.Config) App {
	return &httpServer{
		cfg: cfg,
	}
}

func (s *httpServer) Start() {
	// Create server
	mux := http.NewServeMux()
	mux.HandleFunc("/flowcharts/", s.handleFlowcharts)

	handler := s.enableCORS(mux)
	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: handler,
	}

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}
