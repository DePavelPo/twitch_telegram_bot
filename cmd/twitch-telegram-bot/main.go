package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	fileClient "twitch_telegram_bot/internal/client/file"
	telegramClient "twitch_telegram_bot/internal/client/telegram-client"
	twitchClient "twitch_telegram_bot/internal/client/twitch-client"
	twitchOauthClient "twitch_telegram_bot/internal/client/twitch-oauth-client"
	"twitch_telegram_bot/internal/middleware"

	telegramHandler "twitch_telegram_bot/internal/handlers/telegram"
	twitchHandler "twitch_telegram_bot/internal/handlers/twitch"

	notificationService "twitch_telegram_bot/internal/service/notification"
	telegramService "twitch_telegram_bot/internal/service/telegram"
	teleUpdatesCheckService "twitch_telegram_bot/internal/service/telegram_updates_check"
	twitchService "twitch_telegram_bot/internal/service/twitch"
	twitchUserAuthservice "twitch_telegram_bot/internal/service/twitch-user-authorization"
	twitchTokenService "twitch_telegram_bot/internal/service/twitch_token"

	dbRepository "twitch_telegram_bot/db/repository"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

const (
	localENV = "local"
	prodENV  = "prod"
)

type config struct {
	APIAddr      string `envconfig:"API_ADDR" required:"true"`
	DebugAddr    string `envconfig:"DEBUG_ADDR" required:"true"`
	RedirectAddr string `envconfig:"REDIRECT_ADDR" required:"true"`
	Env          string `envconfig:"CURRENT_ENV" required:"true"`
	DBConn       string `envconfig:"DB_CONN" required:"true"`
}

func loadConfig() (config, error) {
	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		logrus.WithError(err).Fatal("failed to load config")
		return config{}, err
	}
	if cfg.APIAddr == "" || cfg.DebugAddr == "" || cfg.RedirectAddr == "" || cfg.Env == "" || cfg.DBConn == "" {
		return config{}, envconfig.ErrInvalidSpecification
	}
	return cfg, nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := loadConfig()
	if err != nil {
		logrus.Fatalf("missing or invalid config: %v", err)
	}

	var protocol string
	switch cfg.Env {
	case localENV:
		protocol = "http"
	case prodENV:
		protocol = "https"
	default:
		logrus.Fatalf("unknown env: %s", cfg.Env)
	}

	db, err := sqlx.Connect("postgres", cfg.DBConn)
	if err != nil {
		logrus.WithError(err).Fatal("cannot connect to db")
	}
	if err := db.Ping(); err != nil {
		logrus.WithError(err).Fatal("cannot ping db")
	}
	dbRepo := dbRepository.NewDBRepository(db)

	// Initialize clients and services
	telegaClient := telegramClient.NewTelegramClient()
	twitchOauthClient := twitchOauthClient.NewTwitchOauthClient(protocol, cfg.RedirectAddr)
	fClient := fileClient.NewFileClient()

	tts, err := twitchTokenService.NewTwitchTokenService(dbRepo, twitchOauthClient)
	if err != nil {
		logrus.WithError(err).Fatal("cannot init twitchTokenService")
	}
	go tts.SyncBg(ctx, 5*time.Minute)

	twitchClient := twitchClient.NewTwitchClient(tts)
	telegaService := telegramService.NewService(telegaClient)
	twitchService := twitchService.NewService(twitchClient, twitchOauthClient)

	tns, err := notificationService.NewTwitchNotificationService(dbRepo, twitchClient, twitchOauthClient)
	if err != nil {
		logrus.WithError(err).Fatal("cannot init notificationService")
	}
	go tns.SyncBg(ctx, 5*time.Minute)

	tuas, err := twitchUserAuthservice.NewTwitchUserAuthorizationService(dbRepo, twitchOauthClient, protocol, cfg.RedirectAddr)
	if err != nil {
		logrus.WithError(err).Fatal("cannot init twitchUserAuthservice")
	}
	tucs, err := teleUpdatesCheckService.NewTelegramUpdatesCheckService(twitchClient, fClient, dbRepo, tns, tuas, telegaService, twitchOauthClient)
	if err != nil {
		logrus.WithError(err).Fatal("cannot init teleUpdatesCheckService")
	}
	go tucs.SyncBg(ctx, time.Second)

	_ = telegramHandler.NewTelegramHandler(telegaService)
	twitchHandler := twitchHandler.NewTwitchHandler(twitchService, tuas)

	// Routers
	apiRouter := mux.NewRouter()
	apiRouter.HandleFunc("/api/user/token/set", twitchHandler.GetUserToken).Methods("POST").Schemes("http")
	apiHandlerWithCORS := middleware.ConfigureCORS(apiRouter)

	debugRouter := mux.NewRouter()
	debugRouter.HandleFunc("/admin/twitch/user", twitchHandler.GetUser).Methods("POST").Schemes("http")
	debugRouter.HandleFunc("/admin/twitch/stream", twitchHandler.GetActiveStreamInfoByUser).Methods("POST").Schemes("http")

	logrus.Info("server start...")

	debugSrv := &http.Server{
		Handler:      debugRouter,
		Addr:         cfg.DebugAddr,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}
	apiSrv := &http.Server{
		Handler:      apiHandlerWithCORS,
		Addr:         cfg.APIAddr,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := debugSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("debug server error")
		}
	}()

	go func() {
		defer wg.Done()
		if err := apiSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("api server error")
		}
	}()

	<-ctx.Done()
	logrus.Info("shutting down servers...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := debugSrv.Shutdown(shutdownCtx); err != nil {
		logrus.WithError(err).Fatal("debug server shutdown error")
	}
	if err := apiSrv.Shutdown(shutdownCtx); err != nil {
		logrus.WithError(err).Fatal("api server shutdown error")
	}

	wg.Wait()
	logrus.Info("servers stopped successfully")
}
