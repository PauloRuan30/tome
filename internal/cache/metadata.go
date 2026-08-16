package cache

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{httpClient: &http.Client{Timeout: 10 * time.Second}}
}

type Result struct {
	Title  string
	Author string
}

type olResponse struct {
	Docs []struct {
		Title  string   `json:"title"`
		Author []string `json:"author_name"`
	} `json:"docs"`
}

// Search queries the Open Library Search API for canonical metadata.
func (c *Client) Search(query string) (*Result, error) {
	u := fmt.Sprintf("https://openlibrary.org/search.json?q=%s&limit=1", url.QueryEscape(query))
	resp, err := c.httpClient.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ol olResponse
	if err := json.NewDecoder(resp.Body).Decode(&ol); err != nil {
		return nil, err
	}
	if len(ol.Docs) == 0 {
		return nil, fmt.Errorf("no results for %q", query)
	}

	res := &Result{Title: ol.Docs[0].Title}
	if len(ol.Docs[0].Author) > 0 {
		res.Author = strings.Join(ol.Docs[0].Author, ", ")
	}
	return res, nil
}
