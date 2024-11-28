package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/longln/go-social-media/internal/store"
)

type postKey string

const postCTX postKey = "post"

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required"`
	Content string   `json:"content" validate:"required"`
	Tags    []string `json:"tags" validate:"required"`
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(&payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	userID := 1
	post := store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		// TODO: change after auth
		UserID: int64(userID),
		User:   store.User{ID: int64(userID)},
	}

	ctx := r.Context()
	if err := app.store.Posts.Create(ctx, &post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	comments, err := app.store.Comments.GetByPostID(r.Context(), post.ID)
	app.logger.Infof("comments: %v", comments)
	app.logger.Infof("err: %v", err)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	post.Comments = comments

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	result := chi.URLParam(r, "postID")
	postID, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	err = app.store.Posts.Delete(r.Context(), postID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrRecordNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type UpdatePostPayload struct {
	Title   *string `json:"title" validate:"omitempty"`
	Content *string `json:"content" validate:"omitempty"`
}

func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)
	var payload UpdatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(&payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if payload.Title != nil {
		post.Title = *payload.Title
	}

	if payload.Content != nil {
		post.Content = *payload.Content
	}

	err := app.store.Posts.Update(r.Context(), post)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrRecordNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := chi.URLParam(r, "postID")

		postIDInt, err := strconv.ParseInt(postID, 10, 64)
		if err != nil {
			app.badRequestResponse(w, r, err)
			return
		}

		ctx := r.Context()
		post, err := app.store.Posts.GetByID(ctx, postIDInt)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrRecordNotFound):
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}
		ctx = context.WithValue(ctx, postCTX, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value(postCTX).(*store.Post)
	return post
}
