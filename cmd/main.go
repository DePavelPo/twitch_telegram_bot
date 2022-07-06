package main

import (
	"context"
	"net/http"
	"time"

	teleUpdatesCheckService "twitch_telegram_bot/internal/service/telegram_updates_check"

	telegramClient "twitch_telegram_bot/internal/client/telegram-client"
	twitchClient "twitch_telegram_bot/internal/client/twitch-client"

	telegramHandler "twitch_telegram_bot/internal/handlers/telegram"
	twitchHandler "twitch_telegram_bot/internal/handlers/twitch"

	telegramService "twitch_telegram_bot/internal/service/telegram"
	twitchService "twitch_telegram_bot/internal/service/twitch"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		logrus.Fatal("Error loading .env file")
	}

	tucs, err := teleUpdatesCheckService.NewTelegramUpdatesCheckService()
	if err != nil {
		logrus.Fatalf("cannot init teleUpdatesCheckService: %v", err)
	}
	go tucs.SyncBg(ctx, time.Second*1)

	var (
		telegaClient = telegramClient.NewTelegramClient()
		twitchClient = twitchClient.NewTwitchClient()
	)

	telegaService := telegramService.NewService(telegaClient)
	twitchService := twitchService.NewService(twitchClient)

	telegaHandler := telegramHandler.NewTelegramHandler(telegaService)
	twitchHandler := twitchHandler.NewTwitchHandler(twitchService)

	router := mux.NewRouter()

	router.HandleFunc("/commands", telegaHandler.GetBotData).Methods("GET").Schemes("HTTP")

	// TODO: add handle with return twitch user info, after connect with telegram req
	router.HandleFunc("/twitch/oauth", twitchHandler.GetOAuthToken).Methods("POST").Schemes("HTTP")

	logrus.Info("server start...")

	srv := &http.Server{
		Handler:      router,
		Addr:         "localhost:8084",
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	logrus.Fatal(srv.ListenAndServe())
}
