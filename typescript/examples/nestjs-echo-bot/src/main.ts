import 'reflect-metadata';

import { Injectable, Module } from '@nestjs/common';
import { NestFactory } from '@nestjs/core';
import { ElphBotModule, ElphBotService } from '@yaroslavfed/bot-sdk-nestjs';
import type { BotUpdate } from '@yaroslavfed/bot-sdk';

let updates: UpdatesService;

@Injectable()
class UpdatesService {
  public constructor(private readonly bot: ElphBotService) {}

  public async handle(update: BotUpdate): Promise<void> {
    const chatId = update.chat.id;
    const text = update.message?.text;
    if (!chatId || !text) return;

    await this.bot.sendMessage({
      chatId,
      requestId: update.updateId,
      text: text === '/ping' ? 'pong' : `Вы написали: ${text}`,
    });
  }
}

@Module({
  providers: [UpdatesService],
  imports: [
    ElphBotModule.register({
      apiUrl: process.env.ELPH_API_URL!,
      token: process.env.ELPH_BOT_TOKEN!,
      webhookUrl: process.env.ELPH_WEBHOOK_URL!,
      webhookSecret: process.env.ELPH_WEBHOOK_SECRET!,
      onUpdate: async (update) => updates.handle(update),
    }),
  ],
})
class AppModule {}

const app = await NestFactory.create(AppModule, { rawBody: true });
updates = app.get(UpdatesService);
await app.listen(process.env.PORT ?? 3000);
