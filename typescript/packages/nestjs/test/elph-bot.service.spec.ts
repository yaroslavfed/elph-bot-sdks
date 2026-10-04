import assert from 'node:assert/strict';
import { createHmac } from 'node:crypto';
import test from 'node:test';

import { ElphBotService } from '../src/elph-bot.service.js';
import { ElphWebhookController } from '../src/elph-webhook.controller.js';
import { HttpException } from '@nestjs/common';

test('registers the webhook on application bootstrap', async () => {
  const requests: Request[] = [];
  const bot = new ElphBotService({
    apiUrl: 'https://elph.example',
    token: 'bot-token',
    webhookUrl: 'https://bot.example/webhooks/elph',
    webhookSecret: 'secret',
    fetch: async (input, init) => {
      requests.push(new Request(input, init));
      return new Response(JSON.stringify({ success: true }));
    },
  });

  await bot.onApplicationBootstrap();

  assert.equal(requests[0]?.method, 'PUT');
  assert.deepEqual(await requests[0]?.json(), {
    url: 'https://bot.example/webhooks/elph',
    secret: 'secret',
  });
});

test('delivers a verified update to the configured handler', async () => {
  let handledUpdateId: string | undefined;
  const secret = 'secret';
  const body = Buffer.from(JSON.stringify({
    schemaVersion: 1,
    updateId: 'update-1',
    botId: 'bot-1',
    chat: { id: 'chat-1', type: 'direct' },
    from: { id: 'user-1' },
    createdAt: '2026-10-04T00:00:00Z',
    inlineCommand: { sourceMessageId: 'message-1', commandId: 'help' },
  }));
  const signature = `sha256=${createHmac('sha256', secret).update(body).digest('hex')}`;
  const bot = new ElphBotService({
    apiUrl: 'https://elph.example',
    token: 'bot-token',
    webhookSecret: secret,
    autoRegisterWebhook: false,
    onUpdate: async (update) => {
      handledUpdateId = update.updateId;
    },
  });

  await bot.handleWebhook(body, signature);
  assert.equal(handledUpdateId, 'update-1');
});

test('returns 401 for an invalid webhook signature', async () => {
  const bot = new ElphBotService({
    apiUrl: 'https://elph.example',
    token: 'bot-token',
    webhookSecret: 'secret',
    autoRegisterWebhook: false,
  });
  const controller = new ElphWebhookController(bot);
  const body = Buffer.from(JSON.stringify({ schemaVersion: 1, updateId: 'update-1' }));

  await assert.rejects(
    controller.receive({ rawBody: body }, 'sha256=invalid'),
    (error: unknown) => error instanceof HttpException && error.getStatus() === 401,
  );
});
