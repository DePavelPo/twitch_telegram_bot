package main

import (
	"context"
	"net/http"
	"os"
	"time"

	teleUpdatesCheckService "twitch_telegram_bot/internal/service/telegram_updates_check"

	telegramClient "twitch_telegram_bot/internal/client/telegram-client"
	twitchClient "twitch_telegram_bot/internal/client/twitch-client"

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
		telegaClient = telegramClient.NewTelegramClient()
		twitchClient = twitchClient.NewTwitchClient()
	)

	tucs, err := teleUpdatesCheckService.NewTelegramUpdatesCheckService(twitchClient)
	if err != nil {
		logrus.Fatalf("cannot init teleUpdatesCheckService: %v", err)
	}
	go tucs.SyncBg(ctx, time.Second*1)

	telegaService := telegramService.NewService(telegaClient)
	twitchService := twitchService.NewService(twitchClient)

	telegaHandler := telegramHandler.NewTelegramHandler(telegaService)
	twitchHandler := twitchHandler.NewTwitchHandler(twitchService)

	router := mux.NewRouter()

	router.HandleFunc("/commands", telegaHandler.GetBotData).Methods("GET").Schemes("HTTP")

	router.HandleFunc("/twitch/oauth", twitchHandler.GetOAuthToken).Methods("POST").Schemes("HTTP")
	router.HandleFunc("/twitch/user", twitchHandler.GetUser).Methods("POST").Schemes("HTTP")
	router.HandleFunc("/twitch/stream", twitchHandler.GetActiveStreamInfoByUser).Methods("POST").Schemes("HTTP")

	logrus.Info("server start...")

	srv := &http.Server{
		Handler:      router,
		Addr:         "localhost:8084",
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	logrus.Fatal(srv.ListenAndServe())
}
