package main

import (
	"context"
	"net/http"
	"os"
	"sync"
	"time"

	notificationService "twitch_telegram_bot/internal/service/notification"
	teleUpdatesCheckService "twitch_telegram_bot/internal/service/telegram_updates_check"

	twitchTokenService "twitch_telegram_bot/internal/service/twitch_token"

	telegramClient "twitch_telegram_bot/internal/client/telegram-client"
	twitchClient "twitch_telegram_bot/internal/client/twitch-client"
	twitchOauthClient "twitch_telegram_bot/internal/client/twitch-oauth-client"

	telegramHandler "twitch_telegram_bot/internal/handlers/telegram"
	twitchHandler "twitch_telegram_bot/internal/handlers/twitch"

	telegramService "twitch_telegram_bot/internal/service/telegram"
	twitchService "twitch_telegram_bot/internal/service/twitch"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		logrus.Fatal("Error loading .env file")
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
		twitchOauthClient = twitchOauthClient.NewTwitchOauthClient()
	)

	tts, err := twitchTokenService.NewTwitchTokenService(db, twitchOauthClient)
	if err != nil {
		logrus.Fatalf("cannot init twitchTokenService: %v", err)
	}
	go tts.SyncBg(ctx, time.Minute*5)

	twitchClient := twitchClient.NewTwitchClient(tts)

	telegaService := telegramService.NewService(telegaClient)
	twitchService := twitchService.NewService(twitchClient, twitchOauthClient)

	tns, err := notificationService.NewTwitchNotificationService(db, twitchClient)
	if err != nil {
		logrus.Fatalf("cannot init notificationService: %v", err)
	}
	go tns.SyncBg(ctx, time.Minute*5)

	tucs, err := teleUpdatesCheckService.NewTelegramUpdatesCheckService(twitchClient, tns, telegaService, twitchOauthClient)
	if err != nil {
		logrus.Fatalf("cannot init teleUpdatesCheckService: %v", err)
	}
	go tucs.SyncBg(ctx, time.Second*1)

	telegaHandler := telegramHandler.NewTelegramHandler(telegaService)
	twitchHandler := twitchHandler.NewTwitchHandler(twitchService)

	router1 := mux.NewRouter()
	router2 := mux.NewRouter()

	router2.HandleFunc("/", twitchHandler.GetUserToken)

	router1.HandleFunc("/commands", telegaHandler.GetBotData).Methods("GET").Schemes("HTTP")

	router1.HandleFunc("/twitch/oauth", twitchHandler.GetOAuthToken).Methods("POST").Schemes("HTTP")
	router1.HandleFunc("/twitch/user", twitchHandler.GetUser).Methods("POST").Schemes("HTTP")
	router1.HandleFunc("/twitch/stream", twitchHandler.GetActiveStreamInfoByUser).Methods("POST").Schemes("HTTP")

	logrus.Info("server start...")

	wg := new(sync.WaitGroup)

	wg.Add(2)

	go func() {
		srv := &http.Server{
			Handler:      router1,
			Addr:         "localhost:8084",
			WriteTimeout: 5 * time.Second,
			ReadTimeout:  5 * time.Second,
		}

		logrus.Fatal(srv.ListenAndServe())
		wg.Done()

	}()

	go func() {
		srv := &http.Server{
			Handler:      router2,
			Addr:         "localhost:3000",
			WriteTimeout: 5 * time.Second,
			ReadTimeout:  5 * time.Second,
		}

		logrus.Fatal(srv.ListenAndServe())
		wg.Done()

	}()

	wg.Wait()
}
