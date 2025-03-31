import { Context } from '../core/core.js'
import { Handler, Instance } from '../directive/directive.js'

/**
 * Info is information about a controller.
 */
export interface Info {
  // ID is the controller ID.
  id: string
  // Version is the semver version of the controller.
  version: string
  // Description is a short description of the controller.
  description?: string
}

/**
 * Controller tracks a particular process.
 */
export interface Controller extends Handler {
  /**
   * GetControllerInfo returns information about the controller.
   */
  getControllerInfo(): Info

  /**
   * Execute executes the controller goroutine.
   * Returning null ends execution.
   * Returning an error triggers a retry with backoff.
   * Retry will NOT re-construct the controller, just re-start Execute.
   */
  execute(ctx: Context): Promise<Error | null>

  /**
   * Close releases any resources used by the controller.
   * Error indicates any issue encountered releasing.
   */
  close(): Promise<Error | null>
}

/**
 * Creates a new controller information object.
 *
 * @param id - The controller ID
 * @param version - The semver version
 * @param description - Description of the controller
 * @returns The controller info object
 */
export function newInfo(
  id: string,
  version: string,
  description?: string,
): Info {
  return {
    id,
    version,
    description,
  }
}
