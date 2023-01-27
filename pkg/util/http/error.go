package utilhttp

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"k8s.io/klog/v2"

	utilerror "github.com/mjudeikis/portal/pkg/util/error"
)

// TODO: Set correlation ID in request context and use it in all messages.

var correlationIDKey = "correlationID"

type correlationID string

func WriteErrorInternalServerError(w http.ResponseWriter, serverErr error) {
	correlationID := correlationID(uuid.New().String())
	userError := fmt.Errorf("internal server error")
	klog.ErrorS(serverErr, userError.Error(), correlationIDKey, correlationID)
	utilerror.WriteCloudError(w, utilerror.NewCloudError(http.StatusBadRequest, utilerror.CloudErrorCodeInvalidParameter, "Error: %s. %s: %s", userError.Error(), correlationIDKey, correlationID))
}

func WriteErrorBadRequest(w http.ResponseWriter, serverErr error) {
	userError := fmt.Errorf("bad request")
	WriteErrorBadRequestWithReason(w, userError, serverErr)
}

func WriteErrorBadRequestWithReason(w http.ResponseWriter, userError, serverErr error) {
	correlationID := correlationID(uuid.New().String())
	klog.ErrorS(serverErr, userError.Error(), correlationIDKey, correlationID)
	utilerror.WriteCloudError(w, utilerror.NewCloudError(http.StatusBadRequest, utilerror.CloudErrorCodeBadRequest, "Error: %s. %s: %s", userError.Error(), correlationIDKey, correlationID))
}

func WriteErrorUnauthorized(w http.ResponseWriter, serverErr error) {
	correlationID := correlationID(uuid.New().String())
	userError := fmt.Errorf("unauthorized")
	klog.ErrorS(serverErr, userError.Error(), correlationIDKey, correlationID)
	utilerror.WriteCloudError(w, utilerror.NewCloudError(http.StatusUnauthorized, utilerror.CloudErrorCodeUnauthorized, "Error: %s. %s: %s", userError.Error(), correlationIDKey, correlationID))
}

func WriteErrorConflict(w http.ResponseWriter, serverErr error) {
	userError := fmt.Errorf("conflict")
	WriteErrorConflictWithReason(w, userError, serverErr)
}

func WriteErrorConflictWithReason(w http.ResponseWriter, userError, serverErr error) {
	correlationID := correlationID(uuid.New().String())
	klog.ErrorS(serverErr, userError.Error(), correlationIDKey, correlationID)
	utilerror.WriteCloudError(w, utilerror.NewCloudError(http.StatusUnauthorized, utilerror.CloudErrorCodeUnauthorized, "Error: %s. %s: %s", userError.Error(), correlationIDKey, correlationID))
}
