package server

import (
	"crypto/tls"
	"fmt"
	"github.com/g-portal/pmproxy/gateway/pkg/middleware"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
)

func websocketProxy(c *gin.Context) {
	cloudflareClientWebsocketKey := c.Request.Header.Get("Sec-Websocket-Key")
	log.Printf("Cloudflare Client Websocket Key: %s", cloudflareClientWebsocketKey)
	if err := middleware.HeaderMapperRequest(c.Request); err != nil {
		log.Printf("Could not map headers: %s", err.Error())
		c.Status(http.StatusBadRequest)
		c.Abort()
		return
	}

	host := c.Param("host")
	port := c.Param("port")
	path := c.Param("path")

	if !validateToken(c, host) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	targetURL := fmt.Sprintf("wss://%s:%s%s", host, port, path)

	// Extract incoming request headers
	upstreamHeaders := c.Request.Header.Clone()
	for key, values := range upstreamHeaders {
		log.Printf("Key: %s, Values: %s", key, values)
		if strings.HasPrefix(key, "Sec-Websocket") ||
			strings.HasPrefix(key, "Upgrade") ||
			strings.HasPrefix(key, "Connection") {
			upstreamHeaders.Del(key)
		}
	}

	log.Printf("Target URL: %s", targetURL)
	log.Printf("Headers: %+v", c.Request.Header)

	dialer := websocket.DefaultDialer
	dialer.Subprotocols = splitAndTrimHeader(c.Request.Header.Get("Sec-Websocket-Protocol"))
	dialer.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	// Connect to upstream (platform management)
	upstreamConnection, httpResp, err := dialer.Dial(targetURL, upstreamHeaders)
	if err != nil {
		log.Printf("Error connecting to WebSocket upstream: %s", err.Error())
		log.Printf("HTTP Response: %+v", httpResp)
		c.AbortWithError(httpResp.StatusCode, err)
		return
	}
	log.Println("Successfully connected to WebSocket upstream")
	log.Printf("Upstream response headers: %+v", httpResp.Header)

	defer upstreamConnection.Close()

	upgrader := websocket.Upgrader{
		Subprotocols: splitAndTrimHeader(c.Request.Header.Get("Sec-Websocket-Protocol")),
		Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
			log.Printf("Error upgrading WebSocket connection: %s", reason.Error())
		},
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins, you may want to restrict this in production
		},
	}

	upgradeHeader := http.Header{}
	whitelist := []string{
		"Sec-Websocket-Protocol",
	}

	for _, key := range whitelist {
		if hdr := httpResp.Header.Get(key); hdr != "" {
			upgradeHeader.Set(key, hdr)
		}
	}
	c.Request.Header["Sec-Websocket-Key"] = []string{cloudflareClientWebsocketKey}

	// Upgrade client connection to WebSocket
	clientConnection, err := upgrader.Upgrade(c.Writer, c.Request, upgradeHeader)
	if err != nil {
		fmt.Println("Error upgrading connection for WebSocket:", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	log.Println("Successfully upgraded client connection to WebSocket")
	defer clientConnection.Close()

	errClient := make(chan error, 1)
	errBackend := make(chan error, 1)

	go replicateWebsocketConn(clientConnection, upstreamConnection, errClient)
	go replicateWebsocketConn(upstreamConnection, clientConnection, errBackend)

	var message string
	select {
	case err = <-errClient:
		message = "Error when copying from client to upstream: %v"
	case err = <-errBackend:
		message = "Error when copying from upstream to client: %v"

	}
	if e, ok := err.(*websocket.CloseError); !ok || e.Code == websocket.CloseAbnormalClosure {
		log.Printf(message, err)
	}

}

// copyMessages copies messages from source to target, until an error occurs
func replicateWebsocketConn(src, dst *websocket.Conn, errc chan error) {
	for {
		msgType, msg, err := src.ReadMessage()
		if err != nil {
			m := websocket.FormatCloseMessage(websocket.CloseNormalClosure, fmt.Sprintf("%v", err))
			if e, ok := err.(*websocket.CloseError); ok {
				if e.Code != websocket.CloseNoStatusReceived {
					m = websocket.FormatCloseMessage(e.Code, e.Text)
				}
			}
			errc <- err
			dst.WriteMessage(websocket.CloseMessage, m)
			break
		}
		err = dst.WriteMessage(msgType, msg)
		if err != nil {
			errc <- err
			break
		}
	}
}

func splitAndTrimHeader(value string) []string {
	split := strings.Split(value, ",")
	for i, v := range split {
		split[i] = strings.TrimSpace(v)
	}
	return split
}
