package main

import "net/http"

func (srv *gameServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	data := envelope{
		"environment": srv.config.env,
		"version":     version,
	}

	err := srv.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		srv.serverErrorResponse(w, r, err)
		return
	}
}
