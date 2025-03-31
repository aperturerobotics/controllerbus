import { Context } from '../../core/core'
import { Bus } from '../../bus/bus'
import { Controller, Info, newInfo } from '../controller'
import { Factory } from '../factory'
import { Config } from '../../config/config'
import {
  Directive,
  Instance,
  Reference,
  Resolver,
  ResolverHandler,
} from '../../directive/directive'
import { Logger } from '../../core/core'

/**
 * ControllerID is the controller ID.
 */
export const ControllerID = 'controllerbus/controller/loader'

/**
 * ControllerVersion is the controller version.
 */
export const ControllerVersion = '0.0.1'

/**
 * ExecControllerDirective is a directive to execute a controller.
 */
export class ExecControllerDirective implements Directive {
  private readonly factory: Factory
  private readonly config: Config

  constructor(factory: Factory, config: Config) {
    this.factory = factory
    this.config = config
  }

  /**
   * Validate validates the directive.
   */
  validate(): Error | null {
    if (!this.factory) {
      return new Error('factory is required')
    }
    if (!this.config) {
      return new Error('config is required')
    }
    return this.config.validate()
  }

  /**
   * GetValueOptions returns options relating to value handling.
   */
  getValueOptions() {
    return {
      maxValueCount: 1, // Only want one controller
      maxValueHardCap: true,
      unrefDisposeDur: 0,
      unrefDisposeEmptyImmediate: true,
    }
  }

  /**
   * GetName returns the directive type name.
   */
  getName(): string {
    return 'ExecController'
  }

  /**
   * IsEquivalent checks if the other directive is equivalent.
   */
  isEquivalent(other: Directive): boolean {
    if (!(other instanceof ExecControllerDirective)) {
      return false
    }

    const otherDir = other as ExecControllerDirective
    return (
      otherDir.factory.getControllerID() === this.factory.getControllerID() &&
      otherDir.config.getConfigID() === this.config.getConfigID() &&
      otherDir.config.equalsConfig(this.config)
    )
  }

  /**
   * GetFactory returns the controller factory.
   */
  getFactory(): Factory {
    return this.factory
  }

  /**
   * GetConfig returns the controller configuration.
   */
  getConfig(): Config {
    return this.config
  }

  /**
   * GetDebugVals returns the directive arguments as key/value pairs.
   */
  getDebugVals() {
    return {
      'config-id': [this.config.getConfigID()],
    }
  }
}

/**
 * ExecControllerResolver resolves the execution of a controller.
 */
export class ExecControllerResolver implements Resolver {
  private readonly bus: Bus
  private readonly factory: Factory
  private readonly config: Config
  private readonly logger: Logger
  private release: (() => void) | null = null

  constructor(bus: Bus, factory: Factory, config: Config, logger: Logger) {
    this.bus = bus
    this.factory = factory
    this.config = config
    this.logger = logger
  }

  /**
   * Resolve resolves the values, emitting them to the handler.
   */
  async resolve(ctx: Context, handler: ResolverHandler): Promise<Error | null> {
    try {
      // Construct the controller
      const [ctrl, constructErr] = await this.factory.construct(
        ctx,
        this.config,
        {
          getLogger: () => this.logger,
        },
      )

      if (constructErr) {
        return constructErr
      }

      // Add controller to the bus
      const [releaseFunc, addErr] = await this.bus.addController(
        ctx,
        ctrl,
        (exitErr) => {
          if (exitErr) {
            // If controller exits with error, add the error
            handler.markIdle(true)
          }
        },
      )

      if (addErr) {
        return addErr
      }

      this.release = releaseFunc

      // Register cleanup when the resolver is removed
      handler.addResolverRemovedCallback(() => {
        if (this.release) {
          this.release()
          this.release = null
        }
      })

      // Add the controller as a value
      const [valueId, accepted] = handler.addValue(ctrl)
      if (!accepted) {
        return new Error('controller value was rejected')
      }

      // Register cleanup when the value is removed
      handler.addValueRemovedCallback(valueId, () => {
        if (this.release) {
          this.release()
          this.release = null
        }
      })

      // Mark as not idle since controller is running
      handler.markIdle(false)

      // Wait for context to be done
      await ctx.done()
      return null
    } catch (e) {
      return e instanceof Error ? e : new Error(String(e))
    }
  }
}

