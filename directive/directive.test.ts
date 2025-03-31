import { describe, it, expect, vi } from 'vitest'
import {
  newDebugValues,
  newProtoDebugValues,
  newAttachedValue,
  newTypedAttachedValue,
  newTransformedAttachedValue,
} from './directive'

describe('DebugValues', () => {
  it('should create an empty object', () => {
    const debugVals = newDebugValues()
    expect(debugVals).toEqual({})
  })

  it('should convert to ProtoDebugValue array', () => {
    const debugVals = newDebugValues()
    debugVals['key1'] = ['value1', 'value2']
    debugVals['key2'] = ['value3']

    const protoVals = newProtoDebugValues(debugVals)
    expect(protoVals).toHaveLength(2)
    expect(protoVals).toEqual([
      { key: 'key1', values: ['value1', 'value2'] },
      { key: 'key2', values: ['value3'] },
    ])
  })

  it('should sort ProtoDebugValue array by key', () => {
    const debugVals = newDebugValues()
    debugVals['z'] = ['z-value']
    debugVals['a'] = ['a-value']
    debugVals['m'] = ['m-value']

    const protoVals = newProtoDebugValues(debugVals)
    expect(protoVals).toHaveLength(3)
    expect(protoVals[0].key).toBe('a')
    expect(protoVals[1].key).toBe('m')
    expect(protoVals[2].key).toBe('z')
  })
})

describe('AttachedValue', () => {
  it('should create an attached value', () => {
    const value = { test: 'value' }
    const attached = newAttachedValue(123, value)

    expect(attached.getValueID()).toBe(123)
    expect(attached.getValue()).toBe(value)
  })

  it('should create a typed attached value', () => {
    const value = { test: 'value' }
    const attached = newTypedAttachedValue<{ test: string }>(123, value)

    expect(attached.getValueID()).toBe(123)
    expect(attached.getValue()).toBe(value)
    expect(attached.getValue().test).toBe('value')
  })

  it('should create a transformed attached value', () => {
    const value = { original: 'value' }
    const typedValue = newTypedAttachedValue<{ original: string }>(123, value)
    const transformed = { transformed: 'data' }

    const transformedValue = newTransformedAttachedValue(
      typedValue,
      transformed,
    )

    expect(transformedValue.getValueID()).toBe(123)
    expect(transformedValue.getValue()).toBe(value)
    expect(transformedValue.getValue().original).toBe('value')
    expect(transformedValue.getTransformedValue()).toBe(transformed)
    expect(transformedValue.getTransformedValue().transformed).toBe('data')
  })
})
