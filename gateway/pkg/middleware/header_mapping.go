package middleware

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

// HeaderMapperRequest uses the base64 encoded request headers in get parameter to apply them
func HeaderMapperRequest(req *http.Request) error {
	ClearHeaders(&req.Header)
	if req.URL.Query().Has("headers") {
		b64Headers := req.URL.Query().Get("headers")
		decoded, err := base64.StdEncoding.DecodeString(b64Headers)
		if err != nil {
			return fmt.Errorf("could not decode headers: %s", err.Error())
		}

		headers := map[string]string{}
		if err := json.Unmarshal(decoded, &headers); err != nil {
			return fmt.Errorf("could not unmarshal headers: %s", err.Error())
		}

		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	return nil
}
