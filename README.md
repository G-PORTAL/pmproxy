# Worker and Gateway: Infrastructure Proxies

## Overview

This repository contains two essential components for managing and routing incoming requests within your infrastructure: the **Worker** and the **Gateway**. These components work together to ensure secure and efficient handling of requests to your internal network services.

### Worker

The **Worker** is a Cloudflare Worker designed to act as a traffic router based on JSON Web Tokens (JWT). It receives incoming requests and determines the appropriate **Gateway** to forward the request to. The Worker makes use of the JWT token to authenticate and authorize the request before directing it to the chosen Gateway for further processing.

### Gateway

The **Gateway** is responsible for applying customization rules and ensuring the security of incoming requests. Once a request is routed by the Worker, the Gateway applies specific rules and configurations as needed by the platform management system. It acts as a gatekeeper, making sure that requests comply with predefined requirements before being forwarded to services within the internal network.

## Concept

The base requirement behind this project is providing a secure and efficient way to the platform management to the bare metal customers. To provide the best latency, we use Cloudflare Workers. Most crucial tasks are already being done there. Most importantly, the JWT Token includes the upstream (gateway) to speak with which is the closest entrypoint into the internal network the platform management is running on.

As an example: Platform Management Location is London and the Customer that is trying to access the Platform Management is located elsewhere. The customer is connecting to the cloudflare datacenter closest which then proxies the traffic towards the gateway that is located in London which then proxies the traffic towards the platform management system.

To archive something similar without Cloudflare Workers would require a lot of infrastructure around the world and a lot of time to set up. With Cloudflare Workers, we can do this with a few lines of code.

## Infrastructure

Below is a simplified ASCII representation of the infrastructure:

```
+-------------------+
|    Cloudflare     |
|      Worker       |
+---------+---------+
          |
          | (1) Authenticate and authorize
          |
+---------v---------+
|                   |
|      Gateway      |
|                   |
+---------+---------+
          |
          | (2) Apply customization rules
          |
+---------v---------+
|                   |
|   Internal        |
|   Network         |
|                   |
+-------------------+

```


1. **Worker**: Receives incoming requests, validates JWT tokens, and routes requests to the appropriate Gateway.

2. **Gateway**: Applies customization rules and policies to incoming requests before forwarding them to the internal network.

## Supported Platforms

The pmproxy currently supports the following platform management systems:

- IDRAC8
- IDRAC9
- ILO5
- ASRockRack


## Getting Started

To set up and deploy the Worker and Gateway, follow the instructions in their respective directories:
- [Gateway Setup](./gateway/)
- [Worker Setup](./worker/)
- [JWT Generation](./jwt/)

## Contributing

Contributions are welcome! Feel free to open issues and pull requests for bug fixes, improvements, or new features.

---

**Note:** This is a high-level overview. For detailed setup instructions, usage guidelines, and more, refer to the individual directories of the **Worker** and **Gateway** components.
