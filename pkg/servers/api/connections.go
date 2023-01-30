package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"k8s.io/klog/v2"

	"github.com/faroshq/faros-ingress/pkg/api"
	"github.com/faroshq/faros-ingress/pkg/models"
	utilhash "github.com/faroshq/faros-ingress/pkg/util/hash"
	utilhttp "github.com/faroshq/faros-ingress/pkg/util/http"
	utilpassword "github.com/faroshq/faros-ingress/pkg/util/password"
)

func (s *Service) getConnection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authenticated, user, err := s.authenticate(w, r)
	if err != nil || !authenticated {
		return
	}

	vars := mux.Vars(r)
	connID := vars["connection"]
	if connID == "" {
		utilhttp.WriteErrorBadRequestWithReason(w, fmt.Errorf("connection id is required"), fmt.Errorf("connection id is required"))
		return
	}

	connectionRef, err := s.store.GetConnection(ctx, models.Connection{
		ID:     connID,
		UserID: user.ID,
	})
	if err != nil {
		utilhttp.WriteErrorInternalServerError(w, err)
		return
	}

	result := api.Connection{
		ID:       connectionRef.ID,
		Name:     connectionRef.Name,
		TTL:      connectionRef.TTL,
		Token:    connectionRef.Token,
		Hostname: connectionRef.Hostname,
		Secure:   connectionRef.Secure,
		LastUsed: connectionRef.LastUsedAt,
	}

	utilhttp.Respond(w, result)
}

func (s *Service) listConnections(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authenticated, user, err := s.authenticate(w, r)
	if err != nil || !authenticated {
		return
	}

	connectionsRef, err := s.store.ListConnections(ctx, models.Connection{
		UserID: user.ID,
	})
	if err != nil {
		utilhttp.WriteErrorInternalServerError(w, err)
		return
	}

	result := api.ConnectionList{}
	for _, connectionRef := range connectionsRef {
		result.Items = append(result.Items, api.Connection{
			ID:       connectionRef.ID,
			Name:     connectionRef.Name,
			LastUsed: connectionRef.LastUsedAt,
			Token:    connectionRef.Token,
			TTL:      connectionRef.TTL,
			Hostname: connectionRef.Hostname,
			Secure:   connectionRef.Secure,
		})
	}

	utilhttp.Respond(w, result)
}

func (s *Service) createConnection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authenticated, user, err := s.authenticate(w, r)
	if err != nil || !authenticated {
		return
	}

	request := &api.Connection{}
	err = utilhttp.Read(r, request)
	if err != nil {
		utilhttp.WriteErrorInternalServerError(w, err)
		return
	}

	connections, err := s.store.ListConnections(ctx, models.Connection{
		UserID: user.ID,
	})
	if err != nil {
		utilhttp.WriteErrorInternalServerError(w, err)
		return
	}

	// check connection quota
	if len(connections) >= s.config.ConnectionQuota && s.config.ConnectionQuota != 0 {
		utilhttp.WriteErrorBadRequestWithReason(w, fmt.Errorf("connection quota exceeded"), nil)
		return
	}

	// name is unique per user
	for _, connection := range connections {
		if connection.Name == request.Name {
			utilhttp.WriteErrorConflictWithReason(w, fmt.Errorf("connection already exists"), nil)
			return
		}
	}

	connection := models.Connection{
		Token:  uuid.New().String(),
		UserID: user.ID,
		Name:   request.Name,
		TTL:    request.TTL,
		Secure: request.Secure,
	}

	// clean up hostname
	request.Hostname = strings.Replace(request.Hostname, "https://", "", 1)
	request.Hostname = strings.Replace(request.Hostname, "http://", "", 1)
	request.Hostname = strings.Replace(request.Hostname, s.config.HostnameSuffix, "", 1)

	// if hostname is not provided, generate one
	if request.Hostname == "" {
		request.Hostname = fmt.Sprintf("https://%s.%s", utilhash.GetHash(uuid.New().String()), s.config.HostnameSuffix)
	} else {
		request.Hostname = fmt.Sprintf("https://%s.%s", request.Hostname, s.config.HostnameSuffix)
		_, err := url.Parse(request.Hostname)
		if err != nil {
			utilhttp.WriteErrorBadRequestWithReason(w, fmt.Errorf("hostname is not valid"), err)
			return
		}
	}

	if !strings.HasSuffix(request.Hostname, s.config.HostnameSuffix) {
		utilhttp.WriteErrorBadRequestWithReason(w, fmt.Errorf("hostname '%s' must end with '%s'", request.Hostname, s.config.HostnameSuffix), nil)
		return
	}

	_, err = s.store.GetConnection(ctx, models.Connection{
		Hostname: connection.Hostname,
	})
	if err == nil {
		utilhttp.WriteErrorConflictWithReason(w, fmt.Errorf("hostname already taken"), nil)
		return
	}

	var username, password string
	if request.Secure {
		if request.Username == "" {
			username = "faros"
		} else {
			username = request.Username
		}

		if request.Password == "" {
			password = uuid.New().String()
		} else {
			password = request.Password
		}

		hashedPassword, err := utilpassword.GeneratePasswordHash([]byte(username + ":" + password))
		if err != nil {
			utilhttp.WriteErrorInternalServerError(w, err)
			return
		}
		connection.BasicAuthHash = hashedPassword
		connection.Secure = true
	} else {
		connection.BasicAuthHash = []byte{}
		connection.Secure = false

	}

	connection.GatewayURL = s.config.DefaultGateway
	connection.Hostname = request.Hostname
	connection.TTL = request.TTL

	connectionCreated, err := s.store.CreateConnection(ctx, connection)
	if err != nil {
		utilhttp.WriteErrorInternalServerError(w, err)
		return
	}

	utilhttp.Respond(w, api.Connection{
		ID:       connectionCreated.ID,
		Name:     connectionCreated.Name,
		Token:    connectionCreated.Token,
		Hostname: connectionCreated.Hostname,
		TTL:      connectionCreated.TTL,
		Username: username,
		Password: password,
		Secure:   connectionCreated.Secure,
	})
}

