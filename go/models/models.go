package models

import "io"

type Update struct {
	ID            string         `json:"update_id"`
	Message       *Message       `json:"message,omitempty"`
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}

type Message struct {
	ID        string `json:"message_id"`
	From      *User  `json:"from,omitempty"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text,omitempty"`
	ReplyToID string `json:"reply_to_message_id,omitempty"`
}

type Chat struct {
	ID   any    `json:"id"`
	Type string `json:"type,omitempty"`
}

type User struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name,omitempty"`
}

type CallbackQuery struct {
	ID      string   `json:"id"`
	From    User     `json:"from"`
	Message *Message `json:"message,omitempty"`
	Data    string   `json:"data,omitempty"`
}

type ReplyMarkup any

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
}

type ReplyKeyboardMarkup struct {
	Keyboard [][]KeyboardButton `json:"keyboard"`
}

type KeyboardButton struct {
	Text string `json:"text"`
}

type BotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

type InputFile struct {
	Filename string
	Content  io.Reader
}
