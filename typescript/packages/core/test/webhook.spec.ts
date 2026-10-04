import assert from 'node:assert/strict';
import { createHmac } from 'node:crypto';
import test from 'node:test';

import { WebhookVerificationError } from '../src/errors.js';
import { parseWebhookUpdate } from '../src/webhook.js';

const secret = 'webhook-secret';

function signature(body: Uint8Array): string {
  return `sha256=${createHmac('sha256', secret).update(body).digest('hex')}`;
}

test('parses a signed message update', () => {
  const body = Buffer.from(JSON.stringify({
    schemaVersion: 1,
    updateId: 'update-1',
    botId: 'bot-1',
    chat: { id: 'chat-1', type: 'direct' },
    from: { id: 'user-1' },
    createdAt: '2026-10-04T00:00:00Z',
    message: { id: 'message-1', text: 'Hello' },
  }));

  const update = parseWebhookUpdate(body, signature(body), secret);
  assert.equal(update.message?.text, 'Hello');
});

test('parses a signed inline command update', () => {
  const body = Buffer.from(JSON.stringify({
    schemaVersion: 1,
    updateId: 'update-1',
    botId: 'bot-1',
    chat: { id: 'chat-1', type: 'direct' },
    from: { id: 'user-1' },
    createdAt: '2026-10-04T00:00:00Z',
    inlineCommand: { sourceMessageId: 'message-1', commandId: 'button-1' },
  }));
  assert.equal(parseWebhookUpdate(body, signature(body), secret).inlineCommand?.commandId, 'button-1');
});

test('rejects an invalid signature and unsupported version', () => {
  const signedBody = Buffer.from(JSON.stringify({ schemaVersion: 1, updateId: 'update-1' }));
  assert.throws(
    () => parseWebhookUpdate(signedBody, 'sha256=wrong', secret),
    (error: unknown) => error instanceof WebhookVerificationError && error.statusCode === 401,
  );

  const unsupported = Buffer.from(JSON.stringify({ schemaVersion: 2, updateId: 'update-1' }));
  assert.throws(
    () => parseWebhookUpdate(unsupported, signature(unsupported), secret),
    (error: unknown) => error instanceof WebhookVerificationError && error.statusCode === 400,
  );
});