func (s *Service) updateConnection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authenticated, user, err := s.authenticate(w, r)
	if err != nil || !authenticated {
		return
	}

	connectionID := mux.Vars(r)["connection"]

	request := &api.Connection{}
	err = utilhttp.Read(r, request)
	if err != nil {
		utilhttp.WriteErrorInternalServerError(w, err)
		return
	}

	// name is unique per user
	Connection := &models.Connection{
		Name:   request.Name,
		UserID: user.ID,
	}

	if connectionID != Connection.ID {
		utilhttp.WriteErrorBadRequestWithReason(w, fmt.Errorf("connection name is required"), fmt.Errorf("connection name is required"))
		return
	}

	current, err := s.store.GetConnection(ctx, *Connection)
	if err == nil {
		utilhttp.WriteErrorConflictWithReason(w, fmt.Errorf("connection already exists"), nil)
		return
	}

	var username, password string
	var hashedPassword []byte

	if request.Username != "" {
		username = request.Username
	}
	if request.Password != "" {
		password = request.Password
	}
	if request.Hostname != "" {
		current.Hostname = request.Hostname
	}
	if request.Secure != current.Secure {
		current.Secure = request.Secure
	}

	if request.Username != "" && request.Password != "" {
		hashedPassword, err = utilpassword.GeneratePasswordHash([]byte(username + ":" + password))
		if err != nil {
			utilhttp.WriteErrorInternalServerError(w, err)
			return
		}
		current.BasicAuthHash = hashedPassword
		current.Secure = true
	}

	connectionUpdated, err := s.store.UpdateConnection(ctx, *current)
	if err != nil {
		utilhttp.WriteErrorInternalServerError(w, err)
		return
	}

	utilhttp.Respond(w, api.Connection{
		ID:       connectionUpdated.ID,
		Name:     connectionUpdated.Name,
		Hostname: connectionUpdated.Hostname,
		Username: request.Username,
		Password: request.Password,
		Secure:   connectionUpdated.Secure,
	})
}

func (s *Service) deleteConnection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authenticated, user, err := s.authenticate(w, r)
	if err != nil || !authenticated {
		return
	}

	vars := mux.Vars(r)
	connectionID := vars["connection"]
	if connectionID == "" {
		utilhttp.WriteErrorBadRequestWithReason(w, fmt.Errorf("connectionID is required"), fmt.Errorf("connectionID name is required"))
		return
	}

	Connection, err := s.store.GetConnection(ctx, models.Connection{
		UserID: user.ID,
		ID:     connectionID,
	})
	if err != nil {
		klog.Error(err)
		utilhttp.WriteErrorInternalServerError(w, err)
		return
	}

	if err := s.store.DeleteConnection(ctx, *Connection); err != nil {
		klog.Error(err)
		utilhttp.WriteErrorInternalServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
