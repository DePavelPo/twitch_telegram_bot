package main

import (
	"context"
	"net/http"
	"time"

	teleUpdatesCheckService "twitch_telegram_bot/internal/service/telegram_updates_check"

	telegreamClient "twitch_telegram_bot/internal/client/telegram-client"
	telegramHandler "twitch_telegram_bot/internal/handlers/telegram"
	telegramService "twitch_telegram_bot/internal/service/telegram"

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
	go tucs.SyncBg(ctx, time.Hour*1)

	var (
		telegaClient = telegreamClient.NewTelegramClient()
	)

	telegaService := telegramService.NewService(telegaClient)

	telegaHandler := telegramHandler.NewTelegramHandler(telegaService)

	router := mux.NewRouter()

	router.HandleFunc("/commands", telegaHandler.GetBotData).Methods("GET").Schemes("HTTP")

	logrus.Info("server start...")

	srv := &http.Server{
		Handler:      router,
		Addr:         "localhost:8084",
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	logrus.Fatal(srv.ListenAndServe())
}
