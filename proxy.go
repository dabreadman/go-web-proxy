package main

import (
	"io"
	"log"
	"net"
	"net/http"
)

var blockList map[string]bool
var cache map[string]byte

func httpsHandler(w http.ResponseWriter, req *http.Request) {
	req.URL.Scheme = "https"

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}
	clientCon, buffrw, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	server := req.URL.Host
	serverCon, err := net.Dial("tcp", server)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Doesn't work because connect is hijacked
	// w.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	buffrw.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n")
	buffrw.Flush()
	// Connection to client
	go io.Copy(serverCon, buffrw)
	io.Copy(buffrw, serverCon)

	serverCon.Close()
	clientCon.Close()

}

func httpHandler(w http.ResponseWriter, req *http.Request) {
	return
	// If request/response is cached
	if cachedRes, ok := cache[req.URL.RawPath]; ok {
		log.Printf("%s\n", string(cachedRes))
		// If-Not-Modified request
		// If 304{
		//		Send cached response back to client
		//		Close
		// } else{
		// 		delete(cache, url)
		// }
		//
	}
	log.Printf("http %s \n", req.Host)
	// Make request
	// Buffer response
	// Make header
	// Cache
	/*
		cache[url]=res
		go func() {
		    time.Sleep(90 * time.Second)
		    delete(cache,res)
		    }()
	*/

	// Send to client
	resp, err := http.Get(req.URL.RawPath)
	if err != nil {
		log.Printf("%s\n", err)
	}
	log.Printf("%d", resp.StatusCode)
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
