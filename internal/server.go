package internal

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
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
	idParamName := "id"
	var invalidParams []pkg.InvalidParam
	idStr := r.PathValue(idParamName)
	if idStr == "" {
		invalidParams = append(invalidParams, pkg.InvalidParam{
			Name:   idParamName,
			Reason: "param not provided",
		})
		return uuid.Nil, invalidParams
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		invalidParams = append(invalidParams, pkg.InvalidParam{
			Name:   idParamName,
			Reason: err.Error(),
		})
	}
	return id, invalidParams
}

func parseShuffled(r *http.Request) (bool, []pkg.InvalidParam) {
	shuffledParamName := "shuffled"
	var invalidParams []pkg.InvalidParam
	shuffled := false

	if !r.URL.Query().Has(shuffledParamName) {
		return shuffled, invalidParams
	}

	shuffledStr := r.URL.Query().Get(shuffledParamName)
	var err error
	shuffled, err = strconv.ParseBool(shuffledStr)
	if err != nil {
		invalidParams = append(invalidParams, pkg.InvalidParam{
			Name:   shuffledParamName,
			Reason: err.Error(),
		})
	}
	return shuffled, invalidParams
}

func parseCards(r *http.Request) ([]string, []pkg.InvalidParam) {
	cardsParamName := "cards"
	var cards []string
	var invalidParams []pkg.InvalidParam

	cardsStr := r.URL.Query().Get("cards")
	if cardsStr == "" {
		return cards, invalidParams
	}
	cards = strings.Split(cardsStr, ",")
	cardsByCode := generateAllCardsCombinationsByCode()
	for _, card := range cards {
		if _, ok := cardsByCode[card]; !ok {
			invalidParams = append(invalidParams, pkg.InvalidParam{
				Name:   cardsParamName,
				Reason: fmt.Sprintf("unrecognised card: %s", card),
			})
		}
	}
	return cards, invalidParams
}

func parseCount(r *http.Request) (int, []pkg.InvalidParam) {
	countParamName := "count"
	var invalidParams []pkg.InvalidParam

	countStr := r.URL.Query().Get(countParamName)
	if countStr == "" {
		invalidParams = append(invalidParams, pkg.InvalidParam{
			Name:   countParamName,
			Reason: "parameter missing",
		})
		return 0, invalidParams
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		invalidParams = append(invalidParams, pkg.InvalidParam{
			Name:   countParamName,
			Reason: err.Error(),
		})
	}
	if count < 1 {
		invalidParams = append(invalidParams, pkg.InvalidParam{
			Name:   countParamName,
			Reason: "count should be greater or equal to 1",
		})
	}
	return count, invalidParams
}
