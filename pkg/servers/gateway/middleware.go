package gateway

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gorilla/handlers"
	"k8s.io/klog/v2"
)

func Panic() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if e := recover(); e != nil {
					klog.Error("panic")
					klog.Error(e)

					klog.Error(string(debug.Stack()))
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			h.ServeHTTP(w, r)
		})
	}
}

type logResponseWriter struct {
	http.ResponseWriter

	statusCode int
	path       string
	bytes      int
}

func (w *logResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker := w.ResponseWriter.(http.Hijacker)
	return hijacker.Hijack()
}

func (w *logResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

func (w *logResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func (w *logResponseWriter) Flush() {
	flucher := w.ResponseWriter.(http.Flusher)
	flucher.Flush()
}

func (w *logResponseWriter) CloseNotify() <-chan bool {
	notify := w.ResponseWriter.(http.CloseNotifier)
	return notify.CloseNotify()
}

type logReadCloser struct {
	io.ReadCloser

	bytes int
}

func (rc *logReadCloser) Read(b []byte) (int, error) {
	n, err := rc.ReadCloser.Read(b)
	rc.bytes += n
	return n, err
}

func Log() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t := time.Now()

			ctx := r.Context()

			r.Body = &logReadCloser{ReadCloser: r.Body}
			w = &logResponseWriter{ResponseWriter: w, statusCode: http.StatusOK, path: r.URL.Path}

			logger := klog.FromContext(ctx)

			r = r.WithContext(ctx)

			logger = logger.WithValues(
				"request_method", r.Method,
				"request_path", r.URL.Path,
				"request_proto", r.Proto,
				"request_remote_addr", r.RemoteAddr,
				"request_user_agent", r.UserAgent(),
			)

			// enrich context with logger and add back to request
			ctx = klog.NewContext(ctx, logger)
			r = r.WithContext(ctx)

			defer func() {
				logger.WithValues(
					"body_read_bytes", r.Body.(*logReadCloser).bytes,
					"body_written_bytes", w.(*logResponseWriter).bytes,
					"duration", time.Since(t).Seconds(),
					"response_status_code", w.(*logResponseWriter).statusCode,
				).V(4).Info("sent response")
			}()
			h.ServeHTTP(w, r)
		})
	}
}

func Gzip() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return handlers.CompressHandler(h)
	}
}
