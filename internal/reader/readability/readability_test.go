// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package readability // import "miniflux.app/v2/internal/reader/readability"

import (
	"bytes"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestBaseURL(t *testing.T) {
	html := `
		<html>
			<head>
				<base href="https://example.org/ ">
			</head>
			<body>
				<article>
					Some content
				</article>
			</body>
		</html>`

	baseURL, _, err := ExtractContent(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	if baseURL != "https://example.org/" {
		t.Errorf(`Unexpected base URL, got %q instead of "https://example.org/"`, baseURL)
	}
}

func TestMultipleBaseURL(t *testing.T) {
	html := `
		<html>
			<head>
				<base href="https://example.org/ ">
				<base href="https://example.com/ ">
			</head>
			<body>
				<article>
					Some content
				</article>
			</body>
		</html>`

	baseURL, _, err := ExtractContent(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	if baseURL != "https://example.org/" {
		t.Errorf(`Unexpected base URL, got %q instead of "https://example.org/"`, baseURL)
	}
}

func TestRelativeBaseURL(t *testing.T) {
	html := `
		<html>
			<head>
				<base href="/test/ ">
			</head>
			<body>
				<article>
					Some content
				</article>
			</body>
		</html>`

	baseURL, _, err := ExtractContent(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	if baseURL != "" {
		t.Errorf(`Unexpected base URL, got %q`, baseURL)
	}
}

func TestWithoutBaseURL(t *testing.T) {
	html := `
		<html>
			<head>
				<title>Test</title>
			</head>
			<body>
				<article>
					Some content
				</article>
			</body>
		</html>`

	baseURL, _, err := ExtractContent(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	if baseURL != "" {
		t.Errorf(`Unexpected base URL, got %q instead of ""`, baseURL)
	}
}

func TestRemoveStyleScript(t *testing.T) {
	html := `
		<html>
			<head>
				<title>Test</title>
				    <script src="tololo.js"></script>
			</head>
			<body>
				<script src="tololo.js"></script>
				<style>
			  		h1 {color:red;}
			  		p {color:blue;}
				</style>
				<article>Some content</article>
			</body>
		</html>`
	want := `<div><div><article>Somecontent</article></div></div>`

	_, content, err := ExtractContent(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	content = strings.ReplaceAll(content, "\n", "")
	content = strings.ReplaceAll(content, " ", "")
	content = strings.ReplaceAll(content, "\t", "")

	if content != want {
		t.Errorf(`Invalid content, got %s instead of %s`, content, want)
	}
}

func TestRemoveBlacklist(t *testing.T) {
	html := `
		<html>
			<head>
				<title>Test</title>
			</head>
			<body>
				<article class="super-ad">Some content</article>
				<article class="g-plus-crap">Some other thing</article>
				<article class="stuff popupbody">And more</article>
				<article class="legit">Valid!</article>
			</body>
		</html>`
	want := `<div><div><articleclass="legit">Valid!</article></div></div>`

	_, content, err := ExtractContent(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	content = strings.ReplaceAll(content, "\n", "")
	content = strings.ReplaceAll(content, " ", "")
	content = strings.ReplaceAll(content, "\t", "")

	if content != want {
		t.Errorf(`Invalid content, got %s instead of %s`, content, want)
	}
}

func TestNestedSpanInCodeBlock(t *testing.T) {
	html := `
		<html>
			<head>
				<title>Test</title>
			</head>
			<body>
				<article><p>Some content</p><pre><code class="hljs-built_in">Code block with <span class="hljs-built_in">nested span</span> <span class="hljs-comment"># exit 1</span></code></pre></article>
			</body>
		</html>`
	want := `<div><div><p>Some content</p><pre><code class="hljs-built_in">Code block with <span class="hljs-built_in">nested span</span> <span class="hljs-comment"># exit 1</span></code></pre></div></div>`

	_, result, err := ExtractContent(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	if result != want {
		t.Errorf(`Invalid content, got %s instead of %s`, result, want)
	}
}

func BenchmarkExtractContent(b *testing.B) {
	var testCases = map[string][]byte{
		"miniflux_github.html":    {},
		"miniflux_wikipedia.html": {},
	}
	for filename := range testCases {
		data, err := os.ReadFile("testdata/" + filename)
		if err != nil {
			b.Fatalf(`Unable to read file %q: %v`, filename, err)
		}
		testCases[filename] = data
	}
	for range b.N {
		for _, v := range testCases {
			ExtractContent(bytes.NewReader(v))
		}
	}
}

func TestGetLinkDensity(t *testing.T) {
	tests := []struct {
		html  string
		score float32
	}{
		{
			`<p>no links at all here</p>`,
			0.0,
		},
		{
			`<p>this is a great <a>test</a>.</p>`,
			0.19,
		},
		{
			`<p>this is a test
			with some <i>nested<b>elements<a>and a link</a>deeply</b>nested</i>.</p>
			and a bunch of padding to make the score a bit low`,
			0.084,
		},
	}

	const float64EqualityThreshold = 1e-3
	for _, test := range tests {
		document, err := goquery.NewDocumentFromReader(strings.NewReader(test.html))
		if err != nil {
			t.Fatal(err)
		}
		f := getLinkDensity(document.Selection)

		if math.Abs(float64(f-test.score)) > float64EqualityThreshold {
			t.Errorf(`Invalid count, got %f instead of %f`, f, test.score)
		}
	}
}
