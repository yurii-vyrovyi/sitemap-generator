package loader

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoader_getLinksAndBase(t *testing.T) {
	t.Parallel()

	type Test struct {
		page     []byte
		expLinks []string
		expBases []string
	}

	// html.Parse() accepts any input and never returns an error (empty page, unknown tags, non-closed tags).
	// Errors may happen only in case of Reader errors – io.EOF or other errors.
	// Tests don't cover those cases.

	tests := map[string]Test{
		"OK": {
			page:     pageOK,
			expLinks: []string{"http://abs.link.com", "/rel/link"},
			expBases: []string{"http://test.com"},
		},

		"non-HTML": {
			page:     nonHTMLPage,
			expLinks: nil,
			expBases: nil,
		},

		"multiple bases HTML": {
			page:     badHTMLPage,
			expLinks: []string{"http://abs.link.com", "/rel/link"},
			expBases: []string{"http://test.com", "http://another.test.com"},
		},
	}

	//nolint:paralleltest
	for description, test := range tests {
		test := test

		t.Run(description, func(t *testing.T) {
			t.Parallel()

			links, bases := getLinksAndBase(test.page)

			require.Equal(t, test.expLinks, links)
			require.Equal(t, test.expBases, bases)
		})
	}
}

func TestLoader_updateLinksWithBase(t *testing.T) {
	t.Parallel()

	type Test struct {
		links    []string
		pageURL  string
		baseURL  string
		expLinks []string
	}

	tests := map[string]Test{
		"OK": {
			links: []string{
				"http://abs.link.com",
				"/rel/link",
			},
			pageURL: "http://hello.com",
			baseURL: "http://test.com/some/more/",
			expLinks: []string{
				"http://abs.link.com",
				"http://test.com/rel/link",
			},
		},

		"No base url": {
			links: []string{
				"http://abs.link.com",
				"/rel/link",
			},
			pageURL: "http://hello.com",
			baseURL: "",
			expLinks: []string{
				"http://abs.link.com",
				"http://hello.com/rel/link",
			},
		},

		"Relative base URL": {
			links: []string{
				"http://abs.link.com",
				"/rel/link",
			},
			pageURL: "http://hello.com/",
			baseURL: "/some/more",
			expLinks: []string{
				"http://abs.link.com",
				"http://hello.com/rel/link",
			},
		},

		"non HTTP links": {
			links: []string{
				"mailto:someone@home.com",
				"someone@home.com",
				"#page-anchor",
			},
			pageURL:  "http://hello.com/",
			baseURL:  "some/more",
			expLinks: nil,
		},
	}

	//nolint:paralleltest
	for description, test := range tests {
		test := test

		t.Run(description, func(t *testing.T) {
			t.Parallel()

			links := updateLinksWithBase(test.links, test.baseURL, test.pageURL)

			require.Equal(t, test.expLinks, links)
		})
	}

}

func TestLoader_getPageLinks(t *testing.T) {
	t.Parallel()

	type Test struct {
		srcPage []byte
		expRes  []string
	}

	tests := map[string]Test{
		"OK": {
			srcPage: pageOK,
			expRes: []string{
				"http://abs.link.com",
				"http://test.com/rel/link",
			},
		},

		"Multiple bases": {
			srcPage: pageWithTwoBases,
			expRes: []string{
				"http://abs.link.com",
				"http://test.com/rel/link",
			},
		},
	}

	ctx := context.Background()

	//nolint:paralleltest
	for description, test := range tests {
		test := test

		t.Run(description, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(test.srcPage)
			}))

			ldr := New()
			res := ldr.GetPageLinks(ctx, server.URL)

			require.Equal(t, test.expRes, res)
		})
	}
}
