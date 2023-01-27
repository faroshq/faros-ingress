package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/faroshq/faros-ingress/pkg/api"
	"github.com/faroshq/faros-ingress/pkg/models"
	utilhttp "github.com/faroshq/faros-ingress/pkg/util/http"
)

func (s *Service) getConnectionGateway(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// TODO: Connection authentication

	vars := mux.Vars(r)
	connID := vars["connection"]
	if connID == "" {
		utilhttp.WriteErrorBadRequestWithReason(w, fmt.Errorf("connection id is required"), fmt.Errorf("connection id is required"))
		return
	}

	connectionRef, err := s.store.GetConnection(ctx, models.Connection{
		ID: connID,
	})
	if err != nil {
		utilhttp.WriteErrorInternalServerError(w, err)
		return
	}

	result := api.ConnectionGateway{
		Hostname: connectionRef.GatewayURL,
	}

	utilhttp.Respond(w, result)
}
