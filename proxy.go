package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var blockList map[string]bool = map[string]bool{
	// "example.com": true,
}

var cache map[string]cacheItem = make(map[string]cacheItem)
var CACHE_EXPIRY = 90 * time.Second

// Struct to store cache with identifier
type cacheItem struct {
	body    []byte
	headers http.Header
	date    string
}

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

	URI := req.RequestURI
	cachedRes, exist := cache[URI]
	log.Printf("Cache Exist for URI %s? %t", URI, exist)
	// If request/response is cached
	if exist {
		// Add header to check if cache is fresh
		location, _ := time.LoadLocation("GMT")
		time := time.Now().In(location).Format(http.TimeFormat)

		client := &http.Client{}
		req, err := http.NewRequest("GET", URI, nil)
		if err != nil {
			log.Printf("%s\n", err)
		}
		req.Header.Add("If-Modified-Since", time)
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("%s\n", err)
		}
		defer resp.Body.Close()
		// If not modified, use cache
		log.Printf("%d %s", resp.StatusCode, URI)
		if resp.StatusCode == 304 {
			for k, v := range cachedRes.headers {
				for _, vv := range v {
					w.Header().Set(k, vv)
				}
			}
			fmt.Fprint(w, string(cachedRes.body))
			log.Printf("Used cache, returning")
			return
		}
	}

	// Send to client
	resp, err := http.Get(URI)
	if err != nil {
		log.Printf("%s\n", err)
	}
	defer resp.Body.Close()
	// Copies all headers
	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Set(k, vv)
		}
	}

	// Copy body and pass to client
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	fmt.Fprint(w, string(body))
	// Cache and set cache expiry
	respDate := resp.Header.Get("date")
	cachedData := cacheItem{
		body:    body,
		headers: resp.Header,
		date:    respDate,
	}
	cache[URI] = cachedData

	go func(date string) {
		time.Sleep(CACHE_EXPIRY)
		if cache[URI].date != date {
			delete(cache, URI)
		}
		log.Printf("Killing cache for %s at %s \n%v\n", URI, cachedData.date, cachedData.headers)
	}(respDate)
}

func networkHandler(w http.ResponseWriter, req *http.Request) {
	host := req.Host
	log.Printf("Host: %s, %s", host, req.RequestURI)
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

func blockListHandler() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Web Proxy Console")
	fmt.Println("---------------------")

	for {
		fmt.Print("-> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Println(err)
		}
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		arguments := strings.Split(text, " ")
		fmt.Printf("\nPrinting\n%v\n", arguments)
		switch arguments[0] {
		case "list":
			log.Printf("%v\n", blockList)
		case "block":
			blockList[arguments[1]] = true
		case "unblock":
			delete(blockList, arguments[1])
		default:
			log.Println("# Wrong input: <block/unblock> <URI>", text)
		}
		log.Printf("Entered line: %s\n", text)
	}
}

func main() {
	go blockListHandler()
	networkHandler := http.HandlerFunc(networkHandler)
	// Create a thread of networkHandler for each connection
	http.ListenAndServe(":8080", networkHandler)
}
