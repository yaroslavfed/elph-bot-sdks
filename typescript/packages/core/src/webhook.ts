import { createHmac, timingSafeEqual } from 'node:crypto';

import { WebhookVerificationError } from './errors.js';
import type { BotUpdate } from './types.js';

export const WEBHOOK_SIGNATURE_HEADER = 'x-elph-bot-signature';

export function parseWebhookUpdate(rawBody: Uint8Array, signature: string | undefined, secret: string): BotUpdate {
  if (!signature) {
    throw new WebhookVerificationError('Webhook signature is missing', 401);
  }

  const expected = `sha256=${createHmac('sha256', secret).update(rawBody).digest('hex')}`;
  const actualBytes = Buffer.from(signature);
  const expectedBytes = Buffer.from(expected);
  if (actualBytes.length !== expectedBytes.length || !timingSafeEqual(actualBytes, expectedBytes)) {
    throw new WebhookVerificationError('Webhook signature is invalid', 401);
  }

  let value: unknown;
  try {
    value = JSON.parse(Buffer.from(rawBody).toString('utf8'));
  } catch {
    throw new WebhookVerificationError('Webhook body is not valid JSON');
  }

  if (!isBotUpdate(value)) {
    throw new WebhookVerificationError('Webhook body does not contain a supported BotUpdate');
  }
  return value;
}

function isBotUpdate(value: unknown): value is BotUpdate {
  if (
    !isRecord(value) || value.schemaVersion !== 1 || typeof value.updateId !== 'string' ||
    typeof value.botId !== 'string' || !isChat(value.chat) || !isUser(value.from) ||
    typeof value.createdAt !== 'string' || (value.message === undefined) === (value.inlineCommand === undefined)
  ) {
    return false;
  }
  if (value.message !== undefined && (!isRecord(value.message) || typeof value.message.id !== 'string')) {
    return false;
  }
  return value.inlineCommand === undefined || (
    isRecord(value.inlineCommand) && typeof value.inlineCommand.commandId === 'string' &&
    typeof value.inlineCommand.sourceMessageId === 'string'
  );
}

function isChat(value: unknown): boolean {
  return isRecord(value) && typeof value.id === 'string' && typeof value.type === 'string';
}

function isUser(value: unknown): boolean {
  return isRecord(value) && typeof value.id === 'string';
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}
