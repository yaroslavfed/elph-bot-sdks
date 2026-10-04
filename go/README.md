# Elph Bot SDK for Go

SDK позволяет backend-приложению принимать события Elph через webhook и отправлять сообщения от имени бота.

## Пакеты

- `bot` содержит основной API для бота
- `models` содержит модели обновлений, сообщений и клавиатур
- `elphsdk` содержит низкоуровневый HTTP client для нестандартных интеграций

```go
import (
    "github.com/yaroslavfed/elph-bot-sdks/go/bot"
    "github.com/yaroslavfed/elph-bot-sdks/go/models"
)
```

## Установка

После публикации версии добавьте модуль в проект:

```bash
go get github.com/yaroslavfed/elph-bot-sdks/go@v1.0.0
```

## Настройка

```env
ELPH_API_URL=https://elph-dev.app.eltex.loc
ELPH_BOT_TOKEN=<bot-token>
ELPH_WEBHOOK_URL=https://bot.example.com/webhooks/elph
ELPH_WEBHOOK_SECRET=<random-secret>
```

`ELPH_WEBHOOK_URL` должен быть публичным HTTPS-адресом. `ELPH_WEBHOOK_SECRET` используется для подписи каждого входящего webhook. Для dev-стенда с внутренним сертификатом передайте собственный `http.Client` через `bot.WithHTTPClient`.

## Быстрый старт

Пример запускает HTTP-server, регистрирует webhook и отвечает `pong` на `/ping`.

```go
client, err := bot.New(
    os.Getenv("ELPH_BOT_TOKEN"),
    bot.WithServerURL(os.Getenv("ELPH_API_URL")),
    bot.WithWebhookSecretToken(os.Getenv("ELPH_WEBHOOK_SECRET")),
    bot.WithDefaultHandler(func(ctx context.Context, client *bot.Bot, update *models.Update) {
        if update.Message == nil || update.Message.Text != "/ping" {
            return
        }
        _, err := client.SendMessage(ctx, &bot.SendMessageParams{
            ChatID:    update.Message.Chat.ID,
            RequestID: update.ID,
            Text:      "pong",
        })
        if err != nil {
            log.Printf("send message: %v", err)
        }
    }),
)
if err != nil {
    log.Fatal(err)
}

go client.StartWebhook(context.Background())
server := &http.Server{Addr: ":8080", Handler: client.WebhookHandler()}
go server.ListenAndServe()

_, err = client.SetWebhook(context.Background(), &bot.SetWebhookParams{
    URL:         os.Getenv("ELPH_WEBHOOK_URL"),
    SecretToken: os.Getenv("ELPH_WEBHOOK_SECRET"),
})
if err != nil {
    log.Fatal(err)
}
```

`WebhookHandler` проверяет HMAC-SHA256 заголовок `x-elph-bot-signature`, разбирает `BotUpdate` v1 и кладёт обновление в очередь. `StartWebhook` читает очередь и вызывает handler. Запускайте `StartWebhook` до получения первого webhook.

## Обработка обновлений

| Поле `models.Update` | Когда заполняется | Основные данные |
|---|---|---|
| `Message` | Пользователь отправил сообщение | `ID`, `Chat`, `From`, `Text` |
| `CallbackQuery` | Нажата inline-кнопка | `Data`, `From`, `Message` |

`update.ID` является строковым идентификатором доставки. Передавайте его как `RequestID` при ответе на обновление. Это делает повтор одного исходящего запроса идемпотентным.

## Отправка текста

```go
_, err := client.SendMessage(ctx, &bot.SendMessageParams{
    ChatID:    "chat-id",
    RequestID: "operation-42",
    Text:      "Сообщение отправлено",
})
```

`ChatID` принимает строку и числовые значения. Elph хранит идентификатор чата как строку. Если `RequestID` не задан, SDK создаст новый идентификатор. Для повторяемой бизнес-операции передавайте устойчивый `RequestID` самостоятельно.

## Inline-клавиатура

```go
_, err := client.SendMessage(ctx, &bot.SendMessageParams{
    ChatID:    "chat-id",
    RequestID: "choose-plan",
    Text:      "Выберите вариант",
    ReplyMarkup: models.InlineKeyboardMarkup{
        InlineKeyboard: [][]models.InlineKeyboardButton{{
            {Text: "A", CallbackData: "plan:a"},
            {Text: "B", CallbackData: "plan:b"},
        }},
    },
})
```

После нажатия SDK передаст `CallbackQuery.Data` со значением `plan:a` или `plan:b`.

## Reply keyboard и menu commands

Reply keyboard и menu commands являются настройками всего бота. Они не принадлежат одному сообщению или одному чату.

```go
err := client.SetReplyKeyboard(ctx, models.ReplyKeyboardMarkup{
    Keyboard: [][]models.KeyboardButton{{
        {Text: "/ping"},
        {Text: "/help"},
    }},
})

_, err = client.SetMyCommands(ctx, &bot.SetMyCommandsParams{
    Commands: []models.BotCommand{
        {Command: "/ping", Description: "Проверить доступность"},
        {Command: "/help", Description: "Получить справку"},
    },
})
```

`ReplyKeyboardMarkup` можно также передать в `SendMessageParams.ReplyMarkup`. SDK сначала заменит reply keyboard бота, затем отправит сообщение.

## Отправка файла

```go
file, err := os.Open("report.pdf")
if err != nil {
    return err
}
defer file.Close()

_, err = client.SendDocument(ctx, &bot.SendDocumentParams{
    ChatID:    "chat-id",
    RequestID: "report-42",
    Caption:   "Отчёт",
    Document: models.InputFile{
        Filename: "report.pdf",
        Content:  file,
    },
})
```

Файл передаётся потоково и не загружается в память целиком.

## Методы `bot.Bot`

| Метод | Назначение |
|---|---|
| `SetWebhook` | Сохраняет публичный адрес webhook и secret у бота |
| `WebhookHandler` | Возвращает `http.Handler` для входящих webhook |
| `StartWebhook` | Запускает обработку принятых обновлений |
| `SendMessage` | Отправляет текст и inline-клавиатуру |
| `SendDocument` | Отправляет файл потоком |
| `SetReplyKeyboard` | Заменяет reply keyboard для бота |
| `SetMyCommands` | Заменяет menu commands для бота |

## Низкоуровневый API

`elphsdk.Client` полезен, если приложение использует свой HTTP router, свой lifecycle или хочет обращаться к Bot API напрямую. Он предоставляет `SetWebhook`, `SetReplyKeyboard`, `SetMenuCommands`, `SendMessage` и `SendFile`. `elphsdk.WebhookHandler` принимает `BotUpdate` напрямую.

## Ограничения текущей версии

SDK поддерживает текстовые сообщения, inline-команды, клавиатуры и файлы. Вложения во входящем сообщении, групповые сценарии и дополнительные типы событий доступны в низкоуровневом `elphsdk.BotUpdate` и будут добавляться в публичные `models` по мере расширения Bot API.

Runnable пример находится в `../examples/go-echo-bot`.
