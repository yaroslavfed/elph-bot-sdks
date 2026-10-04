# Elph Bot SDK for TypeScript and NestJS

SDK помогает backend-приложению принимать события Elph через webhook и отправлять сообщения от имени бота.

В репозитории есть два пакета:

- `@elph-chat/bot-sdk` содержит TypeScript client для Bot API, модели и проверку подписи webhook
- `@elph-chat/bot-sdk-nestjs` добавляет NestJS-модуль, HTTP endpoint и регистрацию webhook при старте приложения

## Установка

После публикации пакетов добавьте зависимости в проект:

```bash
yarn add @elph-chat/bot-sdk @elph-chat/bot-sdk-nestjs
```

Для NestJS-проекта обычно достаточно второго пакета. Первый пригодится, если приложение не использует NestJS или если нужен прямой доступ к Bot API.

## Настройка

Создайте переменные окружения:

```env
ELPH_API_URL=https://elph-dev.app.eltex.loc
ELPH_BOT_TOKEN=<bot-token>
ELPH_WEBHOOK_URL=https://bot.example.com/webhooks/elph
ELPH_WEBHOOK_SECRET=<random-secret>
```

`ELPH_WEBHOOK_URL` должен вести на публичный HTTPS-адрес. `ELPH_WEBHOOK_SECRET` используется для HMAC-SHA256 подписи каждого входящего запроса. Не помещайте token и secret в код или журнал приложения.

## Быстрый старт с NestJS

В `main.ts` включите сохранение исходного тела запроса. Это необходимо для проверки подписи, потому что подпись вычисляется по байтам тела до разбора JSON.

```ts
import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module.js';

const app = await NestFactory.create(AppModule, { rawBody: true });
await app.listen(process.env.PORT ?? 3000);
```

Подключите модуль в `AppModule`. По умолчанию после старта он вызывает `PUT /bot-api/v1/webhook`, передавая `webhookUrl` и `webhookSecret`.

```ts
import { Module } from '@nestjs/common';
import { ElphBotModule } from '@elph-chat/bot-sdk-nestjs';

@Module({
  imports: [
    ElphBotModule.register({
      apiUrl: process.env.ELPH_API_URL!,
      token: process.env.ELPH_BOT_TOKEN!,
      webhookUrl: process.env.ELPH_WEBHOOK_URL!,
      webhookSecret: process.env.ELPH_WEBHOOK_SECRET!,
      onUpdate: async (update) => {
        if (update.message?.text === '/ping') {
          return;
        }
      },
    }),
  ],
})
export class AppModule {}
```

Модуль добавляет `POST /webhooks/elph`. Если публичный URL использует другой путь, укажите этот путь в reverse proxy или создайте отдельный endpoint, который передаст `rawBody` в `ElphBotService.handleWebhook`.

## Автоматическая и ручная регистрация webhook

Автоматический режим включён по умолчанию. При каждом запуске приложения SDK устанавливает актуальные URL и secret:

```ts
ElphBotModule.register({ apiUrl, token, webhookUrl, webhookSecret });
```

Для ручного режима отключите регистрацию. Тогда адрес можно установить отдельной административной командой через client:

```ts
ElphBotModule.register({
  apiUrl,
  token,
  webhookSecret,
  autoRegisterWebhook: false,
});
```

```ts
import { ElphBotClient } from '@elph-chat/bot-sdk';

const client = new ElphBotClient({ apiUrl, token });
await client.setWebhook({ url: webhookUrl, secret: webhookSecret });
```

## Входящие обновления

`BotUpdate` версии 1 содержит `updateId` и один из сценариев:

- `message` для текстового сообщения или сообщения с вложениями
- `inlineCommand` для нажатия inline-кнопки

Обработчик должен использовать `updateId` как `requestId` для исходящего ответа. Это делает повтор одного и того же обновления идемпотентным для Bot API.

```ts
if (update.inlineCommand?.commandId === 'help') {
  await handleHelp();
}

for (const attachment of update.message?.attachments ?? []) {
  await processAttachment(attachment.id);
}
```

Некорректная, отсутствующая или неподписанная webhook-доставка отклоняется до обработки JSON с ошибкой `WebhookVerificationError`.

## Отправка сообщения и inline-кнопок

`sendMessage` отправляет текст, а `inlineKeyboard` прикрепляет кнопки к конкретному сообщению. Значение `command` возвращается в `inlineCommand.commandId` после нажатия.

```ts
await bot.sendMessage({
  chatId,
  requestId: update.updateId,
  text: 'Выберите действие',
  inlineKeyboard: {
    rows: [
      { buttons: [{ text: 'Помощь', command: 'help' }] },
      { buttons: [{ text: 'Статус', command: 'status' }] },
    ],
  },
});
```

## Reply keyboard и команды меню

`setReplyKeyboard` обновляет общую клавиатуру бота. `setMyCommands` обновляет общий список команд меню. Эти настройки применяются ко всем чатам бота, поэтому вызывайте их при конфигурации бота или по явной административной команде.

```ts
await bot.setReplyKeyboard({
  rows: [{
    buttons: [
      { text: '/ping', command: '/ping' },
      { text: '/help', command: '/help' },
    ],
  }],
});

await bot.setMyCommands({
  buttons: [
    { command: '/ping', text: 'Проверить доступность' },
    { command: '/help', text: 'Показать справку' },
  ],
});
```

## Отправка файла

`sendFile` принимает объект `Blob`, имя файла и стабильный `requestId`.

```ts
const file = new Blob(['Содержимое файла'], { type: 'text/plain' });
await bot.sendFile({
  chatId,
  requestId: update.updateId,
  file,
  filename: 'report.txt',
  caption: 'Отчёт',
});
```

## TypeScript client без NestJS

```ts
import { ElphBotClient, parseWebhookUpdate } from '@elph-chat/bot-sdk';

const client = new ElphBotClient({ apiUrl, token });
const update = parseWebhookUpdate(rawBody, signature, webhookSecret);
await client.sendMessage({ chatId: update.chat.id, requestId: update.updateId, text: 'Принято' });
```

`ElphBotClient` предоставляет методы `setWebhook`, `sendMessage`, `sendFile`, `setReplyKeyboard` и `setMyCommands`. HTTP-ошибки Bot API выбрасываются как `ElphBotApiError` с `statusCode`, `errorCode` и `details`.

## Разработка SDK

Из каталога `typescript`:

```bash
yarn install
yarn build
yarn test
```

Пакетный scope `@elph-chat` нужно подтвердить до первой публикации в npm registry. До публикации примеры используют локальные workspace-зависимости.
