import { Config } from '../../config/config'
import { ToyControllerConfig } from './controller.pb'
import { equalsConfig } from '../../config/config'
import { ControllerID } from './controller'

// Extend the proto-generated ToyControllerConfig with the Config interface methods
declare module './controller.pb' {
  interface ToyControllerConfig extends Config {}
}

// Add the Config interface methods to ToyControllerConfig objects
// Since ToyControllerConfig is an interface not a class, we can't use prototype
// Instead, extend a global validator object that can be used to implement these functions
export const toyControllerConfigImpl = {
  validate: function(this: ToyControllerConfig): Error | null {
    if (!this.name) {
      return new Error('name cannot be empty')
    }
    return null
  },
  
  getConfigID: function(this: ToyControllerConfig): string {
    return ControllerID
  },
  
  equalsConfig: function(this: ToyControllerConfig, other: Config): boolean {
    return equalsConfig<ToyControllerConfig>(this, other as ToyControllerConfig)
  }
}

// Helper function to create a ToyControllerConfig with the Config interface implementations
export function createToyControllerConfig(name: string): ToyControllerConfig & Config {
  return Object.assign(
    { name },
    {
      validate: toyControllerConfigImpl.validate,
      getConfigID: toyControllerConfigImpl.getConfigID,
      equalsConfig: toyControllerConfigImpl.equalsConfig
    }
  )
}
