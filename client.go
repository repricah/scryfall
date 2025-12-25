package scryfall

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"
	"golang.org/x/time/rate"
)

const (
	defaultBaseURL           = "https://api.scryfall.com"
	defaultUserAgent         = "repricah-scryfall/0.1"
	defaultTimeout           = 15 * time.Second
	defaultRequestsPerSecond = 10
)

// Client interacts with the public Scryfall API while enforcing basic rate
// limiting and structured logging.
type Client struct {
	httpClient *http.Client
	baseURL    *url.URL
	limiter    *rate.Limiter
	userAgent  string
	logger     *log.Logger
}

// Option configures the Scryfall client.
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client instance.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// WithBaseURL overrides the API base URL.
func WithBaseURL(rawURL string) Option {
	return func(c *Client) {
		if rawURL == "" {
			return
		}
		parsed, err := url.Parse(rawURL)
		if err != nil {
			c.logger.Warn("invalid scryfall base url supplied", "base_url", rawURL, "error", err)
			return
		}
		c.baseURL = parsed
	}
}

// WithLimiter injects a custom rate limiter configuration.
func WithLimiter(limiter *rate.Limiter) Option {
	return func(c *Client) {
		if limiter != nil {
			c.limiter = limiter
		}
	}
}

// WithUserAgent overrides the default user agent header.
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		if ua != "" {
			c.userAgent = ua
		}
	}
}

// WithLogger sets a custom structured logger.
func WithLogger(logger *log.Logger) Option {
	return func(c *Client) {
		if logger != nil {
			c.logger = logger
		}
	}
}

