package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// wrapResponseWriter 包装 gin.ResponseWriter 实现 http.ResponseWriter 接口
type responseWriterWrapper struct {
	gin.ResponseWriter
}

func wrapResponseWriter(w gin.ResponseWriter) http.ResponseWriter {
	return &responseWriterWrapper{ResponseWriter: w}
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}
