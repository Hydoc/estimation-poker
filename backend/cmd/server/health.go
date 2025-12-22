package main

import "net/http"

func (app *application) healthcheckHandler(writer http.ResponseWriter, request *http.Request) {
	data := envelope{
		"status": "available",
		"systemInfo": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	err := app.writeJSON(writer, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(writer, request, err)
	}
}
