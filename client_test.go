package scryfall

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestGetCardByID(t *testing.T) {
	t.Parallel()

	card := Card{ID: "abc-123", Name: "Test Card"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/cards/abc-123", r.URL.Path)
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(card))
	}))
	t.Cleanup(server.Close)

	client := NewClient(
		WithBaseURL(server.URL),
		WithLimiter(rate.NewLimiter(rate.Inf, 0)),
	)

	got, err := client.GetCardByID(context.Background(), "abc-123")
	require.NoError(t, err)
	require.Equal(t, card.ID, got.ID)
	require.Equal(t, card.Name, got.Name)
}

func TestListBulkData(t *testing.T) {
	t.Parallel()

	payload := map[string]any{
		"data": []CardBulkData{{ID: "bulk-1", Type: "default_cards"}},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/bulk-data", r.URL.Path)
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(payload))
	}))
	t.Cleanup(server.Close)

	client := NewClient(
		WithBaseURL(server.URL),
		WithLimiter(rate.NewLimiter(rate.Inf, 0)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)

	data, err := client.ListBulkData(ctx)
	require.NoError(t, err)
	require.Len(t, data, 1)
	require.Equal(t, "bulk-1", data[0].ID)
	require.Equal(t, "default_cards", data[0].Type)
}

func TestAPIErrorDecoding(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"details": "slow down",
			"type":    "rate_limited",
		})
	}))
	t.Cleanup(server.Close)

	client := NewClient(
		WithBaseURL(server.URL),
		WithLimiter(rate.NewLimiter(rate.Inf, 0)),
	)

	_, err := client.GetCardByID(context.Background(), "abc")
	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	require.Equal(t, http.StatusTooManyRequests, apiErr.StatusCode)
	require.Contains(t, apiErr.Error(), "slow down")
}

func TestDownloadBulkData(t *testing.T) {
	t.Parallel()

	// Create test data with multiple cards
	testCards := []Card{
		{ID: "card-1", Name: "Test Card 1", Prices: CardPrices{USD: "1.99"}},
		{ID: "card-2", Name: "Test Card 2", Prices: CardPrices{USD: "2.99"}},
		{ID: "card-3", Name: "Test Card 3", Prices: CardPrices{USD: "3.99"}},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(testCards))
	}))
	t.Cleanup(server.Close)

	client := NewClient(
		WithBaseURL(server.URL),
		WithLimiter(rate.NewLimiter(rate.Inf, 0)),
	)

	cards, err := client.DownloadBulkData(context.Background(), server.URL)
	require.NoError(t, err)
	require.Len(t, cards, 3)
	require.Equal(t, "card-1", cards[0].ID)
	require.Equal(t, "Test Card 1", cards[0].Name)
	require.Equal(t, "1.99", cards[0].Prices.USD)
}

func TestDownloadBulkData_EmptyURI(t *testing.T) {
	t.Parallel()

	client := NewClient()
	_, err := client.DownloadBulkData(context.Background(), "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "download URI is required")
}

func TestDownloadBulkData_HTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	client := NewClient(
		WithBaseURL(server.URL),
		WithLimiter(rate.NewLimiter(rate.Inf, 0)),
	)

	_, err := client.DownloadBulkData(context.Background(), server.URL)
	require.Error(t, err)
	require.Contains(t, err.Error(), "download failed with status 404")
}

func TestDownloadBulkDataStream_Progress(t *testing.T) {
	t.Parallel()

	cards := []Card{
		{ID: "card-1", Name: "Test Card 1"},
		{ID: "card-2", Name: "Test Card 2"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(cards))
	}))
	t.Cleanup(server.Close)

	client := NewClient(
		WithLimiter(rate.NewLimiter(rate.Inf, 0)),
	)

	var progressCalls int32
	var seen []string
	err := client.DownloadBulkDataStream(context.Background(), server.URL, func(card Card) error {
		seen = append(seen, card.ID)
		return nil
	}, func(current, total int64) {
		if current > 0 {
			atomic.AddInt32(&progressCalls, 1)
		}
		_ = total
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"card-1", "card-2"}, seen)
	require.NotZero(t, atomic.LoadInt32(&progressCalls))
}

func TestProcessBulkDataStream_InvalidPayload(t *testing.T) {
	t.Parallel()

	client := NewClient()
	err := client.ProcessBulkDataStream(bytes.NewBufferString("{}"), func(card Card) error {
		return nil
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected '['")
}
