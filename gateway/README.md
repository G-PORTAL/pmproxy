# pmproxy: Gateway

Welcome to the pmproxy gateway documentation! This component is a crucial part of the pmproxy software suite. It's responsible for ensuring the security of incoming requests and routing logic. The Gateway acts as a gatekeeper, playing a pivotal role in validating JWT Tokens and making specific modifications required for the platform management system to accept proxied traffic. It also addresses certain limitations arising from the Cloudflare environment.

## Overview

The Gateway component intercepts incoming WebSocket connections, employing base64-encoded headers provided as query parameters. These actions prepare the request for seamless forwarding to the platform management system. The Sentinel ensures the integrity of the connection and paves the way for efficient and secure communication between the client and the management system.

## Key Features

- Secure Request Processing: The Gateway validates JWT Tokens, ensuring the authenticity of incoming requests.
- Customization Rules: Apply specific modifications tailored to the platform management system's requirements.
- WebSocket Handling: Intercept WebSocket connections, allowing base64-encoded headers to be passed as query parameter.
- Cloudflare Workarounds: Address limitations induced by the Cloudflare environment, ensuring smooth operation.

## Getting Started

1. Clone this repository
2. Install dependencies: `go get -d ./...`
3. Build the binary: `go build -o gateway`
4. Run the binary: `./gateway`
5. The Gateway is now running on port 8443
6. Configure firewall rules to prevent direct communication with the gateway

Alternatively, if you do not want to build the binary yourself, you can use the docker image we build directly:
```bash
docker run -d \
  -p 8443:8443 \
  -v /etc/letsencrypt/live/pm.example.com/fullchain.pem:/etc/ssl/origin.crt:ro \
  -v /etc/letsencrypt/live/pm.example.com/privkey.pem:/etc/ssl/origin.key:ro \
  -e JWT_KEY=changeme \
  -e GIN_MODE=release \
  -e LISTEN=0.0.0.0:8443 \
  gportal/pmproxy-gateway:latest
```

## JWT Authentication
The Worker component is responsible for pre-validating JWT Tokens. However, the Gateway is validating the JWT Tokens again in case somebody discovered a way to bypass the firewall rules. The Worker and the Gateway are both also ensuring that the token is not expired.
In addition to the default claims, the Gateway also checks for the following claims:
* `ip` - The IP address of the platform management system to reach

## Configuration
There are a few environment variables to define how the Gateway should behave. The following table lists all available variables and their default values.

| Variable     | Description                                                               | Default        |
|--------------|---------------------------------------------------------------------------|----------------|
| `LISTEN`     | The address to listen on                                                  | `0.0.0.0:8443` |
| `CERT_PATH`  | Path where the SSL Certificate is located                                 | `origin.crt`   |
| `KEY_PATH`   | Path where the SSL Private Key is located                                 | `origin.key`   |
| `GIN_MODE`   | Defines the mode the HTTP server is running in (set to release for prod)  | `development`  |
| `JWT_KEY`    | Defines the JWT signing key secret                                        | `changeme`     |


## Troubleshooting
If Cloudflare declines the connection to your upstream running on :8443, make sure that the DNS entry belongs to the same domain.
Cloudflare only allows connections to custom ports when being served from the same domain as the worker is running on.
If that is not possible within your environment, you may want to use port 443 instead.

