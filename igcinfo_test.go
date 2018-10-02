package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func Test_formatISO801(t *testing.T) {
	expected := "P0Y0M1DT17H0M52S"

	var timeT time.Duration
	timeT += (60 * 60 * 24) // 1 day
	timeT += (17 * 60 * 60) // 17 hours
	timeT += 52             // 52 seconds
	timeT *= 1000000000

	actual := formatISO8601(timeT)

	if actual != expected {
		t.Error(actual, " differs from expected: ", expected)
	}
}

func Test_removeEmpty(t *testing.T) {
	testValues := []string{"", "AB", "BRE", "", "CT", ""}
	actual := removeEmpty(testValues)

	for _, val := range actual {
		if val == "" {
			t.Errorf("Value '%s' is an empty string", val)
		}
	}
}

func Test_handlerAPI_generic(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(handlerAPI))
	defer testServer.Close()

	url := testServer.URL + "/igcinfo/api/"

	response, err := http.Get(url)
	if err != nil {
		t.Errorf("Error with constructing GET method. %s", err)
	}

	var res []string
	json.NewDecoder(response.Body).Decode(res)

	if res != nil { // The response back should be the empty array of IDs (nothing is POSTed yet to the server)
		t.Error("Did not get back an empty array")
	}
}

func Test_handlerAPIIGC_POST(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(handlerAPIIGC))
	defer testServer.Close()

	url := testServer.URL + "/igcinfo/api/igc/"

	postURL := "{\"url\":\"http://skypolaris.org/wp-content/uploads/IGS%20Files/Madrid%20to%20Jerez.igc\"}"

	response, err := http.Post(url, "application/json", strings.NewReader(postURL))
	if err != nil {
		t.Errorf("Error making POST request %s", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Status code is not OK: %d", response.StatusCode)
	}
	respBody, _ := ioutil.ReadAll(response.Body)
	idMap := make(map[string]string)
	json.Unmarshal(respBody, &idMap)

	id := idMap["id"]
	if id != "25S" {
		t.Errorf("Expected id 25S differs from actual %s", id)
	}
}

func Test_handlerAPIIGC_(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(handlerAPI))
	defer testServer.Close()

	url := testServer.URL + "/igcinfo/api/"

	/*

		post to server

	*/

	response, err := http.Get(url)
	if err != nil {
		t.Errorf("Error with constructing GET method. %s", err)
	}

	var res []string
	json.NewDecoder(response.Body).Decode(res)

	if res != nil { // The response back should be the empty array of IDs (nothing is POSTed yet to the server)
		t.Error("Did not get back an empty array")
	}
}
