package main

import (
	"context"
	"net/http"
	"os"
	"sync"
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
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

const (
	localENV = "local"
	prodENV  = "prod"
)

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		logrus.Fatal("Error loading .env file")
	}

	directAddr, debugAddr, redirectAddr := os.Getenv("DIRECT_ADDR"), os.Getenv("DEBUG_ADDR"), os.Getenv("REDIRECT_ADDR")
	env := os.Getenv("CURRENT_ENV")

	var protocol string
	switch env {
	case localENV:
		protocol = "http"
	case prodENV:
		protocol = "https"
	default:
		logrus.Fatalf("unknown env: %s", env)
	}

	db, err := sqlx.Connect("postgres", os.Getenv("DB_CONN"))
	if err != nil {
		logrus.Fatalf("cannot connect to db: %v", err)
	}

	err = db.Ping()
	if err != nil {
		logrus.Fatalf("cannot ping db: %v", err)
	}

	var (
		telegaClient      = telegramClient.NewTelegramClient()
		twitchOauthClient = twitchOauthClient.NewTwitchOauthClient(protocol, redirectAddr)
		fClient           = fileClient.NewFileClient()
	)

	dbRepo := dbRepository.NewDBRepository(db)

	tts, err := twitchTokenService.NewTwitchTokenService(dbRepo, twitchOauthClient)
	if err != nil {
		logrus.Fatalf("cannot init twitchTokenService: %v", err)
	}
	go tts.SyncBg(ctx, time.Minute*5)

	twitchClient := twitchClient.NewTwitchClient(tts)

	telegaService := telegramService.NewService(telegaClient)
	twitchService := twitchService.NewService(twitchClient, twitchOauthClient)

	tns, err := notificationService.NewTwitchNotificationService(dbRepo, twitchClient, twitchOauthClient)
	if err != nil {
		logrus.Fatalf("cannot init notificationService: %v", err)
	}
	go tns.SyncBg(ctx, time.Minute*5)

	tuas, err := twitchUserAuthservice.NewTwitchUserAuthorizationService(dbRepo, twitchOauthClient, protocol, redirectAddr)
	if err != nil {
		logrus.Fatalf("cannot init twitchUserAuthservice: %v", err)
	}

	tucs, err := teleUpdatesCheckService.NewTelegramUpdatesCheckService(twitchClient, fClient, dbRepo, tns, tuas, telegaService, twitchOauthClient)
	if err != nil {
		logrus.Fatalf("cannot init teleUpdatesCheckService: %v", err)
	}
	go tucs.SyncBg(ctx, time.Second*1)

	telegaHandler := telegramHandler.NewTelegramHandler(telegaService)
	twitchHandler := twitchHandler.NewTwitchHandler(twitchService, tuas)

	debugRouter, directRouter := mux.NewRouter(), mux.NewRouter()

	debugRouter.HandleFunc("/commands", telegaHandler.GetBotData).Methods("GET").Schemes("HTTP")
	debugRouter.HandleFunc("/twitch/oauth", twitchHandler.GetOAuthToken).Methods("POST").Schemes("HTTP")
	debugRouter.HandleFunc("/twitch/user", twitchHandler.GetUser).Methods("POST").Schemes("HTTP")
	debugRouter.HandleFunc("/twitch/stream", twitchHandler.GetActiveStreamInfoByUser).Methods("POST").Schemes("HTTP")

	directRouter.HandleFunc("/user/token/get", twitchHandler.GetUserToken).Methods("POST").Schemes("HTTP")

	// Configure CORS
	handler := middleware.ConfigureCORS(directRouter)

	logrus.Info("server start...")

	wg := new(sync.WaitGroup)

	wg.Add(2)
	go func() {
		srv := &http.Server{
			Handler:      debugRouter,
			Addr:         debugAddr,
			WriteTimeout: 5 * time.Second,
			ReadTimeout:  5 * time.Second,
		}

		logrus.Fatal(srv.ListenAndServe())
		wg.Done()
	}()

	go func() {
		srv := &http.Server{
			Handler:      handler,
			Addr:         directAddr,
			WriteTimeout: 5 * time.Second,
			ReadTimeout:  5 * time.Second,
		}

		if env == prodENV {
			logrus.Fatal(srv.ListenAndServeTLS(os.Getenv("CRT_DIR"), os.Getenv("TLS_KEY_DIR")))
		} else {
			logrus.Fatal(srv.ListenAndServe())
		}

		wg.Done()
	}()

	wg.Wait()
}
