package reporter

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/yurii-vyrovyi/sitemap-generator/internal/core"
)

const Xlmns = "http://www.sitemaps.org/schemas/sitemap/0.9"

type (
	Reporter struct {
		config Config
	}

	Config struct {
		FileName string
	}

	URLSet struct {
		XMLName xml.Name  `xml:"urlset"`
		URLSet  []URLItem `xml:"url"`
		Xmlns   string    `xml:"xlmns,attr"`
	}

	URLItem struct {
		Loc string `xml:"loc"`
	}
)

func New(config Config) *Reporter {
	return &Reporter{
		config: config,
	}
}

func (r *Reporter) Save(tree *core.PageItem) error {

	links := treeToList(tree)

	us := URLSet{
		URLSet: make([]URLItem, 0, len(links)),
		Xmlns:  Xlmns,
	}

	for _, link := range links {

		u := escapeLink(link)

		urlItem := URLItem{
			Loc: u,
		}

		us.URLSet = append(us.URLSet, urlItem)
	}

	buf, err := xml.MarshalIndent(us, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	resBuffer := append([]byte(xml.Header), buf...)

	//nolint:gosec
	if err := ioutil.WriteFile(r.config.FileName, resBuffer, 0644); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

func escapeLink(link string) string {

	escapeSymbols := map[string]string{
		`&`: "&amp;",
		`'`: "&apos;",
		`\`: "&quot;",
		`>`: "&gt;",
		`<`: "&lt;",
	}

	escapedLink := link
	for k, v := range escapeSymbols {
		escapedLink = strings.ReplaceAll(escapedLink, k, v)
	}

	return escapedLink
}

func treeToList(root *core.PageItem) []string {

	var addBranch func(lst []string, root *core.PageItem) []string
	addBranch = func(lst []string, root *core.PageItem) []string {

		for _, child := range root.Children {
			lst = addBranch(lst, child)
		}

		return append(lst, root.URL)
	}

	return addBranch(nil, root)

}
