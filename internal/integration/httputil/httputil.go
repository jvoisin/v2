// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package httputil // import "miniflux.app/v2/internal/integration/httputil"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"miniflux.app/v2/internal/http/client"
	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second

// DoJSONRequest marshals payload as a JSON request body and sends it to the
// given endpoint using the shared integration HTTP client. It always sets the
// Content-Type and User-Agent headers; entries from extraHeaders are applied on
// top of them (for authentication or other integration-specific headers).
//
// The caller owns the returned response: it must close the body and is
// responsible for inspecting the status code.
func DoJSONRequest(method, endpoint string, payload any, extraHeaders map[string]string) (*http.Response, error) {
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("unable to encode request body: %w", err)
	}

	return DoJSONRequestWithBody(method, endpoint, requestBody, extraHeaders)
}

// DoJSONRequestWithBody sends a pre-marshaled JSON request body to the given
// endpoint using the shared integration HTTP client.
func DoJSONRequestWithBody(method, endpoint string, requestBody []byte, extraHeaders map[string]string) (*http.Response, error) {
	request, err := http.NewRequest(method, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)
	for key, value := range extraHeaders {
		request.Header.Set(key, value)
	}

	response, err := client.NewClientWithOptions(client.Options{Timeout: defaultClientTimeout}).Do(request)
	if err != nil {
		return nil, fmt.Errorf("unable to send request: %w", err)
	}

	return response, nil
}
