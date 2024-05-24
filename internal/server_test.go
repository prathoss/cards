package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/prathoss/cards/pkg"
)

func TestParseCount(t *testing.T) {
	tests := []struct {
		name          string
		param         string
		expectedCount int
		expectedError bool
	}{
		{
			name:          "EmptyParameter",
			param:         "",
			expectedCount: 0,
			expectedError: true,
		},
		{
			name:          "ValidParameter",
			param:         "10",
			expectedCount: 10,
			expectedError: false,
		},
		{
			name:          "NonIntParameter",
			param:         "not-a-number",
			expectedCount: 0,
			expectedError: true,
		},
		{
			name:          "NegativeParameter",
			param:         "-5",
			expectedCount: -5,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/?count="+tt.param, nil)
			gotCount, gotInvalidParams := parseCount(req)

			if gotCount != tt.expectedCount {
				t.Errorf("parseCount() gotCount = %v, expectedCount = %v", gotCount, tt.expectedCount)
			}

			if (len(gotInvalidParams) > 0) != tt.expectedError {
				t.Errorf("parseCount() gotInvalidParams = %v, expectedError = %v", gotInvalidParams, tt.expectedError)
			}
		})
	}
}

func TestParseCards(t *testing.T) {
	tests := []struct {
		desc           string
		reqURL         string
		expectedCards  []string
		expectedErrors []pkg.InvalidParam
	}{
		{
			desc:           "Valid card set",
			reqURL:         "/?cards=KH,KD,KS,KC",
			expectedCards:  []string{"KH", "KD", "KS", "KC"},
			expectedErrors: nil,
		},
		{
			desc:          "Invalid card set",
			reqURL:        "/?cards=KH,XD,KS,YC",
			expectedCards: []string{"KH", "XD", "KS", "YC"},
			expectedErrors: []pkg.InvalidParam{
				{
					Name:   "cards",
					Reason: "unrecognised card: XD",
				},
				{
					Name:   "cards",
					Reason: "unrecognised card: YC",
				},
			},
		},
		{
			desc:           "Empty card parameter",
			reqURL:         "/?cards=",
			expectedCards:  []string{},
			expectedErrors: nil,
		},
		{
			desc:           "No card parameter",
			reqURL:         "/",
			expectedCards:  []string{},
			expectedErrors: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.reqURL, nil)
			cards, parseErrors := parseCards(req)

			if !slices.Equal(cards, tt.expectedCards) {
				t.Errorf("Expected cards %v, but got %v", tt.expectedCards, cards)
			}
			if !slices.Equal(parseErrors, tt.expectedErrors) {
				t.Errorf("Expected errors to be %v, but got %v", tt.expectedErrors, parseErrors)
			}
		})
	}
}

var _ DeckProcessor = (*DeckProcessorMock)(nil)

type DeckProcessorMock struct {
	storage map[uuid.UUID]*Deck
}

func (d *DeckProcessorMock) Create(_ context.Context, cardsCodes []string, shuffled bool) (Deck, error) {
	deck, err := NewDeck(cardsCodes, shuffled)
	if err != nil {
		return Deck{}, err
	}
	d.storage[deck.ID] = &deck
	return deck, nil
}

func (d *DeckProcessorMock) Get(_ context.Context, deckID uuid.UUID) (Deck, error) {
	deck, ok := d.storage[deckID]
	if !ok {
		return Deck{}, pkg.NewNotFoundError("deck not found")
	}
	return *deck, nil
}

func (d *DeckProcessorMock) DrawCards(ctx context.Context, deckID uuid.UUID, count int) ([]Card, error) {
	deck, err := d.Get(ctx, deckID)
	if err != nil {
		return nil, err
	}

	cards := deck.Cards[0:count]
	deck.Cards = deck.Cards[count:]
	return cards, nil
}

func TestDeck_Draw_Concurrency(t *testing.T) {
	s := &Server{
		config: Config{},
		deckProcessor: &DeckProcessorMock{
			storage: map[uuid.UUID]*Deck{},
		},
	}

	deck, err := s.deckProcessor.Create(context.Background(), nil, false)
	if err != nil {
		t.Fatal(err)
	}

	concurrencyDegree := 200
	wg := &sync.WaitGroup{}

	cardsReceived := atomic.Int32{}
	expectedError := newNotEnoughCardsError()
	nonExpectedErrorsCount := atomic.Int32{}
	for range concurrencyDegree {
		wg.Add(1)
		go func() {
			defer wg.Done()
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/deck/%s/draw?count=1", deck.ID.String()), nil)
			req.SetPathValue("id", deck.ID.String())

			response, err := s.drawCards(recorder, req)

			if err != nil {
				var badRequestError *pkg.BadRequestError
				if !errors.As(err, &badRequestError) && badRequestError.Error() != expectedError.Error() {
					nonExpectedErrorsCount.Add(1)
				}
				return
			}
			cardsResponse := response.(CardsResponse)
			cardsReceived.Add(int32(len(cardsResponse.Cards)))
		}()
	}
	wg.Wait()

	if nonExpectedErrorsCount.Load() > 0 {
		t.Fatalf("handler returns unexpected errors")
	}

	if cardsReceived.Load() > int32(len(deck.Cards)) {
		t.Fatalf("drew more cards than possible")
	}
}
