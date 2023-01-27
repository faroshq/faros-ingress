package api

import (
	"net/http"
)

// oidcLogin is a http handler for oidc login
// /faros.sh/oidc/login
func (s *Service) oidcLogin(w http.ResponseWriter, r *http.Request) {
	s.authenticator.OIDCLogin(w, r)
}

// oidcCallback is a http handler for oidc login callback
// /faros.sh/oidc/callback
func (s *Service) oidcCallback(w http.ResponseWriter, r *http.Request) {
	s.authenticator.OIDCCallback(w, r)
}
