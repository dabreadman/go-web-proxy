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
	//"example.com": true,
}

//var blockList sync.Map
var cacheSavings map[string]*cacheSaving = make(map[string]*cacheSaving)
var cache map[string]cacheItem = make(map[string]cacheItem)
var CACHE_EXPIRY = 90 * time.Second

// Struct to store cache with identifier
type cacheItem struct {
	body    []byte
	headers http.Header
	date    string
}

// Struct to store information about cache savings
type cacheSaving struct {
	dataSaved        int
	timeSaved        time.Duration
	lastUncachedTime time.Duration
}

func colorOutput(str string, color string) string {
	colorCode := ""
	switch color {
	case "green":
		colorCode = "32"
	case "gray":
		colorCode = "37"
	case "red":
		colorCode = "91"
	case "yellow":
		colorCode = "93"
	case "cyan":
		colorCode = "96"
	}
	colorCode = "\033[" + colorCode + "m"
	return colorCode + str + "\033[0m"
}

func httpsHandler(w http.ResponseWriter, req *http.Request) {
	//startTimer := time.Now()
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
	//elapsed := time.Since(startTimer)
	// Logging
	//log.Printf("%s Connection Established: %s [%s]",
	//colorOutput("HTTPS", "green"), colorOutput(server, "yellow"), colorOutput(elapsed.String(), "cyan"))
	// Logging

	// Connection to client
	go io.Copy(serverCon, buffrw)
	io.Copy(buffrw, serverCon)

	serverCon.Close()
	clientCon.Close()
}

func httpHandler(w http.ResponseWriter, req *http.Request) {
	startTimer := time.Now()
	URI := req.RequestURI
	cachedRes, exist := cache[URI]
	// If request/response is cached
	if exist {
		// Add header to check if cache is fresh
		location, _ := time.LoadLocation("GMT")
		formattedTime := time.Now().In(location).Format(http.TimeFormat)

		client := &http.Client{}
		req, err := http.NewRequest("GET", URI, nil)
		if err != nil {
			log.Printf("%s\n", err)
		}
		req.Header.Add("If-Modified-Since", formattedTime)
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("%s\n", err)
		}
		defer resp.Body.Close()
		// If not modified, use cache
		if resp.StatusCode == 304 {
			for k, v := range cachedRes.headers {
				for _, vv := range v {
					w.Header().Set(k, vv)
				}
			}
			fmt.Fprint(w, string(cachedRes.body))

			// Update time and bandwidth saved
			elapsed := time.Since(startTimer)
			savingPointer := cacheSavings[URI]
			timeSaved := savingPointer.lastUncachedTime - elapsed
			cachedBodyLen := len(cachedRes.body)
			savingPointer.timeSaved += timeSaved
			savingPointer.dataSaved += cachedBodyLen

			// Logging
			//if timeSaved > 0 {
			//log.Printf("%s  Completed Transmission: %s [%s] %s %s",
			//colorOutput("HTTP", "green"), colorOutput(URI, "yellow"), colorOutput(elapsed.String(), "cyan"),
			//colorOutput(strconv.Itoa(cachedBodyLen)+" bytes CACHED", "gray"), colorOutput("["+timeSaved.String()+"]", "green"))
			//} else {
			//log.Printf("%s  Completed Transmission: %s [%s] %s %s",
			//colorOutput("HTTP", "green"), colorOutput(URI, "yellow"), colorOutput(elapsed.String(), "cyan"),
			//colorOutput(strconv.Itoa(cachedBodyLen)+" bytes CACHED", "gray"), colorOutput("["+(-timeSaved).String()+"]", "red"))
			//}
			// Logging

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
	elapsed := time.Since(startTimer)
	// Logging
	//log.Printf("%s  Completed Transmission: %s [%s]",
	//colorOutput("HTTP", "green"), colorOutput(URI, "yellow"), colorOutput(elapsed.String(), "cyan"))
	// Logging

	// Cache and set cache expiry
	respDate := resp.Header.Get("date")
	cachedData := cacheItem{
		body:    body,
		headers: resp.Header,
		date:    respDate,
	}
	cache[URI] = cachedData
	savingPointer, exist := cacheSavings[URI]
	if !exist {
		resourceSavedPointer := &cacheSaving{
			dataSaved:        0,
			timeSaved:        0,
			lastUncachedTime: elapsed,
		}
		cacheSavings[URI] = resourceSavedPointer
	} else {
		savingPointer.lastUncachedTime = elapsed
	}

	go func(date string) {
		time.Sleep(CACHE_EXPIRY)
		if cache[URI].date == date {
			delete(cache, URI)

			// Logging
			//log.Printf("%s for %s registered at %s\n", colorOutput("Killing cache", "red"), colorOutput(URI, "yellow"), colorOutput(cachedData.date, "cyan"))
			// Logging
		}
	}(respDate)
}

func networkHandler(w http.ResponseWriter, req *http.Request) {
	host := req.Host
	exist, ok := blockList[host]
	log.Printf("Host: %s, %s\nCache exist? %t set to %t", host, req.RequestURI, exist, ok)

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

func CLIHandler() {
	fmt.Println("Proxy Console")
	reader := bufio.NewReader(os.Stdin)

	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Println(err)
		}
		// convert CRLF to LF
		text = strings.Replace(text, "\r\n", "", -1)
		log.Println(text)
		arguments := strings.Split(text, " ")

		if arguments[0] == "list" {
			log.Printf("%v\n", blockList)
		} else if arguments[0] == "block" {
			blockList[arguments[1]] = true
		} else if arguments[0] == "unblock" {
			delete(blockList, arguments[1])
		} else {
			log.Println("# Wrong input: <block/unblock> <URI>", text)
		}
	}
}

func main() {
	go CLIHandler()
	networkHandler := http.HandlerFunc(networkHandler)
	// Create a thread of networkHandler for each connection
	http.ListenAndServe(":8080", networkHandler)
}
