import { describe, it, expect, vi } from 'vitest'
import {
  ToyFactory,
  newToyFactory,
  toyFactoryVersion,
  toyControllerID,
} from './controller_factory'
import { ToyControllerConfig } from './controller.pb'
import { createToyControllerConfig } from './controller_config'
import { ToyController } from './controller'
import { background } from '../../core/core'
import { Config } from '../../config/config'

// Mock logger
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

describe('ToyFactory', () => {
  it('should create a new factory', () => {
    const factory = newToyFactory()
    expect(factory).toBeInstanceOf(ToyFactory)
  })

  it('should return correct controller and config IDs', () => {
    const factory = newToyFactory()
    expect(factory.getControllerID()).toBe(toyControllerID)
    expect(factory.getConfigID()).toBe(toyControllerID)
  })

  it('should return correct version', () => {
    const factory = newToyFactory()
    expect(factory.getVersion()).toBe(toyFactoryVersion)
  })

  it('should construct a ToyControllerConfig instance', () => {
    const factory = newToyFactory()
    const config = factory.constructConfig()
    expect(config).toBeInstanceOf(ToyControllerConfig)
  })

  it('should construct a controller with valid config', async () => {
    const factory = newToyFactory()
    const config = new ToyControllerConfig()
    config.name = 'Test User'

    const ctx = background()
    const logger = new MockLogger()
    const opts = {
      getLogger: () => logger,
    }

    const [controller, err] = await factory.construct(ctx, config, opts)

    expect(err).toBeNull()
    expect(controller).toBeInstanceOf(ToyController)
  })

  it('should return error with invalid config', async () => {
    const factory = newToyFactory()
    const invalidConfig = {} as Config

    const ctx = background()
    const logger = new MockLogger()
    const opts = {
      getLogger: () => logger,
    }

    const [controller, err] = await factory.construct(ctx, invalidConfig, opts)

    expect(err).not.toBeNull()
    expect(err?.message).toBe('wrong type of config')
    expect(controller).toBeNull()
  })
})
