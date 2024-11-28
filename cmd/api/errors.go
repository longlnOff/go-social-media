package main

import (
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorf("internal server error: %s, path: %s, error: %s", r.Method, r.URL.Path, err)
	writeJSONError(w, http.StatusInternalServerError, "the server encountered an internal error")
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("bad request: %s, path: %s, error: %s", r.Method, r.URL.Path, err)
	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorf("not found: %s, path: %s, error: %s", r.Method, r.URL.Path, err)
	writeJSONError(w, http.StatusNotFound, err.Error())
}

func (app *application) conflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorf("conflict: %s, path: %s, error: %s", r.Method, r.URL.Path, err)
	writeJSONError(w, http.StatusConflict, err.Error())
}
