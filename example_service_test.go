package scryfall_test

import (
	"context"
	"fmt"

	"github.com/repricah/scryfall"
)

type CardLookupService struct {
	client scryfall.ClientAPI
}

func NewCardLookupService(client scryfall.ClientAPI) *CardLookupService {
	return &CardLookupService{client: client}
}

func (s *CardLookupService) CardName(ctx context.Context, id string) (string, error) {
	card, err := s.client.GetCardByID(ctx, id)
	if err != nil {
		return "", err
	}
	return card.Name, nil
}

type fakeClient struct{
	card *scryfall.Card
}

func (f fakeClient) GetCardByID(ctx context.Context, id string) (*scryfall.Card, error) {
	return f.card, nil
}

func (f fakeClient) ListBulkData(ctx context.Context) ([]scryfall.CardBulkData, error) {
	return nil, nil
}

func (f fakeClient) ListSets(ctx context.Context) ([]scryfall.CardSet, error) {
	return nil, nil
}

func (f fakeClient) GetBulkDataByType(ctx context.Context, bulkType string) (*scryfall.CardBulkData, error) {
	return nil, nil
}

func (f fakeClient) DownloadBulkDataStream(ctx context.Context, downloadURI string, cardCallback func(scryfall.Card) error, progressFn scryfall.ProgressFunc) error {
	return nil
}

func (f fakeClient) DownloadBulkData(ctx context.Context, downloadURI string) ([]scryfall.Card, error) {
	return nil, nil
}

func ExampleCardLookupService() {
	service := NewCardLookupService(fakeClient{card: &scryfall.Card{Name: "Black Lotus"}})
	name, _ := service.CardName(context.Background(), "abc-123")
	fmt.Println(name)

	// Output: Black Lotus
}
