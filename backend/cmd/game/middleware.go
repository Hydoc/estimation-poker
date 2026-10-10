package main

import (
	"fmt"
	"net/http"
)

func (srv *gameServer) withRequiredQueryParam(param string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			queryParam := r.URL.Query().Get(param)
			if len(queryParam) == 0 || !r.URL.Query().Has(param) {
				err := srv.writeJSON(w, http.StatusBadRequest, envelope{"message": fmt.Sprintf("%s is missing in query", param)}, nil)
				if err != nil {
					srv.serverErrorResponse(w, r, err)
				}
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (srv *gameServer) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				srv.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
