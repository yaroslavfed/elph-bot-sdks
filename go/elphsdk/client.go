package elphsdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type APIError struct {
	StatusCode int
	ErrorCode  string
	Message    string
	Details    any
}

func (e *APIError) Error() string {
	if e.ErrorCode == "" {
		return fmt.Sprintf("Elph Bot API returned HTTP %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("Elph Bot API returned HTTP %d (%s): %s", e.StatusCode, e.ErrorCode, e.Message)
}

type Client struct {
	baseURL    *url.URL
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, token string, httpClient *http.Client) (*Client, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse Elph API URL: %w", err)
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, errors.New("Elph API URL must be absolute")
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{baseURL: parsedURL, token: token, httpClient: httpClient}, nil
}

func (c *Client) SetWebhook(ctx context.Context, webhookURL, secret string) error {
	return c.doJSON(ctx, http.MethodPut, "/bot-api/v1/webhook", struct {
		URL    string `json:"url"`
		Secret string `json:"secret,omitempty"`
	}{webhookURL, secret}, nil)
}

func (c *Client) SetReplyKeyboard(ctx context.Context, keyboard ReplyKeyboard) error {
	return c.doJSON(ctx, http.MethodPut, "/bot-api/v1/reply-keyboard", struct {
		ReplyKeyboard ReplyKeyboard `json:"replyKeyboard"`
	}{keyboard}, nil)
}

func (c *Client) SetMenuCommands(ctx context.Context, commands MenuCommands) error {
	return c.doJSON(ctx, http.MethodPut, "/bot-api/v1/menu-commands", struct {
		MenuCommands MenuCommands `json:"menuCommands"`
	}{commands}, nil)
}

func (c *Client) SendMessage(ctx context.Context, request SendMessageRequest) (MessageResponse, error) {
	var response MessageResponse
	err := c.doJSON(ctx, http.MethodPost, "/bot-api/v1/messages", request, &response)
	return response, err
}

func (c *Client) SendFile(ctx context.Context, request SendFileRequest, content io.Reader) (MessageResponse, error) {
	reader, writer := io.Pipe()
	multipartWriter := multipart.NewWriter(writer)
	errCh := make(chan error, 1)
	go func() {
		defer writer.Close()
		defer multipartWriter.Close()
		for key, value := range map[string]string{"chatId": request.ChatID, "requestId": request.RequestID, "caption": request.Caption, "replyToMessageId": request.ReplyToMessageID} {
			if value != "" {
				if err := multipartWriter.WriteField(key, value); err != nil {
					errCh <- err
					return
				}
			}
		}
		name := request.FileName
		if name == "" {
			name = "file"
		}
		part, err := multipartWriter.CreateFormFile("file", path.Base(name))
		if err == nil {
			_, err = io.Copy(part, content)
		}
		errCh <- err
	}()
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint("/bot-api/v1/files"), reader)
	if err != nil {
		return MessageResponse{}, err
	}
	httpRequest.Header.Set("Authorization", "Bearer "+c.token)
	httpRequest.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	response, err := c.httpClient.Do(httpRequest)
	writeErr := <-errCh
	if err != nil {
		return MessageResponse{}, err
	}
	if writeErr != nil {
		return MessageResponse{}, fmt.Errorf("write multipart request: %w", writeErr)
	}
	defer response.Body.Close()
	if err := apiError(response); err != nil {
		return MessageResponse{}, err
	}
	var result MessageResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return MessageResponse{}, fmt.Errorf("decode Bot API response: %w", err)
	}
	return result, nil
}

func (c *Client) doJSON(ctx context.Context, method, endpoint string, body, output any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode Bot API request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, method, c.endpoint(endpoint), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if err := apiError(response); err != nil {
		return err
	}
	if output == nil {
		return nil
	}
	if err := json.NewDecoder(response.Body).Decode(output); err != nil {
		return fmt.Errorf("decode Bot API response: %w", err)
	}
	return nil
}

func (c *Client) endpoint(endpoint string) string {
	copy := *c.baseURL
	copy.Path = path.Join(c.baseURL.Path, endpoint)
	return copy.String()
}

func apiError(response *http.Response) error {
	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		return nil
	}
	var payload struct {
		ErrorCode string `json:"errorCode"`
		Message   string `json:"message"`
		Details   any    `json:"details"`
	}
	_ = json.NewDecoder(io.LimitReader(response.Body, 64*1024)).Decode(&payload)
	if payload.Message == "" {
		payload.Message = strings.TrimSpace(response.Status)
	}
	return &APIError{StatusCode: response.StatusCode, ErrorCode: payload.ErrorCode, Message: payload.Message, Details: payload.Details}
}
