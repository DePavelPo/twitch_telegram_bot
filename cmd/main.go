package main

import (
	"log"
	"net/http"
	"time"
	telegreamClient "twitch_telegram_bot/internal/client/telegram-client"
	telegramHandler "twitch_telegram_bot/internal/handlers/telegram"
	telegramService "twitch_telegram_bot/internal/service/telegram"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	// ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

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

	log.Fatal(srv.ListenAndServe())
}
