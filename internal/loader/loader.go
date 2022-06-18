package loader

import (
	"bytes"
	"context"
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	TagA     = "a"
	TagBase  = "base"
	AttrHref = "href"
)

type Loader struct{}

func New() *Loader {
	return &Loader{}
}

func (l *Loader) GetPage(ctx context.Context, pageURL string) ([]byte, error) {

	return getPage(ctx, pageURL)
}

// GetPageLinks returns all URLs of <a> tags on the page. URLs are absolute.
// GetPageLinks ignores invalid URLs including <base> href URL.
func (l *Loader) GetPageLinks(ctx context.Context, pageURL string) []string {
	page, err := getPage(ctx, pageURL)
	if err != nil {
		log.Printf("failed to load page: %v", err)
		return nil
	}

	links, bases := getLinksAndBase(page)
	if err != nil {
		log.Printf("failed to extract links from the page: %v", err)
		return nil
	}

	var baseURL string

	if len(bases) == 0 {
		baseURL = pageURL
	} else {
		baseURL = bases[0]

		if len(bases) > 1 {
			baseValue := fmt.Sprintf(`<base href="%v"/>`, bases[0])
			log.Printf("ERR: page [%v] has more than one <base>. Applying %v", pageURL, baseValue)
		}
	}

	absLinks := updateLinksWithBase(links, baseURL, pageURL)

	return absLinks
}

func getPage(ctx context.Context, pageURL string) ([]byte, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to run request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(" GET request failed: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failde to read page from response: %w", err)
	}

	return body, nil
}

// getLinksAndBase extracts all <a> tag links and all <base> href links.
func getLinksAndBase(page []byte) ([]string, []string) {

	node, err := html.Parse(bytes.NewReader(page))
	if err != nil {
		log.Printf("ERR: failed to parse page: %v", err)
		return nil, nil
	}

	var extractAnchorsFunc func(*html.Node) ([]string, []string)

	extractAnchorsFunc = func(n *html.Node) ([]string, []string) {

		var links []string
		var bases []string

		if n.Type == html.ElementNode {

			switch n.Data {

			case TagA:
				for _, attr := range n.Attr {
					if attr.Key == AttrHref && len(attr.Val) > 0 {
						links = append(links, attr.Val)
					}
				}

			case TagBase:
				for _, attr := range n.Attr {
					if attr.Key == AttrHref && len(attr.Val) > 0 {
						bases = append(bases, attr.Val)
					}
				}

			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			l, b := extractAnchorsFunc(c)
			links = append(links, l...)
			bases = append(bases, b...)
		}

		return links, bases

	}

	links, bases := extractAnchorsFunc(node)

	return links, bases
}

// updateLinksWithBase transforms all links to an absolute form.
// <base> URL is resolved against a page URL (if base URL is relative).
// Then every link is resolved against absolute base URL.
func updateLinksWithBase(links []string, base, page string) []string {

	res := make([]string, 0, len(links))

	urlPage, err := url.Parse(page)
	if err != nil {
		log.Printf("ERR: invalid base url [%v]: %v", base, err)
		return nil
	}

	// in case base path is relative it will be resolved against the page URL
	urlBase, err := urlPage.Parse(base)
	if err != nil {
		log.Printf("ERR: invalid base url [%v]: %v", base, err)
		return nil
	}

	for _, link := range links {

		noHashLink, _, _ := strings.Cut(link, "#")

		if len(noHashLink) == 0 {
			continue
		}

		urlLink, err := url.Parse(link)
		if err != nil || (urlLink.Scheme != "http" && urlLink.Scheme != "https") {
			continue
		}

		resURL, err := urlBase.Parse(noHashLink)
		if err != nil {
			log.Printf("ERR: failed to join base url [%v] and link [%v]: %v", base, link, err)
			continue
		}

		res = append(res, resURL.String())
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
