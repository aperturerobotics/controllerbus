/**
 * Config is a configuration object that can be validated and compared.
 */
export interface Config {
  /**
   * Validate validates the configuration.
   * This is a cursory validation to see if the values "look correct."
   */
  validate(): Error | null

  /**
   * GetConfigID returns the unique string for this configuration type.
   * This string is stored with the encoded config.
   */
  getConfigID(): string

  /**
   * EqualsConfig checks if the config is equal to another.
   */
  equalsConfig(other: Config): boolean
}

/**
 * Checks if two configs are equal by comparing their JSON representations.
 * This is a generic implementation that should work for most configs.
 *
 * @param a - First config to compare
 * @param b - Second config to compare
 * @returns true if the configs are equal
 */
export function equalsConfig<T extends Config>(
  a: T,
  b: Config | null | undefined,
): boolean {
  if (!b) {
    return false
  }

  if (a.getConfigID() !== b.getConfigID()) {
    return false
  }

  // Compare by serializing to JSON
  return JSON.stringify(a) === JSON.stringify(b)
}
