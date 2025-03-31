import { describe, it, expect } from 'vitest'
import { Config, equalsConfig } from './config'

// Mock config implementation
class MockConfig implements Config {
  constructor(
    private id: string,
    public field1: string,
    public field2: number,
  ) {}

  validate(): Error | null {
    return null
  }

  getConfigID(): string {
    return this.id
  }

  equalsConfig(other: Config): boolean {
    return equalsConfig(this, other)
  }
}

describe('Config', () => {
  describe('equalsConfig', () => {
    it('should return false if second config is null or undefined', () => {
      const config = new MockConfig('test-config', 'value', 42)

      expect(equalsConfig(config, null)).toBe(false)
      expect(equalsConfig(config, undefined)).toBe(false)
    })

    it('should return false if config IDs are different', () => {
      const config1 = new MockConfig('config-1', 'value', 42)
      const config2 = new MockConfig('config-2', 'value', 42)

      expect(equalsConfig(config1, config2)).toBe(false)
    })

    it('should return false if configs have different field values', () => {
      const config1 = new MockConfig('test-config', 'value1', 42)
      const config2 = new MockConfig('test-config', 'value2', 42)
      const config3 = new MockConfig('test-config', 'value1', 43)

      expect(equalsConfig(config1, config2)).toBe(false)
      expect(equalsConfig(config1, config3)).toBe(false)
    })

    it('should return true if configs have the same ID and field values', () => {
      const config1 = new MockConfig('test-config', 'value', 42)
      const config2 = new MockConfig('test-config', 'value', 42)

      expect(equalsConfig(config1, config2)).toBe(true)
    })

    it('should work with the Config.equalsConfig implementation', () => {
      const config1 = new MockConfig('test-config', 'value', 42)
      const config2 = new MockConfig('test-config', 'value', 42)
      const config3 = new MockConfig('test-config', 'different', 42)

      expect(config1.equalsConfig(config2)).toBe(true)
      expect(config1.equalsConfig(config3)).toBe(false)
    })
  })
})
