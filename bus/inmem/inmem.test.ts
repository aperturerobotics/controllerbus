import { describe, it, expect, vi, beforeEach } from 'vitest'
import { InMemBus, newBus } from './inmem'
import { background } from '../../core/core'
import {
  Controller as DirectiveController,
  Directive,
  Instance,
  Reference,
  Handler,
  ReferenceHandler,
  Resolver,
  ResolverHandler,
  ValueOptions,
} from '../../directive/directive'
import { Controller, Info } from '../../controller/controller'

// Mock directive implementation
class MockDirective implements Directive {
  constructor(private name: string) {}

  validate(): Error | null {
    return null
  }

  getValueOptions(): ValueOptions {
    return {
      maxValueCount: 0,
      maxValueHardCap: false,
      unrefDisposeDur: 0,
      unrefDisposeEmptyImmediate: true,
    }
  }

  getName(): string {
    return this.name
  }
}

// Mock controller implementation
class MockController implements Controller {
  public readonly executePromise: Promise<Error | null>
  private readonly resolveExecute: (err: Error | null) => void
  private readonly info: Info = {
    id: 'mock-controller',
    version: '1.0.0',
    description: 'Mock Controller',
  }

  constructor(public readonly handleDirectiveMock = vi.fn()) {
    let resolveExec: (err: Error | null) => void = () => {}
    this.executePromise = new Promise<Error | null>((resolve) => {
      resolveExec = resolve
    })
    this.resolveExecute = resolveExec
  }

  async execute(_ctx: any): Promise<Error | null> {
    return this.executePromise
  }

  finish(err: Error | null = null): void {
    this.resolveExecute(err)
  }

  getControllerInfo(): Info {
    return this.info
  }

  async handleDirective(ctx: any, instance: any): Promise<Resolver[] | null> {
    return this.handleDirectiveMock(ctx, instance)
  }

  async close(): Promise<Error | null> {
    return null
  }
}

// Mock directive controller
class MockDirectiveController implements DirectiveController {
  addDirectiveMock = vi.fn().mockImplementation((dir, handler) => {
    const mockInstance = {
      getDirective: () => dir,
      getDirectiveIdent: () => dir.getName(),
    } as Instance

    const mockRef = {
      release: vi.fn(),
    } as Reference

    return Promise.resolve([mockInstance, mockRef, null])
  })

  addHandlerMock = vi.fn().mockImplementation((handler) => {
    return Promise.resolve([() => {}, null])
  })

  getDirectivesMock = vi.fn().mockReturnValue([])

  async addDirective(
    dir: Directive,
    handler: ReferenceHandler | null,
  ): Promise<[Instance, Reference, Error | null]> {
    return this.addDirectiveMock(dir, handler)
  }

  async addHandler(handler: Handler): Promise<[() => void, Error | null]> {
    return this.addHandlerMock(handler)
  }

  getDirectives(): Instance[] {
    return this.getDirectivesMock()
  }
}

describe('InMemBus', () => {
  let directiveCtrl: MockDirectiveController
  let bus: InMemBus

  beforeEach(() => {
    directiveCtrl = new MockDirectiveController()
    bus = newBus(directiveCtrl) as InMemBus
  })

  it('should delegate directive methods to the directive controller', async () => {
    const directive = new MockDirective('test-directive')
    const handler = {} as ReferenceHandler

    await bus.addDirective(directive, handler)
    expect(directiveCtrl.addDirectiveMock).toHaveBeenCalledWith(
      directive,
      handler,
    )

    const mockHandler = {} as Handler
    await bus.addHandler(mockHandler)
    expect(directiveCtrl.addHandlerMock).toHaveBeenCalledWith(mockHandler)

    bus.getDirectives()
    expect(directiveCtrl.getDirectivesMock).toHaveBeenCalled()
  })

  it('should add and get controllers', async () => {
    const ctx = background()
    const controller = new MockController()

    // Add controller
    const [release, err] = await bus.addController(ctx, controller)
    expect(err).toBeNull()
    expect(typeof release).toBe('function')

    // Controller should be registered
    const controllers = bus.getControllers()
    expect(controllers).toContain(controller)

    // Release the controller
    release()

    // Controller should be removed
    const controllersAfterRelease = bus.getControllers()
    expect(controllersAfterRelease).not.toContain(controller)
  })

  it('should execute controllers', async () => {
    const ctx = background()
    const controller = new MockController()

    // Execute controller
    const executePromise = bus.executeController(ctx, controller)

    // Finish execution
    controller.finish(null)

    // Should complete without error
    const err = await executePromise
    expect(err).toBeNull()
  })

  it('should propagate controller execution errors', async () => {
    const ctx = background()
    const controller = new MockController()
    const testError = new Error('test error')

    // Execute controller
    const executePromise = bus.executeController(ctx, controller)

    // Finish with error
    controller.finish(testError)

    // Should propagate error
    const err = await executePromise
    expect(err).toBe(testError)
  })

  it('should call callback when controller exits', async () => {
    const ctx = background()
    const controller = new MockController()
    const callback = vi.fn()

    // Add controller with callback
    const [release, err] = await bus.addController(ctx, controller, callback)
    expect(err).toBeNull()

    // Finish execution
    controller.finish(null)

    // Wait for async operations
    await vi.waitFor(() => {
      expect(callback).toHaveBeenCalledWith(null)
    })
  })

  it('should call callback with error when controller fails', async () => {
    const ctx = background()
    const controller = new MockController()
    const callback = vi.fn()
    const testError = new Error('test error')

    // Add controller with callback
    const [release, err] = await bus.addController(ctx, controller, callback)
    expect(err).toBeNull()

    // Finish with error
    controller.finish(testError)

    // Wait for async operations
    await vi.waitFor(() => {
      expect(callback).toHaveBeenCalledWith(testError)
    })
  })
})
