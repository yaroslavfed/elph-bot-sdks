package elphsdk

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebhookHandlerAcceptsSignedMessage(t *testing.T) {
	body := []byte(`{"schemaVersion":1,"updateId":"update-1","botId":"bot-1","chat":{"id":"chat-1","type":"direct"},"from":{"id":"user-1"},"createdAt":"2026-10-04T00:00:00Z","message":{"id":"message-1"}}`)
	called := false
	handler := WebhookHandler("secret", func(_ context.Context, update BotUpdate) error { called = update.Message != nil; return nil })
	request := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	request.Header.Set(SignatureHeader, signatureForTest(body, "secret"))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent || !called {
		t.Fatalf("status = %d, called = %t", response.Code, called)
	}
}

func TestWebhookHandlerRejectsInvalidSignature(t *testing.T) {
	body := []byte(`{"schemaVersion":1,"updateId":"update-1","botId":"bot-1","chat":{"id":"chat-1"},"from":{"id":"user-1"},"message":{"id":"message-1"}}`)
	request := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	request.Header.Set(SignatureHeader, "sha256=wrong")
	response := httptest.NewRecorder()
	WebhookHandler("secret", func(context.Context, BotUpdate) error { return nil }).ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", response.Code)
	}
}

func signatureForTest(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
