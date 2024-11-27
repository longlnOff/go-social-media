package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/longln/go-social-media/internal/store"
)

type userKey string

const userCTX userKey = "user"

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getUserFromContext(r)

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
	}
}

type FollowerUser struct {
	UserID int64 `json:"user_id"`
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := app.getUserFromContext(r)

	//TODO: Revert back to auth userID from ctx
	var payload FollowerUser
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	ctx := r.Context()
	if err := app.store.Followers.Follow(ctx, followerUser.ID, payload.UserID); err != nil {
		switch {
		case errors.Is(err, store.ErrConflict):
			app.conflictResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	unfollowedUser := app.getUserFromContext(r)

	// TODO: Revert back to auth userID from ctx
	var payload FollowerUser
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	ctx := r.Context()
	if err := app.store.Followers.Unfollow(ctx, unfollowedUser.ID, payload.UserID); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getUserFromContext(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCTX).(*store.User)
	return user
}


func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")

		userIDInt, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			app.badRequestResponse(w, r, err)
			return
		}

		ctx := r.Context()
		user, err := app.store.Users.GetByID(ctx, userIDInt)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrRecordNotFound):
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}
		ctx = context.WithValue(ctx, userCTX, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}