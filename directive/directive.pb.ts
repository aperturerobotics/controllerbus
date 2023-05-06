/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'

export const protobufPackage = 'directive'

/** DirectiveInfo contains directive information in protobuf form. */
export interface DirectiveInfo {
  /** Name is the directive name. */
  name: string
  /** DebugVals contains the directive debug values. */
  debugVals: ProtoDebugValue[]
}

/** DirectiveState contains directive info and state info in protobuf form. */
export interface DirectiveState {
  /** Info is the directive info. */
  info: DirectiveInfo | undefined
}

/** ProtoDebugValue is a debug value. */
export interface ProtoDebugValue {
  /** Key is the debug value key. */
  key: string
  /** Values are the debug value values. */
  values: string[]
}

function createBaseDirectiveInfo(): DirectiveInfo {
  return { name: '', debugVals: [] }
}

export const DirectiveInfo = {
  encode(
    message: DirectiveInfo,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.name !== '') {
      writer.uint32(10).string(message.name)
    }
    for (const v of message.debugVals) {
      ProtoDebugValue.encode(v!, writer.uint32(18).fork()).ldelim()
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): DirectiveInfo {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseDirectiveInfo()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.name = reader.string()
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.debugVals.push(
            ProtoDebugValue.decode(reader, reader.uint32())
          )
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
  // Transform<DirectiveInfo, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<DirectiveInfo | DirectiveInfo[]>
      | Iterable<DirectiveInfo | DirectiveInfo[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [DirectiveInfo.encode(p).finish()]
        }
      } else {
        yield* [DirectiveInfo.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, DirectiveInfo>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<DirectiveInfo> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [DirectiveInfo.decode(p)]
        }
      } else {
        yield* [DirectiveInfo.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): DirectiveInfo {
    return {
      name: isSet(object.name) ? String(object.name) : '',
      debugVals: Array.isArray(object?.debugVals)
        ? object.debugVals.map((e: any) => ProtoDebugValue.fromJSON(e))
        : [],
    }
  },

  toJSON(message: DirectiveInfo): unknown {
    const obj: any = {}
    message.name !== undefined && (obj.name = message.name)
    if (message.debugVals) {
      obj.debugVals = message.debugVals.map((e) =>
        e ? ProtoDebugValue.toJSON(e) : undefined
      )
    } else {
      obj.debugVals = []
    }
    return obj
  },

  create<I extends Exact<DeepPartial<DirectiveInfo>, I>>(
    base?: I
  ): DirectiveInfo {
    return DirectiveInfo.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<DirectiveInfo>, I>>(
    object: I
  ): DirectiveInfo {
    const message = createBaseDirectiveInfo()
    message.name = object.name ?? ''
    message.debugVals =
      object.debugVals?.map((e) => ProtoDebugValue.fromPartial(e)) || []
    return message
  },
}

function createBaseDirectiveState(): DirectiveState {
  return { info: undefined }
}

export const DirectiveState = {
  encode(
    message: DirectiveState,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.info !== undefined) {
      DirectiveInfo.encode(message.info, writer.uint32(10).fork()).ldelim()
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): DirectiveState {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseDirectiveState()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.info = DirectiveInfo.decode(reader, reader.uint32())
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
  // Transform<DirectiveState, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<DirectiveState | DirectiveState[]>
      | Iterable<DirectiveState | DirectiveState[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [DirectiveState.encode(p).finish()]
        }
      } else {
        yield* [DirectiveState.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, DirectiveState>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<DirectiveState> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [DirectiveState.decode(p)]
        }
      } else {
        yield* [DirectiveState.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): DirectiveState {
    return {
      info: isSet(object.info)
        ? DirectiveInfo.fromJSON(object.info)
        : undefined,
    }
  },

  toJSON(message: DirectiveState): unknown {
    const obj: any = {}
    message.info !== undefined &&
      (obj.info = message.info ? DirectiveInfo.toJSON(message.info) : undefined)
    return obj
  },

  create<I extends Exact<DeepPartial<DirectiveState>, I>>(
    base?: I
  ): DirectiveState {
    return DirectiveState.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<DirectiveState>, I>>(
    object: I
  ): DirectiveState {
    const message = createBaseDirectiveState()
    message.info =
      object.info !== undefined && object.info !== null
        ? DirectiveInfo.fromPartial(object.info)
        : undefined
    return message
  },
}

function createBaseProtoDebugValue(): ProtoDebugValue {
  return { key: '', values: [] }
}

export const ProtoDebugValue = {
  encode(
    message: ProtoDebugValue,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.key !== '') {
      writer.uint32(10).string(message.key)
    }
    for (const v of message.values) {
      writer.uint32(18).string(v!)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ProtoDebugValue {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseProtoDebugValue()
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

          message.values.push(reader.string())
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
  // Transform<ProtoDebugValue, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<ProtoDebugValue | ProtoDebugValue[]>
      | Iterable<ProtoDebugValue | ProtoDebugValue[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [ProtoDebugValue.encode(p).finish()]
        }
      } else {
        yield* [ProtoDebugValue.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, ProtoDebugValue>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<ProtoDebugValue> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [ProtoDebugValue.decode(p)]
        }
      } else {
        yield* [ProtoDebugValue.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): ProtoDebugValue {
    return {
      key: isSet(object.key) ? String(object.key) : '',
      values: Array.isArray(object?.values)
        ? object.values.map((e: any) => String(e))
        : [],
    }
  },

  toJSON(message: ProtoDebugValue): unknown {
    const obj: any = {}
    message.key !== undefined && (obj.key = message.key)
    if (message.values) {
      obj.values = message.values.map((e) => e)
    } else {
      obj.values = []
    }
    return obj
  },

  create<I extends Exact<DeepPartial<ProtoDebugValue>, I>>(
    base?: I
  ): ProtoDebugValue {
    return ProtoDebugValue.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<ProtoDebugValue>, I>>(
    object: I
  ): ProtoDebugValue {
    const message = createBaseProtoDebugValue()
    message.key = object.key ?? ''
    message.values = object.values?.map((e) => e) || []
    return message
  },
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

function isSet(value: any): boolean {
  return value !== null && value !== undefined
}
