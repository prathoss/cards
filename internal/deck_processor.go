package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/prathoss/cards/pkg"
	"go.mongodb.org/mongo-driver/mongo"
)

type DeckProcessor interface {
	Create(ctx context.Context, cardsCodes []string, shuffled bool) (Deck, error)
	Get(ctx context.Context, deckID uuid.UUID) (Deck, error)
	DrawCards(ctx context.Context, deckID uuid.UUID, count int) ([]Card, error)
}

var _ DeckProcessor = (*DeckRepository)(nil)

type DeckRepository struct {
	db *mongo.Collection
}

func NewDeckRepository(client *mongo.Client) *DeckRepository {
	return &DeckRepository{
		db: client.Database("cards").Collection("decks"),
	}
}

func (d *DeckRepository) Create(ctx context.Context, cardsCodes []string, shuffled bool) (Deck, error) {
	deck, err := NewDeck(cardsCodes, shuffled)
	if err != nil {
		return Deck{}, err
	}

	_, err = d.db.InsertOne(ctx, deck)
	if err != nil {
		return Deck{}, err
	}
	return deck, nil
}

func (d *DeckRepository) Get(ctx context.Context, deckID uuid.UUID) (Deck, error) {
	var deck Deck
	err := d.db.FindOne(ctx, Deck{
		ID: deckID,
	}).Decode(&deck)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Deck{}, pkg.NewNotFoundError(fmt.Sprintf("deck with ID %s not found", deckID))
		}
		return Deck{}, err
	}

	return deck, nil
}

func (d *DeckRepository) DrawCards(ctx context.Context, deckID uuid.UUID, count int) ([]Card, error) {
	deck, err := d.Get(ctx, deckID)
	if err != nil {
		return nil, err
	}

	cards, err := deck.DrawCards(count)
	if err != nil {
		return nil, err
	}

	_, err = d.db.ReplaceOne(ctx, Deck{ID: deck.ID}, deck)
	if err != nil {
		return nil, err
	}
	return cards, nil
}
