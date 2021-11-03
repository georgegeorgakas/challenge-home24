package parsers

import (
	"golang.org/x/net/html"
	"io"
	"net/url"
	"regexp"
	"strings"
)

var internalLinksFound int
var externalLinksFound int
var websiteTitle string
var htmlVersion string
var doctype = make(map[string]string)

// Initialize HTML versions
func init() {
	doctype["HTML 4.01 Strict"] = `"-//W3C//DTD HTML 4.01//EN"`
	doctype["HTML 4.01 Transitional"] = `"-//W3C//DTD HTML 4.01 Transitional//EN"`
	doctype["HTML 4.01 Frameset"] = `"-//W3C//DTD HTML 4.01 Frameset//EN"`
	doctype["XHTML 1.0 Strict"] = `"-//W3C//DTD XHTML 1.0 Strict//EN"`
	doctype["XHTML 1.0 Transitional"] = `"-//W3C//DTD XHTML 1.0 Transitional//EN"`
	doctype["XHTML 1.0 Frameset"] = `"-//W3C//DTD XHTML 1.0 Frameset//EN"`
	doctype["XHTML 1.1"] = `"-//W3C//DTD XHTML 1.1//EN"`
	doctype["HTML 5"] = `html`
}

// ParseWebsiteData Parse all website data
func ParseWebsiteData(body io.Reader, validUrls *[]string, givenUrl string) (
	int, int, string, string, map[string]int) {
	internalLinksFound = 0
	externalLinksFound = 0
	headingByLevel := make(map[string]int, 0)
	z := html.NewTokenizer(body)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return internalLinksFound, externalLinksFound, websiteTitle, htmlVersion, headingByLevel
		case tt == html.DoctypeToken:
			// Find the html version here
			t := z.Token()
			checkDoctype(t.Data)
		case tt == html.StartTagToken:
			t := z.Token()
			// Check if we have a link element
			isAnchor := t.Data == "a"
			if isAnchor {
				getValidUrls(validUrls, t.Attr, givenUrl)
				continue
			}

			// Check if we have a title element
			isTitle := t.Data == "title"
			if isTitle {
				title := z.Next()
				// Get the actual text of title
				if title == html.TextToken {
					t1 := z.Token()
					websiteTitle = t1.Data
					continue
				}
			}

			// Check for headings with a regex and store them in an array based on level
			isHeading, _ := regexp.MatchString("h[1-6]", t.Data)
			if isHeading {
				headingByLevel[t.Data] += 1
			}
		}
	}
}

// Check if the url is valid
// and if it is internal or not
func getValidUrls(validUrls *[]string, attr []html.Attribute, givenUrl string) {
	for _, a := range attr {
		if a.Key == "href" {
			href := a.Val
			if !isUrl(href) {
				continue
			}
			websiteUrl, _ := url.Parse(href)
			givenUrlParsed, _ := url.Parse(givenUrl)
			if websiteUrl.Host == givenUrlParsed.Host {
				internalLinksFound++
			} else {
				externalLinksFound++
			}
			*validUrls = append(*validUrls, href)
		}
	}
}

// Check if we receive a valid url
func isUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// Check what html version the website has
func checkDoctype(doctypeFound string) {
	htmlVersion = "UNKNOWN"

	for doctype, matcher := range doctype {
		match := strings.Contains(doctypeFound, matcher)

		if match == true {
			htmlVersion = doctype
			break
		}
	}
}
