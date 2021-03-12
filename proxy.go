package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

var blockList map[string]bool
var cache map[string]byte

func httpsHandler(w http.ResponseWriter, req *http.Request) {
	req.URL.Scheme = "https"

	server := req.URL.Host
	serverCon, err := net.Dial("tcp", server)
	fmt.Println(serverCon.RemoteAddr())
	// Connection to client
	// io.Copy(serverCon,clientCon)
	// io.Copy(clientCon,serverCon)

	if err != nil {
		log.Panic(err)
	}
}

func httpHandler(w http.ResponseWriter, req *http.Request) {
	// If request/response is cached
	if cachedRes, ok := cache[req.Host]; ok {
		fmt.Printf("%s\n", string(cachedRes))
		// If-Not-Modified request
		// If 304{
		//		Send cached response back to client
		//		Close
		// } else{
		// 		delete(cache, url)
		// }
		//
	}
	fmt.Printf("http %s \n", req.Host)
	// Make request
	// Buffer response
	// Make header
	// Cache
	// Send to client
	resp, err := http.Get(req.URL.RawPath)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	fmt.Printf("%d", resp.StatusCode)
	defer resp.Body.Close()
}

func networkHandler(w http.ResponseWriter, req *http.Request) {
	host := req.Host
	// If not in blockList
	if !blockList[host] {
		// If HTTPS
		if req.Method == "CONNECT" {
			httpsHandler(w, req)
			// If HTTP
		} else {
			// Handles connections to server
			httpHandler(w, req)
		}
		// Return 403 if in blockList
	} else {
		log.Printf("Blocked %s\n", req.Host)
		w.WriteHeader(http.StatusForbidden)
	}
}

func main() {
	networkHandler := http.HandlerFunc(networkHandler)
	// Create a thread of networkHandler for each connection
	http.ListenAndServe(":8080", networkHandler)
}
