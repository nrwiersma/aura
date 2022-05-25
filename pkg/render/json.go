package render

import (
	"fmt"
	"net/http"

	jsoniter "github.com/json-iterator/go"
)

type apiError struct {
	Msg string `json:"error"`
}

// JSONInternalServerError writes a json internal server error.
func JSONInternalServerError(rw http.ResponseWriter) {
	JSONError(rw, http.StatusInternalServerError, "internal server error")
}

// JSONErrorf writes a json error message.
func JSONErrorf(rw http.ResponseWriter, code int, msg string, args ...interface{}) {
	JSONError(rw, code, fmt.Sprintf(msg, args...))
}

// JSONError writes a json error message.
func JSONError(rw http.ResponseWriter, code int, msg string) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(code)

	apiErr := apiError{Msg: msg}
	b, err := jsoniter.Marshal(apiErr)
	if err != nil {
		_, _ = rw.Write([]byte(`{"error":"internal server error"}`))
	}

	_, _ = rw.Write(b)
}

// JSON writes a json response.
func JSON(rw http.ResponseWriter, code int, v interface{}) error {
	b, err := jsoniter.Marshal(v)
	if err != nil {
		return err
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(code)
	_, _ = rw.Write(b)
	return nil
}
