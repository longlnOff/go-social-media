package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/longln/go-social-media/docs"
	"github.com/longln/go-social-media/internal/store"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type application struct {
	config config
	store  store.Storage
	logger *zap.SugaredLogger
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
	apiURL       string
	mail         mailConfig
}

type mailConfig struct {
	exp time.Duration
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

		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.address)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

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
			r.Put("/activate/{token}", app.activateUserHandler)
			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.userContextMiddleware)
				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})

			r.Group(func(r chi.Router) {
				// TODO: Authentication middleware
				r.Get("/feed", app.getUserFeedHandler)
			})

		})

		// public routes
		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
			// r.Post("/token", app.loginUserHandler)
		})
	})

	return r
}

func (app *application) serve(mux http.Handler) error {
	// Docs
	docs.SwaggerInfo.Version = app.config.version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"

	server := http.Server{
		Addr:         app.config.address,
		Handler:      mux,
		WriteTimeout: app.config.writeTimeout,
		ReadTimeout:  app.config.readTimeout,
		IdleTimeout:  app.config.idleTimeout,
	}
	app.logger.Infof("Starting server on %s", app.config.address)
	return server.ListenAndServe()
}
