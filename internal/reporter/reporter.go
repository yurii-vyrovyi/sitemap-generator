package reporter

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"strings"
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
		UrlSet  []UrlItem `xml:"url"`
		Xmlns   string    `xml:"xlmns,attr"`
	}

	UrlItem struct {
		Loc string `xml:"loc"`
	}
)

func New(config Config) *Reporter {
	return &Reporter{
		config: config,
	}
}

func (r *Reporter) Save(links []string) error {

	us := URLSet{
		UrlSet: make([]UrlItem, 0, len(links)),
		Xmlns:  Xlmns,
	}

	for _, link := range links {

		u := escapeLink(link)

		urlItem := UrlItem{
			Loc: u,
		}

		us.UrlSet = append(us.UrlSet, urlItem)
	}

	buf, err := xml.MarshalIndent(us, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	resBuffer := append([]byte(xml.Header), buf...)

	if err := ioutil.WriteFile(r.config.FileName, resBuffer, 0666); err != nil {
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
		escapedLink = strings.Replace(escapedLink, k, v, -1)
	}

	return escapedLink
}
