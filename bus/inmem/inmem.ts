import { Bus } from '../bus'
import { Context } from '../../core/core'
import { Controller } from '../../controller/controller'
import {
  Directive,
  Instance,
  Reference,
  Handler,
  ReferenceHandler,
  Controller as DirectiveController,
} from '../../directive/directive'

interface ControllerHandle {
  controller: Controller
  execCtx: Context
  cancelCtx: () => void
  callback: ((exitErr: Error | null) => void) | null
}

/**
 * InMemBus is an in-memory implementation of Bus.
 */
export class InMemBus implements Bus {
  private readonly directiveCtrl: DirectiveController
  private readonly controllers: Map<Controller, ControllerHandle> = new Map()
  private readonly handlerReleases: Map<Handler, () => void> = new Map()

  constructor(directiveCtrl: DirectiveController) {
    this.directiveCtrl = directiveCtrl
  }

  /**
   * GetDirectives returns a list of all currently executing directives.
   */
  getDirectives(): Instance[] {
    return this.directiveCtrl.getDirectives()
  }

  /**
   * AddDirective adds a directive to the controller.
   */
  async addDirective(
    dir: Directive,
    handler: ReferenceHandler | null,
  ): Promise<[Instance, Reference, Error | null]> {
    return this.directiveCtrl.addDirective(dir, handler)
  }

  /**
   * AddHandler adds a directive handler.
   */
  async addHandler(handler: Handler): Promise<[() => void, Error | null]> {
    const [release, err] = await this.directiveCtrl.addHandler(handler)
    if (!err && release) {
      this.handlerReleases.set(handler, release)
    }
    return [release, err]
  }

  /**
   * GetControllers returns a list of all currently active controllers.
   */
  getControllers(): Controller[] {
    return Array.from(this.controllers.keys())
  }

  /**
   * AddController adds a controller to the bus and calls Execute().
   */
  async addController(
    ctx: Context,
    ctrl: Controller,
    cb: ((exitErr: Error | null) => void) | null = null,
  ): Promise<[() => void, Error | null]> {
    if (this.controllers.has(ctrl)) {
      return [() => {}, new Error('controller already added to bus')]
    }

    const [execCtx, cancelExecCtx] = ctx.withCancel()

    // Add the controller as a handler
    const [handlerRelease, handlerErr] = await this.addHandler(ctrl)
    if (handlerErr) {
      cancelExecCtx()
      return [() => {}, handlerErr]
    }

    // Store the controller handle
    this.controllers.set(ctrl, {
      controller: ctrl,
      execCtx,
      cancelCtx: cancelExecCtx,
      callback: cb,
    })

    // Start the controller execution
    this.executeControllerInternal(execCtx, ctrl, cancelExecCtx, cb)

    // Return a release function
    return [
      () => {
        this.removeController(ctrl)
        if (cb) cb(null)
      },
      null,
    ]
  }

  /**
   * ExecuteController adds a controller to the bus and calls Execute().
   */
  async executeController(
    ctx: Context,
    ctrl: Controller,
  ): Promise<Error | null> {
    const existing = this.controllers.get(ctrl)
    if (existing) {
      return new Error('controller already executing')
    }

    const [handlerRelease, handlerErr] = await this.addHandler(ctrl)
    if (handlerErr) {
      return handlerErr
    }

    // Execute the controller
    try {
      const err = await ctrl.execute(ctx)
      if (err) {
        return err
      }
      return null
    } finally {
      handlerRelease()
    }
  }

  /**
   * RemoveController removes the controller from the bus.
   */
  removeController(ctrl: Controller): void {
    const handle = this.controllers.get(ctrl)
    if (!handle) {
      return
    }

    // Cancel the execution context
    handle.cancelCtx()

    // Remove the handler
    const handlerRelease = this.handlerReleases.get(ctrl)
    if (handlerRelease) {
      handlerRelease()
      this.handlerReleases.delete(ctrl)
    }

    // Remove from controllers map
    this.controllers.delete(ctrl)
  }

  private async executeControllerInternal(
    ctx: Context,
    ctrl: Controller,
    cancelCtx: () => void,
    cb: ((exitErr: Error | null) => void) | null,
  ): Promise<void> {
    try {
      const err = await ctrl.execute(ctx)

      // If we're still in the controllers map (not removed)
      if (this.controllers.has(ctrl)) {
        if (err) {
          // Execution failed with an error
          if (cb) cb(err)

          // Remove controller from bus since it returned an error
          this.removeController(ctrl)
        } else {
          // Normal exit
          if (cb) cb(null)
        }
      }
    } catch (e) {
      const err = e instanceof Error ? e : new Error(String(e))
      if (cb) cb(err)

      // Remove controller from bus on exception
      this.removeController(ctrl)
    }
  }
}

/**
 * Creates a new in-memory bus.
 *
 * @param directiveCtrl - The directive controller to use
 * @returns A new in-memory bus
 */
export function newBus(directiveCtrl: DirectiveController): Bus {
  return new InMemBus(directiveCtrl)
}
