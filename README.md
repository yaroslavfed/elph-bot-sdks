# Elph Bot SDKs

SDK и общие контракты для backend-ботов Elph.

## Состав

- `contracts/` содержит версионированные схемы и fixtures Bot API
- `go/` содержит Go SDK
- `go/examples/` содержит runnable пример Go-бота
- `typescript/` содержит TypeScript client, NestJS integration и runnable пример NestJS-бота

## Выпуск версий

GitHub Actions проверяет Go и TypeScript на каждом pull request и push в `master`. Тег `vX.Y.Z` запускает публикацию обоих npm-пакетов и создаёт GitHub Release. Версия в `typescript/packages/core/package.json` и `typescript/packages/nestjs/package.json` должна совпадать с тегом без префикса `v`.

Перед первым выпуском добавьте repository secret `NPM_TOKEN` с правом публикации в npm scope `@elph-chat`. Для предварительных версий используйте тег и package version с суффиксом, например `v0.2.0-rc.1`: workflow опубликует их под npm tag `next`.

Go SDK не требует отдельной публикации. Для него создавайте тег `go/vX.Y.Z`, например `go/v0.1.0`: Go распознаёт такой тег как версию модуля `github.com/yaroslavfed/elph-bot-sdks/go` и workflow создаёт GitHub Release.
