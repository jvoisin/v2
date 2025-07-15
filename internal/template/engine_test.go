package template // import "miniflux.app/v2/internal/template"

import (
	"testing"
)

func TestParseTemplates(t *testing.T) {
	if err := NewEngine(nil).ParseTemplates(); err != nil {
		t.Fatal(err)
	}
}
