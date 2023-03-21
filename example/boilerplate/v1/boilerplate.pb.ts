/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'

export const protobufPackage = 'boilerplate.v1'

/** Boilerplate implements the boilerplate directive. */
export interface Boilerplate {
  /**
   * MessageText is the message to print with the boilerplate.
   * This is an example field.
   * The keyword "message" prevents us from using that as the field name.
   */
  messageText: string
}

/** BoilerplateResult implements the boilerplate directive result. */
export interface BoilerplateResult {
  /** PrintedLen is the final length of the printed message. */
  printedLen: number
}

function createBaseBoilerplate(): Boilerplate {
  return { messageText: '' }
}

export const Boilerplate = {
  encode(
    message: Boilerplate,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.messageText !== '') {
      writer.uint32(10).string(message.messageText)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Boilerplate {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseBoilerplate()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag != 10) {
            break
          }

          message.messageText = reader.string()
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
  // Transform<Boilerplate, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<Boilerplate | Boilerplate[]>
      | Iterable<Boilerplate | Boilerplate[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Boilerplate.encode(p).finish()]
        }
      } else {
        yield* [Boilerplate.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Boilerplate>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<Boilerplate> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Boilerplate.decode(p)]
        }
      } else {
        yield* [Boilerplate.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): Boilerplate {
    return {
      messageText: isSet(object.messageText) ? String(object.messageText) : '',
    }
  },

  toJSON(message: Boilerplate): unknown {
    const obj: any = {}
    message.messageText !== undefined && (obj.messageText = message.messageText)
    return obj
  },

  create<I extends Exact<DeepPartial<Boilerplate>, I>>(base?: I): Boilerplate {
    return Boilerplate.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<Boilerplate>, I>>(
    object: I
  ): Boilerplate {
    const message = createBaseBoilerplate()
    message.messageText = object.messageText ?? ''
    return message
  },
}

function createBaseBoilerplateResult(): BoilerplateResult {
  return { printedLen: 0 }
}

export const BoilerplateResult = {
  encode(
    message: BoilerplateResult,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.printedLen !== 0) {
      writer.uint32(8).uint32(message.printedLen)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): BoilerplateResult {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseBoilerplateResult()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag != 8) {
            break
          }

          message.printedLen = reader.uint32()
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
  // Transform<BoilerplateResult, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<BoilerplateResult | BoilerplateResult[]>
      | Iterable<BoilerplateResult | BoilerplateResult[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [BoilerplateResult.encode(p).finish()]
        }
      } else {
        yield* [BoilerplateResult.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, BoilerplateResult>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<BoilerplateResult> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [BoilerplateResult.decode(p)]
        }
      } else {
        yield* [BoilerplateResult.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): BoilerplateResult {
    return {
      printedLen: isSet(object.printedLen) ? Number(object.printedLen) : 0,
    }
  },

  toJSON(message: BoilerplateResult): unknown {
    const obj: any = {}
    message.printedLen !== undefined &&
      (obj.printedLen = Math.round(message.printedLen))
    return obj
  },

  create<I extends Exact<DeepPartial<BoilerplateResult>, I>>(
    base?: I
  ): BoilerplateResult {
    return BoilerplateResult.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<BoilerplateResult>, I>>(
    object: I
  ): BoilerplateResult {
    const message = createBaseBoilerplateResult()
    message.printedLen = object.printedLen ?? 0
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
