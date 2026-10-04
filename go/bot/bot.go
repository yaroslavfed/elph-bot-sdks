package bot

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/yaroslavfed/elph-bot-sdks/go/elphsdk"
	"github.com/yaroslavfed/elph-bot-sdks/go/models"
)

const defaultUpdatesChannelCap = 1024

type HandlerFunc func(context.Context, *Bot, *models.Update)
type ErrorsHandler func(error)
type Option func(*Bot)

type Bot struct {
	token              string
	baseURL            string
	httpClient         *http.Client
	client             *elphsdk.Client
	webhookSecretToken string
	defaultHandlerFunc HandlerFunc
	errorsHandler      ErrorsHandler
	updates            chan *models.Update
	workers            int
	startOnce          sync.Once
}

func New(token string, options ...Option) (*Bot, error) {
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("empty token")
	}
	b := &Bot{token: token, baseURL: "https://elph.app", webhookSecretToken: "", defaultHandlerFunc: defaultHandler, errorsHandler: defaultErrorsHandler, updates: make(chan *models.Update, defaultUpdatesChannelCap), workers: 1}
	for _, option := range options {
		option(b)
	}
	client, err := elphsdk.NewClient(b.baseURL, token, b.httpClient)
	if err != nil {
		return nil, err
	}
	b.client = client
	return b, nil
}

func WithServerURL(baseURL string) Option {
	return func(bot *Bot) { bot.baseURL = baseURL }
}

func WithHTTPClient(client *http.Client) Option { return func(bot *Bot) { bot.httpClient = client } }

func WithWebhookSecretToken(secret string) Option {
	return func(bot *Bot) { bot.webhookSecretToken = secret }
}
func WithDefaultHandler(handler HandlerFunc) Option {
	return func(bot *Bot) {
		if handler != nil {
			bot.defaultHandlerFunc = handler
		}
	}
}
func WithErrorsHandler(handler ErrorsHandler) Option {
	return func(bot *Bot) {
		if handler != nil {
			bot.errorsHandler = handler
		}
	}
}
func WithUpdatesChannelCap(capacity int) Option {
	return func(bot *Bot) {
		if capacity > 0 {
			bot.updates = make(chan *models.Update, capacity)
		}
	}
}
func WithWorkers(workers int) Option {
	return func(bot *Bot) {
		if workers > 0 {
			bot.workers = workers
		}
	}
}

type SetWebhookParams struct {
	URL         string `json:"url"`
	SecretToken string `json:"secret_token,omitempty"`
}

func (b *Bot) SetWebhook(ctx context.Context, params *SetWebhookParams) (bool, error) {
	if params == nil || params.URL == "" {
		return false, fmt.Errorf("webhook URL is required")
	}
	secret := params.SecretToken
	if secret == "" {
		secret = b.webhookSecretToken
	} else {
		b.webhookSecretToken = secret
	}
	if err := b.client.SetWebhook(ctx, params.URL, secret); err != nil {
		return false, err
	}
	return true, nil
}

func (b *Bot) WebhookHandler() http.Handler {
	return elphsdk.WebhookHandler(b.webhookSecretToken, func(_ context.Context, update elphsdk.BotUpdate) error {
		select {
		case b.updates <- toModelsUpdate(update):
			return nil
		default:
			return fmt.Errorf("updates queue is full")
		}
	})
}

func (b *Bot) StartWebhook(ctx context.Context) {
	b.startOnce.Do(func() {
		var waitGroup sync.WaitGroup
		for range b.workers {
			waitGroup.Add(1)
			go func() { defer waitGroup.Done(); b.waitUpdates(ctx) }()
		}
		waitGroup.Wait()
	})
}

type SendMessageParams struct {
	ChatID      any                `json:"chat_id"`
	Text        string             `json:"text"`
	RequestID   string             `json:"request_id,omitempty"`
	ReplyMarkup models.ReplyMarkup `json:"reply_markup,omitempty"`
}

type SendDocumentParams struct {
	ChatID    any              `json:"chat_id"`
	Document  models.InputFile `json:"document"`
	Caption   string           `json:"caption,omitempty"`
	RequestID string           `json:"request_id,omitempty"`
}

func (b *Bot) SendMessage(ctx context.Context, params *SendMessageParams) (*models.Message, error) {
	if params == nil {
		return nil, fmt.Errorf("send message params are required")
	}
	chatID, err := chatIDString(params.ChatID)
	if err != nil {
		return nil, err
	}
	if keyboard, ok := replyKeyboardOf(params.ReplyMarkup); ok {
		if err := b.SetReplyKeyboard(ctx, keyboard); err != nil {
			return nil, err
		}
	}
	inlineKeyboard, err := inlineKeyboardOf(params.ReplyMarkup)
	if err != nil {
		return nil, err
	}
	requestID := params.RequestID
	if requestID == "" {
		requestID = randomRequestID()
	}
	_, err = b.client.SendMessage(ctx, elphsdk.SendMessageRequest{ChatID: chatID, RequestID: requestID, Text: params.Text, InlineKeyboard: inlineKeyboard})
	if err != nil {
		return nil, err
	}
	return &models.Message{Chat: models.Chat{ID: chatID}, Text: params.Text}, nil
}

