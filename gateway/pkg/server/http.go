package server

import (
	"crypto/tls"
	"github.com/g-portal/pmproxy/gateway/pkg/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httputil"
)

func httpProxy(c *gin.Context) {
	if c.Request.Header.Get("X-Pm-Host") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Pm-Host header missing"})
		return
	}

	if !validateToken(c, c.Request.Header.Get("X-Pm-Host")) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	middleware.ProxyRequest(c.Request)

	proxyClient := httputil.ReverseProxy{}
	defaultTransport := http.DefaultTransport.(*http.Transport)
	defaultTransport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	proxyClient.Director = func(req *http.Request) {}
	proxyClient.Transport = defaultTransport

	proxyClient.ServeHTTP(c.Writer, c.Request)
}
