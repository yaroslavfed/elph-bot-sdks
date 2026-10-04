package elphsdk

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientSendsAuthorizedWebhookRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPut || request.URL.Path != "/bot-api/v1/webhook" {
			t.Fatalf("request = %s %s", request.Method, request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("authorization = %q", request.Header.Get("Authorization"))
		}
		body, _ := io.ReadAll(request.Body)
		if !bytes.Contains(body, []byte(`"url":"https://bot.example/webhook"`)) {
			t.Fatalf("body = %s", body)
		}
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "token", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if err := client.SetWebhook(context.Background(), "https://bot.example/webhook", "secret"); err != nil {
		t.Fatal(err)
	}
}

func TestClientSendsFileAsMultipart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := request.ParseMultipartForm(1024 * 1024); err != nil {
			t.Fatal(err)
		}
		if request.FormValue("chatId") != "chat-1" || request.FormValue("requestId") != "request-1" {
			t.Fatalf("form = %#v", request.MultipartForm.Value)
		}
		file, header, err := request.FormFile("file")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		content, _ := io.ReadAll(file)
		if header.Filename != "hello.txt" || string(content) != "hello" {
			t.Fatalf("file = %q %q", header.Filename, content)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "token", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.SendFile(context.Background(), SendFileRequest{ChatID: "chat-1", RequestID: "request-1", FileName: "hello.txt"}, bytes.NewBufferString("hello")); err != nil {
		t.Fatal(err)
	}
}