func (b *Bot) SendDocument(ctx context.Context, params *SendDocumentParams) (*models.Message, error) {
	if params == nil || params.Document.Content == nil {
		return nil, fmt.Errorf("document is required")
	}
	chatID, err := chatIDString(params.ChatID)
	if err != nil {
		return nil, err
	}
	requestID := params.RequestID
	if requestID == "" {
		requestID = randomRequestID()
	}
	_, err = b.client.SendFile(ctx, elphsdk.SendFileRequest{
		ChatID: chatID, RequestID: requestID, Caption: params.Caption, FileName: params.Document.Filename,
	}, params.Document.Content)
	if err != nil {
		return nil, err
	}
	return &models.Message{Chat: models.Chat{ID: chatID}, Text: params.Caption}, nil
}

func (b *Bot) SetReplyKeyboard(ctx context.Context, keyboard models.ReplyKeyboardMarkup) error {
	rows := make([]elphsdk.ReplyKeyboardRow, 0, len(keyboard.Keyboard))
	for _, row := range keyboard.Keyboard {
		buttons := make([]elphsdk.ReplyKeyboardButton, 0, len(row))
		for _, button := range row {
			buttons = append(buttons, elphsdk.ReplyKeyboardButton{Text: button.Text, Command: button.Text})
		}
		rows = append(rows, elphsdk.ReplyKeyboardRow{Buttons: buttons})
	}
	return b.client.SetReplyKeyboard(ctx, elphsdk.ReplyKeyboard{Rows: rows})
}

func (b *Bot) SetMenuCommands(ctx context.Context, commands []models.KeyboardButton) error {
	buttons := make([]elphsdk.CommandButton, 0, len(commands))
	for _, command := range commands {
		buttons = append(buttons, elphsdk.CommandButton{Text: command.Text, Command: command.Text})
	}
	return b.client.SetMenuCommands(ctx, elphsdk.MenuCommands{Buttons: buttons})
}

type SetMyCommandsParams struct {
	Commands []models.BotCommand `json:"commands"`
}

func (b *Bot) SetMyCommands(ctx context.Context, params *SetMyCommandsParams) (bool, error) {
	if params == nil {
		return false, fmt.Errorf("set commands params are required")
	}
	buttons := make([]elphsdk.CommandButton, 0, len(params.Commands))
	for _, command := range params.Commands {
		buttons = append(buttons, elphsdk.CommandButton{Text: command.Description, Command: command.Command})
	}
	if err := b.client.SetMenuCommands(ctx, elphsdk.MenuCommands{Buttons: buttons}); err != nil {
		return false, err
	}
	return true, nil
}

func (b *Bot) waitUpdates(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-b.updates:
			b.defaultHandlerFunc(ctx, b, update)
		}
	}
}
func defaultHandler(_ context.Context, _ *Bot, update *models.Update) {
	log.Printf("[ELPHBOT] [UPDATE] %+v", update)
}
func defaultErrorsHandler(err error) { log.Printf("[ELPHBOT] [ERROR] %v", err) }

func toModelsUpdate(update elphsdk.BotUpdate) *models.Update {
	chat := models.Chat{ID: update.Chat.ID, Type: update.Chat.Type}
	from := models.User{ID: update.From.ID, DisplayName: update.From.DisplayName}
	result := &models.Update{ID: update.UpdateID}
	if update.Message != nil {
		result.Message = &models.Message{ID: update.Message.ID, From: &from, Chat: chat, Text: update.Message.Text, ReplyToID: update.Message.ReplyToMessageID}
		return result
	}
	result.CallbackQuery = &models.CallbackQuery{ID: update.UpdateID, From: from, Message: &models.Message{ID: update.InlineCommand.SourceMessageID, Chat: chat}, Data: update.InlineCommand.CommandID}
	return result
}

func inlineKeyboardOf(markup models.ReplyMarkup) (*elphsdk.InlineKeyboard, error) {
	if markup == nil {
		return nil, nil
	}
	keyboard, ok := markup.(models.InlineKeyboardMarkup)
	if !ok {
		if pointer, isPointer := markup.(*models.InlineKeyboardMarkup); isPointer && pointer != nil {
			keyboard = *pointer
			ok = true
		}
	}
	if !ok {
		if _, isReply := replyKeyboardOf(markup); isReply {
			return nil, nil
		}
		return nil, fmt.Errorf("unsupported reply markup %T", markup)
	}
	rows := make([]elphsdk.InlineKeyboardRow, 0, len(keyboard.InlineKeyboard))
	for _, row := range keyboard.InlineKeyboard {
		buttons := make([]elphsdk.InlineButton, 0, len(row))
		for _, button := range row {
			if button.CallbackData == "" {
				return nil, fmt.Errorf("inline button callback data is required")
			}
			buttons = append(buttons, elphsdk.InlineButton{Text: button.Text, Command: button.CallbackData})
		}
		rows = append(rows, elphsdk.InlineKeyboardRow{Buttons: buttons})
	}
	return &elphsdk.InlineKeyboard{Rows: rows}, nil
}

func replyKeyboardOf(markup models.ReplyMarkup) (models.ReplyKeyboardMarkup, bool) {
	switch keyboard := markup.(type) {
	case models.ReplyKeyboardMarkup:
		return keyboard, true
	case *models.ReplyKeyboardMarkup:
		if keyboard != nil {
			return *keyboard, true
		}
	}
	return models.ReplyKeyboardMarkup{}, false
}

func chatIDString(value any) (string, error) {
	if value == nil {
		return "", fmt.Errorf("chat ID is required")
	}
	result := fmt.Sprint(value)
	if result == "" {
		return "", fmt.Errorf("chat ID is required")
	}
	return result, nil
}
func randomRequestID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("elph-%d", len(bytes))
	}
	return "elph-" + hex.EncodeToString(bytes)
}
