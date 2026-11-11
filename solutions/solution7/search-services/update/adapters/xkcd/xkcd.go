// Package xkcd is an adapter to xkcd site
package xkcd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"yadro.com/course/closers"
	"yadro.com/course/update/core"
)

const (
	lastPath   = "/info.0.json"
	maxRetries = 5
	backoff    = 1 * time.Second
)

type Client struct {
	log    *slog.Logger
	client http.Client
	url    string
}

func NewClient(url string, timeout time.Duration, log *slog.Logger) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("empty base url specified")
	}
	return &Client{
		client: http.Client{Timeout: timeout},
		log:    log,
		url:    url,
	}, nil
}

func (c Client) Get(ctx context.Context, id int) (core.XKCDInfo, error) {
	return c.get(ctx, fmt.Sprintf("%s/%d/%s", c.url, id, lastPath))
}

func (c Client) LastID(ctx context.Context) (int, error) {
	comics, err := c.get(ctx, c.url+lastPath)
	if err != nil {
		return 0, err
	}
	return comics.ID, nil
}

func (c Client) get(ctx context.Context, url string) (core.XKCDInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return core.XKCDInfo{}, fmt.Errorf("failed to create request: %v", err)
	}

	var resp *http.Response
	var attempts int
	for {
		resp, err = c.client.Do(req)
		if err == nil {
			break
		}
		attempts++
		if attempts == maxRetries {
			return core.XKCDInfo{}, fmt.Errorf("failed to request comics: %v", err)
		}
		c.log.Error("failed to connect, sleeping and retrying", "error", err)
		time.Sleep(backoff)
	}

	defer closers.CloseOrLog(resp.Body, c.log)
	if resp.StatusCode == http.StatusNotFound {
		return core.XKCDInfo{}, core.ErrNotFound
	}
	info := struct {
		ID         int    `json:"num"`
		URL        string `json:"img"`
		Title      string `json:"title"`
		SafeTitle  string `json:"safe_title"`
		Transcript string `json:"transcript"`
		Alt        string `json:"alt"`
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return core.XKCDInfo{}, fmt.Errorf("failed to decode comics: %v", err)
	}

	return core.XKCDInfo{
		ID:  info.ID,
		URL: info.URL,
		Description: strings.Join(
			[]string{info.Title, info.SafeTitle, info.Transcript, info.Alt},
			" ",
		),
	}, nil
}
