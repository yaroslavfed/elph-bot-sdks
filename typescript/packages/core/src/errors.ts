export class ElphBotApiError extends Error {
  public constructor(
    message: string,
    public readonly statusCode: number,
    public readonly errorCode?: string,
    public readonly details?: unknown,
  ) {
    super(message);
    this.name = 'ElphBotApiError';
  }
}

export class WebhookVerificationError extends Error {
  public constructor(message: string, public readonly statusCode: 400 | 401 = 400) {
    super(message);
    this.name = 'WebhookVerificationError';
  }
}
