# pmproxy: Worker

Welcome to the pmproxy Worker documentation! This component is a crucial part of the pmproxy software suite. It's responsible for ensuring the security of incoming requests and routing logic. The Worker acts as a router, playing a pivotal role in validating JWT Tokens and making specific modifications required for the platform management system to accept proxied traffic. It also uses the `upstream` claim included within the JWT token to determine which gateway should be used providing the best latency to the platform management.

## Overview

The Worker component intercepts incoming WebSocket connections, skillfully employing base64-encoded headers provided as query parameters. These actions prepare the request for seamless forwarding to the platform management system. The Gateway ensures the integrity of the connection and paves the way for efficient, secure, and optimized communication between the client and the management system.

## Key Features

- Secure Request Processing: The Worker validates JWT Tokens, ensuring the authenticity of incoming requests.
- Customization Rules: Apply specific modifications tailored to the platform management system's requirements.
- WebSocket Handling: Intercept WebSocket connections, allowing base64-encoded headers to be passed as query parameter.
- Cloudflare Workarounds: Address limitations induced by the Cloudflare environment, ensuring smooth operation.

## Getting Started

1. Clone this repository
2. Copy the `wrangler.example.toml` to `wrangler.toml` and fill in the required values
3. Install dependencies: `yarn install`
4. Authorize with cloudflare: `wrangler login`
5. Deploy to Cloudflare Worker: `wrangler deploy`

## JWT Authentication
The Worker uses JWT Tokens to authenticate and authorize incoming requests. The Worker component is responsible for pre-validating those JWT Tokens. The Worker validates the token and ensures that the request is authorized to access the platform management system. The Sentinel and the Commander are both also ensuring that the token is not expired.
In addition to the default claims, the Worker also checks for the following claims:
* `upstream` - The address of the gateway system to reach

More details: [JWT Generation](./jwt/)


## Initiate Session
To initiate the session with your platform management system, you need to send a request to the worker. The worker will then validate the JWT Token and forward the request to the appropriate gateway. The gateway will then apply the customization rules and forward the request to the platform management system.
The entrypoint should always be the `/gpsession` path, which is handled by the worker to parse the JWT token and set the appropriate environment.
https://my.pm.example.com/gpsession?token=<JWT_TOKEN>