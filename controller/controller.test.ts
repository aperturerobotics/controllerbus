import { describe, it, expect, vi } from 'vitest'
import { Controller, Info, newInfo } from './controller'
import { Context } from '../core/core'
import { Directive, Instance, Resolver } from '../directive/directive'

describe('Controller interface', () => {
  it('should create controller info correctly', () => {
    const id = 'test-controller'
    const version = '1.0.0'
    const description = 'Test Controller'

    const info = newInfo(id, version, description)

    expect(info).toEqual({
      id,
      version,
      description,
    })
  })

  it('should create controller info with default description', () => {
    const id = 'test-controller'
    const version = '1.0.0'

    const info = newInfo(id, version)

    expect(info).toEqual({
      id,
      version,
      description: '',
    })
  })

  it('should define the expected methods in the interface', () => {
    // Define a mock controller implementing the Controller interface
    class MockController implements Controller {
      getControllerInfo = vi.fn().mockReturnValue({
        id: 'mock-controller',
        version: '1.0.0',
        description: 'Mock Controller',
      })

      handleDirective = vi.fn().mockResolvedValue(null)

      execute = vi.fn().mockResolvedValue(null)

      close = vi.fn().mockResolvedValue(null)
    }

    const controller = new MockController()

    // Check that it has all the expected methods
    expect(typeof controller.getControllerInfo).toBe('function')
    expect(typeof controller.handleDirective).toBe('function')
    expect(typeof controller.execute).toBe('function')
    expect(typeof controller.close).toBe('function')

    // Verify method signatures by calling them with expected arguments
    controller.getControllerInfo()
    expect(controller.getControllerInfo).toHaveBeenCalled()

    const mockCtx = {} as Context
    const mockInstance = {} as Instance
    controller.handleDirective(mockCtx, mockInstance)
    expect(controller.handleDirective).toHaveBeenCalledWith(
      mockCtx,
      mockInstance,
    )

    controller.execute(mockCtx)
    expect(controller.execute).toHaveBeenCalledWith(mockCtx)

    controller.close()
    expect(controller.close).toHaveBeenCalled()
  })
})
