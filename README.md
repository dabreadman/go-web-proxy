# go-web-proxy
A simple web proxy server implemented in Golang.

## Features
- **HTTP Caching**  
	The proxy caches HTTP response to save time and bandwidth. Expires every 90 seconds.  
	This is usually done on your browser.
- **Blocklist**  
The proxy maintains a blocklist to stop requests to specified URLs.

Read more about the proxy [here](https://github.com/dabreadman/go-web-proxy/blob/main/Web%20Proxy%20Documentation.pdf).

## Flowchart
![Flowchart](https://github.com/dabreadman/go-web-proxy/blob/main/action%20flow.png)

# Use
<sup>Please make sure [Go](https://golang.org/dl/) is present.</sup>
1. Clone go-web-proxy `git clone https://github.com/dabreadman/go-web-proxy.git`.
2. Run  `go run proxy.go`.
This will serve the proxy server on `127.0.0.1:8080`.

These are several commands to use for the proxy CLI.  
- **list**  
This command lists all the domains that is in the blocklist.
- **block \<domain>**  
This blocks any request to **domain** as host.
E.g. **domain**, or **domain**/some/url
- **unblock \<domain>**  
This remove entry of **domain** from the blocklist, unblocking requests to **domain**.
- **saved**  
This command lists the bandwidth and time saved from using the HTTP cache.

# Development
What's important to understand is that `go <func>` creates a thread of `func`, making it non-blocking.

The program is mainly consisted in [proxy.go](https://github.com/dabreadman/go-web-proxy/blob/main/proxy.go).

The program starts from `main`.
```go
networkHandler := http.HandlerFunc(networkHandler)
http.ListenAndServe(":8080", networkHandler)
```
This serves a listener to port `8080`, and assign a `networkHandler` for every connection to the port.

`networkHandler` handles blocking, and calls to `httpHandler` or `httpsHandler` depending on the request method.
A HTTPS request has a `CONNECT` method.

`httpsHandler` establish a connection to both the client and the server, and forward connections, forming a HTTPS tunnel.

`httpHandler` checks if a HTTP request is cached, 
If it is cached, a request with `If-Modified-Since` header is sent to the server. 
- A 200 (OK) will indicate that the cache is now 	unusable, and new response will be pass to the proxy, which will pass to the client and cache the response.
A cache killer is hired to kill the cache in `CACHE_EXPIRY` time if it was not updated, indicated by the `date` property.

- A 304 (Not Modified) will be returned if cache is still usable, in which the cached response will be forwarded to the client.
As a 304 response does not carry body, it saves bandwidth and time as connection between **intranet** *(proxy to client)* is significantly faster than connection between **internet** *(proxy to server)*.

`CLIhandler` reads input from `os.Stdin`, and parse it to perform operations based on the input arguments.  
I removed `\r\n` from input as this was developed on `Windows` environment. For Linux, remove `\n` instead.

`colorOutput` colors the terminal output by surrounding strings with special characters.  
Read more [here](https://misc.flogisoft.com/bash/tip_colors_and_formatting)  
`\e` is replaced with `\033` in this program.  
Read more about the difference between the two [here](https://unix.stackexchange.com/questions/89812/the-difference-between-e-and).

<sub>Unit test is still in shambles</sub>

