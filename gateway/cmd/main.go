package main

import (
	"github.com/g-portal/pmproxy/gateway/pkg/env"
	"github.com/g-portal/pmproxy/gateway/pkg/server"
)

func main() {
	listenAddr := env.GetWithDefault("LISTEN", "0.0.0.0:8443")
	certPath := env.GetWithDefault("CERT_PATH", "origin.crt")
	keyPath := env.GetWithDefault("KEY_PATH", "origin.key")
	server.RunServer(listenAddr, certPath, keyPath)
}
