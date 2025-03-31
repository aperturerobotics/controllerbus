import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ToyController, ControllerID, Version } from './controller'
import { ToyControllerConfig } from './controller.pb'
import { createToyControllerConfig } from './controller_config'
import { background } from '../../core/core'

// Mock logger implementation
class MockLogger {
  debug = vi.fn()
  info = vi.fn()
  warn = vi.fn()
  error = vi.fn()
  debugf = vi.fn()
  infof = vi.fn()
  warnf = vi.fn()
  errorf = vi.fn()
}

describe('ToyController', () => {
  let logger: MockLogger
  let config: ToyControllerConfig
  let controller: ToyController

  beforeEach(() => {
    logger = new MockLogger()
    config = createToyControllerConfig('Test User')
    controller = new ToyController(logger, config)
  })

  it('should return correct controller information', () => {
    const info = controller.getControllerInfo()

    expect(info.id).toBe(ControllerID)
    expect(info.version).toBe(Version)
    expect(info.description).toBe('toy controller')
  })

  it('should handle directive and return null', async () => {
    const ctx = background()
    const instance = { getDirective: vi.fn() } as any

    const resolvers = await controller.handleDirective(ctx, instance)
    expect(resolvers).toBeNull()
  })

  it('should execute and log debug message', async () => {
    const ctx = background()
    const [cancelCtx, cancelFn] = ctx.withCancel()

    // Start execution (which should block on ctx.done())
    const executePromise = controller.execute(cancelCtx)

    // Verify the log message
    expect(logger.debug).toHaveBeenCalledWith('toy controller executed')

    // Cancel the context to complete execution
    cancelFn()

    // Should complete without error
    const err = await executePromise
    expect(err).toBeNull()
  })

  it('should say hello correctly', () => {
    controller.sayHello()

    expect(logger.debugf).toHaveBeenCalledWith('Hello %s!', 'Test User')
  })

  it('should close without error', async () => {
    const err = await controller.close()
    expect(err).toBeNull()
  })
})
