package api

import (
	"net/http"

	"github.com/faroshq/faros-ingress/pkg/models"
	utilhttp "github.com/faroshq/faros-ingress/pkg/util/http"
)

func (s *Service) authenticate(w http.ResponseWriter, r *http.Request) (bool, *models.User, error) {
	authenticated, user, err := s.authenticator.Authenticate(r)
	if err != nil {
		utilhttp.WriteErrorUnauthorized(w, err)
		return false, nil, err
	}

	if !authenticated {
		utilhttp.WriteErrorUnauthorized(w, err)
		return false, nil, nil
	}

	return true, user, nil
}