/**
 * LoaderController is a controller that loads and executes other controllers.
 */
export class LoaderController implements Controller {
  private readonly logger: Logger
  private readonly bus: Bus

  constructor(logger: Logger, bus: Bus) {
    this.logger = logger
    this.bus = bus
  }

  /**
   * GetControllerInfo returns information about the controller.
   */
  getControllerInfo(): Info {
    return newInfo(
      ControllerID,
      ControllerVersion,
      'controller execution manager',
    )
  }

  /**
   * HandleDirective asks if the handler can resolve the directive.
   */
  async handleDirective(
    _ctx: Context,
    instance: Instance,
  ): Promise<Resolver[] | null> {
    const directive = instance.getDirective()

    // Only handle ExecControllerDirective
    if (!(directive instanceof ExecControllerDirective)) {
      return null
    }

    const execDir = directive as ExecControllerDirective
    return [
      new ExecControllerResolver(
        this.bus,
        execDir.getFactory(),
        execDir.getConfig(),
        this.logger,
      ),
    ]
  }

  /**
   * Execute executes the controller goroutine.
   */
  async execute(ctx: Context): Promise<Error | null> {
    // This controller doesn't need to do anything in its execute method
    // It just needs to be on the bus to handle directives
    await ctx.done()
    return null
  }

  /**
   * Close releases any resources used by the controller.
   */
  async close(): Promise<Error | null> {
    return null
  }
}

/**
 * Creates a new loader controller.
 *
 * @param logger - The logger to use
 * @param bus - The bus to use
 * @returns A new loader controller and any error
 */
export async function newController(
  logger: Logger,
  bus: Bus,
): Promise<[LoaderController, Error | null]> {
  return [new LoaderController(logger, bus), null]
}

/**
 * Creates a new ExecController directive.
 *
 * @param factory - The controller factory to use
 * @param config - The controller configuration
 * @returns A new ExecController directive
 */
export function newExecController(
  factory: Factory,
  config: Config,
): ExecControllerDirective {
  return new ExecControllerDirective(factory, config)
}

/**
 * Waits for a controller to be running from an ExecController directive.
 *
 * @param ctx - Context for the operation
 * @param bus - The bus to use
 * @param directive - The ExecController directive
 * @param cb - Optional callback for directive events
 * @returns The controller, instance, value reference, and any error
 */
export async function waitExecControllerRunning(
  ctx: Context,
  bus: Bus,
  directive: ExecControllerDirective,
  cb: any | null,
): Promise<[Controller, Instance, Reference, Error | null]> {
  return new Promise<[Controller, Instance, Reference, Error | null]>(
    (resolve, reject) => {
      const handler = cb
        ? cb
        : {
            handleValueAdded: (instance: Instance, value: any) => {
              // Resolve with the controller value
              // Create a mock reference object since null is not assignable to Reference type
              const mockRef: Reference = {
                release: () => {},
                getReference: () => 1
              }
              resolve([value.getValue() as Controller, instance, mockRef, null])
            },
            handleValueRemoved: () => {},
            handleInstanceDisposed: () => {
              reject(
                new Error(
                  'directive instance disposed before controller was running',
                ),
              )
            },
          }

      // Add the directive
      bus
        .addDirective(directive, handler)
        .then(([_instance, _reference, err]) => {
          if (err) {
            reject(err)
            return
          }

          // Wait for context to be done
          ctx.done().then(() => {
            reject(new Error('context done before controller was running'))
          })
        })
        .catch(reject)
    },
  )
}
