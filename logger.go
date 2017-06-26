package adapter

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type loggingResponseWriter interface {
	http.ResponseWriter
	http.Flusher
	Status() int
	Size() int
}

// responseLogger is wrapper of http.ResponseWriter that keeps track of its HTTP
// status code and body size
//
// steel from https://github.com/gorilla/handlers/blob/master/handlers.go
type responseLogger struct {
	w      http.ResponseWriter
	status int
	size   int
}

func (l *responseLogger) Header() http.Header {
	return l.w.Header()
}

func (l *responseLogger) Write(b []byte) (int, error) {
	if l.status == 0 {
		// The status will be StatusOK if WriteHeader has not been called yet
		l.status = http.StatusOK
	}
	size, err := l.w.Write(b)
	l.size += size
	return size, err
}

func (l *responseLogger) WriteHeader(s int) {
	l.w.WriteHeader(s)
	l.status = s
}

func (l *responseLogger) Status() int {
	return l.status
}

func (l *responseLogger) Size() int {
	return l.size
}

func (l *responseLogger) Flush() {
	f, ok := l.w.(http.Flusher)
	if ok {
		f.Flush()
	}
}

type closeNotifyWriter struct {
	loggingResponseWriter
	http.CloseNotifier
}

func makeLogger(w http.ResponseWriter) loggingResponseWriter {
	var logger loggingResponseWriter = &responseLogger{w: w}
	if c, ok := w.(http.CloseNotifier); ok {
		return &closeNotifyWriter{logger, c}
	}
	return logger
}

type loggingHandler struct {
	handler http.Handler
}

func (h loggingHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	start := time.Now()
	logger := makeLogger(w)
	url := *req.URL
	h.handler.ServeHTTP(logger, req)
	fields := logrus.Fields{
		"uri":          url.String(),
		"status":       logger.Status(),
		"size":         logger.Size(),
		"method":       req.Method,
		"user":         logger.Header().Get("x-ngx-omniauth-user"),
		"provider":     logger.Header().Get("x-ngx-omniauth-provider"),
		"original_uri": req.Header.Get("x-ngx-omniauth-original-uri"),
	}

	end := time.Now()
	fields["reqtime_microsec"] = end.Sub(start).Nanoseconds() / 1000
	logrus.WithFields(fields).Info("access log")
}

// LoggingHandler logs HTTP requests.
func LoggingHandler(h http.Handler) http.Handler {
	return loggingHandler{h}
}
