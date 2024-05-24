package internal

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"

	"github.com/google/uuid"
	"github.com/prathoss/cards/pkg"
)

const (
	CardSuitClubs    = "CLUBS"
	CardSuitDiamonds = "DIAMONDS"
	CardSuitHearths  = "HEARTS"
	CardSuitSpades   = "SPADES"

	CardValueAce   = "ACE"
	CardValueTwo   = "2"
	CardValueThree = "3"
	CardValueFour  = "4"
	CardValueFive  = "5"
	CardValueSix   = "6"
	CardValueSeven = "7"
	CardValueEight = "8"
	CardValueNine  = "9"
	CardValueTen   = "10"
	CardValueJack  = "JACK"
	CardValueQueen = "QUEEN"
	CardValueKing  = "KING"
)

type Card struct {
	Value string `json:"value"`
	Suit  string `json:"suit"`
}

func (c Card) Code() string {
	v := ""
	s := ""
	if _, err := strconv.Atoi(c.Value); err == nil {
		v = c.Value
	} else if len(c.Value) > 0 {
		v = string(c.Value[0])
	}
	if len(c.Suit) > 0 {
		s = string(c.Suit[0])
	}
	return fmt.Sprintf("%s%s", v, s)
}

type Deck struct {
	ID       uuid.UUID `json:"deck_id"`
	Shuffled bool      `json:"shuffled"`
	Cards    []Card    `json:"cards"`
}

func NewDeck(cardCodes []string, shuffled bool) (Deck, error) {
	cards := generateCards(cardCodes)
	deck := Deck{
		ID:       uuid.New(),
		Shuffled: shuffled,
		Cards:    cards,
	}
	if shuffled {
		if err := deck.ShuffleCards(); err != nil {
			return deck, err
		}
	}
	return deck, nil
}

// ShuffleCards shuffles cards in deck (in place) using Fisher-Yates' algorithm
func (d *Deck) ShuffleCards() error {
	for i := len(d.Cards) - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i)+1))
		if err != nil {
			return err
		}
		j := int(jBig.Int64())
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	}
	return nil
}

// generateCards generates a slice of cards based on the provided codes.
// If the codes slice is empty, it generates all possible combinations of cards.
// It uses generateAllCardsCombinationsByCode to get the mapping of codes to cards.
// Unknown card codes will be ignored
func generateCards(codes []string) []Card {
	if len(codes) == 0 {
		return generateAllCardsCombinations()
	}
	cardsByCode := generateAllCardsCombinationsByCode()
	cards := make([]Card, 0, len(codes))
	for _, code := range codes {
		if card, ok := cardsByCode[code]; ok {
			cards = append(cards, card)
		}
	}
	return cards
}

// generateAllCardsCombinationsByCode generates a mapping of card codes to cards.
// It uses generateAllCardsCombinations to get all possible combinations of cards.
// The resulting map maps each code to its corresponding card.
func generateAllCardsCombinationsByCode() map[string]Card {
	cards := generateAllCardsCombinations()
	cardsByCode := make(map[string]Card)
	for _, c := range cards {
		cardsByCode[c.Code()] = c
	}
	return cardsByCode
}

// generateAllCardsCombinations generates all possible combinations of cards.
// It creates cards for each possible suit and value combination and returns them as a slice.
func generateAllCardsCombinations() []Card {
	suits := []string{CardSuitClubs, CardSuitDiamonds, CardSuitHearths, CardSuitSpades}
	values := []string{CardValueAce, CardValueTwo, CardValueThree, CardValueFour, CardValueFive, CardValueSix, CardValueSeven, CardValueEight, CardValueNine, CardValueTen, CardValueJack, CardValueQueen, CardValueKing}
	cards := make([]Card, 0, len(suits)*len(values))
	for _, suit := range suits {
		for _, value := range values {
			cards = append(cards, Card{Value: value, Suit: suit})
		}
	}
	return cards
}

type CreateDeckResponse struct {
	ID        uuid.UUID `json:"deck_id"`
	Shuffled  bool      `json:"shuffled"`
	Remaining int       `json:"remaining"`
}

func NewCreateDeckResponse(deck Deck) CreateDeckResponse {
	return CreateDeckResponse{
		ID:        deck.ID,
		Shuffled:  deck.Shuffled,
		Remaining: len(deck.Cards),
	}
}

type OpenDeckResponse struct {
	CreateDeckResponse
	Cards []CardResponse `json:"cards"`
}

func NewOpenDeckResponse(deck Deck) OpenDeckResponse {
	cardsResponse := NewCardsResponse(deck.Cards)
	return OpenDeckResponse{
		CreateDeckResponse: NewCreateDeckResponse(deck),
		Cards:              cardsResponse.Cards,
	}
}

type CardsResponse struct {
	Cards []CardResponse `json:"cards"`
}

func NewCardsResponse(cards []Card) CardsResponse {
	cardsResponse := make([]CardResponse, 0, len(cards))
	for _, card := range cards {
		cardsResponse = append(cardsResponse, NewCardResponse(card))
	}
	return CardsResponse{
		Cards: cardsResponse,
	}
}

type CardResponse struct {
	Card
	Code string `json:"code"`
}

func NewCardResponse(c Card) CardResponse {
	return CardResponse{
		Card: c,
		Code: c.Code(),
	}
}

type DeckProcessor interface {
	Create(ctx context.Context, cardsCodes []string, shuffled bool) (Deck, error)
	Get(ctx context.Context, deckID uuid.UUID) (Deck, error)
	DrawCards(ctx context.Context, deckID uuid.UUID, count int) ([]Card, error)
}

func newNotEnoughCardsError() *pkg.BadRequestError {
	return pkg.NewBadRequestError(pkg.InvalidParam{
		Name:   "deck",
		Reason: "deck does not have enough cards",
	})
}
