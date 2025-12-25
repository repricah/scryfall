package scryfall

import "context"

// ClientAPI exposes the Scryfall client methods used by downstream services.
// It enables testing with lightweight fakes without pulling in extra deps.
type ClientAPI interface {
	GetCardByID(ctx context.Context, id string) (*Card, error)
	ListBulkData(ctx context.Context) ([]CardBulkData, error)
	ListSets(ctx context.Context) ([]CardSet, error)
	GetBulkDataByType(ctx context.Context, bulkType string) (*CardBulkData, error)
	DownloadBulkDataStream(ctx context.Context, downloadURI string, cardCallback func(Card) error, progressFn ProgressFunc) error
	DownloadBulkData(ctx context.Context, downloadURI string) ([]Card, error)
}
