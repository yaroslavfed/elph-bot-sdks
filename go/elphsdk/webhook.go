package elphsdk

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

const SignatureHeader = "x-elph-bot-signature"

type UpdateHandler func(context.Context, BotUpdate) error

func WebhookHandler(secret string, handler UpdateHandler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			writer.Header().Set("Allow", http.MethodPost)
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(io.LimitReader(request.Body, 1024*1024+1))
		if err != nil {
			http.Error(writer, "read webhook payload", http.StatusBadRequest)
			return
		}
		if len(body) > 1024*1024 {
			http.Error(writer, "webhook payload too large", http.StatusRequestEntityTooLarge)
			return
		}
		if secret != "" && !validSignature(body, request.Header.Get(SignatureHeader), secret) {
			http.Error(writer, "invalid webhook signature", http.StatusUnauthorized)
			return
		}
		var update BotUpdate
		if err := json.Unmarshal(body, &update); err != nil {
			http.Error(writer, "invalid update payload", http.StatusBadRequest)
			return
		}
		if update.SchemaVersion != 1 || update.UpdateID == "" || update.BotID == "" || update.Chat.ID == "" || update.From.ID == "" || (update.Message == nil) == (update.InlineCommand == nil) {
			http.Error(writer, "unsupported update payload", http.StatusBadRequest)
			return
		}
		if err := handler(request.Context(), update); err != nil {
			http.Error(writer, "update handling failed", http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	})
}

func validSignature(body []byte, value string, secret string) bool {
	const prefix = "sha256="
	if !strings.HasPrefix(value, prefix) {
		return false
	}
	provided, err := hex.DecodeString(strings.TrimPrefix(value, prefix))
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return hmac.Equal(provided, mac.Sum(nil))
}
