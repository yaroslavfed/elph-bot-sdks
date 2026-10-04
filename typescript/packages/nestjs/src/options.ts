import type { BotUpdate, ElphBotClientOptions } from '@yaroslavfed/bot-sdk';

export interface ElphBotModuleOptions extends ElphBotClientOptions {
  webhookUrl?: string;
  webhookSecret: string;
  autoRegisterWebhook?: boolean;
  onUpdate?: (update: BotUpdate) => void | Promise<void>;
}

export const ELPH_BOT_OPTIONS = Symbol('ELPH_BOT_OPTIONS');
