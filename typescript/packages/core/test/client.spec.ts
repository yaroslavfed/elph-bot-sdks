import assert from 'node:assert/strict';
import test from 'node:test';

import { ElphBotApiError } from '../src/errors.js';
import { ElphBotClient } from '../src/client.js';

test('sends a message with bearer authentication and requestId', async () => {
  let request: Request | undefined;
  const client = new ElphBotClient({
    apiUrl: 'https://elph.example/',
    token: 'bot-token',
    fetch: async (input, init) => {
      request = new Request(input, init);
      return new Response(JSON.stringify({ success: true }), { status: 200 });
    },
  });

  await client.sendMessage({ chatId: 'chat-1', requestId: 'update-1', text: 'Hello' });

  assert.equal(request?.url, 'https://elph.example/bot-api/v1/messages');
  assert.equal(request?.headers.get('authorization'), 'Bearer bot-token');
  assert.deepEqual(await request?.json(), { chatId: 'chat-1', requestId: 'update-1', text: 'Hello' });
});

test('turns an API error into ElphBotApiError', async () => {
  const client = new ElphBotClient({
    apiUrl: 'https://elph.example',
    token: 'bot-token',
    fetch: async () => new Response(JSON.stringify({ errorCode: 'CHAT_NOT_FOUND', message: 'Chat was not found' }), { status: 404 }),
  });

  await assert.rejects(
    client.sendMessage({ chatId: 'chat-1', requestId: 'update-1', text: 'Hello' }),
    (error: unknown) => error instanceof ElphBotApiError && error.statusCode === 404 && error.errorCode === 'CHAT_NOT_FOUND',
  );
});

test('sends a file as multipart form data', async () => {
  let request: Request | undefined;
  const client = new ElphBotClient({
    apiUrl: 'https://elph.example',
    token: 'bot-token',
    fetch: async (input, init) => {
      request = new Request(input, init);
      return new Response(JSON.stringify({ success: true }));
    },
  });

  await client.sendFile({
    chatId: 'chat-1',
    requestId: 'update-1',
    file: new Blob(['file content'], { type: 'text/plain' }),
    filename: 'report.txt',
    caption: 'Report',
  });

  assert.match(request?.headers.get('content-type') ?? '', /^multipart\/form-data; boundary=/);
  const body = await request?.formData();
  assert.equal(body?.get('chatId'), 'chat-1');
  assert.equal(body?.get('requestId'), 'update-1');
  assert.equal(body?.get('caption'), 'Report');
  assert.equal((body?.get('file') as File).name, 'report.txt');
});

test('wraps keyboard and menu command updates in their API envelopes', async () => {
  const requests: Request[] = [];
  const client = new ElphBotClient({
    apiUrl: 'https://elph.example',
    token: 'bot-token',
    fetch: async (input, init) => {
      requests.push(new Request(input, init));
      return new Response(JSON.stringify({ success: true }));
    },
  });

  await client.setReplyKeyboard({ rows: [{ buttons: [{ text: '/ping', command: '/ping' }] }] });
  await client.setMyCommands({ buttons: [{ text: 'Проверка', command: '/ping' }] });

  assert.deepEqual(await requests[0]?.json(), {
    replyKeyboard: { rows: [{ buttons: [{ text: '/ping', command: '/ping' }] }] },
  });
  assert.deepEqual(await requests[1]?.json(), {
    menuCommands: { buttons: [{ text: 'Проверка', command: '/ping' }] },
  });
});
