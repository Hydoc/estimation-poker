package main

import (
	"fmt"
	"net/http"
)

func (srv *gameServer) withRequiredQueryParam(param string, next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		queryParam := request.URL.Query().Get(param)
		if len(queryParam) == 0 || !request.URL.Query().Has(param) {
			err := srv.writeJSON(writer, http.StatusBadRequest, envelope{"message": fmt.Sprintf("%s is missing in query", param)}, nil)
			if err != nil {
				srv.serverErrorResponse(writer, request, err)
			}
			return
		}
		next.ServeHTTP(writer, request)
	}
}

func (srv *gameServer) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				writer.Header().Set("Connection", "close")
				srv.serverErrorResponse(writer, request, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(writer, request)
	})
}
