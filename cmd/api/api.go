package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/longln/go-social-media/internal/store"
)

type application struct {
	config config
	store  store.Storage
}

type dbConfig struct {
	address      string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type config struct {
	address      string
	writeTimeout time.Duration
	readTimeout  time.Duration
	idleTimeout  time.Duration
	db           dbConfig
	env          string
	version      string
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// A good base middleware stack
	// r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthcheckHandler)

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", app.createPostHandler)

			r.Route("/{postID}", func(r chi.Router) {
				r.Use(app.postsContextMiddleware)
				r.Get("/", app.getPostHandler)
				r.Delete("/", app.deletePostHandler)
				r.Patch("/", app.updatePostHandler)
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.userContextMiddleware)
				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})
		})
	})

	return r
}

func (app *application) serve(mux http.Handler) error {

	server := http.Server{
		Addr:         app.config.address,
		Handler:      mux,
		WriteTimeout: app.config.writeTimeout,
		ReadTimeout:  app.config.readTimeout,
		IdleTimeout:  app.config.idleTimeout,
	}
	log.Printf("Starting server on %s", app.config.address)
	return server.ListenAndServe()
}
