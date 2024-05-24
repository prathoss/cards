package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/prathoss/cards/pkg"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Server struct {
	config            Config
	deckProcessor     DeckProcessor
	drawingCardsMutex sync.Mutex
}

func NewServer(config Config) (*Server, error) {
	ctx, cFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cFunc()

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(config.MongoConnection).SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Server{
		config:            config,
		deckProcessor:     NewDeckRepository(client),
		drawingCardsMutex: sync.Mutex{},
	}, nil
}

func (s *Server) createDeck(_ http.ResponseWriter, r *http.Request) (any, error) {
	var invalidParams []pkg.InvalidParam

	shuffled, shuffledErrors := parseShuffled(r)
	invalidParams = append(invalidParams, shuffledErrors...)

	cards, cardsErrors := parseCards(r)
	invalidParams = append(invalidParams, cardsErrors...)

	if len(invalidParams) > 0 {
		return nil, pkg.NewBadRequestError(invalidParams...)
	}

	deck, err := s.deckProcessor.Create(r.Context(), cards, shuffled)
	if err != nil {
		return nil, err
	}
	return NewCreateDeckResponse(deck), nil
}

func (s *Server) openDeck(_ http.ResponseWriter, r *http.Request) (any, error) {
	var invalidParams []pkg.InvalidParam

	id, idErrors := parseID(r)
	invalidParams = append(invalidParams, idErrors...)

	if len(invalidParams) > 0 {
		return nil, pkg.NewBadRequestError(invalidParams...)
	}

	deck, err := s.deckProcessor.Get(r.Context(), id)
	if err != nil {
		return nil, err
	}
	return NewOpenDeckResponse(deck), nil
}

func (s *Server) drawCards(_ http.ResponseWriter, r *http.Request) (any, error) {
	var invalidParams []pkg.InvalidParam

	id, idErrors := parseID(r)
	invalidParams = append(invalidParams, idErrors...)

	count, countErrors := parseCount(r)
	invalidParams = append(invalidParams, countErrors...)

	if len(invalidParams) > 0 {
		return nil, pkg.NewBadRequestError(invalidParams...)
	}

	s.drawingCardsMutex.Lock()
	defer s.drawingCardsMutex.Unlock()

	cards, err := s.deckProcessor.DrawCards(r.Context(), id, count)
	if err != nil {
		return nil, err
	}
	return NewCardsResponse(cards), nil
}

func (s *Server) Run() {
	mux := http.NewServeMux()

	mux.Handle("POST /api/v1/deck", pkg.HttpHandler(s.createDeck))
	mux.Handle("POST /api/v1/deck/{id}/open", pkg.HttpHandler(s.openDeck))
	mux.Handle("POST /api/v1/deck/{id}/draw", pkg.HttpHandler(s.drawCards))

	server := &http.Server{
		Addr:              s.config.Address,
		Handler:           pkg.CorrelationHandler(pkg.LoggingHandler(mux)),
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

	if !r.URL.Query().Has(shuffledParamName) {
		return false, invalidParams
	}

	shuffledStr := r.URL.Query().Get(shuffledParamName)
	shuffled, err := strconv.ParseBool(shuffledStr)
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
