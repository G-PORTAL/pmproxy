package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
)

// Debugger prints request URL, request headers and response headers to the console
func Debugger(c *gin.Context) {
	reqHeaders := c.Request.Header
	c.Next()
	if c.Writer.Status() != http.StatusOK {
		resHeaders := c.Writer.Header()
		log.Printf("%s %s", c.Request.Method, c.Request.URL.String())
		log.Println("Request headers:")
		for k, v := range reqHeaders {
			log.Printf("\t %s: %s\n", k, v)
		}
		log.Println("Response headers:")
		for k, v := range resHeaders {
			log.Printf("\t %s: %s\n", k, v)
		}
	}
}

// ClearHeaders removes all headers from request
func ClearHeaders(headers *http.Header) {
	for k, _ := range *headers {
		headers.Del(k)
	}
}

// ClearCfHeaders removes all headers from request
func ClearCfHeaders(headers *http.Header) {
	for k, _ := range *headers {
		if strings.HasPrefix(k, "Cf-") ||
			strings.HasPrefix(k, "Cdn-Loop") {
			headers.Del(k)
		}
	}
}

// ProxyRequest modifies the request to be proxied
func ProxyRequest(req *http.Request) {
	req.URL.Scheme = "https"
	req.URL.Host = req.Header.Get("X-Pm-Host")
	if proto := req.Header.Get("X-Forwarded-Proto"); proto != "" {
		req.URL.Scheme = proto
	}

	if req.Header.Get("X-Pm-Port") != "" {
		req.URL.Host += ":" + req.Header.Get("X-Pm-Port")
	}

	req.Host = req.Header.Get("X-Forwarded-Host")
	req.Header.Set("Host", req.Header.Get("X-Forwarded-Host"))
	req.Header.Set("X-Forwarded-Host", req.Header.Get("X-Forwarded-Host"))
	req.Header.Set("X-Forwarded-For", req.Header.Get("Cf-Connecting-Ip"))
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("Origin", req.Header.Get("X-Forwarded-Host"))

	// Hackaround for ILO5 and Cloudflare Workers for the X-Auth-Token header which
	// is going to be removed on incoming and outgoing requests.
	if req.Header.Get("X-Pm-Token") != "" {
		req.Header.Set("X-Auth-Token", req.Header.Get("X-Pm-Token"))
	}

	req.Header.Del("X-Forwarded-Port")

	ClearCfHeaders(&req.Header)
}
