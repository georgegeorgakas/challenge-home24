package main

import (
	"challenge-home24/parsers"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
)

var inaccessibleLinksFound int

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
	response.HasLoginForm = getFormsOfWebsite(urls.Url)

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

// This is a function in which we receive the forms of a website
// and with custom conditions we decide if we have a login form or not
// Test cases for this function to work are
// https://github.com/login &&
// https://gitlab.com/users/sign_in
func getFormsOfWebsite(givenUrl string) bool {
	response, err := http.Get(givenUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	forms := parsers.ParseForms(response.Body)
	if len(forms) == 0 {
		print("no forms found at %q")
	}

	for _, formData := range forms {
		if strings.Contains(formData.Action, "sign_in") || strings.Contains(formData.Action, "log_in") ||
			strings.Contains(formData.Action, "login") || strings.Contains(formData.Action, "signin") {
			return true
		}
		for key, _ := range formData.Values {
			if key == "login" {
				return true
			}
		}
	}
	return false
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
