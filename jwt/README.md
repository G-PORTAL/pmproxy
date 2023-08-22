# pmproxy: JWT

This is a basic command line utility to generate JWT Tokens for the pmproxy software suite.
For production use, you may want to generate the Token within your own backend.
Make sure to have the JWT_SECRET set inside the wrangler.toml in the worker directory.

## Structure

The JWT Token is structured as follows:

```json
{
  "exp": 1692702990,
  "iat": 1692701190,
  "ip": "127.0.0.1", // Platform Management IP
  "nbf": 1692701180,
  "upstream": "my.pm.example.com" // Upstream (gateway) to speak with
}
```

## Generating a JWT Token

To generate a JWT Token, run the following command:

```bash
go run main.go -ip 127.0.0.1 -upstream my.pm.example.com
```