package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func (c *Conveyor) WithGzipSupport(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" &&
			r.Header.Get("Content-Type") != "text/html" {
			next.ServeHTTP(w, r)
			return
		}

		// Check if the request body is Gzip-encoded
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to decode gzip request body", http.StatusInternalServerError)
				return
			}
			r.Body = gr
			if err := gr.Close(); err != nil {
				http.Error(w, "Failed to decode gzip request body", http.StatusInternalServerError)
				return
			}
		}

		// Check if the client supports Gzip encoding
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {

			w.Header().Set("Content-Encoding", "gzip")
			//w.Header().Set("Content-Length", "10")
			gw := &gzipWriter{
				ResponseWriter: w,
				Writer:         gzip.NewWriter(w),
			}
			next.ServeHTTP(gw, r)
			if err := gw.Writer.(*gzip.Writer).Close(); err != nil {
				http.Error(w, "Failed to decode gzip request body", http.StatusInternalServerError)
				return
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.WriteCloser
}

func (w *gzipWriter) Write(b []byte) (int, error) {
	defer w.Writer.Close()
	return w.Writer.Write(b)
}

func (w *gzipWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}
