package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

// CreateTestServer creates a test HTTP server with the given handler
func CreateTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// MakeTestRequest creates and executes a test HTTP request
func MakeTestRequest(method, url string, body interface{}) (*http.Response, error) {
	var reqBody *bytes.Buffer

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	return client.Do(req)
}

// ParseResponse parses the response body into the given interface
func ParseResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}

// CreateTestRequest creates a test request with the given parameters
func CreateTestRequest(method, path string, body interface{}) (*http.Request, error) {
	var reqBody *bytes.Buffer

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// ExecuteTestRequest executes a test request and returns the response
func ExecuteTestRequest(handler http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}
