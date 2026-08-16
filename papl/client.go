package papl

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// DefaultUserAgent identifies the client to papl.cs.brown.edu.
const DefaultUserAgent = "papl/dev (+https://github.com/tamnd/papl-cli)"

// Client makes HTTP requests to the PAPL site with pacing and retries.
type Client struct {
	http  *http.Client
	delay time.Duration
	mu    sync.Mutex
	last  time.Time
}

// NewClient returns a Client with the given delay and timeout.
// Uses InsecureSkipVerify to handle institutional TLS certificates.
func NewClient(delay, timeout time.Duration) *Client {
	transport := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 90 * time.Second,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
	}
	return &Client{
		http: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		delay: delay,
	}
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if wait := c.delay - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

// FetchPage fetches a URL and returns the body bytes.
func (c *Client) FetchPage(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= 3; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoffDur(attempt)):
			}
		}
		body, retry, err := c.do(ctx, rawURL)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("fetch %s: %w", rawURL, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) (body []byte, retry bool, err error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", DefaultUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d fetching %s", resp.StatusCode, rawURL)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d fetching %s", resp.StatusCode, rawURL)
	}

	b, err := io.ReadAll(io.LimitReader(resp.Body, 16*1024*1024))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func backoffDur(attempt int) time.Duration {
	d := time.Duration(1<<uint(attempt-1)) * time.Second
	if d > 8*time.Second {
		d = 8 * time.Second
	}
	return d
}
