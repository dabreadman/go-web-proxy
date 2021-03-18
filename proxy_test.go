package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleHTTP(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://www.example.com", nil)

	if err != nil {
		t.Fatal(err)
	}
	req.RequestURI = "http://www.example.com"

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(networkHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("blocklist: %v\nstatus: %v", blockList, status)
	}
	cleanResp, err := http.Get("http://example.com")
	if err != nil {
		t.Errorf("HTTPs failed, err: %q", err)
	}

	cleanBody, err := io.ReadAll(cleanResp.Body)
	if err != nil {
		t.Errorf("HTTPs failed, err: %q", err)
	}
	if string(cleanBody) != string(rr.Body.String()) {
		t.Errorf("HTTPs failed, body different.\nw/o proxy: %s\nw/ proxy: %s", cleanBody, rr.Body.String())
	}
}

// func TestHandleHTTPS(t *testing.T) {
// 	req, err := http.NewRequest(http.MethodConnect, "www.example.com", nil)

// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	req.RequestURI = "https://www.example.com"

// 	rr := httptest.NewRecorder()
// 	handler := http.HandlerFunc(networkHandler)

// 	handler.ServeHTTP(rr, req)

// 	if status := rr.Code; status != http.StatusOK {
// 		t.Errorf("blocklist: %v\nstatus: %v", blockList, status)
// 	}

// 	cleanResp, err := http.Get("https://www.example.com")
// 	if err != nil {
// 		t.Errorf("HTTPs failed, err: %q", err)
// 	}

// 	cleanBody, err := io.ReadAll(cleanResp.Body)
// 	if err != nil {
// 		t.Errorf("HTTPs failed, err: %q", err)
// 	}
// 	if string(cleanBody) != string(rr.Body.String()) {
// 		t.Errorf("HTTPs failed, body different.\nw/o proxy: %s\nw/ proxy: %s", cleanBody, rr.Body.String())
// 	}
// }
