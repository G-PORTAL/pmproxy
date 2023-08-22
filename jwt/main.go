package main

import (
	"flag"
	"fmt"
	jwt "github.com/golang-jwt/jwt/v5"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

// GenerateToken generates a JWT token for the given platform management IP address and upstream.
func GenerateToken(ip net.IP, upstream string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"ip":       ip.String(),
		"upstream": upstream,
		"iat":      time.Now().Unix(),
		"nbf":      time.Now().Add(-time.Second * 10).Unix(),
		"exp":      time.Now().Add(time.Minute * 30).Unix(),
	})

	// read JWT_SECRET from wrangler.toml
	data, err := os.ReadFile("../worker/wrangler.toml")
	if err != nil {
		log.Fatalln(err.Error())
	}
	var secret string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "JWT_SECRET") {
			secret = strings.Trim(strings.TrimPrefix(line, "JWT_SECRET = "), "\"")
		}
	}
	if secret == "" {
		log.Fatalln("JWT_SECRET not found in wrangler.toml")
	}

	return token.SignedString([]byte(secret))
}

func main() {
	flag.Parse()
	if ipVar == "" || upstreamVar == "" {
		flag.Usage()
		os.Exit(1)
	}
	token, err := GenerateToken(net.ParseIP(ipVar), upstreamVar)
	if err != nil {
		log.Fatalln(err.Error())
	}

	fmt.Println(token)
}

var ipVar string
var upstreamVar string

func init() {
	flag.StringVar(&ipVar, "ip", "", "IP of the platform management")
	flag.StringVar(&upstreamVar, "upstream", "", "Upstream connection (gateway)")
}
