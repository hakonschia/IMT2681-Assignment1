/*
Tests the handler functions of the API
*/

package igcapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// PostURLToServer posts an .igc url to the server and returns the response (only testing function)
func PostURLToServer(t *testing.T, s *httptest.Server) *http.Response {
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

// Tests that /igcinfo/api/ responds with information about the API
func Test_handlerAPI_info(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(HandlerAPI))
	defer testServer.Close()

	url := testServer.URL + "/igcinfo/api/"

	response, err := http.Get(url)
	if err != nil {
		t.Errorf("Error with constructing GET method. %s", err)
	}

	res := make(map[string]interface{})
	json.NewDecoder(response.Body).Decode(&res)

	keys := reflect.ValueOf(res).MapKeys()
	if len(keys) != 3 {
		t.Errorf("There are %d keys in the map, should be 3.", len(keys))
		return
	}

	// Compare the keys
	if keys[0].Interface() != "uptime" { // Convert reflect.Value to interface to compare to string
		t.Errorf("Key 0 expected to be '%s', got '%s'.", "uptime", keys[0])
	}
	if keys[1].Interface() != "info" {
		t.Errorf("Key 0 expected to be '%s', got '%s'.", "info", keys[0])
	}
	if keys[2].Interface() != "version" {
		t.Errorf("Key 0 expected to be '%s', got '%s'.", "version", keys[0])
	}

	// Compare the values
	if res["uptime"] != "P0Y0M0DT0H0M0S" {
		t.Errorf("Uptime expected to be '%s', got '%s'", "P0Y0M0DT0H0M0S", res["uptime"])
	}
	if res["info"] != "Service for IGC tracks" {
		t.Errorf("Info expected to be '%s', got '%s'", "V1", res["info"])
	}
	if res["version"] != "V1" {
		t.Errorf("Version expected to be '%s', got '%s'", "V!", res["version"])
	}
}

// Tests that posting to the server returns the correct response (the ID)
func Test_handlerIGC_POST(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(HandlerIGC))
	defer testServer.Close()

	response := PostURLToServer(t, testServer)

	respBody, _ := ioutil.ReadAll(response.Body)
	idMap := make(map[string]int)
	json.Unmarshal(respBody, &idMap)

	id := idMap["id"]
	if id != 1 { // The response should be the ID that was added
		t.Errorf("Expected id 1 differs from actual %d", id)
	}
}

// Tests that /igcinfo/api/igc/ returns an empty array before anything is posted
func Test_handlerIGC_empty(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(HandlerIGC))
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

// Tests that /igcinfo/api/igc/<ID> returns the correct information about the track with ID 1
func Test_handlerIGC_ID(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(HandlerIGC))
	defer testServer.Close()

	url := testServer.URL + "/igcinfo/api/igc/"

	_ = PostURLToServer(t, testServer) // The response from the POST is not needed for this test

	// Add the expected ID to the url
	url += "1/"

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
func Test_handlerIGC_ID_Field(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(HandlerIGC))
	defer testServer.Close()

	url := testServer.URL + "/igcinfo/api/igc/"

	_ = PostURLToServer(t, testServer) // The response from POST is not needed for this test

	url += "1/" // Check ID 1

	expectedKeys := [5]string{
		"H_date",
		"pilot",
		"glider",
		"glider_id",
		"track_length",
	}

	expectedValues := [5]string{
		"2016-02-19T00:00:00Z",
		"Miguel Angel Gordillo",
		"RV8",
		"EC-XLL",
		"443.2573603705269",
	}

	baseURL := url
	for i := 0; i < 5; i++ {
		url := baseURL + expectedKeys[i] + "/" // Add the field to the URL (/igcinfo/api/igc/1/pilot etc.)

		response, err := http.Get(url)
		if err != nil {
			t.Errorf("Error with constructing GET method. %s", err)
		}
		respBody, _ := ioutil.ReadAll(response.Body) // Read the response
		actual := string(respBody)                   // Convert to a string
		actual = actual[:len(actual)-1]              // Remove the last character of the string (a newline is read as well)

		if expectedValues[i] != actual {
			t.Errorf("Expected '%s' differs from actual '%s'", expectedValues[i], actual)
		}
	}
}
