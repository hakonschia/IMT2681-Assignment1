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

// Posts the .igc url to the server and returns the response
func postURLToServer(t *testing.T, s *httptest.Server) *http.Response {
	url := s.URL + "/igcinfo/api/igc/"

	postURL := "{\"url\":\"http://skypolaris.org/wp-content/uploads/IGS%20Files/Madrid%20to%20Jerez.igc\"}"

	response, err := http.Post(url, "application/json", strings.NewReader(postURL))
	if err != nil {
		t.Errorf("Error making POST request %s", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Status code is not OK: %d", response.StatusCode)
	}

	return response
}

func Test_formatISO801(t *testing.T) {
	expected := "P0Y1M1DT17H0M52S"

	var timeT time.Duration
	timeT += (30 * 60 * 60 * 24) // 1 month
	timeT += (60 * 60 * 24)      // 1 day
	timeT += (17 * 60 * 60)      // 17 hours
	timeT += 52                  // 52 seconds
	timeT *= time.Second         // Convert to seconds

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

	response := postURLToServer(t, testServer)

	respBody, _ := ioutil.ReadAll(response.Body)
	idMap := make(map[string]string)
	json.Unmarshal(respBody, &idMap)

	id := idMap["id"]
	if id != "25S" { // The response should be the ID that was added
		t.Errorf("Expected id 25S differs from actual %s", id)
	}
}

func Test_handlerAPIIGC_empty(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(handlerAPIIGC))
	defer testServer.Close()

	url := testServer.URL + "/igcinfo/api/igc/"

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

func Test_handlerAPIIGC_ID(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(handlerAPIIGC))
	defer testServer.Close()

	url := testServer.URL + "/igcinfo/api/igc/"

	_ = postURLToServer(t, testServer) // The response from the POST is not needed for this test

	// Add the expected ID to the url
	url += "25S/"

	response, err := http.Get(url)
	if err != nil {
		t.Errorf("Error with constructing GET method. %s", err)
	}

	actual := make(map[string]interface{}) // Get the values the server has
	json.NewDecoder(response.Body).Decode(&actual)

	expected := make(map[string]interface{}) // Set the expected values
	expected["H_date"] = "2016-02-19T00:00:00Z"
	expected["pilot"] = "Miguel Angel Gordillo"
	expected["glider"] = "RV8"
	expected["glider_id"] = "EC-XLL"
	expected["track_length"] = 443.2573603705269

	for key := range expected {
		if expected[key] != actual[key] {
			t.Errorf("Expected %s differs from actual %v", expected[key], actual[key])
		}
	}

}

// Checks that all the fields match after posted to the server
func Test_handlerAPIIGC_ID_Field(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(handlerAPIIGC))
	defer testServer.Close()

	url := testServer.URL + "/igcinfo/api/igc/"

	_ = postURLToServer(t, testServer) // The response from POST is not needed for this test

	url += "25S/"

	var keys [5]string
	keys[0] = "H_date"
	keys[1] = "pilot"
	keys[2] = "glider"
	keys[3] = "glider_id"
	keys[4] = "track_length"

	var expected [5]string
	expected[0] = "2016-02-19T00:00:00Z"
	expected[1] = "Miguel Angel Gordillo"
	expected[2] = "RV8"
	expected[3] = "EC-XLL"
	expected[4] = "443.2573603705269"

	baseURL := url
	for i := 0; i < len(keys); i++ {
		url := baseURL + keys[i] + "/"

		response, err := http.Get(url)
		if err != nil {
			t.Errorf("Error with constructing GET method. %s", err)
		}
		respBody, _ := ioutil.ReadAll(response.Body) // Read the response
		actual := string(respBody)                   // Convert to a string
		actual = actual[:len(actual)-1]              // Remove the last character of the string (a newline is read as well)

		if expected[i] != actual {
			t.Errorf("Expected '%s' differs from actual '%s'", expected[i], actual)
		}
	}
}
