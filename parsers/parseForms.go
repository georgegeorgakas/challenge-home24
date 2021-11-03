package parsers

import (
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/url"
	"strings"
)

// htmlForm represents the basic elements of an HTML Form.
type htmlForm struct {
	// Action is the URL where the form will be submitted
	Action string
	// Method is the HTTP method to use when submitting the form
	Method string
	// Values contains form values to be submitted
	Values url.Values
}

// ParseForms function from
// https://github.com/google/go-github/blob/f99e304acee0faf0b8a3e2d006f9725d7542ca14/scrape/forms.go?fbclid=IwAR1RfkygTR6TzCHyv6B3v19fXQ2JtqcLE9xh5I_K-8YABmY-RKGpiZZgE-s#L96
func ParseForms(r io.Reader) (forms []htmlForm) {
	doc, _ := goquery.NewDocumentFromReader(r)
	doc.Find("form").Each(func(_ int, s *goquery.Selection) {
		form := htmlForm{Values: url.Values{}}
		form.Action, _ = s.Attr("action")
		form.Method, _ = s.Attr("method")

		s.Find("input").Each(func(_ int, s *goquery.Selection) {
			name, _ := s.Attr("name")
			if name == "" {
				return
			}

			typ, _ := s.Attr("type")
			typ = strings.ToLower(typ)
			_, checked := s.Attr("checked")
			if (typ == "radio" || typ == "checkbox") && !checked {
				return
			}

			value, _ := s.Attr("value")
			form.Values.Add(name, value)
		})
		s.Find("textarea").Each(func(_ int, s *goquery.Selection) {
			name, _ := s.Attr("name")
			if name == "" {
				return
			}

			value := s.Text()
			form.Values.Add(name, value)
		})
		forms = append(forms, form)
	})
	return forms
}
