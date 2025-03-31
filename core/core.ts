/**
 * Context is an interface for cancelable operations.
 * It mimics Go's context.Context in TypeScript.
 */
export interface Context {
  /**
   * Done returns a Promise that resolves when the context is done (canceled).
   */
  done(): Promise<void>

  /**
   * IsDone returns true if the context is already done.
   */
  isDone(): boolean

  /**
   * WithCancel returns a new context that can be canceled manually.
   */
  withCancel(): [Context, () => void]

  /**
   * WithTimeout returns a new context that will be automatically canceled after the specified timeout.
   */
  withTimeout(timeoutMs: number): [Context, () => void]
}

/**
 * Creates a background context that is never canceled.
 */
export function background(): Context {
  return new BackgroundContext()
}

class BackgroundContext implements Context {
  done(): Promise<void> {
    return new Promise<void>(() => {
      // This promise never resolves
    })
  }

  isDone(): boolean {
    return false
  }

  withCancel(): [Context, () => void] {
    const controller = new AbortController()
    return [new CancelableContext(controller), () => controller.abort()]
  }

  withTimeout(timeoutMs: number): [Context, () => void] {
    const controller = new AbortController()
    const timerId = setTimeout(() => controller.abort(), timeoutMs)

    return [
      new CancelableContext(controller),
      () => {
        clearTimeout(timerId)
        controller.abort()
      },
    ]
  }
}

class CancelableContext implements Context {
  private readonly controller: AbortController

  constructor(controller: AbortController) {
    this.controller = controller
  }

  done(): Promise<void> {
    return new Promise<void>((resolve) => {
      if (this.controller.signal.aborted) {
        resolve()
        return
      }

      this.controller.signal.addEventListener('abort', () => {
        resolve()
      })
    })
  }

  isDone(): boolean {
    return this.controller.signal.aborted
  }

  withCancel(): [Context, () => void] {
    const controller = new AbortController()

    // Link this new controller to the parent controller
    if (!this.controller.signal.aborted) {
      this.controller.signal.addEventListener('abort', () => {
        controller.abort()
      })
    } else {
      controller.abort() // Already canceled
    }

    return [new CancelableContext(controller), () => controller.abort()]
  }

  withTimeout(timeoutMs: number): [Context, () => void] {
    const controller = new AbortController()

    // Link this new controller to the parent controller
    if (!this.controller.signal.aborted) {
      this.controller.signal.addEventListener('abort', () => {
        controller.abort()
      })
    } else {
      controller.abort() // Already canceled
      return [new CancelableContext(controller), () => controller.abort()]
    }

    const timerId = setTimeout(() => controller.abort(), timeoutMs)

    return [
      new CancelableContext(controller),
      () => {
        clearTimeout(timerId)
        controller.abort()
      },
    ]
  }
}

export interface Logger {
  debug(message: string): void
  info(message: string): void
  warn(message: string): void
  error(message: string): void
  debugf(format: string, ...args: any[]): void
  infof(format: string, ...args: any[]): void
  warnf(format: string, ...args: any[]): void
  errorf(format: string, ...args: any[]): void
}

/**
 * Simple string formatter (replaces %s with arguments).
 */
function formatString(format: string, ...args: any[]): string {
  return format.replace(/%s/g, () => String(args.shift()))
}

/**
 * Console-based implementation of the Logger interface.
 */
export class ConsoleLogger implements Logger {
  debug(message: string): void {
    console.debug(message)
  }

  info(message: string): void {
    console.info(message)
  }

  warn(message: string): void {
    console.warn(message)
  }

  error(message: string): void {
    console.error(message)
  }

  debugf(format: string, ...args: any[]): void {
    console.debug(formatString(format, ...args))
  }

  infof(format: string, ...args: any[]): void {
    console.info(formatString(format, ...args))
  }

  warnf(format: string, ...args: any[]): void {
    console.warn(formatString(format, ...args))
  }

  errorf(format: string, ...args: any[]): void {
    console.error(formatString(format, ...args))
  }
}
