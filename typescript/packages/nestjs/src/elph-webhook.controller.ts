import { Controller, Headers, HttpCode, HttpException, InternalServerErrorException, Post, Req } from '@nestjs/common';

import { WEBHOOK_SIGNATURE_HEADER, WebhookVerificationError } from '@yaroslavfed/bot-sdk';
import { ElphBotService } from './elph-bot.service.js';

interface WebhookRequest {
  rawBody?: Buffer;
}

@Controller('webhooks/elph')
export class ElphWebhookController {
  public constructor(private readonly bot: ElphBotService) {}

  @Post()
  @HttpCode(204)
  public async receive(
    @Req() request: WebhookRequest,
    @Headers(WEBHOOK_SIGNATURE_HEADER) signature?: string,
  ): Promise<void> {
    if (!request.rawBody) {
      throw new InternalServerErrorException('NestJS rawBody is required for Elph webhook verification');
    }
    try {
      await this.bot.handleWebhook(request.rawBody, signature);
    } catch (error) {
      if (error instanceof WebhookVerificationError) {
        throw new HttpException(error.message, error.statusCode);
      }
      throw error;
    }
  }
}
