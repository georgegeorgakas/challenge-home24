package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

var doctype = make(map[string]string)

var allLinksFound int
var internalLinksFound int
var externalLinksFound int
var inaccessibleLinksFound int
var websiteTitle string
var htmlVersion string
var headingByLevel map[string]int

// htmlForm represents the basic elements of an HTML Form.
type htmlForm struct {
	// Action is the URL where the form will be submitted
	Action string
	// Method is the HTTP method to use when submitting the form
	Method string
	// Values contains form values to be submitted
	Values url.Values
}

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

type SearchUrls struct {
	Url string
}

type WebsiteData struct {
	Title            string         `json:"title"`
	HtmlVersion      string         `json:"html_version"`
	Headings         map[string]int `json:"headings"`
	InternalUrls     int            `json:"internal_urls"`
	ExternalUrls     int            `json:"external_urls"`
	InaccessibleUrls int            `json:"inaccessible_urls"`
	ValidUrls        []string       `json:"valid_urls"`
	HasLoginForm     bool           `json:"has_login_form"`
}

func getData(w http.ResponseWriter, r *http.Request) {
	var wg sync.WaitGroup
	var urls SearchUrls

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Please enter data with the url in order to return you the details")
		return
	}
	err = json.Unmarshal(reqBody, &urls)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if urls.Url == "" {
		http.Error(w, "Please enter a url field", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)

	validUrls := make([]string, 0)
	getWebsiteDetails(&validUrls, urls.Url)

	for _, href := range validUrls {
		wg.Add(1)
		go checkInaccessibleUrls(&wg, href)
	}
	wg.Wait()
	var response = WebsiteData{
		Title:            websiteTitle,
		HtmlVersion:      htmlVersion,
		Headings:         headingByLevel,
		InternalUrls:     internalLinksFound,
		ExternalUrls:     externalLinksFound,
		InaccessibleUrls: inaccessibleLinksFound,
		ValidUrls:        validUrls,
		HasLoginForm:     false,
	}

	json.NewEncoder(w).Encode(response)
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/event", getData).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func getWebsiteDetails(validUrls *[]string, givenUrl string) {
	allLinksFound = 0
	internalLinksFound = 0
	externalLinksFound = 0
	inaccessibleLinksFound = 0
	headingByLevel = make(map[string]int, 0)
	response, err := http.Get(givenUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	z := html.NewTokenizer(response.Body)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			for heading, headingCounter := range headingByLevel {
				fmt.Printf("%s has %d\n", heading, headingCounter)
			}
			// End of the document, we're done
			return
		case tt == html.DoctypeToken:
			// Find the html version here
			t := z.Token()
			checkDoctype(t.Data)
		case tt == html.StartTagToken:
			t := z.Token()
			// Check if we have a link element
			isAnchor := t.Data == "a"
			if isAnchor {
				allLinksFound++
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

// GET Request to check if we have found any inaccessible URLs
func checkInaccessibleUrls(wg *sync.WaitGroup, href string) {
	defer wg.Done()
	_, err := http.Get(href)
	if err != nil {
		inaccessibleLinksFound++
		println(err.Error())
	}
}

// Check if a url is valid
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

func findLoginForm(r io.Reader) (forms []htmlForm) {
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
