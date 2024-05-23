package internal

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/prathoss/cards/pkg"
)

type Server struct {
	config Config
}

func NewServer(config Config) *Server {
	return &Server{
		config: config,
	}
}

func (s *Server) Run() {
	mux := http.NewServeMux()

	server := &http.Server{
		Addr:              s.config.Address,
		Handler:           mux,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 100 * time.Millisecond,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       15 * time.Second,
		ErrorLog:          slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
	}

	if err := pkg.ServeWithShutdown(server); err != nil {
		slog.Error("server shut down with error", pkg.Err(err))
	}
}

func parseID(r *http.Request) (uuid.UUID, []pkg.InvalidParam) {
	return uuid.Nil, nil
}

func parseShuffled(r *http.Request) (bool, []pkg.InvalidParam) {
	return nil, nil
}

func parseCards(r *http.Request) ([]string, []pkg.InvalidParam) {
	return nil, nil
}

func parseCount(r *http.Request) (int, []pkg.InvalidParam) {
	return 0, nil
}
