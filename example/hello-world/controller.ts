import { Context } from '../../core/core'
import { Controller, Info, newInfo } from '../../controller/controller'
import { Directive, Instance, Resolver } from '../../directive/directive'
import { Logger } from '../../core/core'
import { ToyControllerConfig } from './controller.pb'

// ControllerID is the controller ID.
export const ControllerID = 'controllerbus/toys/load-controller/controller'

// Version is the controller version.
export const Version = '0.0.1'

/**
 * ToyController is an example controller.
 */
export class ToyController implements Controller {
  private readonly logger: Logger
  private readonly conf: ToyControllerConfig

  /**
   * Constructs a new toy controller.
   */
  constructor(logger: Logger, conf: ToyControllerConfig) {
    this.logger = logger
    this.conf = conf
  }

  /**
   * GetControllerInfo returns information about the controller.
   */
  getControllerInfo(): Info {
    return newInfo(ControllerID, Version, 'toy controller')
  }

  /**
   * HandleDirective asks if the handler can resolve the directive.
   */
  async handleDirective(
    ctx: Context,
    instance: Instance,
  ): Promise<Resolver[] | null> {
    // Pass for the example controller
    return null
  }

  /**
   * Execute executes the controller goroutine.
   * Returning null ends execution.
   * Returning an error triggers a retry with backoff.
   */
  async execute(ctx: Context): Promise<Error | null> {
    this.logger.debug('toy controller executed')
    await ctx.done()
    return null
  }

  /**
   * SayHello says hello.
   */
  sayHello(): void {
    this.logger.debugf('Hello %s!', this.conf.name)
  }

  /**
   * Close releases any resources used by the controller.
   */
  async close(): Promise<Error | null> {
    return null
  }
}
