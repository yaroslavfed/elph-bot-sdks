package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/yaroslavfed/elph-bot-sdks/go/bot"
	"github.com/yaroslavfed/elph-bot-sdks/go/models"
)

func main() {
	client, err := bot.New(
		os.Getenv("ELPH_BOT_TOKEN"),
		bot.WithServerURL(os.Getenv("ELPH_API_URL")),
		bot.WithWebhookSecretToken(os.Getenv("ELPH_WEBHOOK_SECRET")),
		bot.WithDefaultHandler(func(ctx context.Context, client *bot.Bot, update *models.Update) {
			if update.Message == nil {
				return
			}
			_, err := client.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    update.Message.Chat.ID,
				RequestID: update.ID,
				Text:      "Echo: " + update.Message.Text,
			})
			if err != nil {
				log.Printf("send message: %v", err)
			}
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := client.SetWebhook(context.Background(), &bot.SetWebhookParams{URL: os.Getenv("ELPH_WEBHOOK_URL")}); err != nil {
		log.Fatal(err)
	}
	go client.StartWebhook(context.Background())
	log.Fatal(http.ListenAndServe(":8080", client.WebhookHandler()))
}
