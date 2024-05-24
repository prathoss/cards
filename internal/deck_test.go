package internal

import (
	"slices"
	"testing"
)

func TestCard_Code(t *testing.T) {
	tests := []struct {
		name string
		card Card
		want string
	}{
		{
			name: "NumericValueAndSingleCharSuit",
			card: Card{
				Value: CardValueTen,
				Suit:  CardSuitSpades,
			},
			want: "10S",
		},
		{
			name: "AlphabeticSingleCharValueAndSuit",
			card: Card{
				Value: CardValueAce,
				Suit:  CardSuitHearths,
			},
			want: "AH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.card.Code(); got != tt.want {
				t.Errorf("Card.Code() = %v, want %v", got, tt.want)
			}
		})
	}
}

func cardsToCodes(cards []Card) []string {
	resultCodes := make([]string, 0, len(cards))
	for _, c := range cards {
		resultCodes = append(resultCodes, c.Code())
	}
	return resultCodes
}

func TestGenerateCards(t *testing.T) {
	tests := []struct {
		name  string
		codes []string
		want  []string
	}{
		{
			name:  "generate with multiple codes",
			codes: []string{"AS", "2H", "3D", "4C"},
			want:  []string{"AS", "2H", "3D", "4C"},
		},
		{
			name:  "generate with single code",
			codes: []string{"AS"},
			want:  []string{"AS"},
		},
		{
			name:  "generate with empty code slice",
			codes: []string{},
			want:  cardsToCodes(generateAllCardsCombinations()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateCards(tt.codes)
			resultCodes := cardsToCodes(result)
			if !slices.Equal(resultCodes, tt.want) {
				t.Fatalf("got %v, want %v", resultCodes, tt.want)
			}

		})
	}
}

func TestShuffleCards(t *testing.T) {
	tests := []struct {
		name  string
		cards []Card
	}{
		{
			name:  "EmptyDeck",
			cards: []Card{},
		},
		{
			name: "SingleCardDeck",
			cards: []Card{
				{Suit: CardSuitHearths, Value: CardValueAce},
			},
		},
		{
			name: "MultipleCardDeck",
			cards: []Card{
				{Suit: CardSuitClubs, Value: CardValueAce},
				{Suit: CardSuitHearths, Value: CardValueTen},
				{Suit: CardSuitSpades, Value: CardValueNine},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deck := &Deck{Cards: tt.cards}
			copyBeforeShuffle := make([]Card, len(tt.cards))
			copy(copyBeforeShuffle, tt.cards)
			if err := deck.ShuffleCards(); err != nil {
				t.Fatalf("could not shuffle deck %s", err.Error())
			}

			if len(tt.cards) > 2 {
				if slices.Equal(deck.Cards, copyBeforeShuffle) {
					t.Errorf("Deck of cards was not shuffled. Initial and final states are the same.")
				}
			} else {
				if !slices.Equal(deck.Cards, copyBeforeShuffle) {
					t.Errorf("Deck of cards with zero or one card should stay the same after shuffle.")
				}
			}
		})
	}
}
