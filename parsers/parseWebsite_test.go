package parsers

import (
	"bytes"
	"io/ioutil"
	"testing"
)

// Test for our main parser based on downloaded data from
// https://home24.career.softgarden.de/en/
func TestParseWebsiteData(t *testing.T) {
	file, _ := ioutil.ReadFile("../output.html")
	htmlFile := bytes.NewReader(file)
	validUrls := make([]string, 0)
	internalLinksFound, externalLinksFound, websiteTitle, htmlVersion, headingByLevel := ParseWebsiteData(htmlFile, &validUrls, "https://home24.career.softgarden.de/en/")
	if internalLinksFound != 0 {
		t.Error("Internal Links should be 0")
	}

	if externalLinksFound != 13 {
		t.Error("External Links should be 13")
	}

	if websiteTitle != "Careers – home24" {
		t.Error("Website title should be : Careers – home24")
	}

	if htmlVersion != "HTML 5" {
		t.Error("HTML version should be : HTML 5")
	}

	for headingName, headingCount := range headingByLevel {
		if headingName == "h1" && headingCount != 2 {
			t.Error("We should have 2 h1 tags")
		}

		if headingName == "h2" && headingCount != 7 {
			t.Error("We should have 7 h2 tags")
		}

		if headingName == "h3" && headingCount != 4 {
			t.Error("We should have 4 h3 tags")
		}

		if headingName == "h4" && headingCount != 0 {
			t.Error("We should have 0 h4 tags")
		}

		if headingName == "h5" && headingCount != 1 {
			t.Error("We should have 1 h5 tags")
		}

		if headingName == "h6" && headingCount != 0 {
			t.Error("We should have 0 h6 tags")
		}
	}
}
