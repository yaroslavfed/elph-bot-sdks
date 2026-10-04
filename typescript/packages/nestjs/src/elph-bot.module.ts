import { DynamicModule, Module } from '@nestjs/common';

import { ElphBotService } from './elph-bot.service.js';
import { ElphWebhookController } from './elph-webhook.controller.js';
import { ELPH_BOT_OPTIONS, type ElphBotModuleOptions } from './options.js';

@Module({})
export class ElphBotModule {
  public static register(options: ElphBotModuleOptions): DynamicModule {
    return {
      module: ElphBotModule,
      controllers: [ElphWebhookController],
      providers: [
        { provide: ELPH_BOT_OPTIONS, useValue: options },
        ElphBotService,
      ],
      exports: [ElphBotService],
    };
  }
}
