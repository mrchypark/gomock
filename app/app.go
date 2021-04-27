package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/smeshkov/gomock/config"
)

// RegisterHandlers registers all handlers of the application.
func RegisterHandlers(version, mockPath string, cfg *config.Config, mck *config.Mock) http.Handler {

	// Use gorilla/mux for rich routing.
	// See http://www.gorillatoolkit.org/pkg/mux
	r := mux.NewRouter()

	// Shows if app is healthy
	r.Methods(http.MethodGet).Path("/healthcheck").Handler(appHandler(healthcheckHandler))

	// Shows current version of the App
	r.Methods(http.MethodGet).Path("/version").Handler(appHandler(versionHandler(version)))

	setupAPI(mockPath, mck, r)

	return r
}

// http://blog.golang.org/error-handling-and-go
type appHandler func(http.ResponseWriter, *http.Request) *appError

type appError struct {
	Error   error
	Message string
	Code    int
	Log     *zap.Logger
}

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil { // e is *appError, not os.Error.
		l := e.Log
		if l == nil {
			l = zap.L()
		}
		l.Error(fmt.Sprintf("handler error: status code: %d, message: %s, underlying err: %#v",
			e.Code, e.Message, e.Error))

		http.Error(w, e.Message, e.Code)
	}
}

// func appErrorf(err error, format string, v ...interface{}) *appError {
// 	return &appError{
// 		Error:   err,
// 		Message: fmt.Sprintf(format, v...),
// 		Code:    500,
// 	}
// }

// writeResponse writes response to provided ResponseWriter in JSON format.
func writeResponse(rw http.ResponseWriter, response interface{}) *appError {
	rw.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(rw).Encode(response)
	if err != nil {
		return &appError{
			Error:   err,
			Message: fmt.Sprintf("error in response write: %v", err),
			Code:    http.StatusInternalServerError,
		}
	}
	return nil
}
