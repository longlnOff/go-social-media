package main

import (
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "available",
		"env":     app.config.env,
		"version": app.config.version,
	}
	if err := app.jsonResponse(w, http.StatusOK, data); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
	}
}