// NewClient constructs a Scryfall API client with sane defaults.
func NewClient(opts ...Option) *Client {
	base, _ := url.Parse(defaultBaseURL)
	c := &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		baseURL:    base,
		limiter:    rate.NewLimiter(rate.Limit(defaultRequestsPerSecond), defaultRequestsPerSecond),
		userAgent:  defaultUserAgent,
		logger:     log.WithPrefix("scryfall"),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// GetCardByID retrieves an individual card using its Scryfall UUID.
func (c *Client) GetCardByID(ctx context.Context, id string) (*Card, error) {
	if id == "" {
		return nil, fmt.Errorf("card id is required")
	}
	endpoint := fmt.Sprintf("/cards/%s", id)
	var card Card
	if err := c.get(ctx, endpoint, &card); err != nil {
		return nil, err
	}
	return &card, nil
}

// ListBulkData fetches metadata about Scryfall bulk data exports.
func (c *Client) ListBulkData(ctx context.Context) ([]CardBulkData, error) {
	var response struct {
		Data []CardBulkData `json:"data"`
	}
	if err := c.get(ctx, "/bulk-data", &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

// ListSets fetches all sets from Scryfall.
func (c *Client) ListSets(ctx context.Context) ([]CardSet, error) {
	var response struct {
		Data []CardSet `json:"data"`
	}
	if err := c.get(ctx, "/sets", &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

// ProgressFunc is a callback for tracking progress in bytes.
type ProgressFunc func(current, total int64)

// GetBulkDataByType retrieves a single bulk data object by its type.
func (c *Client) GetBulkDataByType(ctx context.Context, bulkType string) (*CardBulkData, error) {
	if bulkType == "" {
		return nil, fmt.Errorf("bulk type is required")
	}
	path := fmt.Sprintf("/bulk-data/%s", bulkType)
	var bulkData CardBulkData
	if err := c.get(ctx, path, &bulkData); err != nil {
		return nil, err
	}
	return &bulkData, nil
}

// DownloadBulkDataStream downloads and parses a bulk data file from Scryfall using streaming.
// It calls the provided callback for each card object encountered.
// progressFn, if provided, will be called periodically with the number of bytes read.
func (c *Client) DownloadBulkDataStream(ctx context.Context, downloadURI string, cardCallback func(Card) error, progressFn ProgressFunc) error {
	if downloadURI == "" {
		return fmt.Errorf("download URI is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	c.logger.Info("downloading bulk data (streaming)", "uri", downloadURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURI, http.NoBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("perform request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	var reader io.Reader = resp.Body
	if progressFn != nil {
		reader = &progressReader{
			ReadCloser: resp.Body,
			Total:      resp.ContentLength,
			OnRead:     progressFn,
		}
	}

	return c.ProcessBulkDataStream(reader, cardCallback)
}

// DownloadToFile downloads a bulk data file to a local file path with progress tracking.
func (c *Client) DownloadToFile(ctx context.Context, downloadURI string, filePath string, progress ProgressFunc) error {
	if downloadURI == "" {
		return fmt.Errorf("download URI is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURI, http.NoBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("perform request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(filepath.Clean(filePath)) // #nosec G304
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer func() {
		_ = out.Close()
	}()

	var reader io.Reader = resp.Body
	if progress != nil {
		reader = &progressReader{
			ReadCloser: resp.Body,
			Total:      resp.ContentLength,
			OnRead:     progress,
		}
	}

	if _, err := io.Copy(out, reader); err != nil {
		return fmt.Errorf("copy to file: %w", err)
	}

	return nil
}

// ProcessBulkDataStream handles the streaming JSON parsing from an io.Reader.
func (c *Client) ProcessBulkDataStream(reader io.Reader, cardCallback func(Card) error) error {
	dec := json.NewDecoder(reader)

	// Read opening bracket
	t, err := dec.Token()
	if err != nil {
		return fmt.Errorf("decode opening bracket: %w", err)
	}
	if delim, ok := t.(json.Delim); !ok || delim != '[' {
		return fmt.Errorf("expected '[' at start of bulk data")
	}

	for dec.More() {
		var card Card
		if err := dec.Decode(&card); err != nil {
			return fmt.Errorf("decode card object: %w", err)
		}
		if err := cardCallback(card); err != nil {
			return err
		}
	}

	// Read closing bracket
	t, err = dec.Token()
	if err != nil {
		return fmt.Errorf("decode closing bracket: %w", err)
	}
	if delim, ok := t.(json.Delim); !ok || delim != ']' {
		return fmt.Errorf("expected ']' at end of bulk data")
	}

	return nil
}

type progressReader struct {
	io.ReadCloser
	Total   int64
	Current int64
	OnRead  ProgressFunc
}

func (r *progressReader) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	r.Current += int64(n)
	if r.OnRead != nil {
		r.OnRead(r.Current, r.Total)
	}
	return n, err
}

// DownloadBulkData is a legacy wrapper around DownloadBulkDataStream that loads everything into memory.
// Deprecated: Use DownloadBulkDataStream for large datasets.
func (c *Client) DownloadBulkData(ctx context.Context, downloadURI string) ([]Card, error) {
	var cards []Card
	err := c.DownloadBulkDataStream(ctx, downloadURI, func(card Card) error {
		cards = append(cards, card)
		return nil
	}, nil)
	return cards, err
}

func (c *Client) get(ctx context.Context, path string, dest any) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := c.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("wait for rate limiter: %w", err)
	}

	rel, err := url.Parse(path)
	if err != nil {
		return fmt.Errorf("invalid path %q: %w", path, err)
	}
	fullURL := c.baseURL.ResolveReference(rel)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL.String(), http.NoBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	c.logger.Debug("scryfall api request", "method", req.Method, "url", fullURL.String())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("perform request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		apiErr, readErr := decodeAPIError(resp.Body)
		if readErr != nil {
			return fmt.Errorf("scryfall error status %d: %w", resp.StatusCode, readErr)
		}
		apiErr.StatusCode = resp.StatusCode
		return apiErr
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// APIError represents an error returned by the Scryfall API.
type APIError struct {
	StatusCode int
	Details    string   `json:"details"`
	Type       string   `json:"type"`
	Warnings   []string `json:"warnings"`
}

func (e *APIError) Error() string {
	if e == nil {
		return "scryfall api error"
	}
	if e.Details != "" {
		return fmt.Sprintf("scryfall api error (%d): %s", e.StatusCode, e.Details)
	}
	return fmt.Sprintf("scryfall api error (%d)", e.StatusCode)
}

func decodeAPIError(r io.Reader) (*APIError, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return &APIError{}, nil
	}
	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err != nil {
		return nil, err
	}
	return &apiErr, nil
}
