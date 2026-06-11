// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package raindrop // import "miniflux.app/v2/internal/integration/raindrop"

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"miniflux.app/v2/internal/integration/httputil"
)

type Client struct {
	token        string
	collectionID string
	tags         []string
}

func NewClient(token, collectionID, tags string) *Client {
	return &Client{token: token, collectionID: collectionID, tags: strings.Split(tags, ",")}
}

// https://developer.raindrop.io/v1/raindrops/single#create-raindrop
func (c *Client) CreateRaindrop(entryURL, entryTitle string) error {
	if c.token == "" {
		return errors.New("raindrop: missing token")
	}

	response, err := httputil.DoJSONRequest(http.MethodPost, "https://api.raindrop.io/rest/v1/raindrop", &raindrop{
		Link:       entryURL,
		Title:      entryTitle,
		Collection: collection{Id: c.collectionID},
		Tags:       c.tags,
	}, map[string]string{
		"Authorization": "Bearer " + c.token,
	})
	if err != nil {
		return fmt.Errorf("raindrop: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("raindrop: unable to create bookmark: status=%d", response.StatusCode)
	}

	return nil
}

type raindrop struct {
	Link       string     `json:"link"`
	Title      string     `json:"title"`
	Collection collection `json:"collection"`
	Tags       []string   `json:"tags"`
}

type collection struct {
	Id string `json:"$id"`
}
