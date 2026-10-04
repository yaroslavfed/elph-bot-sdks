import { Injectable, OnApplicationBootstrap } from '@nestjs/common';
import { ElphBotClient, parseWebhookUpdate } from '@yaroslavfed/bot-sdk';
import type { ApiSuccess, BotUpdate, MenuCommands, ReplyKeyboard, SendFileRequest, SendMessageRequest } from '@yaroslavfed/bot-sdk';

import { ELPH_BOT_OPTIONS, type ElphBotModuleOptions } from './options.js';
import { Inject } from '@nestjs/common';

@Injectable()
export class ElphBotService implements OnApplicationBootstrap {
  public readonly client: ElphBotClient;

  public constructor(@Inject(ELPH_BOT_OPTIONS) private readonly options: ElphBotModuleOptions) {
    this.client = new ElphBotClient(options);
  }

  public async onApplicationBootstrap(): Promise<void> {
    if (this.options.autoRegisterWebhook === false) return;
    if (!this.options.webhookUrl) {
      throw new Error('webhookUrl is required when autoRegisterWebhook is enabled');
    }
    await this.client.setWebhook({ url: this.options.webhookUrl, secret: this.options.webhookSecret });
  }

  public async handleWebhook(rawBody: Uint8Array, signature: string | undefined): Promise<BotUpdate> {
    const update = parseWebhookUpdate(rawBody, signature, this.options.webhookSecret);
    await this.options.onUpdate?.(update);
    return update;
  }

  public sendMessage(request: SendMessageRequest): Promise<ApiSuccess> {
    return this.client.sendMessage(request);
  }

  public sendFile(request: SendFileRequest): Promise<ApiSuccess> {
    return this.client.sendFile(request);
  }

  public setReplyKeyboard(replyKeyboard: ReplyKeyboard): Promise<ApiSuccess> {
    return this.client.setReplyKeyboard(replyKeyboard);
  }

  public setMyCommands(menuCommands: MenuCommands): Promise<ApiSuccess> {
    return this.client.setMyCommands(menuCommands);
  }
}
