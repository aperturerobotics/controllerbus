import { Context } from '../core/core'
import { Controller } from '../controller/controller'
import { Controller as DirectiveController } from '../directive/directive'

/**
 * Bus manages running controllers. It has an attached directive controller,
 * which is used to build declarative state requests between controllers.
 */
export interface Bus extends DirectiveController {
  /**
   * GetControllers returns a list of all currently active controllers.
   */
  getControllers(): Controller[]

  /**
   * AddController adds a controller to the bus and calls Execute().
   * The controller will exit if ctx is canceled.
   * Returns a release function for the controller reference.
   * The controller will receive directive callbacks until removed.
   * Any fatal error in the controller is written to cb.
   * If the controller is released, cb will be called with null.
   * cb can be null
   */
  addController(
    ctx: Context,
    ctrl: Controller,
    cb?: ((exitErr: Error | null) => void) | null,
  ): Promise<[() => void, Error | null]>

  /**
   * ExecuteController adds a controller to the bus and calls Execute().
   * The controller will exit if ctx is canceled.
   * Any fatal error in the controller is returned.
   * The controller will receive directive callbacks.
   * If this function returns null, call RemoveController to remove the controller.
   */
  executeController(ctx: Context, ctrl: Controller): Promise<Error | null>

  /**
   * RemoveController removes the controller from the bus.
   * The controller will no longer receive callbacks.
   * Note: this might not cancel the Execute() context automatically.
   */
  removeController(ctrl: Controller): void
}
