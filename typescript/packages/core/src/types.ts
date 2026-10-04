export interface User {
  id: string;
  displayName?: string;
}

export interface Chat {
  id: string;
  type: string;
}

export interface Attachment {
  id: string;
  name?: string;
  mimeType?: string;
  size?: number;
  downloadUrl?: string;
  expiresAt?: string;
}

export interface Message {
  id: string;
  text?: string;
  replyToMessageId?: string;
  attachments?: Attachment[];
}

export interface InlineCommand {
  sourceMessageId: string;
  commandId: string;
}

export interface BotUpdate {
  schemaVersion: 1;
  updateId: string;
  botId: string;
  chat: Chat;
  from: User;
  createdAt: string;
  message?: Message;
  inlineCommand?: InlineCommand;
}

export interface CommandButton {
  text: string;
  command: string;
}

export interface InlineKeyboardButton extends CommandButton {
  children?: InlineKeyboard;
}

export interface InlineKeyboardRow {
  buttons: InlineKeyboardButton[];
}

export interface InlineKeyboard {
  rows: InlineKeyboardRow[];
}

export interface ReplyKeyboardButton extends CommandButton {
  children?: ReplyKeyboard;
}

export interface ReplyKeyboardRow {
  buttons: ReplyKeyboardButton[];
}

export interface ReplyKeyboard {
  rows: ReplyKeyboardRow[];
}

export interface MenuCommands {
  buttons: CommandButton[];
}

export interface SetWebhookRequest {
  url: string;
  secret?: string;
}

export interface SendMessageRequest {
  chatId: string;
  requestId: string;
  text?: string;
  inlineKeyboard?: InlineKeyboard;
  replyToMessageId?: string;
}

export interface SendFileRequest {
  chatId: string;
  requestId: string;
  file: Blob;
  filename: string;
  caption?: string;
  replyToMessageId?: string;
}

export interface ApiSuccess<T = undefined> {
  success: true;
  message?: T;
}
