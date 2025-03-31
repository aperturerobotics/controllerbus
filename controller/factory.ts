import { Context } from '../core/core'
import { Config } from '../config/config'
import { Controller } from './controller'
import { Logger } from '../core/core'

/**
 * ConstructOpts contains options for constructing a controller.
 */
export interface ConstructOpts {
  /**
   * GetLogger returns the logger to use.
   */
  getLogger(): Logger
}

/**
 * Factory constructs Controller objects and their configurations.
 */
export interface Factory {
  /**
   * GetControllerID returns the unique ID for the controller.
   */
  getControllerID(): string

  /**
   * GetConfigID returns the unique ID for the controller config.
   */
  getConfigID(): string

  /**
   * ConstructConfig constructs an instance of the controller configuration.
   */
  constructConfig(): Config

  /**
   * Construct constructs the associated controller given configuration.
   */
  construct(
    ctx: Context,
    config: Config,
    opts: ConstructOpts,
  ): Promise<[Controller, Error | null]>

  /**
   * GetVersion returns the version of this controller.
   */
  getVersion(): string
}
