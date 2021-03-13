package main

import (
	"io"
	"net/http"
	"os"
	"testing"
)

func TestHandleHTTPS(t *testing.T) {
	NetworkHandler := http.HandlerFunc(networkHandler)
	// Create a thread of networkHandler for each connection

	cleanResp, err := http.Get("https://example.com")
	if err != nil {
		t.Errorf("HTTPs failed, err: %q", err)
	}

	go http.ListenAndServe(":8080", NetworkHandler)
	os.Setenv("HTTP_PROXY", "http://localhost:8080")
	proxyResp, err := http.Get("https://example.com")

	if err != nil {
		t.Errorf("HTTPs failed, err: %q", err)
	}

	cleanBody, err := io.ReadAll(cleanResp.Body)
	if err != nil {
		t.Errorf("HTTPs failed, err: %q", err)
	}

	proxyBody, err := io.ReadAll(proxyResp.Body)
	if err != nil {
		t.Errorf("HTTPs failed, err: %q", err)
	}

	if string(cleanBody) != string(proxyBody) {
		t.Errorf("HTTPs failed, body different.\nw/o proxy: %s\nw/ proxy: %s", cleanBody, proxyBody)
	}

}

func TestHandleHTTP(t *testing.T) {
	NetworkHandler := http.HandlerFunc(networkHandler)
	// Create a thread of networkHandler for each connection

	cleanResp, err := http.Get("http://example.com")
	if err != nil {
		t.Errorf("HTTP failed, err: %q", err)
	}

	go http.ListenAndServe(":8080", NetworkHandler)
	os.Setenv("HTTP_PROXY", "http://localhost:8080")
	proxyResp, err := http.Get("https://example.com")

	if err != nil {
		t.Errorf("HTTP failed, err: %q", err)
	}

	cleanBody, err := io.ReadAll(cleanResp.Body)
	if err != nil {
		t.Errorf("HTTP failed, err: %q", err)
	}

	proxyBody, err := io.ReadAll(proxyResp.Body)
	if err != nil {
		t.Errorf("HTTP failed, err: %q", err)
	}

	if string(cleanBody) != string(proxyBody) {
		t.Errorf("HTTP failed, body different.\nw/o proxy: %s\nw/ proxy: %s", cleanBody, proxyBody)
	}

}
