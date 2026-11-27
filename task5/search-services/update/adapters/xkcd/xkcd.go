package xkcd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"yadro.com/course/update/core"
)

type Client struct {
	client http.Client
	url    string
}

func NewClient(url string, timeout time.Duration) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("empty base url specified")
	}
	return &Client{
		client: http.Client{Timeout: timeout},
		url:    url,
	}, nil
}

func (c Client) Get(ctx context.Context, id int) (core.XKCDInfo, error) {
	if id < 0 {
		return core.XKCDInfo{}, core.ErrBadArguments
	}
	reqUrl, err := url.JoinPath(c.url, strconv.Itoa(id), "info.0.json")
	if err != nil {
		return core.XKCDInfo{}, fmt.Errorf("build XKCD URL: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl, nil)
	if err != nil {
		return core.XKCDInfo{}, fmt.Errorf("build XKCD request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return core.XKCDInfo{}, fmt.Errorf("error while making request: %w", err)
	}
	defer resp.Body.Close()
	respCode := resp.StatusCode
	if respCode == http.StatusNotFound {
		return core.XKCDInfo{}, core.ErrNotFound
	}
	if respCode != http.StatusOK {
		return core.XKCDInfo{}, fmt.Errorf("fetch XKCD comic %d: HTTP %s", id, resp.Status)
	}

	var data struct {
		Num   int    `json:"num"`
		Img   string `json:"img"`
		Title string `json:"title"`
		Alt   string `json:"alt"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return core.XKCDInfo{}, fmt.Errorf("decode XKCD comic %d: %w", id, err)
	}

	return core.XKCDInfo{
		ID:          data.Num,
		URL:         data.Img,
		Title:       data.Title,
		Description: data.Alt,
	}, nil
}

func (c Client) LastID(ctx context.Context) (int, error) {
	reqUrl, err := url.JoinPath(c.url, "info.0.json")
	if err != nil {
		return 0, fmt.Errorf("build XKCD URL: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl, nil)
	if err != nil {
		return 0, fmt.Errorf("error while building request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error while making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("fetch latest XKCD comic: HTTP %s", resp.Status)
	}

	var data struct {
		Num int `json:"num"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("decode latest XKCD comic: %w", err)
	}
	return data.Num, nil
}
