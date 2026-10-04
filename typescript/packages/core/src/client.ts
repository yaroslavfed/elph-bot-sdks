import { ElphBotApiError } from './errors.js';
import type {
  ApiSuccess,
  MenuCommands,
  ReplyKeyboard,
  SendFileRequest,
  SendMessageRequest,
  SetWebhookRequest,
} from './types.js';

export interface ElphBotClientOptions {
  apiUrl: string;
  token: string;
  fetch?: typeof fetch;
}

export class ElphBotClient {
  private readonly apiUrl: string;
  private readonly fetchImpl: typeof fetch;

  public constructor(private readonly options: ElphBotClientOptions) {
    if (!options.apiUrl) {
      throw new Error('apiUrl is required');
    }
    if (!options.token) {
      throw new Error('token is required');
    }
    this.apiUrl = options.apiUrl.replace(/\/$/, '');
    this.fetchImpl = options.fetch ?? fetch;
  }

  public setWebhook(request: SetWebhookRequest): Promise<ApiSuccess> {
    return this.request('PUT', '/bot-api/v1/webhook', request);
  }

  public setReplyKeyboard(replyKeyboard: ReplyKeyboard): Promise<ApiSuccess> {
    return this.request('PUT', '/bot-api/v1/reply-keyboard', { replyKeyboard });
  }

  public setMyCommands(menuCommands: MenuCommands): Promise<ApiSuccess> {
    return this.request('PUT', '/bot-api/v1/menu-commands', { menuCommands });
  }

  public sendMessage(request: SendMessageRequest): Promise<ApiSuccess> {
    return this.request('POST', '/bot-api/v1/messages', request);
  }

  public async sendFile(request: SendFileRequest): Promise<ApiSuccess> {
    const body = new FormData();
    body.set('chatId', request.chatId);
    body.set('requestId', request.requestId);
    body.set('file', request.file, request.filename);
    if (request.caption !== undefined) body.set('caption', request.caption);
    if (request.replyToMessageId !== undefined) body.set('replyToMessageId', request.replyToMessageId);
    return this.request('POST', '/bot-api/v1/files', body);
  }

  private async request(method: string, path: string, body?: unknown): Promise<ApiSuccess> {
    const isFormData = body instanceof FormData;
    const response = await this.fetchImpl(`${this.apiUrl}${path}`, {
      method,
      headers: {
        authorization: `Bearer ${this.options.token}`,
        ...(body !== undefined && !isFormData ? { 'content-type': 'application/json' } : {}),
      },
      body: body === undefined ? undefined : isFormData ? body : JSON.stringify(body),
    });

    const payload = await readJson(response);
    if (!response.ok) {
      const error = isRecord(payload) ? payload : {};
      throw new ElphBotApiError(
        typeof error.message === 'string' ? error.message : `Elph Bot API returned HTTP ${response.status}`,
        response.status,
        typeof error.errorCode === 'string' ? error.errorCode : undefined,
        error.details,
      );
    }
    return payload as ApiSuccess;
  }
}

async function readJson(response: Response): Promise<unknown> {
  const text = await response.text();
  if (!text) return { success: true };
  try {
    return JSON.parse(text);
  } catch {
    return undefined;
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}
