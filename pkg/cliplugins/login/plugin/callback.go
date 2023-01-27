package plugin

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/mjudeikis/portal/pkg/models"
)

const LoginResponseHTML = `
	<html>
		<body>
			<script>window.close();</script>
			<p style="font-family: arial, sans-serif; text-align: center; font-size: 20px; margin-top: 100px; font-weight: bold;">
				You are now logged in to the Faros CLI, you can safely close this tab and return to the terminal.
			</p>
		</body>
	</html>
`

// getLocalListener returns a list listener - taken from
// https://golang.org/src/net/http/httptest/server.go?s=2996:3040#L93
func getLocalListener() (net.Listener, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if listener, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			return nil, err
		}
	}

	return listener, nil
}

// handleLoginCallback is used to handle the callback from the api server
func handleLoginCallback(r *http.Request, w http.ResponseWriter) (*models.LoginResponse, error) {
	if r.URL.RawQuery == "" {
		return nil, errors.New("no token found in the authorization request")
	}
	if !strings.HasPrefix(r.URL.RawQuery, "data=") {
		return nil, errors.New("invalid token response from API server")
	}
	raw := strings.TrimPrefix(r.URL.RawQuery, "data=")

	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}

	var response models.LoginResponse
	err = json.Unmarshal(decoded, &response)
	if err != nil {
		return nil, err
	}

	if _, err := w.Write([]byte(LoginResponseHTML)); err != nil {
		return nil, err
	}

	return &response, nil
}
