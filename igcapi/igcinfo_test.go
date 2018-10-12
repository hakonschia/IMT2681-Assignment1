/*
Tests the general functions of the API
*/
package igcapi

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Tests that the same track cannot be added two times
func Test_trackAlreadyAdded(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(HandlerIGC))
	defer testServer.Close()

	// Post a url to the server

	_ = PostURLToServer(t, testServer)         // The response of the first post isn't needed
	response := PostURLToServer(t, testServer) // Add it again, should give an error response back

	resp, _ := ioutil.ReadAll(response.Body)

	respStr := string(resp) // Convert the response to a string and remove the newline at the en
	respStr = respStr[:len(respStr)-1]

	if respStr != "That track has already been added (id: 1)" {
		fmt.Printf("Unexpected response: %s", string(resp))
	}
}

// Tests that a duration gets formatted correctly
func Test_formatISO801(t *testing.T) {
	expected := "P0Y1M1DT17H0M52S"

	var timeT time.Duration
	timeT += (30 * 60 * 60 * 24) // 1 month
	timeT += (60 * 60 * 24)      // 1 day
	timeT += (17 * 60 * 60)      // 17 hours
	timeT += 52                  // 52 seconds
	timeT *= time.Second         // Convert to seconds

	actual := FormatISO8601(timeT)

	if actual != expected {
		t.Error(actual, " differs from expected: ", expected)
	}
}

// Tests that all empty strings are removed from an array
func Test_removeEmpty(t *testing.T) {
	testValues := []string{"", "AB", "BRE", "", "CT", ""}
	actual := RemoveEmpty(testValues)

	for _, val := range actual {
		if val == "" {
			t.Errorf("Value '%s' is an empty string", val)
		}
	}
}
