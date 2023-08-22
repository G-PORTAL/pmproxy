package server

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/g-portal/pmproxy/gateway/pkg/env"
	"github.com/g-portal/pmproxy/gateway/pkg/middleware"
	"github.com/gin-gonic/gin"
	"log"
)

func RunServer(listenAddr, certPath, keyPath string) {
	r := gin.Default()

	// Only add debugger if env variable "DEBUG" is set
	if gin.Mode() == gin.DebugMode {
		r.Use(middleware.Debugger)
		r.Use(middleware.GinBodyLogMiddleware)
	}

	r.GET("/websocket/:host/:port/*path", websocketProxy)
	r.NoRoute(httpProxy)
	log.Fatal(r.RunTLS(listenAddr, certPath, keyPath))
}

// validateToken validates the JWT token from the cookie and compares
// the IP address with the given one to validate access permissions.
func validateToken(c *gin.Context, ip string) bool {
	cookie, err := c.Request.Cookie("pm-session")
	if err != nil || cookie.Valid() != nil {
		return false
	}

	tkn, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		return []byte(env.GetWithDefault("JWT_KEY", "")), nil
	})

	if err != nil || !tkn.Valid {
		return false
	}

	// Check if upstream matches the one in the token
	if tknIP, ok := tkn.Claims.(jwt.MapClaims)["ip"]; ok {
		if _, ok := tknIP.(string); !ok || tknIP.(string) != ip {
			return false
		}
	} else {
		return false
	}

	return true
}
