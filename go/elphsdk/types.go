package elphsdk

type CommandButton struct {
	Text    string `json:"text"`
	Command string `json:"command"`
}
type InlineButton struct {
	Text     string          `json:"text"`
	Command  string          `json:"command"`
	Children *InlineKeyboard `json:"children,omitempty"`
}
type InlineKeyboard struct {
	Rows []InlineKeyboardRow `json:"rows"`
}
type InlineKeyboardRow struct {
	Buttons []InlineButton `json:"buttons"`
}
type ReplyKeyboard struct {
	Rows []ReplyKeyboardRow `json:"rows"`
}
type ReplyKeyboardRow struct {
	Buttons []ReplyKeyboardButton `json:"buttons"`
}
type ReplyKeyboardButton struct {
	Text     string         `json:"text"`
	Command  string         `json:"command"`
	Children *ReplyKeyboard `json:"children,omitempty"`
}
type MenuCommands struct {
	Buttons []CommandButton `json:"buttons"`
}
type Attachment struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	MimeType    string `json:"mimeType"`
	Size        int64  `json:"size"`
	DownloadURL string `json:"downloadUrl"`
	ExpiresAt   string `json:"expiresAt,omitempty"`
}
type Chat struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}
type Sender struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName,omitempty"`
}
type MessageUpdate struct {
	ID               string       `json:"id"`
	Text             string       `json:"text,omitempty"`
	ReplyToMessageID string       `json:"replyToMessageId,omitempty"`
	Attachments      []Attachment `json:"attachments,omitempty"`
}
type InlineCommandUpdate struct {
	SourceMessageID string `json:"sourceMessageId"`
	CommandID       string `json:"commandId"`
}
type BotUpdate struct {
	SchemaVersion int                  `json:"schemaVersion"`
	UpdateID      string               `json:"updateId"`
	BotID         string               `json:"botId"`
	Chat          Chat                 `json:"chat"`
	From          Sender               `json:"from"`
	CreatedAt     string               `json:"createdAt"`
	Message       *MessageUpdate       `json:"message,omitempty"`
	InlineCommand *InlineCommandUpdate `json:"inlineCommand,omitempty"`
}
type SendMessageRequest struct {
	ChatID           string          `json:"chatId"`
	RequestID        string          `json:"requestId"`
	Text             string          `json:"text,omitempty"`
	ReplyToMessageID string          `json:"replyToMessageId,omitempty"`
	InlineKeyboard   *InlineKeyboard `json:"inlineKeyboard,omitempty"`
}
type SendFileRequest struct {
	ChatID           string
	RequestID        string
	Caption          string
	ReplyToMessageID string
	FileName         string
}
type MessageResponse struct {
	Success bool           `json:"success"`
	Message map[string]any `json:"message"`
}
