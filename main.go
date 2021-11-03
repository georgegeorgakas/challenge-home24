package main

import (
	"challenge-home24/parsers"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

var inaccessibleLinksFound int

// htmlForm represents the basic elements of an HTML Form.
type htmlForm struct {
	// Action is the URL where the form will be submitted
	Action string
	// Method is the HTTP method to use when submitting the form
	Method string
	// Values contains form values to be submitted
	Values url.Values
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
	var response = getWebsiteDetails(&validUrls, urls.Url)

	for _, href := range validUrls {
		wg.Add(1)
		go checkInaccessibleUrls(&wg, href)
	}
	wg.Wait()
	response.InaccessibleUrls = inaccessibleLinksFound

	json.NewEncoder(w).Encode(response)
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/event", getData).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func getWebsiteDetails(validUrls *[]string, givenUrl string) WebsiteData {
	response, err := http.Get(givenUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	internalLinksFound, externalLinksFound, websiteTitle, htmlVersion, headingByLevel := parsers.ParseWebsiteData(response.Body, validUrls, givenUrl)
	return WebsiteData{
		Title:            websiteTitle,
		HtmlVersion:      htmlVersion,
		Headings:         headingByLevel,
		InternalUrls:     internalLinksFound,
		ExternalUrls:     externalLinksFound,
		InaccessibleUrls: 0,
		ValidUrls:        *validUrls,
		HasLoginForm:     false,
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
