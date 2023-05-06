/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'

export const protobufPackage = 'configset.proto'

/** ConfigSet contains a configuration set. */
export interface ConfigSet {
  /** Configurations contains the controller configurations. */
  configurations: { [key: string]: ControllerConfig }
}

export interface ConfigSet_ConfigurationsEntry {
  key: string
  value: ControllerConfig | undefined
}

/** ControllerConfig contains a controller configuration. */
export interface ControllerConfig {
  /** Id is the config ID. */
  id: string
  /** Rev is the revision number of the configuration. */
  rev: Long
  /**
   * Config is the configuration object.
   * Supports: protobuf and json (must start with {).
   */
  config: Uint8Array
}

function createBaseConfigSet(): ConfigSet {
  return { configurations: {} }
}

export const ConfigSet = {
  encode(
    message: ConfigSet,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    Object.entries(message.configurations).forEach(([key, value]) => {
      ConfigSet_ConfigurationsEntry.encode(
        { key: key as any, value },
        writer.uint32(10).fork()
      ).ldelim()
    })
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ConfigSet {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseConfigSet()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          const entry1 = ConfigSet_ConfigurationsEntry.decode(
            reader,
            reader.uint32()
          )
          if (entry1.value !== undefined) {
            message.configurations[entry1.key] = entry1.value
          }
          continue
      }
      if ((tag & 7) === 4 || tag === 0) {
        break
      }
      reader.skipType(tag & 7)
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<ConfigSet, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<ConfigSet | ConfigSet[]>
      | Iterable<ConfigSet | ConfigSet[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [ConfigSet.encode(p).finish()]
        }
      } else {
        yield* [ConfigSet.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, ConfigSet>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<ConfigSet> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [ConfigSet.decode(p)]
        }
      } else {
        yield* [ConfigSet.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): ConfigSet {
    return {
      configurations: isObject(object.configurations)
        ? Object.entries(object.configurations).reduce<{
            [key: string]: ControllerConfig
          }>((acc, [key, value]) => {
            acc[key] = ControllerConfig.fromJSON(value)
            return acc
          }, {})
        : {},
    }
  },

  toJSON(message: ConfigSet): unknown {
    const obj: any = {}
    obj.configurations = {}
    if (message.configurations) {
      Object.entries(message.configurations).forEach(([k, v]) => {
        obj.configurations[k] = ControllerConfig.toJSON(v)
      })
    }
    return obj
  },

  create<I extends Exact<DeepPartial<ConfigSet>, I>>(base?: I): ConfigSet {
    return ConfigSet.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<ConfigSet>, I>>(
    object: I
  ): ConfigSet {
    const message = createBaseConfigSet()
    message.configurations = Object.entries(
      object.configurations ?? {}
    ).reduce<{ [key: string]: ControllerConfig }>((acc, [key, value]) => {
      if (value !== undefined) {
        acc[key] = ControllerConfig.fromPartial(value)
      }
      return acc
    }, {})
    return message
  },
}

function createBaseConfigSet_ConfigurationsEntry(): ConfigSet_ConfigurationsEntry {
  return { key: '', value: undefined }
}

export const ConfigSet_ConfigurationsEntry = {
  encode(
    message: ConfigSet_ConfigurationsEntry,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.key !== '') {
      writer.uint32(10).string(message.key)
    }
    if (message.value !== undefined) {
      ControllerConfig.encode(message.value, writer.uint32(18).fork()).ldelim()
    }
    return writer
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): ConfigSet_ConfigurationsEntry {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseConfigSet_ConfigurationsEntry()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.key = reader.string()
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.value = ControllerConfig.decode(reader, reader.uint32())
          continue
      }
      if ((tag & 7) === 4 || tag === 0) {
        break
      }
      reader.skipType(tag & 7)
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<ConfigSet_ConfigurationsEntry, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<
          ConfigSet_ConfigurationsEntry | ConfigSet_ConfigurationsEntry[]
        >
      | Iterable<
          ConfigSet_ConfigurationsEntry | ConfigSet_ConfigurationsEntry[]
        >
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [ConfigSet_ConfigurationsEntry.encode(p).finish()]
        }
      } else {
        yield* [ConfigSet_ConfigurationsEntry.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, ConfigSet_ConfigurationsEntry>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<ConfigSet_ConfigurationsEntry> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [ConfigSet_ConfigurationsEntry.decode(p)]
        }
      } else {
        yield* [ConfigSet_ConfigurationsEntry.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): ConfigSet_ConfigurationsEntry {
    return {
      key: isSet(object.key) ? String(object.key) : '',
      value: isSet(object.value)
        ? ControllerConfig.fromJSON(object.value)
        : undefined,
    }
  },

  toJSON(message: ConfigSet_ConfigurationsEntry): unknown {
    const obj: any = {}
    message.key !== undefined && (obj.key = message.key)
    message.value !== undefined &&
      (obj.value = message.value
        ? ControllerConfig.toJSON(message.value)
        : undefined)
    return obj
  },

  create<I extends Exact<DeepPartial<ConfigSet_ConfigurationsEntry>, I>>(
    base?: I
  ): ConfigSet_ConfigurationsEntry {
    return ConfigSet_ConfigurationsEntry.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<ConfigSet_ConfigurationsEntry>, I>>(
    object: I
  ): ConfigSet_ConfigurationsEntry {
    const message = createBaseConfigSet_ConfigurationsEntry()
    message.key = object.key ?? ''
    message.value =
      object.value !== undefined && object.value !== null
        ? ControllerConfig.fromPartial(object.value)
        : undefined
    return message
  },
}

function createBaseControllerConfig(): ControllerConfig {
  return { id: '', rev: Long.UZERO, config: new Uint8Array() }
}

export const ControllerConfig = {
  encode(
    message: ControllerConfig,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.id !== '') {
      writer.uint32(10).string(message.id)
    }
    if (!message.rev.isZero()) {
      writer.uint32(16).uint64(message.rev)
    }
    if (message.config.length !== 0) {
      writer.uint32(26).bytes(message.config)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ControllerConfig {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseControllerConfig()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.id = reader.string()
          continue
        case 2:
          if (tag !== 16) {
            break
          }

          message.rev = reader.uint64() as Long
          continue
        case 3:
          if (tag !== 26) {
            break
          }

          message.config = reader.bytes()
          continue
      }
      if ((tag & 7) === 4 || tag === 0) {
        break
      }
      reader.skipType(tag & 7)
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<ControllerConfig, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<ControllerConfig | ControllerConfig[]>
      | Iterable<ControllerConfig | ControllerConfig[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [ControllerConfig.encode(p).finish()]
        }
      } else {
        yield* [ControllerConfig.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, ControllerConfig>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<ControllerConfig> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [ControllerConfig.decode(p)]
        }
      } else {
        yield* [ControllerConfig.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): ControllerConfig {
    return {
      id: isSet(object.id) ? String(object.id) : '',
      rev: isSet(object.rev) ? Long.fromValue(object.rev) : Long.UZERO,
      config: isSet(object.config)
        ? bytesFromBase64(object.config)
        : new Uint8Array(),
    }
  },

  toJSON(message: ControllerConfig): unknown {
    const obj: any = {}
    message.id !== undefined && (obj.id = message.id)
    message.rev !== undefined &&
      (obj.rev = (message.rev || Long.UZERO).toString())
    message.config !== undefined &&
      (obj.config = base64FromBytes(
        message.config !== undefined ? message.config : new Uint8Array()
      ))
    return obj
  },

  create<I extends Exact<DeepPartial<ControllerConfig>, I>>(
    base?: I
  ): ControllerConfig {
    return ControllerConfig.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<ControllerConfig>, I>>(
    object: I
  ): ControllerConfig {
    const message = createBaseControllerConfig()
    message.id = object.id ?? ''
    message.rev =
      object.rev !== undefined && object.rev !== null
        ? Long.fromValue(object.rev)
        : Long.UZERO
    message.config = object.config ?? new Uint8Array()
    return message
  },
}

declare var self: any | undefined
declare var window: any | undefined
declare var global: any | undefined
var tsProtoGlobalThis: any = (() => {
  if (typeof globalThis !== 'undefined') {
    return globalThis
  }
  if (typeof self !== 'undefined') {
    return self
  }
  if (typeof window !== 'undefined') {
    return window
  }
  if (typeof global !== 'undefined') {
    return global
  }
  throw 'Unable to locate global object'
})()

function bytesFromBase64(b64: string): Uint8Array {
  if (tsProtoGlobalThis.Buffer) {
    return Uint8Array.from(tsProtoGlobalThis.Buffer.from(b64, 'base64'))
  } else {
    const bin = tsProtoGlobalThis.atob(b64)
    const arr = new Uint8Array(bin.length)
    for (let i = 0; i < bin.length; ++i) {
      arr[i] = bin.charCodeAt(i)
    }
    return arr
  }
}

function base64FromBytes(arr: Uint8Array): string {
  if (tsProtoGlobalThis.Buffer) {
    return tsProtoGlobalThis.Buffer.from(arr).toString('base64')
  } else {
    const bin: string[] = []
    arr.forEach((byte) => {
      bin.push(String.fromCharCode(byte))
    })
    return tsProtoGlobalThis.btoa(bin.join(''))
  }
}

type Builtin =
  | Date
  | Function
  | Uint8Array
  | string
  | number
  | boolean
  | undefined

export type DeepPartial<T> = T extends Builtin
  ? T
  : T extends Long
  ? string | number | Long
  : T extends Array<infer U>
  ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U>
  ? ReadonlyArray<DeepPartial<U>>
  : T extends { $case: string }
  ? { [K in keyof Omit<T, '$case'>]?: DeepPartial<T[K]> } & {
      $case: T['$case']
    }
  : T extends {}
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>

type KeysOfUnion<T> = T extends T ? keyof T : never
export type Exact<P, I extends P> = P extends Builtin
  ? P
  : P & { [K in keyof P]: Exact<P[K], I[K]> } & {
      [K in Exclude<keyof I, KeysOfUnion<P>>]: never
    }

if (_m0.util.Long !== Long) {
  _m0.util.Long = Long as any
  _m0.configure()
}

function isObject(value: any): boolean {
  return typeof value === 'object' && value !== null
}

function isSet(value: any): boolean {
  return value !== null && value !== undefined
}
