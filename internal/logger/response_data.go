package logger

import "net/http"

type responseData struct {
	status int
	size   int
}

// ResponseLoggerWriter - stores the size of the response body and the response code.
type ResponseLoggerWriter struct {
	http.ResponseWriter
	responseData *responseData
}

// NewResponseLoggerWriter - Object Constructor.
func NewResponseLoggerWriter(w http.ResponseWriter) *ResponseLoggerWriter {
	return &ResponseLoggerWriter{
		ResponseWriter: w,
		responseData:   &responseData{},
	}
}

func (r *ResponseLoggerWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	if err != nil {
		return 0, err
	}

	r.responseData.size += size

	return size, nil
}

func (r *ResponseLoggerWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
