package main

import (
	"time"

	"github.com/longln/go-social-media/internal/db"
	"github.com/longln/go-social-media/internal/env"
	"github.com/longln/go-social-media/internal/mailer"
	"github.com/longln/go-social-media/internal/store"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var version string = "1.0.0"

//	@title			Social Media
//	@description	API for Go Social Media
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	longln
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@BasePath	/v1

//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description

func main() {
	cfg := config{
		apiURL:       env.GetString("EXTERNAL_URL", "localhost:4000"),
		address:      env.GetString("ADDRESS", ":4000"),
		frontendURL:  env.GetString("FRONTEND_URL", "http://localhost:3000"),
		writeTimeout: 10 * time.Second,
		readTimeout:  5 * time.Second,
		idleTimeout:  60 * time.Second,
		env:          env.GetString("ENV", "development"),
		version:      version,
		mail: mailConfig{
			exp: 15 * time.Minute,
			fromEmail: env.GetString("FROM_EMAIL", "3t4H9@example.com"),
			sendGrid: sendGridConfig{
				apiKey: env.GetString("SENDGRID_API_KEY", ""),
			},
		},

		db: dbConfig{
			address: env.GetString("DB_ADDRESS",
				"postgres://admin:adminpassword@localhost:5432/socialnetwork?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 25),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 25),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
	}

	// Logger
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	// turn off stack trace
	config.DisableStacktrace = true

	logger := zap.Must(config.Build()).Sugar()
	defer logger.Sync()

	// Database
	db, err := db.New(cfg.db.address, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()
	logger.Infof("Connected to database %s", cfg.db.address)
	logger.Info("database connection pool established")
	store := store.NewStorage(db)

	mailer := mailer.NewSendGridMailer(
						cfg.mail.fromEmail,
						cfg.mail.sendGrid.apiKey,
					)

	app := application{
		config: cfg,
		store:  store,
		logger: logger,
		mailer: mailer,
	}

	logger.Fatal(app.serve(app.mount()))

}
