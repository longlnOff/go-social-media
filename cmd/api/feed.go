package main

import (
	"net/http"

	"github.com/longln/go-social-media/internal/store"
)

func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: pagination, filters, sort
	fq := store.PaginatedFeedQuery{
		Limit:  10,
		Offset: 0,
		Sort:   "desc",
	}

	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(fq); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()
	feeds, err := app.store.Posts.GetUserFeed(ctx, int64(6), fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if err := app.jsonResponse(w, http.StatusOK, feeds); err != nil {
		app.internalServerError(w, r, err)
	}
}
