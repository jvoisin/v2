// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package httputil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/version"
)

func TestDoJSONRequest(t *testing.T) {
	configureIntegrationAllowPrivateNetworksOption(t)

	var gotMethod, gotContentType, gotUserAgent, gotAuth, gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		gotUserAgent = r.Header.Get("User-Agent")
		gotAuth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	payload := map[string]string{"hello": "world"}
	response, err := DoJSONRequest(http.MethodPost, server.URL, payload, map[string]string{
		"Authorization": "Bearer secret",
	})
	if err != nil {
		t.Fatalf("DoJSONRequest returned an error: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, response.StatusCode)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	if gotContentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", gotContentType)
	}
	if want := "Miniflux/" + version.Version; gotUserAgent != want {
		t.Errorf("expected User-Agent %q, got %q", want, gotUserAgent)
	}
	if gotAuth != "Bearer secret" {
		t.Errorf("expected Authorization %q, got %q", "Bearer secret", gotAuth)
	}
	if gotBody != `{"hello":"world"}` {
		t.Errorf("expected body %q, got %q", `{"hello":"world"}`, gotBody)
	}
}

func TestDoJSONRequestWithInvalidEndpoint(t *testing.T) {
	_, err := DoJSONRequest(http.MethodPost, "://invalid", nil, nil)
	if err == nil {
		t.Fatal("expected an error for an invalid endpoint, got nil")
	}
}

func TestDoJSONRequestWithBody(t *testing.T) {
	configureIntegrationAllowPrivateNetworksOption(t)

	const requestBody = `{"hello":"world"}`

	var gotBody, gotContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	response, err := DoJSONRequestWithBody(http.MethodPost, server.URL, []byte(requestBody), nil)
	if err != nil {
		t.Fatalf("DoJSONRequestWithBody returned an error: %v", err)
	}
	defer response.Body.Close()

	if gotContentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", gotContentType)
	}
	if gotBody != requestBody {
		t.Errorf("expected body %q, got %q", requestBody, gotBody)
	}
}

func configureIntegrationAllowPrivateNetworksOption(t *testing.T) {
	t.Helper()

	t.Setenv("INTEGRATION_ALLOW_PRIVATE_NETWORKS", "1")

	configParser := config.NewConfigParser()
	parsedOptions, err := configParser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Unable to configure test options: %v", err)
	}

	previousOptions := config.Opts
	config.Opts = parsedOptions
	t.Cleanup(func() {
		config.Opts = previousOptions
	})
}
