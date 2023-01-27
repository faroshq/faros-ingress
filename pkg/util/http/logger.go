package utilhttp

import (
	"log"
	"strings"
)

type serverErrorLogWriter struct{}

func (*serverErrorLogWriter) Write(p []byte) (int, error) {
	m := string(p)
	// https://github.com/golang/go/issues/26918
	if strings.HasPrefix(m, "http: TLS handshake error") && strings.HasSuffix(m, ": EOF\n") {
		// handle EOF error
	} else {
		// handle other errors
	}
	return len(p), nil
}

func NewServerErrorLog() *log.Logger {
	return log.New(&serverErrorLogWriter{}, "", 0)
}
