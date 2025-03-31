import { describe, it, expect, vi } from 'vitest'
import { Bus } from './bus'
import { Context } from '../core/core'
import { Controller } from '../controller/controller'
import {
  Directive,
  Instance,
  Reference,
  ReferenceHandler,
  Handler,
} from '../directive/directive'

class MockBus implements Bus {
  getControllersMock = vi.fn().mockReturnValue([])
  addControllerMock = vi.fn().mockReturnValue(Promise.resolve([() => {}, null]))
  executeControllerMock = vi.fn().mockReturnValue(Promise.resolve(null))
  removeControllerMock = vi.fn()

  getDirectivesMock = vi.fn().mockReturnValue([])
  addDirectiveMock = vi
    .fn()
    .mockReturnValue(Promise.resolve([{} as Instance, {} as Reference, null]))
  addHandlerMock = vi.fn().mockReturnValue(Promise.resolve([() => {}, null]))

  getControllers(): Controller[] {
    return this.getControllersMock()
  }

  async addController(
    ctx: Context,
    ctrl: Controller,
    cb?: ((exitErr: Error | null) => void) | null,
  ): Promise<[() => void, Error | null]> {
    return this.addControllerMock(ctx, ctrl, cb)
  }

  async executeController(
    ctx: Context,
    ctrl: Controller,
  ): Promise<Error | null> {
    return this.executeControllerMock(ctx, ctrl)
  }

  removeController(ctrl: Controller): void {
    this.removeControllerMock(ctrl)
  }

  getDirectives(): Instance[] {
    return this.getDirectivesMock()
  }

  async addDirective(
    dir: Directive,
    handler: ReferenceHandler | null,
  ): Promise<[Instance, Reference, Error | null]> {
    return this.addDirectiveMock(dir, handler)
  }

  async addHandler(handler: Handler): Promise<[() => void, Error | null]> {
    return this.addHandlerMock(handler)
  }
}

describe('Bus interface', () => {
  it('should define the expected methods', () => {
    const bus = new MockBus()

    // Test that the interface has all required methods
    expect(typeof bus.getControllers).toBe('function')
    expect(typeof bus.addController).toBe('function')
    expect(typeof bus.executeController).toBe('function')
    expect(typeof bus.removeController).toBe('function')
    expect(typeof bus.getDirectives).toBe('function')
    expect(typeof bus.addDirective).toBe('function')
    expect(typeof bus.addHandler).toBe('function')
  })

  it('should delegate directive methods correctly', async () => {
    const bus = new MockBus()
    const mockDir = {} as Directive
    const mockHandler = {} as ReferenceHandler

    await bus.addDirective(mockDir, mockHandler)
    expect(bus.addDirectiveMock).toHaveBeenCalledWith(mockDir, mockHandler)

    const mockDirHandler = {} as Handler
    await bus.addHandler(mockDirHandler)
    expect(bus.addHandlerMock).toHaveBeenCalledWith(mockDirHandler)

    bus.getDirectives()
    expect(bus.getDirectivesMock).toHaveBeenCalled()
  })

  it('should handle controller methods correctly', async () => {
    const bus = new MockBus()
    const mockCtx = {} as Context
    const mockCtrl = {} as Controller
    const mockCb = vi.fn()

    await bus.addController(mockCtx, mockCtrl, mockCb)
    expect(bus.addControllerMock).toHaveBeenCalledWith(
      mockCtx,
      mockCtrl,
      mockCb,
    )

    await bus.executeController(mockCtx, mockCtrl)
    expect(bus.executeControllerMock).toHaveBeenCalledWith(mockCtx, mockCtrl)

    bus.removeController(mockCtrl)
    expect(bus.removeControllerMock).toHaveBeenCalledWith(mockCtrl)

    bus.getControllers()
    expect(bus.getControllersMock).toHaveBeenCalled()
  })
})
