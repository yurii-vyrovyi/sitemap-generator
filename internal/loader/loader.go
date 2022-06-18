package loader

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
)

type (
	Loader struct {
	}
)

const (
	TagA     = "a"
	TagBase  = "base"
	AttrHref = "href"
)

func New() *Loader {
	return &Loader{}
}

func (l *Loader) GetPage(pageURL string) ([]byte, error) {
	return l.getPage(pageURL)
}

func (l *Loader) getPage(pageURL string) ([]byte, error) {

	resp, err := http.Get(pageURL)
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

// GetPageLinks returns all URLs of <a> tags on the page. URLs are absolute.
// GetPageLinks ignores invalid URLs including <base> href URL.
func (l *Loader) GetPageLinks(pageURL string) []string {

	// TODO: should ignore pages with invalid BASE href url

	page, err := l.getPage(pageURL)
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

		// TODO: test
		if len(bases) > 1 {
			baseValue := fmt.Sprintf(`<base href="%v"/>`, bases[0])
			log.Printf("ERR: page [%v] has more than one <base>. Applying %v", pageURL, baseValue)
		}
	}

	absLinks := updateLinksWithBase(links, baseURL, pageURL)

	return absLinks
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
// First <base> URL is resolved against a page URL (if base URL is relative).
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

		resURL, err := urlBase.Parse(link)
		if err != nil {
			log.Printf("ERR: failed to join base url [%v] and link [%v]: %v", base, link, err)
			continue
		}
		res = append(res, resURL.String())
	}

	return res
}
