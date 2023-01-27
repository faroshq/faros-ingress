package utilhttp

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	utilerror "github.com/faroshq/faros-ingress/pkg/util/error"
)

const (
	ProxiedFromDeviceHeader = "proxied-from-device"
)

var errInvalidReferrer = errors.New("invalid referrer")

type ResponseWriter struct {
	HasWrittenStatus bool

	Headers http.Header
	Writer  io.Writer
	Status  int
}

func (w *ResponseWriter) Write(b []byte) (n int, err error) {
	if !w.HasWrittenStatus {
		w.WriteHeader(http.StatusOK)
	}

	return w.Writer.Write(b)
}

func (w *ResponseWriter) Header() http.Header {
	return w.Headers
}

func (w *ResponseWriter) WriteHeader(code int) {
	w.Status = code
	w.HasWrittenStatus = true
}

func Respond(w http.ResponseWriter, ret interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ret)
}

func ProxyResponseFromDevice(w http.ResponseWriter, resp *http.Response) {
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.Header().Set(ProxiedFromDeviceHeader, "")

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	resp.Body.Close()
}

func ProxyResponse(w http.ResponseWriter, resp *http.Response) {
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	resp.Body.Close()
}

func WithReferrer(w http.ResponseWriter, r *http.Request, f func(referrer *url.URL)) {
	referrer, err := url.Parse(r.Referer())
	if err != nil {
		utilerror.WriteCloudError(w, utilerror.NewCloudError(http.StatusBadRequest, utilerror.CloudErrorCodeInvalidParameter, errInvalidReferrer.Error()))
		return
	}
	if referrer.Scheme != "http" && referrer.Scheme != "https" {
		utilerror.WriteCloudError(w, utilerror.NewCloudError(http.StatusBadRequest, utilerror.CloudErrorCodeInvalidParameter, errInvalidReferrer.Error()))
		return
	}
	f(referrer)
}
