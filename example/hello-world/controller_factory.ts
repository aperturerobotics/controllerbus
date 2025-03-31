import { Context } from '../../core/core'
import { Config } from '../../config/config'
import { Controller } from '../../controller/controller'
import { Factory, ConstructOpts } from '../../controller/factory'
import { ToyControllerConfig } from './controller.pb'
import { ToyController } from './controller'

// ToyFactoryVersion is the compiled-in version
export const toyFactoryVersion = '0.1.0'

// ToyControllerID is the controller id
export const toyControllerID = 'hello-world'

/**
 * ToyFactory implements the toy controller factory.
 */
export class ToyFactory implements Factory {
  /**
   * GetControllerID returns the unique ID for the controller.
   */
  getControllerID(): string {
    return toyControllerID
  }

  /**
   * GetConfigID returns the unique ID for the controller config.
   */
  getConfigID(): string {
    return this.getControllerID()
  }

  /**
   * ConstructConfig constructs an instance of the controller configuration.
   */
  constructConfig(): Config {
    return new ToyControllerConfig()
  }

  /**
   * Construct constructs the associated controller given configuration.
   */
  async construct(
    ctx: Context,
    config: Config,
    opts: ConstructOpts,
  ): Promise<[Controller, Error | null]> {
    if (!(config instanceof ToyControllerConfig)) {
      return [null as any, new Error('wrong type of config')]
    }

    const tc = new ToyController(opts.getLogger(), config)
    return [tc, null]
  }

  /**
   * GetVersion returns the version of this controller.
   */
  getVersion(): string {
    return toyFactoryVersion
  }
}

/**
 * Creates a new toy factory.
 */
export function newToyFactory(): ToyFactory {
  return new ToyFactory()
}
