package bot

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yaroslavfed/elph-bot-sdks/go/models"
)

func TestWebhookHandlerMapsInlineCommandToCallbackQuery(t *testing.T) {
	updates := make(chan *models.Update, 1)
	bot, err := New("token", WithWebhookSecretToken("secret"), WithDefaultHandler(func(_ context.Context, _ *Bot, update *models.Update) { updates <- update }))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go bot.StartWebhook(ctx)
	body := []byte(`{"schemaVersion":1,"updateId":"update-1","botId":"bot-1","chat":{"id":"chat-1","type":"direct"},"from":{"id":"user-1"},"createdAt":"2026-10-04T00:00:00Z","inlineCommand":{"sourceMessageId":"message-1","commandId":"choice-a"}}`)
	request := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	request.Header.Set("x-elph-bot-signature", signature(body, "secret"))
	response := httptest.NewRecorder()
	bot.WebhookHandler().ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}
	select {
	case update := <-updates:
		if update.CallbackQuery == nil || update.CallbackQuery.Data != "choice-a" || update.CallbackQuery.Message.Chat.ID != "chat-1" {
			t.Fatalf("update = %#v", update)
		}
	case <-time.After(time.Second):
		t.Fatal("handler was not called")
	}
}

func TestSendMessageMapsInlineKeyboard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/bot-api/v1/messages" {
			t.Fatalf("request = %s %s", request.Method, request.URL.Path)
		}
		body, _ := io.ReadAll(request.Body)
		for _, expected := range []string{`"chatId":"42"`, `"command":"choice-a"`} {
			if !bytes.Contains(body, []byte(expected)) {
				t.Fatalf("body = %s", body)
			}
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()
	bot, err := New("token", WithServerURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	_, err = bot.SendMessage(context.Background(), &SendMessageParams{ChatID: 42, RequestID: "request-1", Text: "Choose", ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{{{Text: "A", CallbackData: "choice-a"}}}}})
	if err != nil {
		t.Fatal(err)
	}
}

func signature(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
