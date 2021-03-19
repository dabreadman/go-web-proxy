package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

func TestCLIHandle(t *testing.T) {
	priorBlockListLen := len(blockList)

	r := strings.NewReader("block www.example.com")
	go CLIHandler(r)
	time.Sleep(1 * time.Second)
	blockedListLen := len(blockList)

	if _, exist := blockList["www.example.com"]; blockedListLen-priorBlockListLen != 1 || !exist {
		t.Errorf("Block list addition failed\n%v\n", blockList)
	}

	r = strings.NewReader("unblock www.example.com")
	go CLIHandler(r)
	time.Sleep(1 * time.Second)
	unblockedListLen := len(blockList)

	if _, exist := blockList["www.example.com"]; blockedListLen-unblockedListLen != 1 || exist {
		t.Errorf("Block list addition failed\n%v\n", blockList)
	}
}

func TestHTTPBlock(t *testing.T) {
	// Blocks example.com
	r := strings.NewReader("block www.example.com")
	go CLIHandler(r)
	time.Sleep(1 * time.Second)

	req, err := http.NewRequest(http.MethodGet, "http://www.example.com", nil)

	if err != nil {
		t.Fatal(err)
	}
	req.RequestURI = "http://www.example.com"

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(networkHandler)

	handler.ServeHTTP(rr, req)

	// Check if request blocked
	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("HTTP block failed\nblocklist: %v\nstatus: %v", blockList, status)
	}

	// Unblocks example.com
	r = strings.NewReader("unblock www.example.com")
	go CLIHandler(r)
	time.Sleep(1 * time.Second)

	req, err = http.NewRequest(http.MethodGet, "http://www.example.com", nil)

	if err != nil {
		t.Fatal(err)
	}
	req.RequestURI = "http://www.example.com"

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(networkHandler)

	handler.ServeHTTP(rr, req)

	// Checks if OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("HTTP unblock failed\nblocklist: %v\nstatus: %v", blockList, status)
	}

	cleanResp, err := http.Get("http://example.com")
	if err != nil {
		t.Errorf("HTTPs failed, err: %q", err)
	}

	cleanBody, err := io.ReadAll(cleanResp.Body)
	if err != nil {
		t.Errorf("HTTPs failed, err: %q", err)
	}
	// Checks if response body is the same
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
