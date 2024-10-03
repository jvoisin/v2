// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	"net/http"
	"runtime"

	"miniflux.app/v2/internal/http/response/json"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/version"
	"miniflux.app/v2/internal/worker"
)

type handler struct {
	store  *storage.Storage
	pool   *worker.Pool
	router *http.ServeMux
}

// Serve declares API routes for the application.
func Serve(router *http.ServeMux, store *storage.Storage, pool *worker.Pool) {
	handler := &handler{store, pool, router}

	//sr := router.PathPrefix("/v1").Subrouter()
	middleware := newMiddleware(store)
	sr.Use(middleware.handleCORS)
	sr.Use(middleware.apiKeyAuth)
	sr.Use(middleware.basicAuth)
	sr.Methods(http.MethodOptions)
	router.HandleFunc("POST /v1/users", handler.createUser)
	router.HandleFunc("GET /v1/users", handler.users)
	router.HandleFunc("GET /v1/users/{userID:[0-9]+}", handler.userByID)
	router.HandleFunc("PUT /v1/users/{userID:[0-9]+}", handler.updateUser)
	router.HandleFunc("DELETE /v1/users/{userID:[0-9]+}", handler.removeUser)
	router.HandleFunc("PUT /v1/users/{userID:[0-9]+}/mark-all-as-read", handler.markUserAsRead)
	router.HandleFunc("GET /v1/users/{username}", handler.userByUsername)
	router.HandleFunc("GET /v1/me", handler.currentUser)
	router.HandleFunc("POST /v1/categories", handler.createCategory)
	router.HandleFunc("GET /v1/categories", handler.getCategories)
	router.HandleFunc("PUT /v1/categories/{categoryID}", handler.updateCategory)
	router.HandleFunc("DELETE /v1/categories/{categoryID}", handler.removeCategory)
	router.HandleFunc("PUT /v1/categories/{categoryID}/mark-all-as-read", handler.markCategoryAsRead)
	router.HandleFunc("GET /v1/categories/{categoryID}/feeds", handler.getCategoryFeeds)
	router.HandleFunc("PUT /v1/categories/{categoryID}/refresh", handler.refreshCategory)
	router.HandleFunc("GET /v1/categories/{categoryID}/entries", handler.getCategoryEntries)
	router.HandleFunc("GET /v1/categories/{categoryID}/entries/{entryID}", handler.getCategoryEntry)
	router.HandleFunc("POST /v1/discover", handler.discoverSubscriptions)
	router.HandleFunc("POST /v1/feeds", handler.createFeed)
	router.HandleFunc("GET /v1/feeds", handler.getFeeds)
	router.HandleFunc("GET /v1/feeds/counters", handler.fetchCounters)
	router.HandleFunc("PUT /v1/feeds/refresh", handler.refreshAllFeeds)
	router.HandleFunc("PUT /v1/feeds/{feedID}/refresh", handler.refreshFeed)
	router.HandleFunc("GET /v1/feeds/{feedID}", handler.getFeed)
	router.HandleFunc("PUT /v1/feeds/{feedID}", handler.updateFeed)
	router.HandleFunc("DELETE /v1/feeds/{feedID}", handler.removeFeed)
	router.HandleFunc("GET /v1/feeds/{feedID}/icon", handler.getIconByFeedID)
	router.HandleFunc("PUT /v1/feeds/{feedID}/mark-all-as-read", handler.markFeedAsRead)
	router.HandleFunc("GET /v1/export", handler.exportFeeds)
	router.HandleFunc("POST /v1/import", handler.importFeeds)
	router.HandleFunc("GET /v1/feeds/{feedID}/entries", handler.getFeedEntries)
	router.HandleFunc("GET /v1/feeds/{feedID}/entries/{entryID}", handler.getFeedEntry)
	router.HandleFunc("GET /v1/entries", handler.getEntries)
	router.HandleFunc("PUT /v1/entries", handler.setEntryStatus)
	router.HandleFunc("GET /v1/entries/{entryID}", handler.getEntry)
	router.HandleFunc("PUT /v1/entries/{entryID}", handler.updateEntry)
	router.HandleFunc("PUT /v1/entries/{entryID}/bookmark", handler.toggleBookmark)
	router.HandleFunc("POST /v1/entries/{entryID}/save", handler.saveEntry)
	router.HandleFunc("GET /v1/entries/{entryID}/fetch-content", handler.fetchContent)
	router.HandleFunc("PUT /v1/flush-history", handler.flushHistory)
	router.HandleFunc("DELETE /v1/flush-history", handler.flushHistory)
	router.HandleFunc("GET /v1/icons/{iconID}", handler.getIconByIconID)
	router.HandleFunc("GET /v1/enclosures/{enclosureID}", handler.getEnclosureByID)
	router.HandleFunc("PUT /v1/enclosures/{enclosureID}", handler.updateEnclosureByID)
	router.HandleFunc("GET /v1/version", handler.versionHandler)
}

func (h *handler) versionHandler(w http.ResponseWriter, r *http.Request) {
	json.OK(w, r, &versionResponse{
		Version:   version.Version,
		Commit:    version.Commit,
		BuildDate: version.BuildDate,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Arch:      runtime.GOARCH,
		OS:        runtime.GOOS,
	})
}
