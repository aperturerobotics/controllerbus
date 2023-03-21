/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'

export const protobufPackage = 'controller'

/** Info contains information about a controller. */
export interface Info {
  /** Id contains the identifier of the controller. */
  id: string
  /** Version contains the version string for the controller. */
  version: string
  /** Description contains a descriptive string about the controller. */
  description: string
}

function createBaseInfo(): Info {
  return { id: '', version: '', description: '' }
}

export const Info = {
  encode(message: Info, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.id !== '') {
      writer.uint32(10).string(message.id)
    }
    if (message.version !== '') {
      writer.uint32(18).string(message.version)
    }
    if (message.description !== '') {
      writer.uint32(26).string(message.description)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Info {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseInfo()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag != 10) {
            break
          }

          message.id = reader.string()
          continue
        case 2:
          if (tag != 18) {
            break
          }

          message.version = reader.string()
          continue
        case 3:
          if (tag != 26) {
            break
          }

          message.description = reader.string()
          continue
      }
      if ((tag & 7) == 4 || tag == 0) {
        break
      }
      reader.skipType(tag & 7)
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<Info, Uint8Array>
  async *encodeTransform(
    source: AsyncIterable<Info | Info[]> | Iterable<Info | Info[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Info.encode(p).finish()]
        }
      } else {
        yield* [Info.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Info>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<Info> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Info.decode(p)]
        }
      } else {
        yield* [Info.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): Info {
    return {
      id: isSet(object.id) ? String(object.id) : '',
      version: isSet(object.version) ? String(object.version) : '',
      description: isSet(object.description) ? String(object.description) : '',
    }
  },

  toJSON(message: Info): unknown {
    const obj: any = {}
    message.id !== undefined && (obj.id = message.id)
    message.version !== undefined && (obj.version = message.version)
    message.description !== undefined && (obj.description = message.description)
    return obj
  },

  create<I extends Exact<DeepPartial<Info>, I>>(base?: I): Info {
    return Info.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<Info>, I>>(object: I): Info {
    const message = createBaseInfo()
    message.id = object.id ?? ''
    message.version = object.version ?? ''
    message.description = object.description ?? ''
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
