package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type RequestResponseInfo struct {
	http.ResponseWriter
	RespInfo *ResponseInfo
}

type ResponseInfo struct {
	StatusCode int
	Size       int
}

func (r *RequestResponseInfo) Header() http.Header {
	return r.ResponseWriter.Header()
}

func (r *RequestResponseInfo) WriteHeader(statusCode int) {
	r.RespInfo.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *RequestResponseInfo) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.RespInfo.Size += size
	return size, err
}

func LogRequestResponseMiddleware(logger *zap.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			respInfo := &ResponseInfo{}
			reqInfo := &RequestResponseInfo{
				ResponseWriter: w,
				RespInfo:       respInfo,
			}

			start := time.Now()
			h.ServeHTTP(reqInfo, r)
			elapsed := time.Since(start)

			logger.Info("HTTP request was made",
				zap.String("method", r.Method),
				zap.String("URI", r.RequestURI),
				zap.Int("status_code", reqInfo.RespInfo.StatusCode),
				zap.Int("size", reqInfo.RespInfo.Size),
				zap.Duration("elapsed", elapsed),
			)
		})
	}
}
